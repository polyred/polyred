// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build darwin

package gpu

import (
	"log"

	"poly.red/internal/driver/mtl"
)

var (
	device mtl.Device
	addFn  shaderFn
)

func init() {
	defer handle(func(err error) {
		if err != nil {
			log.Println(err)
		}
	})

	device = try(mtl.CreateSystemDefaultDevice())
	addFn = try(newAddShader(device))
}

type shaderFn struct {
	fn  mtl.Function
	cps mtl.ComputePipelineState
}
