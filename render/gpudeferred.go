// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"errors"
	"fmt"
	"unsafe"

	"poly.red/buffer"
	"poly.red/color"
	"poly.red/gpu"
	"poly.red/gpu/shader/gpumath/kernels"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/shader"
)

// errGPUDeferredUnsupported signals the GPU deferred path cannot handle this
// scene; the caller falls back to the CPU shader.
var errGPUDeferredUnsupported = errors.New("render: scene not supported by GPU deferred path")

// gpuShadowData is the marshaled shadow state for N shadow-casting lights:
// per-light combined matrices (column-major, 16 floats each) and packed depth
// maps (dlen floats each), matching render/shadow.go:shadingVisibility.
type gpuShadowData struct {
	mats   []float32 // n*16
	depths []float32 // n*dlen
	width  int
	dlen   int
	n      int
}

// gpuShadowData builds the shadow state for the GPU path. All source lights
// must cast shadow (the engine's non-casting-light darkening is not modeled
// here); otherwise returns (nil, false) and the caller falls back to the CPU.
// Combined matrix matches shadingVisibility:
// v.Apply(ScreenToWorld).Apply(lightView).Apply(lightProj).Apply(Viewport).
func (r *Renderer) gpuShadowData(uniforms *shader.MVP) (*gpuShadowData, bool) {
	ls, _ := r.cfg.Scene.Lights()
	if len(r.shadowBufs) == 0 {
		return nil, false
	}
	for i := range r.shadowBufs {
		if i >= len(ls) || !ls[i].CastShadow() {
			return nil, false
		}
	}
	width := r.bufs[0].Bounds().Dx()
	dlen := len(r.shadowBufs[0].depths)
	var mats, depths []float32
	for i := range r.shadowBufs {
		sb := &r.shadowBufs[i]
		m := uniforms.Viewport.
			MulM(sb.camera.ProjMatrix()).
			MulM(sb.camera.ViewMatrix()).
			MulM(uniforms.ViewportToWorld)
		for j := 0; j < 4; j++ { // column-major
			for k := 0; k < 4; k++ {
				mats = append(mats, m.Get(k, j))
			}
		}
		depths = append(depths, sb.depths...)
	}
	return &gpuShadowData{mats: mats, depths: depths, width: width, dlen: dlen, n: len(r.shadowBufs)}, true
}

// gpuDeferredShade runs the deferred Blinn-Phong shading on the GPU and writes
// the shaded colours back into buf. Supports point/directional lights +
// ambient and multiple Blinn-Phong materials (ambient-occlusion off, no shadow
// map); otherwise returns errGPUDeferredUnsupported and the caller uses the CPU.
// matAt resolves a flat material index against the per-frame table, returning nil
// for a negative or out-of-range index (use vertex color).
func matAt(table []*material.BlinnPhong, id int64) *material.BlinnPhong {
	if id < 0 || int(id) >= len(table) {
		return nil
	}
	return table[id]
}

func gpuDeferredShade(dev *gpu.Device, buf *buffer.FragmentBuffer, ls []light.Source, es []light.Environment, camPos math.Vec3[float32], bg color.RGBA, shadow *gpuShadowData, matTable []*material.BlinnPhong) error {
	var lightData []float32
	for _, l := range ls {
		switch lt := l.(type) {
		case *light.Point:
			pos := lt.Position()
			c := lt.Color()
			lightData = append(lightData, 0, pos.X, pos.Y, pos.Z, 1,
				float32(c.R), float32(c.G), float32(c.B), float32(c.A), lt.Intensity())
		case *light.Directional:
			d := lt.Dir()
			c := lt.Color()
			lightData = append(lightData, 1, d.X, d.Y, d.Z, 0,
				float32(c.R), float32(c.G), float32(c.B), float32(c.A), lt.Intensity())
		default:
			return errGPUDeferredUnsupported
		}
	}
	if len(ls) == 0 {
		return errGPUDeferredUnsupported
	}
	var ambientI float32
	for _, e := range es {
		ambientI += e.Intensity()
	}

	w := buf.Bounds().Dx()
	h := buf.Bounds().Dy()
	n := w * h

	normals := make([]float32, n*4)
	worldpos := make([]float32, n*4)
	basecol := make([]float32, n*4)
	matidx := make([]float32, n)
	okMask := make([]bool, n)
	passthrough := make([]bool, n)
	passCol := make([]color.RGBA, n)
	fragxyz := make([]float32, n*4) // screen X,Y,Depth for shadow lookup
	recv := make([]float32, n)      // per-fragment ReceiveShadow flag
	aoflag := make([]float32, n)    // per-fragment AmbientOcclusion flag
	depthbuf := make([]float32, n)  // screen-indexed depth (-1 for non-Ok), for SSAO
	anyAO := false

	matIndex := map[*material.BlinnPhong]int{}
	var materials []float32
	anyShaded := false
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*w + x
			info := buf.UnsafeGet(x, y)
			depthbuf[idx] = -1
			if info.Ok {
				depthbuf[idx] = info.Depth
			}
			if !info.Ok {
				continue
			}
			bp := matAt(matTable, info.MaterialID)
			if bp == nil {
				okMask[idx] = true
				passthrough[idx] = true
				passCol[idx] = info.Col
				continue
			}
			if bp.AmbientOcclusion {
				aoflag[idx] = 1
				anyAO = true
			}
			mIdx, seen := matIndex[bp]
			if !seen {
				mIdx = len(matIndex)
				matIndex[bp] = mIdx
				materials = append(materials,
					float32(bp.Diffuse.R), float32(bp.Diffuse.G), float32(bp.Diffuse.B), float32(bp.Diffuse.A),
					float32(bp.Specular.R), float32(bp.Specular.G), float32(bp.Specular.B), float32(bp.Specular.A),
					bp.Shininess)
			}

			anyShaded = true
			okMask[idx] = true
			matidx[idx] = float32(mIdx)
			nor := info.Nor
			if bp.FlatShading {
				nor = info.FaceNor
			}
			normals[idx*4], normals[idx*4+1], normals[idx*4+2], normals[idx*4+3] = nor.X, nor.Y, nor.Z, 0
			worldpos[idx*4], worldpos[idx*4+1], worldpos[idx*4+2], worldpos[idx*4+3] = info.WordPos.X, info.WordPos.Y, info.WordPos.Z, 1
			fragxyz[idx*4], fragxyz[idx*4+1], fragxyz[idx*4+2] = float32(info.X), float32(info.Y), info.Depth
			if bp.ReceiveShadow {
				recv[idx] = 1
			}

			lod := float32(0)
			if bp.Texture.UseMipmap() {
				siz := float32(bp.Texture.Size()) * math.Sqrt(math.Max(info.Du, info.Dv))
				if siz < 1 {
					siz = 1
				}
				lod = math.Log2(siz)
			}
			bc := bp.Texture.Query(lod, info.U, 1-info.V)
			basecol[idx*4], basecol[idx*4+1], basecol[idx*4+2], basecol[idx*4+3] = float32(bc.R), float32(bc.G), float32(bc.B), float32(bc.A)
		}
	}
	if !anyShaded {
		return errGPUDeferredUnsupported
	}

	scene := []float32{camPos.X, camPos.Y, camPos.Z, 1, ambientI, float32(len(ls)), 0, 0}

	shaded, err := runDeferredKernel(dev, n, normals, worldpos, basecol, lightData, matidx, materials, scene)
	if err != nil {
		return err
	}

	if debugDeferredSelfCheck {
		deferredSelfCheck(n, okMask, passthrough, normals, worldpos, basecol, lightData, matidx, materials, scene, shaded)
	}

	// Apply shadows as a second pass over the shaded float buffer.
	if shadow != nil {
		su := []float32{float32(shadow.width), float32(shadow.dlen), float32(shadow.n), 0}
		if err := runShadowKernel(dev, n, fragxyz, recv, shadow.depths, shadow.mats, shaded, su); err != nil {
			return err
		}
	}

	// Apply SSAO as a final pass.
	if anyAO {
		au := []float32{float32(w), float32(h), 0, 0}
		if err := runAOKernel(dev, n, fragxyz, aoflag, depthbuf, shaded, au); err != nil {
			return err
		}
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*w + x
			info := buf.UnsafeGet(x, y)
			switch {
			case passthrough[idx]:
				info.Col = passCol[idx]
			case okMask[idx]:
				info.Col = color.RGBA{
					R: toByte(shaded[idx*4]),
					G: toByte(shaded[idx*4+1]),
					B: toByte(shaded[idx*4+2]),
					A: toByte(shaded[idx*4+3]),
				}
			default:
				info.Col = bg
			}
			buf.UnsafeSet(x, y, info)
		}
	}
	return nil
}

func toByte(v float32) uint8 {
	return uint8(math.Clamp(float32(math.Round(v)), 0, 255))
}

func runDeferredKernel(dev *gpu.Device, n int, normals, worldpos, basecol, lights, matidx, materials, scene []float32) ([]float32, error) {
	mod, err := kernelModule(dev, kernels.ShadeSrc, "Shade")
	if err != nil {
		return nil, err
	}
	sb := func(i int) gpu.BindGroupLayoutEntry {
		return gpu.BindGroupLayoutEntry{Binding: i, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer}
	}
	layout := dev.NewBindGroupLayout(
		sb(0), sb(1), sb(2), sb(3), sb(4), sb(5), sb(6), sb(7),
	)
	pipe, err := dev.NewComputePipeline(gpu.ComputePipelineDescriptor{Layout: dev.NewPipelineLayout(layout), Module: mod, Entry: "Shade"})
	if err != nil {
		return nil, err
	}

	if len(lights) == 0 {
		lights = []float32{0}
	}
	if len(materials) == 0 {
		materials = make([]float32, 9)
	}
	nb := storageBuf(dev, normals)
	wb := storageBuf(dev, worldpos)
	cb := storageBuf(dev, basecol)
	lb := storageBuf(dev, lights)
	mib := storageBuf(dev, matidx)
	mtb := storageBuf(dev, materials)
	scb := storageBuf(dev, scene)
	out, err := dev.NewBuffer(gpu.BufferDescriptor{Size: n * 4 * 4, Usage: gpu.BufferStorage | gpu.BufferMapRead})
	if err != nil {
		return nil, err
	}
	defer func() {
		nb.Release()
		wb.Release()
		cb.Release()
		lb.Release()
		mib.Release()
		mtb.Release()
		scb.Release()
		out.Release()
	}()

	bg := dev.NewBindGroup(layout,
		gpu.BindGroupEntry{Binding: 0, Buffer: nb},
		gpu.BindGroupEntry{Binding: 1, Buffer: wb},
		gpu.BindGroupEntry{Binding: 2, Buffer: cb},
		gpu.BindGroupEntry{Binding: 3, Buffer: lb},
		gpu.BindGroupEntry{Binding: 4, Buffer: mib},
		gpu.BindGroupEntry{Binding: 5, Buffer: mtb},
		gpu.BindGroupEntry{Binding: 6, Buffer: scb},
		gpu.BindGroupEntry{Binding: 7, Buffer: out},
	)
	enc := dev.NewCommandEncoder()
	cp := enc.BeginComputePass()
	cp.SetPipeline(pipe)
	cp.SetBindGroup(0, bg)
	cp.Dispatch(n, 1, 1)
	cp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	res := make([]float32, n*4)
	copy(res, unsafe.Slice((*float32)(unsafe.Pointer(&out.Bytes()[0])), n*4))
	return res, nil
}

func runShadowKernel(dev *gpu.Device, n int, fragxyz, recv, depths, mats, color, su []float32) error {
	mod, err := kernelModule(dev, kernels.ShadowSrc, "Shadow")
	if err != nil {
		return err
	}
	sb := func(i int) gpu.BindGroupLayoutEntry {
		return gpu.BindGroupLayoutEntry{Binding: i, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer}
	}
	layout := dev.NewBindGroupLayout(
		sb(0), sb(1), sb(2), sb(3), sb(4), sb(5),
	)
	pipe, err := dev.NewComputePipeline(gpu.ComputePipelineDescriptor{Layout: dev.NewPipelineLayout(layout), Module: mod, Entry: "Shadow"})
	if err != nil {
		return err
	}
	if len(depths) == 0 {
		depths = []float32{0}
	}
	if len(mats) == 0 {
		mats = []float32{0}
	}
	fb := storageBuf(dev, fragxyz)
	rb := storageBuf(dev, recv)
	db := storageBuf(dev, depths)
	mb := storageBuf(dev, mats)
	cb, err := dev.NewBuffer(gpu.BufferDescriptor{Size: len(color) * 4, Usage: gpu.BufferStorage | gpu.BufferCopyDst | gpu.BufferMapRead, Data: deferredBytes(color)})
	if err != nil {
		return err
	}
	ub := storageBuf(dev, su)
	defer func() {
		fb.Release()
		rb.Release()
		db.Release()
		mb.Release()
		cb.Release()
		ub.Release()
	}()

	bg := dev.NewBindGroup(layout,
		gpu.BindGroupEntry{Binding: 0, Buffer: fb},
		gpu.BindGroupEntry{Binding: 1, Buffer: rb},
		gpu.BindGroupEntry{Binding: 2, Buffer: db},
		gpu.BindGroupEntry{Binding: 3, Buffer: mb},
		gpu.BindGroupEntry{Binding: 4, Buffer: cb},
		gpu.BindGroupEntry{Binding: 5, Buffer: ub},
	)
	enc := dev.NewCommandEncoder()
	cp := enc.BeginComputePass()
	cp.SetPipeline(pipe)
	cp.SetBindGroup(0, bg)
	cp.Dispatch(n, 1, 1)
	cp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()
	copy(color, unsafe.Slice((*float32)(unsafe.Pointer(&cb.Bytes()[0])), len(color)))
	return nil
}

func runAOKernel(dev *gpu.Device, n int, fragxyz, aoflag, depthbuf, color, au []float32) error {
	mod, err := kernelModule(dev, kernels.AOSrc, "AO")
	if err != nil {
		return err
	}
	sb := func(i int) gpu.BindGroupLayoutEntry {
		return gpu.BindGroupLayoutEntry{Binding: i, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer}
	}
	layout := dev.NewBindGroupLayout(
		sb(0), sb(1), sb(2), sb(3), sb(4),
	)
	pipe, err := dev.NewComputePipeline(gpu.ComputePipelineDescriptor{Layout: dev.NewPipelineLayout(layout), Module: mod, Entry: "AO"})
	if err != nil {
		return err
	}
	fb := storageBuf(dev, fragxyz)
	ab := storageBuf(dev, aoflag)
	db := storageBuf(dev, depthbuf)
	cb, err := dev.NewBuffer(gpu.BufferDescriptor{Size: len(color) * 4, Usage: gpu.BufferStorage | gpu.BufferCopyDst | gpu.BufferMapRead, Data: deferredBytes(color)})
	if err != nil {
		return err
	}
	ub := storageBuf(dev, au)
	defer func() { fb.Release(); ab.Release(); db.Release(); cb.Release(); ub.Release() }()

	bg := dev.NewBindGroup(layout,
		gpu.BindGroupEntry{Binding: 0, Buffer: fb},
		gpu.BindGroupEntry{Binding: 1, Buffer: ab},
		gpu.BindGroupEntry{Binding: 2, Buffer: db},
		gpu.BindGroupEntry{Binding: 3, Buffer: cb},
		gpu.BindGroupEntry{Binding: 4, Buffer: ub},
	)
	enc := dev.NewCommandEncoder()
	cp := enc.BeginComputePass()
	cp.SetPipeline(pipe)
	cp.SetBindGroup(0, bg)
	cp.Dispatch(n, 1, 1)
	cp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()
	copy(color, unsafe.Slice((*float32)(unsafe.Pointer(&cb.Bytes()[0])), len(color)))
	return nil
}

func storageBuf(dev *gpu.Device, d []float32) *gpu.Buffer {
	b, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: len(d) * 4, Usage: gpu.BufferStorage | gpu.BufferCopyDst, Data: deferredBytes(d)})
	return b
}

func deferredBytes(d []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}

// debugDeferredSelfCheck enables the GPU deferred output to be compared against
// the author-once kernels.Shade run as Go over the same G-buffer, isolating
// compiler-lowering bugs from marshaling bugs. Set by tests.
var debugDeferredSelfCheck bool

// selfCheckResult records a self-check outcome so a test can hard-assert that
// the GPU deferred path matches the author-once kernel.
type selfCheckResult struct {
	ran     bool
	matched bool
	detail  string
}

// deferredSelfCheckResult holds the most recent self-check outcome.
var deferredSelfCheckResult selfCheckResult

// deferredSelfCheck reruns the author-once kernels.Shade in Go over the same
// G-buffer and compares it to the GPU output. Because the GPU shader is compiled
// from the same source (kernels.ShadeSrc), this proves the compiler lowering:
// GPU(ShadeSrc) == kernels.Shade-as-Go for every shaded fragment.
func deferredSelfCheck(n int, okMask, passthrough []bool, normals, worldpos, basecol, lights, matidx, materials, scene []float32, gpu []float32) {
	replica := make([]float32, len(gpu))
	for idx := 0; idx < n; idx++ {
		if !okMask[idx] || passthrough[idx] {
			continue
		}
		kernels.Shade(uint(idx), normals, worldpos, basecol, lights, matidx, materials, scene, replica)
	}
	for idx := 0; idx < n; idx++ {
		if !okMask[idx] || passthrough[idx] {
			continue
		}
		for c := 0; c < 3; c++ {
			if d := replica[idx*4+c] - gpu[idx*4+c]; d > 1 || d < -1 {
				deferredSelfCheckResult = selfCheckResult{ran: true, detail: fmt.Sprintf("idx %d chan %d kernel %d gpu %d", idx, c, int(replica[idx*4+c]), int(gpu[idx*4+c]))}
				println("deferredSelfCheck:", deferredSelfCheckResult.detail)
				return
			}
		}
	}
	deferredSelfCheckResult = selfCheckResult{ran: true, matched: true}
	println("deferredSelfCheck: GPU matches author-once kernels.Shade for all shaded fragments")
}
