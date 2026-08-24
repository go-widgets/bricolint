// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

// Package allowline exercises the //bricolint:allow line directive: a valid
// (reasoned) directive suppresses on its own line and on the line below it,
// while a reason-less directive is ignored so the violation still fires.
package allowline

import "github.com/go-widgets/painter"

func paint(p *painter.PixelPainter) {
	p.FillRect(painter.Rect{}, painter.RGBA{}) //bricolint:allow engine SVG render blit — a genuine leaf

	// A reason-less directive is ignored, so the next line is still flagged.
	//bricolint:allow
	p.FillRect(painter.Rect{}, painter.RGBA{}) // want `hand-drawn UI: painter primitive "FillRect"`

	//bricolint:allow document render surface — a genuine leaf
	p.StrokeRect(painter.Rect{}, painter.RGBA{}, 1)
}
