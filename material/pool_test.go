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

// TestResolve pins the single material-resolution path, including the negative-ID
// "use vertex color" fallback that the material-ownership refactor must preserve.
func TestResolve(t *testing.T) {
	// Negative ID: the vertex-color hint, must resolve to nil (not the default).
	if material.Resolve(-1) != nil {
		t.Error("Resolve(-1) should be nil (negative ID -> vertex color)")
	}
	// Absent ID: nil.
	if material.Resolve(1 << 40) != nil {
		t.Error("Resolve(absent) should be nil")
	}
	// ID 0: the default material, non-nil.
	if material.Resolve(0) == nil {
		t.Error("Resolve(0) should be the default material")
	}
	// A registered BlinnPhong resolves to itself (not swapped).
	want := color.RGBA{R: 11, G: 22, B: 33, A: 255}
	id := material.NewBlinnPhong(material.Diffuse(want))
	bp := material.Resolve(id)
	if bp == nil {
		t.Fatal("Resolve(registered) returned nil")
	}
	if bp.Diffuse != want {
		t.Errorf("Resolve returned wrong material: Diffuse=%v want %v", bp.Diffuse, want)
	}
}
