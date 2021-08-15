package example_test

import (
	"testing"

	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/render"
	"poly.red/scene"
	"poly.red/texture"
)

func NewCorrectScene(w, h int) (*scene.Scene, camera.Interface) {
	s := scene.NewScene()
	s.Add(light.NewAmbient(light.Intensity(1)))
	m := mesh.NewPlane(1, 1)
	m.SetMaterial(material.NewBlinnPhong(
		material.Texture(texture.NewTexture(
			texture.WithSource(texture.MustLoadImage("../testdata/uvgrid2.png")),
			texture.WithIsotropicMipMap(true),
		)),
		material.Kdiff(0.6), material.Kspec(0.5),
		material.Shininess(150),
	))
	m.Scale(2, 2, 2)
	s.Add(m)
	return s, camera.NewPerspective(
		camera.Position(math.NewVec3(0, 3, 3)),
		camera.ViewFrustum(45, 1, 0.1, 10),
	)
}

func TestPerspectiveCorrection(t *testing.T) {
	tests := []*BasicOpt{
		{
			Name:       "perspect",
			Width:      500,
			Height:     500,
			CPUProf:    false,
			MemProf:    false,
			ExecTracer: false,
			RenderOpts: []render.Opt{
				render.Debug(false),
				render.MSAA(1),
				render.ShadowMap(false),
			},
		},
	}

	for _, test := range tests {
		t.Logf("%s under settings: %#v", test.Name, test)
		s, cam := NewCorrectScene(test.Width, test.Height)
		rendopts := []render.Opt{render.Camera(cam)}
		rendopts = append(rendopts, test.RenderOpts...)
		Render(t, s, test, rendopts...)
	}
}
