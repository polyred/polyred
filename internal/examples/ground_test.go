package example_test

import (
	"testing"

	"poly.red/camera"
	"poly.red/light"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
)

func NewGroundScene(w, h int) (*scene.Scene, camera.Interface) {
	s := scene.NewScene(light.NewAmbient(light.Intensity(1)))
	g := model.MustLoad("../testdata/ground.obj")
	g.Scale(2, 2, 2)
	s.Add(g)
	return s, camera.NewPerspective(
		camera.Position(math.NewVec3[float32](0, 3, 3)),
		camera.ViewFrustum(45, 1, 0.1, 10),
	)
}

func TestGround(t *testing.T) {
	tests := []*BasicOpt{
		{
			Name:       "ground",
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
		s, cam := NewGroundScene(test.Width, test.Height)
		rendopts := []render.Option{render.Camera(cam)}
		rendopts = append(rendopts, test.RenderOpts...)
		Render(t, s, test, rendopts...)
	}
}
