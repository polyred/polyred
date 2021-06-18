// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"image/color"

	"changkun.de/x/ddd/geometry/primitive"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/math"
)

// Material is an interface that represents a mesh material
type Material interface {
	ReceiveShadow() bool
	Texture() *Texture
	VertexShader(
		v primitive.Vertex,
		uniforms map[string]interface{},
	) primitive.Vertex
	FragmentShader(col color.RGBA, x, n, camera math.Vector, ls []light.Source, es []light.Environment) color.RGBA
}
