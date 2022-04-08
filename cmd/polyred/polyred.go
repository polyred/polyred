// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// The polyred command offers a mesh processing facilities.
package main // go install poly.red/cmd/polyred@latest

import (
	"flag"
	"fmt"
	"os"

	"poly.red/app"
)

func main() {
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage:
polyred show /path/to/mesh.obj
`)
	}

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		return
	}
	switch args[0] {
	case "show":
		if len(args) != 2 {
			flag.Usage()
			return
		}

		_, err := os.Stat(args[1])
		if err != nil {
			flag.Usage()
			return
		}
		app.Run(newApp(args[1]),
			app.Title("polyred"),
			app.MinSize(80, 60),
			app.MaxSize(1920*2, 1080*2),
			app.FPS(false),
		)
	default:
		flag.Usage()
		return
	}
}
