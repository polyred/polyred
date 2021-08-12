<img src="./examples/favicon.png" alt="logo" height="255" align="right" />

# polyred [![Go Reference](https://pkg.go.dev/badge/github.com/changkun/polyred.svg)](https://pkg.go.dev/poly.red) [![Latest relsease](https://img.shields.io/github/v/tag/changkun/polyred?label=polyred)](https://github.com/changkun/polyred/releases) ![polyred](https://github.com/changkun/polyred/workflows/polyred/badge.svg?branch=master) ![](https://changkun.de/urlstat?mode=github&repo=changkun/polyred) [![codecov](https://codecov.io/gh/changkun/polyred/branch/master/graph/badge.svg?token=PSCJA90S57)](https://codecov.io/gh/changkun/polyred) [![Go Report Card](https://goreportcard.com/badge/github.com/changkun/polyred)](https://goreportcard.com/report/github.com/changkun/polyred)

3D graphics facilities in Go.

```go
import "poly.red"
```

_Caution: still under experiment, expect it to break at any time. Use it at your own risk._

## About

`polyred` is a 3D graphics facility written in Go that aims to offer state-of-the-art graphics research algorithms, especially geometry processing, rendering, animation, and etc.

The geometry facility offers different types of geometry representations (mostly for meshes), iterators, solvers, and relevant I/O processors.

The rendering facility offers two levels of API set where the low-level API set contains abstract rendering passes for flexible customization, whereas the high-level API set contains pre-defined rendering effects for better usability and performance.

See a full features list [here](./docs/features.md).

## Getting started

```go
// Create a scene graph
s := scene.NewScene()
s.SetCamera(camera.NewPerspective(
    math.NewVec4(0, 0.6, 0.9, 1), // position
    math.NewVec4(0, 0, 0, 1),     // lookAt
    math.NewVec4(0, 1, 0, 0),     // up
    45,                           // fov
    float64(1920)/float64(1080),  // aspect
    0.1,                          // near
    2,                            // far
))

// Add lights
s.Add(light.NewPoint(
    light.WithPointLightIntensity(7),
    light.WithPointLightColor(color.RGBA{255, 255, 255, 255}),
    light.WithPointLightPosition(math.NewVec4(4, 4, 2, 1)),
    light.WithPointLightShadowMap(true)),
    light.NewAmbient(light.WithAmbientIntensity(0.5)))

// Load models and setup materials
m := io.MustLoadMesh("bunny.obj")
m.SetMaterial(material.NewBlinnPhong(
    material.WithBlinnPhongTexture(
        image.NewTexture(
            image.WithSource(io.MustLoadImage("bunny.png",
                io.WithGammaCorrection(true)),
            ),
            image.WithIsotropicMipMap(true),
        ),
    ),
    material.WithBlinnPhongFactors(0.6, 0.5),
    material.WithBlinnPhongShininess(150),
    material.WithBlinnPhongShadow(true),
    material.WithBlinnPhongAmbientOcclusion(true),
))
m.Scale(2, 2, 2)
s.Add(m)
m = io.MustLoadMesh("ground.obj")
m.SetMaterial(material.NewBlinnPhong(
    material.WithBlinnPhongTexture(
        image.NewTexture(
            image.WithSource(io.MustLoadImage("ground.png",
                io.WithGammaCorrection(true)),
            ),
            image.WithIsotropicMipMap(true),
        ),
    ),
    material.WithBlinnPhongFactors(0.6, 0.5),
    material.WithBlinnPhongShininess(150),
    material.WithBlinnPhongShadow(true),
))
m.Scale(2, 2, 2)
s.Add(m)

// Create the renderer then render the scene graph!
r := render.NewRenderer(
    render.WithSize(1920, 1080),
    render.WithMSAA(2),
    render.WithScene(s),
    render.WithShadowMap(true),
    render.WithGammaCorrection(true),
)
utils.Save(r.Render(), "./render.png")
```

The above example results:

![](./examples/teaser.png)

See more full examples in the [`examples` folder](./examples), especially
[polywine](https://changkun.de/s/polywine) runs a window that shows
results produced by polyred.

## Contributes

Easiest way to contribute is to provide feedback! I would love to hear
what you like and what you think is missing.
[Issue](https://github.com/changkun/polyred/issues/new) and
[PRs](https://github.com/changkun/polyred/pulls) are also welcome.

## License

Copyright &copy; 2020-2021 [Changkun Ou](https://changkun.de). All rights reserved.