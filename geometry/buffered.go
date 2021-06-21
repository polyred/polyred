// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package geometry

type BufferedMesh struct {
	vbo        []int64
	attributes map[string][]float64
}

func NewBufferedMesh() *BufferedMesh {
	return &BufferedMesh{}
}

func (bm *BufferedMesh) SetVertexBuffer(vbo []int64) {
	bm.vbo = vbo
}

func (bm *BufferedMesh) SetAttribute(name string, buf []float64) {
	bm.attributes[name] = buf
}

func (bm *BufferedMesh) GetAttribute(name string) []float64 {
	return bm.attributes[name]
}
