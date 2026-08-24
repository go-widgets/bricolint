// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

// Command bricolint is a standalone driver for the bricolint analyzer. It
// doubles as a go vet tool:
//
//	go install github.com/go-widgets/bricolint/cmd/bricolint@latest
//	go vet -vettool=$(which bricolint) ./...
package main

import (
	"github.com/go-widgets/bricolint"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(bricolint.Analyzer)
}
