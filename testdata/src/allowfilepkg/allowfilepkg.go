// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

//bricolint:allowfile painter back-end (go-pdfkit style) — this whole file IS the leaf

// Package allowfilepkg is a genuine painter leaf: a file-level allowfile
// directive exempts every primitive call in it.
package allowfilepkg

import "github.com/go-widgets/painter"

func paint(p *painter.PixelPainter) {
	p.FillRect(painter.Rect{}, painter.RGBA{})
	p.Text(0, 0, "x", painter.RGBA{})
}
