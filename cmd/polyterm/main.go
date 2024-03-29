// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"poly.red/internal/imageutil"
	"poly.red/internal/term"
)

var t *term.Terminal

func init() {
	tw, th, err := term.GetSize()
	if err != nil {
		panic(err)
	}

	// subtract 5 lines of additiona console output.
	t = term.New(term.Size(tw, th-5))
}

func main() {
	// TODO: make this example interactive with controls.
	t.Draw(imageutil.MustLoadImage("../../internal/examples/out/shadow.png"))
	t.Flush()
}
