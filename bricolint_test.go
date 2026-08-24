// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package bricolint

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// setFlag sets an analyzer flag for the duration of a subtest and restores it.
func setFlag(t *testing.T, name, value string) {
	t.Helper()
	f := Analyzer.Flags.Lookup(name)
	if f == nil {
		t.Fatalf("flag %q not found", name)
	}
	prev := f.Value.String()
	if err := Analyzer.Flags.Set(name, value); err != nil {
		t.Fatalf("set -%s=%s: %v", name, value, err)
	}
	t.Cleanup(func() { _ = Analyzer.Flags.Set(name, prev) })
}

// TestDefaults exercises both rules with the built-in configuration:
//
//   - handdrawn:       painter primitives flagged; structural, non-selector,
//     interface and unrelated-type calls left alone,
//   - allowline:       the //bricolint:allow line directive (same line, line
//     above, and a reason-less directive that is ignored),
//   - allowfilepkg:    the //bricolint:allowfile whole-file exemption,
//   - allowfileempty:  a reason-less allowfile is ignored, so it still fires,
//   - throwaway:       toolkit widgets built inside Draw/Paint(/closures) are
//     flagged; Render (allowed), Update, a free Draw, a package-level call and
//     the various non-constructor selectors are not,
//   - testskip:        a primitive call confined to a _test.go file is skipped,
//   - neither:         a package touching neither library reports nothing.
func TestDefaults(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer,
		"handdrawn", "allowline", "allowfilepkg", "allowfileempty",
		"throwaway", "testskip", "neither")
}

// TestIncludeTests opts into analyzing *_test.go files, so a primitive call in a
// test fixture is flagged (and an allow directive there still exempts).
func TestIncludeTests(t *testing.T) {
	setFlag(t, "includetests", "true")
	analysistest.Run(t, analysistest.TestData(), Analyzer, "incltests")
}

// TestPrimitivesFlag overrides the primitive set so only Text is guarded,
// exercising the -primitives flag path.
func TestPrimitivesFlag(t *testing.T) {
	setFlag(t, "primitives", "Text,,")
	analysistest.Run(t, analysistest.TestData(), Analyzer, "primflag")
}

// TestCheckThrowOff turns the throwaway-widget rule off, which must then produce
// no diagnostics even for a widget built inside Draw.
func TestCheckThrowOff(t *testing.T) {
	setFlag(t, "checkthrow", "false")
	analysistest.Run(t, analysistest.TestData(), Analyzer, "checkthrowoff")
}

// TestExportedKnobs guards the exported constants and default primitive set
// against accidental drift.
func TestExportedKnobs(t *testing.T) {
	if PainterPath != "github.com/go-widgets/painter" {
		t.Errorf("PainterPath = %q", PainterPath)
	}
	if ToolkitPath != "github.com/go-widgets/toolkit" {
		t.Errorf("ToolkitPath = %q", ToolkitPath)
	}
	if MVVMPath != "github.com/go-widgets/mvvm" {
		t.Errorf("MVVMPath = %q", MVVMPath)
	}
	want := map[string]bool{"FillRect": true, "Text": true, "PutPixel": true, "DrawImage": true, "FillPath": true}
	have := make(map[string]bool)
	for _, f := range DefaultPrimitives {
		have[f] = true
	}
	for k := range want {
		if !have[k] {
			t.Errorf("DefaultPrimitives missing %q", k)
		}
	}
}
