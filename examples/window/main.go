// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/image"
	"changkun.de/x/ddd/internal/win"
	"changkun.de/x/ddd/io"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/rend"
	"changkun.de/x/ddd/scene"
	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.design/x/mainthread"
)

func loadScene(width, height int) *scene.Scene {
	s := scene.NewScene()

	c := camera.NewPerspective(
		math.NewVector(-0.5, 0.5, 0.5, 1),
		math.NewVector(0, 0, -0.5, 1),
		math.NewVector(0, 1, 0, 0),
		45,
		float64(width)/float64(height),
		0.1,
		3,
	)
	s.SetCamera(c)
	s.Add(light.NewPoint(
		light.WithPointLightIntensity(20),
		light.WithPointLightColor(color.RGBA{0, 0, 0, 255}),
		light.WithPointLightPosition(math.NewVector(-200, 250, 600, 1)),
	), light.NewAmbient(
		light.WithAmbientIntensity(0.5),
	))

	m := io.MustLoadMesh("../../testdata/bunny.obj")
	data := io.MustLoadImage("../../testdata/bunny.png")
	tex := image.NewTexture(
		image.WithData(data),
		image.WithIsotropicMipMap(true),
	)
	mat := material.NewBlinnPhong(
		material.WithBlinnPhongTexture(tex),
		material.WithBlinnPhongFactors(0.6, 1),
		material.WithBlinnPhongShininess(150),
	)
	m.SetMaterial(mat)
	m.Rotate(math.NewVector(0, 1, 0, 0), -math.Pi/6)
	m.Translate(0, -0, -0.4)
	s.Add(m)

	m = io.MustLoadMesh("../../testdata/ground.obj")
	data = io.MustLoadImage("../../testdata/ground.png")
	tex = image.NewTexture(
		image.WithData(data),
		image.WithIsotropicMipMap(true),
	)
	mat = material.NewBlinnPhong(
		material.WithBlinnPhongTexture(tex),
		material.WithBlinnPhongFactors(0.6, 1),
		material.WithBlinnPhongShininess(150),
	)
	m.SetMaterial(mat)
	m.Rotate(math.NewVector(0, 1, 0, 0), -math.Pi/6)
	m.Translate(0, -0, -0.4)
	s.Add(m)

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
	width, height, msaa := 1600, 1000, 2
	w, err := win.NewWindow(
		win.Title("window"),
		win.Size(width, height),
		win.Resizable(), win.ShowFPS(),
	)
	if err != nil {
		log.Fatalf("failed to create a window: %v", err)
	}
	s := loadScene(width, height)

	r := rend.NewRenderer(
		rend.WithSize(width, height),
		rend.WithMSAA(msaa),
		rend.WithScene(s),
	)
	w.SetRenderer(r)

	// cpu pprof
	f, err := os.Create(fmt.Sprintf("cpu-%v.pprof", time.Now().Format(time.RFC3339)))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	// mem pprof
	defer func() {
		mem, err := os.Create(fmt.Sprintf("mem-%v.pprof", time.Now().Format(time.RFC3339)))
		if err != nil {
			panic(err)
		}
		defer mem.Close()
		runtime.GC()
		pprof.WriteHeapProfile(mem)
	}()

	// trace
	t, err := os.Create(fmt.Sprintf("trace-%v.trace", time.Now().Format(time.RFC3339)))
	if err != nil {
		panic(err)
	}
	defer t.Close()
	trace.Start(t)
	defer trace.Stop()

	go func() {
		w.Run()
	}()
	time.Sleep(time.Second * 5)
}
