// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh

import (
	"fmt"
	"path/filepath"
)

func MustLoad(path string) Mesh {
	m, err := Load(path)
	if err != nil {
		panic(fmt.Errorf("mesh: cannot load a given mesh: %w", err))
	}
	return m
}

func Load(path string) (Mesh, error) {
	switch filepath.Ext(path) {
	case ".obj":
		return LoadOBJ(path)
	default:
		panic("mesh: unsupported format")
	}
}
