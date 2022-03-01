// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

import "fmt"

type Action int

const (
	OnCursor Action = iota
	OnKeyUp
	OnKeyDown
	OnKeyRepeat
)

// Key corresponds to a keyboard key.
type Key struct {
	code uint32
	char string
}

func (key Key) String() string {
	return key.char
}

// KeyEvent describes a key event over the window
type KeyEvent struct {
	Keycode Key
	Mods    ModifierKey
	Pressed bool
}

func (kev KeyEvent) String() string {
	return fmt.Sprintf(
		"code:%v(%s);mods:%d(%v);down(%v)",
		kev.Keycode.code, kev.Keycode.char,
		kev.Mods, kev.Mods,
		kev.Pressed,
	)
}

// ModifierKey corresponds to a set of modifier keys (bitmask).
// The provided modifiers are platform dependent.
type ModifierKey int

// MouseEvent describes a mouse event over the window
type MouseEvent struct {
	Action  MouseAction
	Button  MouseButton
	Mods    ModifierKey
	Xpos    float32
	Ypos    float32
	Xoffset float32
	Yoffset float32
}

func (mev MouseEvent) String() string {
	return fmt.Sprintf(
		"act:%v;btn:%v;mods:%d(%v);pos(%v,%v);offset(%v,%v)",
		mev.Action,
		mev.Button,
		mev.Mods, mev.Mods,
		mev.Xpos, mev.Ypos,
		mev.Xoffset, mev.Yoffset,
	)
}

// MouseButton corresponds to a mouse button.
type MouseButton int

// Mouse buttons
const (
	MouseBtnNone MouseButton = iota
	MouseBtnLeft
	MouseBtnRight
	MouseBtnMiddle
)

func (btn MouseButton) String() string {
	switch btn {
	case MouseBtnNone:
		return fmt.Sprintf("none(%d)", btn)
	case MouseBtnLeft:
		return fmt.Sprintf("left(%d)", btn)
	case MouseBtnMiddle:
		return fmt.Sprintf("middle(%d)", btn)
	case MouseBtnRight:
		return fmt.Sprintf("right(%d)", btn)
	default:
		return "unknown button"
	}
}

// MouseAction corresponds to a mouse action.
type MouseAction int

const (
	MouseMove MouseAction = iota
	MouseUp
	MouseDown
	MouseScroll
)

func (a MouseAction) String() string {
	switch a {
	case MouseMove:
		return fmt.Sprintf("move(%d)", a)
	case MouseUp:
		return fmt.Sprintf("up(%d)", a)
	case MouseDown:
		return fmt.Sprintf("down(%d)", a)
	case MouseScroll:
		return fmt.Sprintf("scroll(%d)", a)
	default:
		return fmt.Sprintf("unknown(%d)", a)
	}
}
