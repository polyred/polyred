package persp

import (
	"changkun.de/x/polyred/camera"
	"changkun.de/x/polyred/geometry"
	"changkun.de/x/polyred/image"
	"changkun.de/x/polyred/io"
	"changkun.de/x/polyred/light"
	"changkun.de/x/polyred/material"
	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/scene"
)

func NewCorrectScene(w, h int) interface{} {
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		math.NewVec4(0, 3, 3, 1),
		math.NewVec4(0, 0, 0, 1),
		math.NewVec4(0, 1, 0, 0),
		45,
		1,
		0.1, 10,
	))
	s.Add(light.NewAmbient(light.WithAmbientIntensity(1)))
	m := geometry.NewPlane(1, 1)
	m.SetMaterial(material.NewBlinnPhong(
		material.WithBlinnPhongTexture(image.NewTexture(
			image.WithSource(io.MustLoadImage("../testdata/uvgrid2.png")),
			image.WithIsotropicMipMap(true),
		)),
		material.WithBlinnPhongFactors(0.6, 0.5),
		material.WithBlinnPhongShininess(150),
	))
	m.Scale(2, 2, 2)
	s.Add(m)
	return s
}
