package persp

import (
	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/scene"
	"poly.red/texture"
)

func NewCorrectScene(w, h int) interface{} {
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		camera.WithPosition(math.NewVec3(0, 3, 3)),
		camera.WithPerspFrustum(45, 1, 0.1, 10),
	))
	s.Add(light.NewAmbient(light.WithAmbientIntensity(1)))
	m := mesh.NewPlane(1, 1)
	m.SetMaterial(material.NewBlinnPhong(
		material.WithBlinnPhongTexture(texture.NewTexture(
			texture.WithSource(texture.MustLoadImage("../testdata/uvgrid2.png")),
			texture.WithIsotropicMipMap(true),
		)),
		material.WithBlinnPhongFactors(0.6, 0.5),
		material.WithBlinnPhongShininess(150),
	))
	m.Scale(2, 2, 2)
	s.Add(m)
	return s
}
