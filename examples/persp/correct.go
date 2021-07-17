package persp

import (
	"poly.red/camera"
	"poly.red/geometry"
	"poly.red/image"
	"poly.red/io"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/scene"
)

func NewCorrectScene(w, h int) interface{} {
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		math.NewVec3(0, 3, 3),
		math.NewVec3(0, 0, 0),
		math.NewVec3(0, 1, 0),
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
