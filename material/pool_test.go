package material_test

import (
	"testing"

	"poly.red/color"
	"poly.red/material"
)

func TestDefault(t *testing.T) {
	m := material.Default()
	if m == nil {
		t.Fatal("Default() returned nil")
	}
	if m.Texture.Query(0, 0, 0) != color.Blue {
		t.Fatalf("default material is not the blue texture")
	}
}

// TestNewBlinnPhong checks NewBlinnPhong returns a configured material directly
// (no global pool, no ID).
func TestNewBlinnPhong(t *testing.T) {
	want := color.RGBA{R: 11, G: 22, B: 33, A: 255}
	m := material.NewBlinnPhong(material.Diffuse(want))
	if m == nil {
		t.Fatal("NewBlinnPhong returned nil")
	}
	if m.Diffuse != want {
		t.Errorf("Diffuse = %v, want %v", m.Diffuse, want)
	}
}
