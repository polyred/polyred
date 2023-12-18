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
