<img src="./examples/favicon.png" alt="logo" height="350" align="right" />

# polyred [![Go Reference](https://pkg.go.dev/badge/github.com/changkun/polyred.svg)](https://pkg.go.dev/poly.red) [![Latest relsease](https://img.shields.io/github/v/tag/changkun/polyred?label=polyred)](https://github.com/changkun/polyred/releases) ![polyred](https://github.com/changkun/polyred/workflows/polyred/badge.svg?branch=master) ![](https://changkun.de/urlstat?mode=github&repo=changkun/polyred) [![codecov](https://codecov.io/gh/changkun/polyred/branch/master/graph/badge.svg?token=PSCJA90S57)](https://codecov.io/gh/changkun/polyred) [![Go Report Card](https://goreportcard.com/badge/github.com/changkun/polyred)](https://goreportcard.com/report/github.com/changkun/polyred)

3D graphics facilities in pure Go.

```go
import "poly.red"
```

_Caution: experimenting, expect it to break at any time. Use it at your own risk._

## About

`polyred` is a 3D graphics facility, written in pure Go, aims to
implement graphics research algorithms in real-time and offline
geometry processing, rendering, animation, and etc.

The current implemented features:

- Cross platform
- No dependency
- Cache-aware concurrency optimization
- Mesh I/O
- Built-in geometries
- Scene graph
- Forward rendering
- Mipmapping
- Sutherland Hodgman Clipping
- Back-face culling
- View frustum culling
- Perspective correct interpolation
- Depth testing
- Deferred shading
- Blinn-Phong reflectance model
- Shadow mapping
- Screen-space ambient occlusion
- Supersampling anti-aliasing
- Shader programming

See a full features list [here](./docs/features.md).

## Getting started

```go
package main

import (
	"poly.red/camera"
	"poly.red/gui"
	"poly.red/io"
	"poly.red/light"
	"poly.red/render"
	"poly.red/scene"
)

func main() {
	r := render.NewRenderer()
	s := scene.NewScene()
	s.Add(mesh.New(models.Bunny))
	s.Add(light.NewPoint())
	gui.Show(r.Render(camera.NewPerspective()))
}
```

See more full examples in the [`examples` folder](./examples).

## Contributes

Easiest way to contribute is to provide feedback! I would love to hear what you like and what you think is missing. [Issue](https://github.com/changkun/polyred/issues/new) and [PRs](https://github.com/changkun/polyred/pulls) are also welcome.

## License

Copyright &copy; 2020-2021 [Changkun Ou](https://changkun.de). All rights reserved.