// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

import "strings"

const (
	ModNone     = ModifierKey(0)
	ModShift    = ModifierKey(1 << 0)
	ModCapsLock = ModifierKey(1 << 1)
	ModControl  = ModifierKey(1 << 2)
	Mod1        = ModifierKey(1 << 3)
	Mod2        = ModifierKey(1 << 4)
	Mod3        = ModifierKey(1 << 5)
	Mod4        = ModifierKey(1 << 6)
	Mod5        = ModifierKey(1 << 7)
)

func (mod ModifierKey) Contain(mod2 ModifierKey) bool {
	return mod&mod2 == mod2
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
	if mod.Contain(ModShift) {
		str = append(str, "shift")
	}
	if mod.Contain(Mod1) {
		str = append(str, "mod1")
	}
	if mod.Contain(Mod2) {
		str = append(str, "mod2")
	}
	if mod.Contain(Mod3) {
		str = append(str, "mod3")
	}
	if mod.Contain(Mod4) {
		str = append(str, "mod4")
	}
	if mod.Contain(Mod5) {
		str = append(str, "mod5")
	}

	ret := strings.Join(str, "+")
	if ret == "" {
		return "unknown"
	}
	return ret
}
