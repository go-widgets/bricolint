// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package incltests

import "github.com/go-widgets/painter"

// With -includetests=true this primitive call in a test file is flagged, and
// the reason-less allow directive above it is honoured as a real directive
// (test comments are scanned too), so it stays exempt.
func draw(p *painter.PixelPainter) {
	p.FillRect(painter.Rect{}, painter.RGBA{}) // want `hand-drawn UI: painter primitive "FillRect"`

	//bricolint:allow test-only raster fixture — a genuine leaf
	p.Text(0, 0, "x", painter.RGBA{})
}
