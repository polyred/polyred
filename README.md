<img src="./internal/examples/favicon.png" alt="logo" height="350" align="right" />

# polyred [![Go Reference](https://pkg.go.dev/badge/github.com/changkun/polyred.svg)](https://pkg.go.dev/poly.red) [![Latest relsease](https://img.shields.io/github/v/tag/changkun/polyred?label=polyred)](https://github.com/changkun/polyred/releases) ![polyred](https://github.com/changkun/polyred/workflows/polyred/badge.svg?branch=master) ![](https://changkun.de/urlstat?mode=github&repo=changkun/polyred) [![codecov](https://codecov.io/gh/changkun/polyred/branch/master/graph/badge.svg?token=PSCJA90S57)](https://codecov.io/gh/changkun/polyred) [![Go Report Card](https://goreportcard.com/badge/github.com/changkun/polyred)](https://goreportcard.com/report/github.com/changkun/polyred)

3D graphics facilities in Go.

```go
import "poly.red"
```

_Caution: still under experiment, expect it to break at any time. Use it at your own risk._

## About

`polyred` is a 3D graphics facility written in Go that aims to offer
state-of-the-art graphics research algorithms, especially geometry processing,
rendering, animation, and etc.

See a full features list [here](https://github.com/changkun/polyred/wiki/features).

## Getting started

The following code snippet shows a minimum example:

```go
package main

import (
	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/light"
	"poly.red/render"
	"poly.red/scene"

	"poly.red/internal/gui" // TODO: make this public
)

func main() {
	// Create a scene graph
	s := scene.NewScene()

	// Load and add the mesh to the scene graph
	s.Add(mesh.MustLoad("bunny.obj"))

	// Create and add a point light to the scene graph
	s.Add(light.NewPoint())

	// Create a camera for the rendering
	c := camera.NewPerspective()

	// Create a renderer and specify scene and camera
	r := render.NewRenderer(
		render.Scene(s),
		render.Camera(c),
	)

	// Render and show the result in a window
	gui.Show(r.Render())
}
```

The above example results:

![](./internal/examples/teaser.png)

See more full examples in the [`examples` folder](./internal/examples).

## Contributes

Easiest way to contribute is to provide feedback! I would love to hear
what you like and what you think is missing.
[Issue](https://github.com/changkun/polyred/issues/new) and
[PRs](https://github.com/changkun/polyred/pulls) are also welcome.

## License

Copyright &copy; 2020-2021 [Changkun Ou](https://changkun.de). All rights reserved.