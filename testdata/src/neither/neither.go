// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

// Package neither imports neither the painter nor the toolkit, so the analyzer
// returns immediately and reports nothing.
package neither

import "strings"

var _ = strings.TrimSpace("x")
