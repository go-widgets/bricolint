// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

// Package throwaway imports only the toolkit (no painter), so rule 1 stays off,
// and exercises rule 2: toolkit widgets constructed inside per-frame methods.
package throwaway

import (
	"strings"

	"github.com/go-widgets/toolkit"
)

// A package-level constructor call is not inside any method, so it is not a
// per-frame throwaway and must not be flagged.
var _ = toolkit.NewButton("package level")

type view struct {
	btn *toolkit.Button
}

// Draw builds a widget on every frame — the core violation.
func (v *view) Draw() {
	b := toolkit.NewButton("ok") // want `throwaway widget: NewButton`
	_ = b

	helper() // non-selector call

	v.other() // a non-New selector call is not a constructor

	var s local
	s.NewThing() // New-prefixed, but the object is a field, not a func

	_ = strings.NewReader("x") // New-prefixed func from another package

	_ = toolkit.New() // bare New(): no widget suffix

	// A widget built inside a closure created during Draw is still per-frame.
	f := func() { v.btn = toolkit.NewButton("nested") } // want `throwaway widget: NewButton`
	f()
}

// Paint is also a per-frame method.
func (v *view) Paint() {
	_ = toolkit.NewLabel("hi") // want `throwaway widget: NewLabel`
}

// Render's construction is an explicitly allowed leaf.
func (v *view) Render() {
	_ = toolkit.NewButton("minimap") //bricolint:allow minimap raster overlay — a genuine leaf
}

// Update is not a paint method, so constructing a widget here is fine.
func (v *view) Update() {
	v.btn = toolkit.NewButton("persisted")
}

// Draw as a free function (no receiver) is not a paint method.
func Draw() {
	_ = toolkit.NewButton("free function")
}

func (v *view) other() {}

func helper() {}

type local struct {
	NewThing func()
}
