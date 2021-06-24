// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package io

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"changkun.de/x/polyred/geometry"
)

// MustLoadMesh loads a given file to a triangle mesh.
func MustLoadMesh(path string) geometry.Mesh {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("loader: cannot open file %s, err: %v", path, err))
	}
	m, err := LoadOBJ(f)
	f.Close()
	if err != nil {
		panic(fmt.Errorf("cannot load obj model, path: %s, err: %v", path, err))
	}
	return m
}
