// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh

import (
	"fmt"
	"path/filepath"
)

func MustLoadAs[T Mesh](path string) T {
	m, err := LoadAs[T](path)
	if err != nil {
		panic(fmt.Errorf("mesh: cannot load a given mesh: %w", err))
	}
	return m
}

func LoadAs[T Mesh](path string) (T, error) {
	switch filepath.Ext(path) {
	case ".obj":
		return LoadObjAs[T](path)
	default:
		panic("mesh: unsupported format")
	}
}
