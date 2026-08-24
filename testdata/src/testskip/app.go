// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

// Package testskip imports the painter in production code but only queries the
// surface; its sole primitive call lives in the _test.go file, which is skipped
// by default — so the default run reports nothing here.
package testskip

import "github.com/go-widgets/painter"

func prod(p *painter.PixelPainter) {
	_, _ = p.Size()
}
