// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"os"

	"poly.red/internal/term"
	"poly.red/internal/utils"
	"poly.red/texture"
)

func main() {
	img := texture.MustLoadImage("../out/shadow.png")
	w, h, _ := term.GetSize()
	img = utils.Resize(w, h, img)

	t := term.New(term.Size(w, h))
	t.Draw(img)
	t.Flush(os.Stdout)
}
