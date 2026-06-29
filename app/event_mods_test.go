// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

import "testing"

// TestX11ModsToLogical guards the X11-state -> logical-ModifierKey translation.
// It is pure bit math (no X11/display), so it runs on every platform and pins
// the bug where a raw X11 state was passed straight through as ModifierKey,
// making Contain(ModShift) test the wrong bit and silently killing Shift+drag
// pan on Linux. The X11 mask values are from X11/X.h.
func TestX11ModsToLogical(t *testing.T) {
	const (
		shiftMask   = 1 << 0
		lockMask    = 1 << 1
		controlMask = 1 << 2
		mod1Mask    = 1 << 3 // Alt
		mod2Mask    = 1 << 4 // typically NumLock, intentionally ignored
		mod4Mask    = 1 << 6 // Super / Command
	)
	cases := []struct {
		name  string
		state uint32
		want  ModifierKey
	}{
		{"none", 0, ModNone},
		{"shift", shiftMask, ModShift},
		{"capslock", lockMask, ModCapsLock},
		{"control", controlMask, ModControl},
		{"alt->option", mod1Mask, ModOption},
		{"super->command", mod4Mask, ModCommand},
		{"shift+control", shiftMask | controlMask, ModShift | ModControl},
		{"ignores numlock", mod2Mask, ModNone},
		{"shift survives numlock", shiftMask | mod2Mask, ModShift},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := x11ModsToLogical(c.state)
			if got != c.want {
				t.Fatalf("x11ModsToLogical(%#x) = %v (%#x), want %v (%#x)", c.state, got, uint64(got), c.want, uint64(c.want))
			}
		})
	}

	// The whole point: the logical bit must satisfy Contain, which a raw X11
	// state would not (X11 ShiftMask is bit 0; logical ModShift is bit 17).
	if !x11ModsToLogical(shiftMask).Contain(ModShift) {
		t.Fatal("translated shift state must satisfy Contain(ModShift)")
	}
	if ModifierKey(shiftMask).Contain(ModShift) {
		t.Fatal("raw X11 shift state must NOT satisfy Contain(ModShift); the translation is what fixes this")
	}
}
