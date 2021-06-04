// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/geometry"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/rend"
	"changkun.de/x/ddd/win"
	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.design/x/mainthread"
)

func loadMesh(path string) *geometry.TriangleMesh {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("loader: cannot open file %s, err: %v", path, err))
	}
	m, err := geometry.LoadOBJ(f)
	f.Close()
	if err != nil {
		panic(fmt.Errorf("cannot load obj model, path: %s, err: %v", path, err))
	}
	return m
}

func loadTexture(path string) *material.Texture {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("loader: cannot open file %s, err: %v", path, err))
	}
	img, err := png.Decode(f)
	f.Close()
	if err != nil {
		panic(fmt.Errorf("cannot load obj model, path: %s, err: %v", path, err))
	}
	var data *image.RGBA
	if v, ok := img.(*image.NRGBA); ok {
		data = (*image.RGBA)(v)
	} else if v, ok := img.(*image.RGBA); ok {
		data = v
	}

	return material.NewTexture(data)
}
func loadScene(width, height int) *rend.Scene {
	s := rend.NewScene()

	c := camera.NewPerspectiveCamera(
		math.NewVector(-0.5, 0.5, 0.5, 1),
		math.NewVector(0, 0, -0.5, 1),
		math.NewVector(0, 1, 0, 0),
		45,
		float64(width)/float64(height),
		-0.1,
		-3,
	)
	s.UseCamera(c)

	l := light.NewPointLight(color.RGBA{0, 0, 0, 255}, math.NewVector(-200, 250, 600, 1))
	s.AddLight(l)

	m := loadMesh("../../testdata/bunny.obj")
	tex := loadTexture("../../testdata/bunny.png")
	mat := material.NewBlinnPhongMaterial(tex, color.RGBA{0, 125, 255, 255}, 0.6, 1, 0.5, 150)
	m.UseMaterial(mat)
	m.Rotate(math.NewVector(0, 1, 0, 0), -math.Pi/6)
	m.Translate(0, -0, -0.4)
	s.AddMesh(m)

	m = loadMesh("../../testdata/ground.obj")
	tex = loadTexture("../../testdata/ground.png")
	mat = material.NewBlinnPhongMaterial(tex, color.RGBA{0, 125, 255, 255}, 0.6, 1, 0.5, 150)
	m.UseMaterial(mat)
	m.Rotate(math.NewVector(0, 1, 0, 0), -math.Pi/6)
	m.Translate(0, -0, -0.4)
	s.AddMesh(m)

	return s
}

func main() {
	mainthread.Init(func() {
		mainthread.Call(func() {
			err := glfw.Init()
			if err != nil {
				panic(err)
			}
		})
		defer func() { mainthread.Call(glfw.Terminate) }()

		fn()
	})
}
func fn() {
	width, height, msaa := 1920, 1080, 2
	w, err := win.NewWindow(
		win.Title("window"),
		win.Size(width, height),
		win.Resizable(), win.ShowFPS(),
	)
	if err != nil {
		log.Fatalf("failed to create a window: %v", err)
	}

	r := rend.NewRasterizer(width, height, msaa)
	w.SetRenderer(r)

	s := loadScene(width, height)
	r.SetScene(s)

	// cpu pprof
	f, err := os.Create("cpu.pprof")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	// mem pprof
	defer func() {
		mem, err := os.Create("mem.pprof")
		if err != nil {
			panic(err)
		}
		defer mem.Close()
		runtime.GC()
		pprof.WriteHeapProfile(mem)
	}()

	// trace
	t, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer t.Close()
	trace.Start(t)
	defer trace.Stop()
	w.Run()
}
