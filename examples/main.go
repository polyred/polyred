// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"

	"changkun.de/x/polyred/examples/gopher"
	"changkun.de/x/polyred/examples/mcguire"
	"changkun.de/x/polyred/rend"
	"changkun.de/x/polyred/scene"
	"changkun.de/x/polyred/utils"
)

type sceneExample struct {
	Name          string
	Func          func(w, h int) interface{} // *rend.Scene or []*rend.Scene
	Width, Height int
	MSAA          int
	ShadowMap     bool
	Debug         bool
	CPUProf       bool
	MemProf       bool
	ExecTracer    bool
}

func main() {
	examples := []sceneExample{
		// {"bunny", bunny.NewBunnyScene, 960, 540, 2, false, true, true, true, true},
		// {"dragon", dragon.NewDragonScene, 500, 500, 2, false, true, false, false, false},
		// {"shadow", shadow.NewShadowScene, 960, 540, 2, true, false, true, false, false},
		{"gopher", gopher.NewGopherScene, 500, 500, 1, true, false, false, false, false},
		// {"mcguire", mcguire.NewMcGuireScene, 540, 540, 2, false, true, false, false, false},
	}

	for _, example := range examples {
		log.Printf("rendering... %s under settings: %#v", example.Name, example)
		render(&example)
	}
}

func render(settings *sceneExample) {
	if settings.CPUProf {
		f, err := os.Create(fmt.Sprintf("%s-cpu-%v.pprof", settings.Name, time.Now().Format(time.RFC3339)))
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if settings.MemProf {
		defer func() {
			mem, err := os.Create(fmt.Sprintf("%s-mem-%v.pprof", settings.Name, time.Now().Format(time.RFC3339)))
			if err != nil {
				panic(err)
			}
			defer mem.Close()
			runtime.GC()
			pprof.WriteHeapProfile(mem)
		}()
	}

	if settings.ExecTracer {
		t, err := os.Create(fmt.Sprintf("%s-trace-%v.trace", settings.Name, time.Now().Format(time.RFC3339)))
		if err != nil {
			panic(err)
		}
		defer t.Close()
		trace.Start(t)
		defer trace.Stop()
	}

	s := settings.Func(settings.Width, settings.Height)
	r := rend.NewRenderer(
		rend.WithSize(int(settings.Width), int(settings.Height)),
		rend.WithMSAA(settings.MSAA),
		rend.WithShadowMap(settings.ShadowMap),
		rend.WithDebug(settings.Debug),
	)
	switch ss := s.(type) {
	case *scene.Scene:
		r.UpdateOptions(
			rend.WithScene(ss),
		)
		utils.Save(r.Render(), fmt.Sprintf("./%s/%s.png", settings.Name, settings.Name))
	case []*mcguire.Scene:
		for _, sss := range ss {
			r.UpdateOptions(
				rend.WithScene(sss.Scene),
			)
			log.Printf("scene name: %s", sss.Name)
			utils.Save(r.Render(), fmt.Sprintf("./%s/%s.png", settings.Name, sss.Name))
		}
	}
}
