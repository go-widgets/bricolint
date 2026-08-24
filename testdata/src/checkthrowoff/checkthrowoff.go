// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

// Package checkthrowoff is analyzed with -checkthrow=false, so the per-frame
// construction below produces no diagnostics.
package checkthrowoff

import "github.com/go-widgets/toolkit"

type view struct{}

func (view) Draw() {
	_ = toolkit.NewButton("ok")
}
