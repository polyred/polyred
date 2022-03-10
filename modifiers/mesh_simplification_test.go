package modifiers_test

import (
	"testing"

	"poly.red/geometry/mesh"
	"poly.red/model"
	"poly.red/modifiers"
)

func TestSimplify(t *testing.T) {
	m := model.StanfordBunnyAs[*mesh.TriangleSoup]()
	modifiers.Simplify(m, 1000)
}
