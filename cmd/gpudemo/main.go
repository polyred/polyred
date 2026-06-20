// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command gpudemo renders a colored triangle on the GPU through poly.red/gpu —
// shaders authored in Go (poly.red/gpu/shader), compiled to the backend
// shading language, run cgo-free — and writes the result to a PNG. It is the
// first non-test consumer of the GPU abstraction and doubles as a usage example.
//
// Usage: go run ./cmd/gpudemo [-o out.png] [-size N]
package main

import (
	"errors"
	"flag"
	"image"
	"image/png"
	"log"
	"os"
	"unsafe"

	"poly.red/gpu"
	"poly.red/gpu/shader"
)

// The whole pipeline, authored in Go and compiled to a shading language.
const src = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

type VOut struct {
	Pos   Vec4 ` + "`gpu:\"position\"`" + `
	Color Vec4
}

//gpu:vertex
func VMain(vid uint, pos []float32, col []float32) VOut {
	return VOut{
		Pos:   Vec4{pos[vid*2], pos[vid*2+1], 0, 1},
		Color: Vec4{col[vid*3], col[vid*3+1], col[vid*3+2], 1},
	}
}

//gpu:fragment
func FMain(in VOut) Vec4 {
	return in.Color
}
`

func main() {
	out := flag.String("o", "triangle.png", "output PNG path")
	size := flag.Int("size", 256, "output image size (pixels)")
	flag.Parse()

	if err := run(*out, *size); err != nil {
		if errors.Is(err, gpu.ErrUnsupported) {
			log.Printf("no GPU backend available on this platform: %v", err)
			return
		}
		log.Fatal(err)
	}
	log.Printf("wrote %s", *out)
}

func run(out string, size int) error {
	dev, err := gpu.Open()
	if err != nil {
		return err
	}
	defer dev.Close()
	log.Printf("GPU driver: %s", dev.Driver())

	ks, err := shader.Compile(src)
	if err != nil {
		return err
	}
	vmod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["VMain"].MSL})
	if err != nil {
		return err
	}
	fmod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["FMain"].MSL})
	if err != nil {
		return err
	}
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: vmod, VertexEntry: "VMain",
		FragmentModule: fmod, FragmentEntry: "FMain",
		ColorFormat: gpu.RGBA8Unorm,
	})
	if err != nil {
		return err
	}

	target, err := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.RGBA8Unorm, Width: size, Height: size, RenderTarget: true})
	if err != nil {
		return err
	}

	// A centered triangle, vertices coloured red, green, blue.
	pos := []float32{0, 0.8, -0.8, -0.6, 0.8, -0.6}
	col := []float32{1, 0, 0, 0, 1, 0, 0, 0, 1}
	posBuf, err := dev.NewBuffer(gpu.BufferDescriptor{Size: len(pos) * 4, Usage: gpu.BufferStorage, Data: floatBytes(pos)})
	if err != nil {
		return err
	}
	colBuf, err := dev.NewBuffer(gpu.BufferDescriptor{Size: len(col) * 4, Usage: gpu.BufferStorage, Data: floatBytes(col)})
	if err != nil {
		return err
	}

	enc := dev.NewCommandEncoder()
	rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{
		ColorTexture: target, Load: gpu.LoadClear, ClearColor: [4]float64{0.1, 0.1, 0.12, 1},
	})
	rp.SetPipeline(pipe)
	rp.SetVertexBuffer(0, posBuf)
	rp.SetVertexBuffer(1, colBuf)
	rp.Draw(gpu.TriangleList, 0, 3)
	rp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	copy(img.Pix, target.ReadPixels())

	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func floatBytes(d []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}
