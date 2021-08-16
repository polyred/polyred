// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package example_test

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"testing"
	"time"

	"poly.red/render"
	"poly.red/scene"
	"poly.red/texture/imageutil"
)

type BasicOpt struct {
	Name          string
	Width, Height int
	CPUProf       bool
	MemProf       bool
	ExecTracer    bool
	RenderOpts    []render.Opt
}

func Render(t *testing.T, s *scene.Scene, opt *BasicOpt, opts ...render.Opt) {
	if opt.CPUProf {
		f, err := os.Create(fmt.Sprintf("%s-cpu-%v.pprof", opt.Name, time.Now().Format(time.RFC3339)))
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if opt.MemProf {
		defer func() {
			mem, err := os.Create(fmt.Sprintf("%s-mem-%v.pprof", opt.Name, time.Now().Format(time.RFC3339)))
			if err != nil {
				panic(err)
			}
			defer mem.Close()
			runtime.GC()
			pprof.WriteHeapProfile(mem)
		}()
	}

	if opt.ExecTracer {
		t, err := os.Create(fmt.Sprintf("%s-trace-%v.trace", opt.Name, time.Now().Format(time.RFC3339)))
		if err != nil {
			panic(err)
		}
		defer t.Close()
		trace.Start(t)
		defer trace.Stop()
	}

	allopts := []render.Opt{
		render.Scene(s),
		render.Size(int(opt.Width), int(opt.Height)),
	}
	allopts = append(allopts, opts...)
	r := render.NewRenderer(allopts...)

	path := fmt.Sprintf("./out/%s.png", opt.Name)
	t.Logf("saving result to: %s", path)
	imageutil.Save(r.Render(), path)
}
