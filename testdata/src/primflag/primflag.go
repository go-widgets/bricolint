// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

// Package primflag is analyzed with -primitives=Text, so only Text is treated
// as a hand-drawn primitive and FillRect is left alone.
package primflag

import "github.com/go-widgets/painter"

func paint(p *painter.PixelPainter) {
	p.FillRect(painter.Rect{}, painter.RGBA{}) // not in the custom set
	p.Text(0, 0, "x", painter.RGBA{})          // want `hand-drawn UI: painter primitive "Text"`
}
