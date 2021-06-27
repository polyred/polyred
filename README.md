<img src="./examples/favicon.png" alt="logo" height="255" align="right" />

# polyred [![Go Reference](https://pkg.go.dev/badge/github.com/changkun/polyred.svg)](https://pkg.go.dev/changkun.de/x/polyred) [![Latest relsease](https://img.shields.io/github/v/tag/changkun/polyred?label=polyred)](https://github.com/changkun/polyred/releases) ![polyred](https://github.com/changkun/polyred/workflows/polyred/badge.svg?branch=master) ![](https://changkun.de/urlstat?mode=github&repo=changkun/polyred)

3D graphics facilities in pure Go.

```go
import "changkun.de/x/polyred"
```

_Caution: experimenting, expect it to break at any time. Use it at your own risk._

## About

`polyred` is a 3D graphics facility written in pure Go. It implements
the rasterization as well as ray tracing pipelines for hybrid
rendering. Although it is a software implementation, it remains fast as it
is optimized to utilize the full power of CPUs. The project aims to
provide a software fallback for real-time and offline geometry processing, rendering, and
etc in graphics research. See a full [features list](./docs/features.md).

## Getting started

```go
// Create a scene graph
s := scene.NewScene()
s.SetCamera(camera.NewPerspective(
    math.NewVector(0, 0.6, 0.9, 1),         // position
    math.NewVector(0, 0, 0, 1),             // lookAt
    math.NewVector(0, 1, 0, 0),             // up
    45,                                     // fov
    float64(opt.width)/float64(opt.height), // aspect
    0.1,                                    // near
    2,                                      // far
))

// Add lights
s.Add(light.NewPoint(
    light.WithPointLightIntensity(7),
    light.WithPointLightColor(color.RGBA{255, 255, 255, 255}),
    light.WithPointLightPosition(math.NewVector(4, 4, 2, 1)),
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

The rendered results:

![](./examples/teaser.png)

See more full examples in the [`examples` folder](./examples), and a 
connecting project [polywine](https://changkun.de/s/polywine) to put
polyred results on a window.

## Contributes

Easiest way to contribute is to provide feedback! I would love to hear what you like and what you think is missing. [Issue](https://github.com/changkun/polyred/issues/new) and [PRs](https://github.com/changkun/polyred/pulls) are also welcome.

## License

Copyright &copy; 2020-2021 [Changkun Ou](https://changkun.de). All rights reserved.