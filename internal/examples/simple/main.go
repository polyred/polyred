// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"poly.red/camera"
	"poly.red/light"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"

	"poly.red/internal/gui" // TODO: make this public
)

func main() {
	// Create a scene graph
	s := scene.NewScene()

	// Create and add a point light and a bunny to the scene graph
	s.Add(light.NewPoint(), model.StanfordBunny())

	// Create a camera for the rendering
	c := camera.NewPerspective()

	// Create a renderer and specify scene and camera
	r := render.NewRenderer(render.Scene(s), render.Camera(c))

	// Render and show the result in a window
	gui.Show(r.Render())
}
