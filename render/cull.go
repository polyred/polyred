// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"poly.red/buffer"
	"poly.red/geometry/primitive"
	"poly.red/math"
)

// cullViewFrustum reports whether the triangle is entirely outside the viewport
// AABB and can be skipped.
func (r *Renderer) cullViewFrustum(buf *buffer.FragmentBuffer, v1, v2, v3 math.Vec4[float32]) bool {
	viewportAABB := primitive.NewAABB(
		math.NewVec3(float32(buf.Bounds().Dx()*r.cfg.MSAA), float32(buf.Bounds().Dy()*r.cfg.MSAA), 1),
		math.NewVec3[float32](0, 0, 0),
		math.NewVec3[float32](0, 0, -1),
	)
	triangleAABB := primitive.NewAABB(v1.ToVec3(), v2.ToVec3(), v3.ToVec3())
	return !viewportAABB.Intersect(triangleAABB)
}

// cullBackFace reports whether the triangle faces away from the viewer.
func (r *Renderer) cullBackFace(v1, v2, v3 math.Vec4[float32]) bool {
	return v2.Sub(v1).Cross(v3.Sub(v1)).Z < 0
}
