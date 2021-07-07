// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gui

import (
	"testing"
)

func TestDispatcher(t *testing.T) {
	d := newDispatcher()

	counter := 0
	d.Subscribe(OnResize, func(e Event) {
		counter++
	})

	d.Dispatch(OnResize, SizeEvent{})

	if counter != 1 {
		t.Fatalf("failed to dispatch event.")
	}
}
