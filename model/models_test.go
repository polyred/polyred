package model_test

import (
	"testing"

	"poly.red/geometry/mesh"
	"poly.red/math"
	"poly.red/model"
	"poly.red/scene"
)

func TestStanfordBunny(t *testing.T) {
	g := model.StanfordBunny()

	scene.IterObjects(g, func(o mesh.Mesh[float32], modelMatrix math.Mat4[float32]) bool {
		t.Log(o.Triangles(), modelMatrix)
		return true
	})
}
