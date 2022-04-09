package example_test

import (
	"testing"

	"poly.red/buffer"
	"poly.red/camera"
	"poly.red/color"
	"poly.red/internal/imageutil"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
)

func NewCorrectScene(w, h int) (*scene.Scene, camera.Interface) {
	s := scene.NewScene(light.NewAmbient(light.Intensity(1)))
	m := model.NewPlane(1, 1)
	// FIXME: setup material
	_ = material.NewBlinnPhong(
		material.Texture(buffer.NewTexture(
			buffer.TextureImage(imageutil.MustLoadImage("../testdata/uvgrid2.png")),
			buffer.TextureIsoMipmap(true),
		)),
		material.Kdiff(color.FromValue(0.6, 0.6, 0.6, 1.0)), material.Kspec(color.FromValue(0.5, 0.5, 0.5, 1.0)),
		material.Shininess(150),
	)

	m.Scale(2, 2, 2)
	s.Add(m)
	return s, camera.NewPerspective(
		camera.Position(math.NewVec3[float32](0, 3, 3)),
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
			RenderOpts: []render.Option{
				render.Debug(false),
				render.MSAA(1),
				render.ShadowMap(false),
			},
		},
	}

	for _, test := range tests {
		t.Logf("%s under settings: %#v", test.Name, test)
		s, cam := NewCorrectScene(test.Width, test.Height)
		rendopts := []render.Option{render.Camera(cam)}
		rendopts = append(rendopts, test.RenderOpts...)
		Render(t, s, test, rendopts...)
	}
}
