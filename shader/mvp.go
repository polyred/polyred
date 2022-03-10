package shader

import (
	"poly.red/geometry/primitive"
	"poly.red/math"
)

var MVPAttr primitive.AttrName = "MVP"

type MVP struct {
	Model       math.Mat4[float32]
	ModelInv    math.Mat4[float32]
	View        math.Mat4[float32]
	ViewInv     math.Mat4[float32]
	Proj        math.Mat4[float32]
	ProjInv     math.Mat4[float32]
	Viewport    math.Mat4[float32]
	ViewportInv math.Mat4[float32]
	// NormalMatrix can be ((Tcamera * Tmodel)^(-1))^T or ((Tmodel)^(-1))^T
	// depending on which transformation space. Here we use the 2nd form,
	// i.e. model space normal matrix to save some computation of camera
	// transforamtion in the shading process.
	// The reason we need normal matrix is that normals are transformed
	// incorrectly using MVP matrices. However, a normal matrix helps us
	// to fix the problem.
	Normal          math.Mat4[float32]
	NormalInv       math.Mat4[float32]
	ViewportToWorld math.Mat4[float32]
}
