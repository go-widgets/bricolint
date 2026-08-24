// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

// Package handdrawn imports only the painter (no toolkit), so rule 2 stays off,
// and exercises rule 1: painter drawing primitives called from app code.
package handdrawn

import "github.com/go-widgets/painter"

func helper() {}

/* a block comment is not a line directive and must be skipped */

func paint(p *painter.PixelPainter) {
	helper() // a non-selector call is never a painter primitive

	w, h := p.Size() // Size is structural, not a drawing primitive
	_, _ = w, h

	p.FillRect(painter.Rect{}, painter.RGBA{}) // want `hand-drawn UI: painter primitive "FillRect"`
	p.Text(0, 0, "x", painter.RGBA{})          // want `hand-drawn UI: painter primitive "Text"`

	// An unnamed interface receiver does not resolve to a painter named type,
	// so it must not be flagged.
	var q interface {
		FillRect(painter.Rect, painter.RGBA)
	} = p
	q.FillRect(painter.Rect{}, painter.RGBA{})

	// A same-named method on an unrelated local type must never be flagged.
	var g gadget
	g.FillRect(painter.Rect{}, painter.RGBA{})
}

type gadget struct{}

func (gadget) FillRect(painter.Rect, painter.RGBA) {}
