# bricolint

[![ci](https://github.com/go-widgets/bricolint/actions/workflows/ci.yml/badge.svg)](https://github.com/go-widgets/bricolint/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-widgets/bricolint.svg)](https://pkg.go.dev/github.com/go-widgets/bricolint)
[![License: BSD-3-Clause](https://img.shields.io/badge/License-BSD--3--Clause-blue.svg)](LICENSE)

A [`go/analysis`](https://pkg.go.dev/golang.org/x/tools/go/analysis) analyzer that
guards the fleet-wide **"no hand-drawn UI"** rule: every visible element must be a
[go-widgets/toolkit](https://github.com/go-widgets/toolkit) widget — never chrome
painted by hand onto a raw [painter](https://github.com/go-widgets/painter)
surface, and never a throwaway widget rebuilt on every frame.

*Bricolage* — improvised, hand-rolled drawing — is how a UI quietly loses
press/hover/focus feedback, theming, HiDPI scaling and accessibility. Once an app
has been migrated onto the toolkit, bricolint keeps it that way: a pull request
that reintroduces manual drawing **fails CI** instead of merging.

```go
p.FillRect(bar, theme.Bg)                 // flagged — hand-drawn chrome
toolkit.NewBackdrop(theme.Bg).Draw(p)     // sanctioned — a toolkit widget

func (v *view) Draw(p painter.Painter) {
    b := toolkit.NewButton("ok")          // flagged — a widget rebuilt every frame
    b.Draw(p)
}
```

## Rules

bricolint only fires in packages that import the painter (rule 1) or the toolkit
(rule 2); everywhere else it is a no-op.

1. **Painter primitive in application code.** A method call whose receiver's
   static type resolves — through `go/types` — to a type in
   `github.com/go-widgets/painter` (the `Painter` interface, or a concrete
   `*PixelPainter` / `*CellPainter`), and whose method is a drawing primitive.
   Because the receiver is *type-resolved*, a same-named method on an unrelated
   type (`strings.Builder`, `bytes.Buffer`, a local helper) is **never** touched.

   Guarded primitives (default):

   ```text
   FillRect  FillRoundRect  StrokeRect  StrokeRoundRect  FillPath  StrokePath
   DrawImage  DrawMask  PutPixel  Text  DrawText  DrawGlyph  DrawLine  Blit
   ```

   Structural, non-drawing methods (`Size`, `PushClip`/`PopClip`,
   `PushTranslate`/`PopTranslate`) are deliberately **not** primitives — querying
   or clipping the surface is not bricolage. Override the set with
   `-primitives=FillRect,Text,...` (comma-separated; replaces the default).

2. **Throwaway widget per frame.** A toolkit constructor call (`toolkit.New<X>`)
   made syntactically inside a method named `Draw`, `Paint` or `Render`
   (constructions inside a closure created there count too). A widget built on
   every paint is never persisted, so it can hold no interaction state — the
   anti-pattern that kills press/hover/focus feedback. Build it once, store it as
   a field, and drive it through a
   [go-widgets/mvvm](https://github.com/go-widgets/mvvm) binding. Turn the rule
   off with `-checkthrow=false`.

By default `*_test.go` files are skipped (tests legitimately construct
fixtures); analyze them too with `-includetests`.

## Allowlisting genuine leaves

Some code is a legitimate leaf: a game framebuffer, a painter **back-end** such
as go-pdfkit, a document/minimap raster, the toolkit's own widget internals.
These opt out **explicitly**, so the exemption is a conscious, documented choice.

- **One line** — a trailing or immediately-preceding line comment:

  ```go
  p.FillRect(bg, c) //bricolint:allow engine SVG raster blit — a genuine leaf
  ```

- **A whole file** — a package/file-level comment:

  ```go
  //bricolint:allowfile painter back-end (go-pdfkit style) — this file IS the leaf
  package pdfsurface
  ```

**The reason is mandatory.** A directive with no reason is *ignored*, so the
violation keeps firing until a justification is written down.

## Install & run locally

```sh
go install github.com/go-widgets/bricolint/cmd/bricolint@latest
go vet -vettool="$(go env GOPATH)/bin/bricolint" ./...
```

Flags pass through `go vet` unprefixed (single-analyzer vet tool):

```sh
go vet -vettool="$(which bricolint)" -checkthrow=false ./...
go vet -vettool="$(which bricolint)" -primitives=FillRect,Text ./...
```

## CI gate

### Reusable workflow (recommended)

Reference the shipped reusable workflow from your app repo and mark the resulting
check as **required** in branch protection:

```yaml
# .github/workflows/bricolint.yml in your app repo
name: bricolint
on:
  pull_request:
  push:
    branches: [main]
jobs:
  bricolint:
    uses: go-widgets/bricolint/.github/workflows/bricolint.yml@main
    # optional overrides:
    # with:
    #   go-version: "1.26.4"
    #   packages: "./..."
    #   working-directory: "."
    #   version: "latest"
```

### Standalone step (copy-paste)

```yaml
jobs:
  bricolint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26.4"
      - run: go install github.com/go-widgets/bricolint/cmd/bricolint@latest
      - run: go vet -vettool="$(go env GOPATH)/bin/bricolint" ./...
```

## False-positive posture

The rules are deliberately conservative:

- Rule 1 fires only when the **static receiver type** resolves to a painter type
  **and** the method is on the primitive set — so a same-named method on an
  unrelated type, an unnamed interface receiver, or a structural painter call is
  never flagged.
- Rule 2 fires only on a resolved `toolkit.New<Widget>` **func** (not a
  same-named field, and not a `New*` from another package) inside a paint method.

The remaining intentional edges — a real render leaf, a painter back-end, the
toolkit's own internals — are handled by the `//bricolint:allow` /
`//bricolint:allowfile` directives, and rule 2 can be disabled wholesale with
`-checkthrow=false`.

## License

BSD-3-Clause — see [LICENSE](LICENSE). Copyright (c) 2026 the go-widgets/bricolint
authors.
