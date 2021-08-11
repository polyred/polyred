// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gui

// Key corresponds to a keyboard key.
type Key int

// ModifierKey corresponds to a set of modifier keys (bitmask).
type ModifierKey int

// MouseButton corresponds to a mouse button.
type MouseButton int

// InputMode corresponds to an input mode.
type CursorMode int

// Cursor corresponds to a g3n standard or user-created cursor icon.
type Cursor int

const (
	ModShift = ModifierKey(1 << iota) // Bitmask
	ModControl
	ModAlt
)
