// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh

import (
	"poly.red/buffer"
	"poly.red/geometry/primitive"
)

var _ Mesh[float32] = &PolygonMesh{}

// PolygonMesh is a polygon based mesh structure that can contain
// arbitrary shaped faces, such as triangle and quad mixed mesh.
type PolygonMesh struct {
}

func NewPolygonMesh(faces []primitive.Face) *PolygonMesh {
	panic("unimplemented")
}

func (m *PolygonMesh) AABB() primitive.AABB              { panic("unimplemented") }
func (m *PolygonMesh) Normalize()                        { panic("unimplemented") }
func (m *PolygonMesh) IndexBuffer() buffer.IndexBuffer   { panic("unimplemented") }
func (m *PolygonMesh) VertexBuffer() buffer.VertexBuffer { panic("unimplemented") }
func (m *PolygonMesh) Triangles() []*primitive.Triangle  { panic("unimplemented") }
