package material_test

import (
	"testing"

	"poly.red/color"
	"poly.red/material"
)

func TestGet(t *testing.T) {
	mat := material.Get(0)
	if mat == nil {
		t.Fatalf("cannot find default material.")
	}

	m, ok := mat.(*material.BlinnPhong)
	if !ok {
		t.Fatalf("default material is not Blinn-Phong.")
	}

	if m.Texture.Query(0, 0, 0) != color.Blue {
		t.Fatalf("default material is not with blue texture.")
	}
}
