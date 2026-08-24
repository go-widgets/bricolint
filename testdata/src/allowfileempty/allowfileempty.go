// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

//bricolint:allowfile

// Package allowfileempty carries a reason-less allowfile directive, which is
// ignored — so the primitive call is still flagged.
package allowfileempty

import "github.com/go-widgets/painter"

func paint(p *painter.PixelPainter) {
	p.FillRect(painter.Rect{}, painter.RGBA{}) // want `hand-drawn UI: painter primitive "FillRect"`
}
