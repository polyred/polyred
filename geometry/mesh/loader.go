// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh

import "path/filepath"

func Load(path string) (Mesh, error) {
	switch filepath.Ext(path) {
	case "obj":
		return LoadOBJ(path)
	default:
		panic("mesh: unsupported format")
	}
}
