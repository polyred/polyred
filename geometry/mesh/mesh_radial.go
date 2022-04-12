// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh

import (
	"poly.red/buffer"
	"poly.red/geometry/primitive"
)

var _ Mesh[float32] = &EditMesh{}

// EditMesh implements Radial Edge Structure that permits convinient mesh editing.
//
// Ref: Weiler, K.J. : The Radial Edge Structure: A Topological Representation
// for Non-Manifold Geometric Modeling. in Geometric Modeling for CAD Applications,
// Springer Verlag, May 1986.
type EditMesh struct {
}

func NewEditMesh(faces []primitive.Face) *EditMesh {
	panic("unimplemented")
}

func (m *EditMesh) AABB() primitive.AABB              { panic("unimplemented") }
func (m *EditMesh) Normalize()                        { panic("unimplemented") }
func (m *EditMesh) IndexBuffer() buffer.IndexBuffer   { panic("unimplemented") }
func (m *EditMesh) VertexBuffer() buffer.VertexBuffer { panic("unimplemented") }
func (m *EditMesh) Triangles() []*primitive.Triangle  { panic("unimplemented") }
