// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.
// Modified from https://github.com/g3n/engine/blob/master/loader/obj/obj.go

package example_test

import (
	"testing"

	"poly.red/camera"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
)

func NewCorrectScene(w, h int) (*scene.Scene, camera.Interface) {
	s := scene.NewScene()
	g := model.MustLoad("../testdata/perspect.obj")
	g.Scale(2, 2, 2)
	s.Add(g)
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
				render.GammaCorrection(true),
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
