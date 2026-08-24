// Copyright (c) 2026 the go-widgets/bricolint authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

// Package toolkit is a minimal stand-in for github.com/go-widgets/toolkit used
// only by the analyzer's analysistest fixtures. It declares just enough widget
// shape and constructors to exercise the throwaway-widget rule.
package toolkit

// Button is a clickable widget.
type Button struct{ Label string }

// Label is a text widget.
type Label struct{ Text string }

// NewButton constructs a Button, mirroring the real toolkit's constructor style.
func NewButton(label string) *Button { return &Button{Label: label} }

// NewLabel constructs a Label.
func NewLabel(text string) *Label { return &Label{Text: text} }

// New is a bare, non-widget factory: its name has no widget suffix, so the
// throwaway rule must not treat it as a constructor.
func New() *Button { return &Button{} }
