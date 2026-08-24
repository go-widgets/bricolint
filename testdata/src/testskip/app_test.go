// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package testskip

import "github.com/go-widgets/painter"

// This primitive call lives in a test file. With the default configuration
// (tests skipped) it must NOT be flagged.
func draw(p *painter.PixelPainter) {
	p.FillRect(painter.Rect{}, painter.RGBA{})
}
