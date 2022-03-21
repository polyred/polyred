// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"image/color"

	"poly.red/buffer"
	"poly.red/geometry/primitive"
	"poly.red/light"
	"poly.red/math"
)

// Material is an interface that represents a mesh material
type Material interface {
	ReceiveShadow() bool
	AmbientOcclusion() bool
	Texture() *buffer.Texture
	VertexShader(v *primitive.Vertex) *primitive.Vertex
	FragmentShader(
		info buffer.Fragment,
		camera math.Vec3[float32],
		ls []light.Source,
		es []light.Environment,
	) color.RGBA
}
