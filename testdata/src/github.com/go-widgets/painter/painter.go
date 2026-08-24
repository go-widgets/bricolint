// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

// Package painter is a minimal stand-in for github.com/go-widgets/painter used
// only by the analyzer's analysistest fixtures. It declares just enough of the
// painter surface (a couple of drawing primitives, one structural method, and
// both a concrete painter and the interface) to exercise the rules.
package painter

// Rect is a rectangle in painter units.
type Rect struct{ X, Y, W, H int }

// RGBA is a straight-alpha colour.
type RGBA struct{ R, G, B, A uint8 }

// PixelPainter is a concrete pixel-back-end painter.
type PixelPainter struct{}

// FillRect paints a solid rectangle.
func (p *PixelPainter) FillRect(r Rect, c RGBA) {}

// StrokeRect paints a border.
func (p *PixelPainter) StrokeRect(r Rect, c RGBA, lineW int) {}

// Text paints ink text.
func (p *PixelPainter) Text(x, y int, s string, ink RGBA) {}

// PutPixel paints a single pixel.
func (p *PixelPainter) PutPixel(x, y int, c RGBA) {}

// Size returns the canvas dimensions — a structural, non-drawing method.
func (p *PixelPainter) Size() (int, int) { return 0, 0 }

// Painter is the drawing surface interface.
type Painter interface {
	FillRect(r Rect, c RGBA)
	Text(x, y int, s string, ink RGBA)
	Size() (int, int)
}

// NewPixelPainter constructs a PixelPainter.
func NewPixelPainter() *PixelPainter { return &PixelPainter{} }
