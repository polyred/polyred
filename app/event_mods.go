// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

import "strings"

const (
	ModNone     = ModifierKey(0)
	ModCapsLock = ModifierKey(1 << 16)
	ModShift    = ModifierKey(1 << 17)
	ModControl  = ModifierKey(1 << 18)
	ModOption   = ModifierKey(1 << 19)
	ModCommand  = ModifierKey(1 << 20)
)

func (mod ModifierKey) Contain(mod2 ModifierKey) bool {
	return mod&mod2 == mod2
}

// X11 keyboard/button modifier masks (X.h). The X server delivers these in the
// event state field; they are not the logical ModifierKey bits above.
const (
	x11ShiftMask   = 1 << 0 // ShiftMask
	x11LockMask    = 1 << 1 // LockMask (caps lock)
	x11ControlMask = 1 << 2 // ControlMask
	x11Mod1Mask    = 1 << 3 // Mod1Mask, conventionally Alt
	x11Mod4Mask    = 1 << 6 // Mod4Mask, conventionally Super / the Windows/Command key
)

// x11ModsToLogical maps an X11 event state mask to the platform-independent
// logical ModifierKey bits used across the app (and matched by darwin's native
// cocoa flags). Without this translation a raw X11 state would never satisfy
// Contain(ModShift) and gestures like Shift+drag pan would silently break.
func x11ModsToLogical(state uint32) ModifierKey {
	var mods ModifierKey
	if state&x11ShiftMask != 0 {
		mods |= ModShift
	}
	if state&x11LockMask != 0 {
		mods |= ModCapsLock
	}
	if state&x11ControlMask != 0 {
		mods |= ModControl
	}
	if state&x11Mod1Mask != 0 {
		mods |= ModOption
	}
	if state&x11Mod4Mask != 0 {
		mods |= ModCommand
	}
	return mods
}

func (mod ModifierKey) String() string {
	if mod == ModNone {
		return "none"
	}

	var str []string
	if mod.Contain(ModCapsLock) {
		str = append(str, "capslock")
	}
	if mod.Contain(ModControl) {
		str = append(str, "ctrl")
	}
	if mod.Contain(ModCommand) {
		str = append(str, "command")
	}
	if mod.Contain(ModShift) {
		str = append(str, "shift")
	}
	if mod.Contain(ModOption) {
		str = append(str, "option")
	}

	ret := strings.Join(str, "+")
	if ret == "" {
		return "unknown"
	}
	return ret
}
