// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package throwaway

import "github.com/go-widgets/toolkit"

// fixture has a real per-frame method, but it lives in a _test.go file, which
// bricolint skips by default — so this construction must not be flagged, and
// the allow directive below must itself be skipped by the comment scanner.
type fixture struct{}

//bricolint:allow this directive lives in a test file and must itself be skipped
func (fixture) Draw() {
	_ = toolkit.NewButton("test fixture")
}
