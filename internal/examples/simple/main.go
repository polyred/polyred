// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/internal/gui"
	"poly.red/light"
	"poly.red/render"
	"poly.red/scene"
)

func main() {
	// Create a scene graph
	s := scene.NewScene()

	// Load and add the mesh to the scene graph
	s.Add(mesh.MustLoad("../../testdata/bunny.obj"))

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
