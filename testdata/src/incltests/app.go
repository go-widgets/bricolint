// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

// Package incltests is analyzed with -includetests=true, so a primitive call in
// its _test.go file is flagged.
package incltests

import "github.com/go-widgets/painter"

func prod(p *painter.PixelPainter) {
	_, _ = p.Size()
}
