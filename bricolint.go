// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

// Package bricolint provides a go/analysis analyzer that guards the fleet-wide
// "no hand-drawn UI" rule: every visible element must be a go-widgets/toolkit
// widget, never chrome painted by hand onto a raw painter surface, and never a
// throwaway widget rebuilt on every paint.
//
// "Bricolage" — improvised, hand-rolled drawing — is how a UI quietly loses
// press/hover/focus feedback, theming, HiDPI scaling and accessibility. Once an
// app has been migrated onto the toolkit, this analyzer keeps it that way: a
// pull request that reintroduces manual drawing fails CI instead of merging.
//
// The analyzer reports two rules, and only in packages that actually touch the
// toolkit stack (it is a no-op everywhere else):
//
//  1. Painter primitive in application code. A method call whose receiver's
//     static type resolves to a type in github.com/go-widgets/painter (the
//     Painter interface itself, or a concrete *PixelPainter / *CellPainter),
//     and whose method is a drawing primitive (FillRect, FillRoundRect,
//     StrokeRect, StrokeRoundRect, FillPath, StrokePath, DrawImage, DrawMask,
//     PutPixel, Text, and the conventional aliases DrawText/DrawGlyph/
//     DrawLine/Blit). Because the receiver is resolved through go/types, a
//     same-named method on an unrelated type (strings.Builder, bytes.Buffer,
//     a local helper) is never touched.
//
//  2. Throwaway widget per frame. A toolkit constructor call (toolkit.New<X>)
//     made syntactically inside a method named Draw, Paint or Render. A widget
//     constructed on every paint is never persisted, so it can hold no
//     interaction state — the anti-pattern that kills press/hover/focus
//     feedback. The fix is to build the widget once, store it as a field, and
//     drive it through a go-widgets/mvvm binding.
//
// Genuine render leaves — a game framebuffer, a painter BACK-END such as
// go-pdfkit, a document/minimap raster, the toolkit's own widget internals —
// opt out explicitly, so the exemption is a conscious, documented choice:
//
//   - A trailing (or immediately preceding) line comment
//     //bricolint:allow <reason> exempts that single line.
//   - A file-level comment //bricolint:allowfile <reason> exempts the whole
//     file.
//
// The reason is mandatory: a directive with no reason is ignored, so the
// violation keeps failing until a justification is written down.
package bricolint

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// PainterPath is the import path of the go-widgets painter surface.
const PainterPath = "github.com/go-widgets/painter"

// ToolkitPath is the import path of the go-widgets pixel/cell toolkit.
const ToolkitPath = "github.com/go-widgets/toolkit"

// MVVMPath is the import path of the go-widgets MVVM library, named in the
// throwaway-widget diagnostic as the sanctioned way to hold widget state.
const MVVMPath = "github.com/go-widgets/mvvm"

const (
	allowDirective     = "bricolint:allow"
	allowFileDirective = "bricolint:allowfile"
)

// DefaultPrimitives is the built-in set of painter drawing methods that count
// as hand-drawn chrome when called on a painter surface from application code.
//
// It covers the real painter API (FillRect/StrokeRect/…/PutPixel/Text) plus the
// conventional primitive names an app is tempted to reach for
// (DrawText/DrawGlyph/DrawLine/Blit). Structural, non-drawing methods of the
// painter (Size, PushClip/PopClip, PushTranslate/PopTranslate) are deliberately
// absent: querying the surface or clipping is not bricolage.
var DefaultPrimitives = []string{
	"Blit",
	"DrawGlyph",
	"DrawImage",
	"DrawLine",
	"DrawMask",
	"DrawText",
	"FillPath",
	"FillRect",
	"FillRoundRect",
	"PutPixel",
	"StrokePath",
	"StrokeRect",
	"StrokeRoundRect",
	"Text",
}

// paintMethods is the set of per-frame method names whose bodies must not
// construct toolkit widgets.
var paintMethods = map[string]bool{
	"Draw":   true,
	"Paint":  true,
	"Render": true,
}

// config holds the analyzer's tunable behaviour. It is stored on the Analyzer so
// that flags and programmatic construction share one code path.
type config struct {
	primitivesFlag string // raw -primitives flag value ("" => defaults)
	checkThrow     bool   // emit the throwaway-widget rule
	includeTests   bool   // also analyze *_test.go files (default: skip them)
	painterPath    string // painter import path (overridable for testing)
	toolkitPath    string // toolkit import path (overridable for testing)
}

// Analyzer is the bricolint go/analysis analyzer. Use it with singlechecker, or
// as a go vet tool: go vet -vettool=$(which bricolint) ./...
var Analyzer = newAnalyzer()

func newAnalyzer() *analysis.Analyzer {
	cfg := &config{checkThrow: true, painterPath: PainterPath, toolkitPath: ToolkitPath}
	a := &analysis.Analyzer{
		Name:     "bricolint",
		Doc:      "flag hand-drawn UI: painter drawing primitives in app code, and toolkit widgets constructed per paint frame",
		URL:      "https://github.com/go-widgets/bricolint",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      func(pass *analysis.Pass) (any, error) { return run(pass, cfg) },
	}
	a.Flags.StringVar(&cfg.primitivesFlag, "primitives", "",
		"comma-separated set of painter methods treated as hand-drawn primitives (empty uses the built-in default)")
	a.Flags.BoolVar(&cfg.checkThrow, "checkthrow", true,
		"report toolkit widgets constructed inside Draw/Paint/Render methods (per-frame throwaway widgets)")
	a.Flags.BoolVar(&cfg.includeTests, "includetests", false,
		"also analyze *_test.go files (default: skip them — tests legitimately construct fixtures)")
	a.Flags.StringVar(&cfg.painterPath, "painterpath", PainterPath,
		"import path treated as the painter surface")
	a.Flags.StringVar(&cfg.toolkitPath, "toolkitpath", ToolkitPath,
		"import path treated as the widget toolkit")
	return a
}

// primitiveSet resolves the effective primitive set for a pass: the -primitives
// flag when set, otherwise the built-in default.
func (c *config) primitiveSet() map[string]bool {
	names := DefaultPrimitives
	if strings.TrimSpace(c.primitivesFlag) != "" {
		names = splitFields(c.primitivesFlag)
	}
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	return set
}

// splitFields splits a comma-separated flag value into trimmed, non-empty names.
func splitFields(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func run(pass *analysis.Pass, cfg *config) (any, error) {
	// The analyzer only concerns packages that use the painter (rule 1) or the
	// toolkit (rule 2). Everywhere else it is a no-op.
	usesPainter := importsPath(pass.Pkg, cfg.painterPath)
	usesToolkit := importsPath(pass.Pkg, cfg.toolkitPath)
	if !usesPainter && !usesToolkit {
		return nil, nil
	}

	al := collectAllows(pass, cfg)
	prims := cfg.primitiveSet()
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	insp.WithStack([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}
		call := n.(*ast.CallExpr)
		if !cfg.includeTests && isTestFile(pass, call.Pos()) {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		switch {
		case usesPainter && prims[sel.Sel.Name] &&
			isPainterType(pass.TypesInfo.TypeOf(sel.X), cfg.painterPath):
			if !al.suppressed(pass, sel.Pos()) {
				pass.Reportf(sel.Pos(),
					"hand-drawn UI: painter primitive %q called in application code; "+
						"build this chrome from a go-widgets/toolkit widget "+
						"(Backdrop, Banner, Label, Button, …) instead, or mark a genuine "+
						"render leaf with //bricolint:allow <reason>", sel.Sel.Name)
			}
		case cfg.checkThrow && usesToolkit &&
			isToolkitConstructor(pass, sel, cfg.toolkitPath):
			if fd := enclosingPaintMethod(stack); fd != nil && !al.suppressed(pass, call.Pos()) {
				pass.Reportf(call.Pos(),
					"throwaway widget: %s constructed inside per-frame method %q; a widget "+
						"built each paint loses press/hover/focus state — persist it as a "+
						"field and drive it through a %s binding instead",
					sel.Sel.Name, fd.Name.Name, MVVMPath)
			}
		}
		return true
	})
	return nil, nil
}

// importsPath reports whether pkg directly imports the given path.
func importsPath(pkg *types.Package, path string) bool {
	for _, imp := range pkg.Imports() {
		if imp.Path() == path {
			return true
		}
	}
	return false
}

// isTestFile reports whether the file containing pos is a _test.go file.
func isTestFile(pass *analysis.Pass, pos token.Pos) bool {
	return strings.HasSuffix(pass.Fset.File(pos).Name(), "_test.go")
}

// isPainterType reports whether t is (a pointer to) a named type declared in the
// painter package — the Painter interface, or a concrete PixelPainter /
// CellPainter. Resolving through go/types is what keeps a same-named method on
// an unrelated type from ever being flagged.
func isPainterType(t types.Type, painterPath string) bool {
	// A nil type asserts to false below, so no explicit nil guard is needed.
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	// named.Obj() is never nil; its package is nil only for universe types,
	// which declare no painter methods — pkg != nil short-circuits those.
	pkg := named.Obj().Pkg()
	return pkg != nil && pkg.Path() == painterPath
}

// isToolkitConstructor reports whether sel selects a New<Widget> function
// exported by the toolkit package.
func isToolkitConstructor(pass *analysis.Pass, sel *ast.SelectorExpr, toolkitPath string) bool {
	name := sel.Sel.Name
	if !strings.HasPrefix(name, "New") || name == "New" {
		return false
	}
	fn, ok := pass.TypesInfo.ObjectOf(sel.Sel).(*types.Func)
	if !ok {
		return false
	}
	pkg := fn.Pkg()
	return pkg != nil && pkg.Path() == toolkitPath
}

// enclosingPaintMethod returns the nearest enclosing function declaration when
// it is a method named Draw/Paint/Render, or nil otherwise. Closures created
// inside a paint method still resolve to that method, so a widget built in a
// per-frame callback is caught too.
func enclosingPaintMethod(stack []ast.Node) *ast.FuncDecl {
	for i := len(stack) - 1; i >= 0; i-- {
		fd, ok := stack[i].(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fd.Recv != nil && paintMethods[fd.Name.Name] {
			return fd
		}
		return nil
	}
	return nil
}

// allowIndex records the //bricolint:allow / :allowfile exemptions found in a
// package, resolved to file names and (for line directives) line numbers.
type allowIndex struct {
	files map[string]bool         // file name -> a valid allowfile directive present
	lines map[string]map[int]bool // file name -> lines carrying a valid allow directive
}

// suppressed reports whether a diagnostic at pos is exempted: by a file-level
// allowfile directive, or by an allow directive on the same or the immediately
// preceding line.
func (a *allowIndex) suppressed(pass *analysis.Pass, pos token.Pos) bool {
	p := pass.Fset.Position(pos)
	if a.files[p.Filename] {
		return true
	}
	if lines := a.lines[p.Filename]; lines != nil {
		return lines[p.Line] || lines[p.Line-1]
	}
	return false
}

// collectAllows scans the package's comments for allow / allowfile directives.
// A directive with an empty reason is ignored (the reason is mandatory), so the
// underlying violation keeps firing until a justification is written.
func collectAllows(pass *analysis.Pass, cfg *config) *allowIndex {
	idx := &allowIndex{files: map[string]bool{}, lines: map[string]map[int]bool{}}
	for _, f := range pass.Files {
		for _, group := range f.Comments {
			for _, c := range group.List {
				if !strings.HasPrefix(c.Text, "//") {
					continue
				}
				pos := pass.Fset.Position(c.Pos())
				if !cfg.includeTests && strings.HasSuffix(pos.Filename, "_test.go") {
					continue
				}
				text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
				if reason, ok := directiveReason(text, allowFileDirective); ok {
					if reason != "" {
						idx.files[pos.Filename] = true
					}
					continue
				}
				if reason, ok := directiveReason(text, allowDirective); ok {
					if reason != "" {
						if idx.lines[pos.Filename] == nil {
							idx.lines[pos.Filename] = map[int]bool{}
						}
						idx.lines[pos.Filename][pos.Line] = true
					}
				}
			}
		}
	}
	return idx
}

// directiveReason matches a directive keyword at the start of a comment body and
// returns its (trimmed) reason. The keyword must stand alone or be followed by a
// space, so bricolint:allow does not spuriously match bricolint:allowfile.
func directiveReason(text, keyword string) (string, bool) {
	if text == keyword {
		return "", true
	}
	if strings.HasPrefix(text, keyword+" ") {
		return strings.TrimSpace(text[len(keyword)+1:]), true
	}
	return "", false
}
