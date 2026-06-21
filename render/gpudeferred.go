// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"errors"
	"unsafe"

	"poly.red/buffer"
	"poly.red/color"
	"poly.red/gpu"
	gpushader "poly.red/gpu/shader"
	"poly.red/gpu/shader/gpumath/kernels"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/shader"
)

// errGPUDeferredUnsupported signals the GPU deferred path cannot handle this
// scene; the caller falls back to the CPU shader.
var errGPUDeferredUnsupported = errors.New("render: scene not supported by GPU deferred path")

// shadowKernel multiplies the shaded colour by the shadow factor for a single
// shadow-casting light. It runs after the Blinn-Phong kernel on the shaded
// float buffer, in place. Mirrors render/shadow.go:shadingVisibility for one
// casting buffer: project the fragment to light space, look up the depth, and
// darken by pow(0.5, occluded). Coordinates use the engine's combined matrix.
const shadowKernel = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

type ShadowU struct {
	W        float32
	DepthLen float32
	N        float32
	Pad      float32
}

func Shadow(gid uint, fragxyz []float32, recv []float32, depths []float32, mats []float32, color []float32, s ShadowU) {
	if recv[gid] < 0.5 {
		return
	}
	fx := fragxyz[gid*4]
	fy := fragxyz[gid*4+1]
	fz := fragxyz[gid*4+2]
	occ := float32(0)
	n := int(s.N)
	dl := int(s.DepthLen)
	width := int(s.W)
	for k := 0; k < n; k++ {
		M := Mat4{
			Vec4{mats[k*16], mats[k*16+1], mats[k*16+2], mats[k*16+3]},
			Vec4{mats[k*16+4], mats[k*16+5], mats[k*16+6], mats[k*16+7]},
			Vec4{mats[k*16+8], mats[k*16+9], mats[k*16+10], mats[k*16+11]},
			Vec4{mats[k*16+12], mats[k*16+13], mats[k*16+14], mats[k*16+15]},
		}
		clip := M * Vec4{fx, fy, fz, 1}
		sx := clip.X / clip.W
		sy := clip.Y / clip.W
		sz := clip.Z / clip.W
		idx := int(sx) + int(sy)*width
		if idx > 0 {
			if idx < dl {
				if sz < depths[k*dl+idx]-0.03 {
					occ = occ + 1
				}
			}
		}
	}
	wf := pow(0.5, occ)
	// Match the engine: uint8(clamp(round(blinn),0,255) * w), truncated.
	color[gid*4] = floor(clamp(round(color[gid*4]), 0.0, 255.0) * wf)
	color[gid*4+1] = floor(clamp(round(color[gid*4+1]), 0.0, 255.0) * wf)
	color[gid*4+2] = floor(clamp(round(color[gid*4+2]), 0.0, 255.0) * wf)
}
`

// aoKernel applies screen-space ambient occlusion as a final pass, mirroring
// material/ao.go: for 8 directions, march the depth buffer and accumulate the
// max elevation angle, then darken by pow(total, 10000). NOTE: that exponent
// amplifies any GPU/CPU float difference, so exact parity is not expected.
const aoKernel = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

type AOU struct {
	W    float32
	H    float32
	Pad1 float32
	Pad2 float32
}

func AO(gid uint, fragxyz []float32, aoflag []float32, depthbuf []float32, color []float32, s AOU) {
	if aoflag[gid] < 0.5 {
		return
	}
	px := fragxyz[gid*4]
	py := fragxyz[gid*4+1]
	traceDepth := fragxyz[gid*4+2]
	width := int(s.W)
	height := int(s.H)
	total := float32(0)
	for d := 0; d < 8; d++ {
		ang := float32(d) * 0.78539816339744830961
		dirX := cos(ang)
		dirY := sin(ang)
		maxangle := float32(0)
		for t := 0; t < 100; t++ {
			ft := float32(t)
			dx := dirX * ft
			dy := dirY * ft
			distance := sqrt(dx*dx + dy*dy)
			if distance >= 1.0 {
				ix := int(px + dx)
				iy := int(py + dy)
				if ix >= 0 {
					if ix < width {
						if iy >= 0 {
							if iy < height {
								elevation := depthbuf[iy*width+ix] - traceDepth
								maxangle = max(maxangle, atan(elevation/distance))
							}
						}
					}
				}
			}
		}
		total = total + (1.57079632679489661923 - maxangle)
	}
	total = total / (1.57079632679489661923 * 8.0)
	total = pow(total, 10000.0)
	color[gid*4] = floor(clamp(round(color[gid*4]), 0.0, 255.0) * total)
	color[gid*4+1] = floor(clamp(round(color[gid*4+1]), 0.0, 255.0) * total)
	color[gid*4+2] = floor(clamp(round(color[gid*4+2]), 0.0, 255.0) * total)
}
`

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
func gpuDeferredShade(dev *gpu.Device, buf *buffer.FragmentBuffer, ls []light.Source, es []light.Environment, camPos math.Vec3[float32], bg color.RGBA, shadow *gpuShadowData) error {
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
			m := material.Get(material.ID(info.MaterialID))
			if m == nil {
				okMask[idx] = true
				passthrough[idx] = true
				passCol[idx] = info.Col
				continue
			}
			bp, ok := m.(*material.BlinnPhong)
			if !ok {
				return errGPUDeferredUnsupported
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
	ks, err := gpushader.Compile(kernels.ShadeSrc)
	if err != nil {
		return nil, err
	}
	mod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["Shade"].MSL})
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
	ks, err := gpushader.Compile(shadowKernel)
	if err != nil {
		return err
	}
	mod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["Shadow"].MSL})
	if err != nil {
		return err
	}
	sb := func(i int) gpu.BindGroupLayoutEntry {
		return gpu.BindGroupLayoutEntry{Binding: i, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer}
	}
	layout := dev.NewBindGroupLayout(
		sb(0), sb(1), sb(2), sb(3), sb(4),
		gpu.BindGroupLayoutEntry{Binding: 5, Visibility: gpu.StageCompute, Kind: gpu.UniformBuffer},
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
	ub := uniformBuf(dev, su)
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
	ks, err := gpushader.Compile(aoKernel)
	if err != nil {
		return err
	}
	mod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["AO"].MSL})
	if err != nil {
		return err
	}
	sb := func(i int) gpu.BindGroupLayoutEntry {
		return gpu.BindGroupLayoutEntry{Binding: i, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer}
	}
	layout := dev.NewBindGroupLayout(
		sb(0), sb(1), sb(2), sb(3),
		gpu.BindGroupLayoutEntry{Binding: 4, Visibility: gpu.StageCompute, Kind: gpu.UniformBuffer},
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
	ub := uniformBuf(dev, au)
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

func uniformBuf(dev *gpu.Device, d []float32) *gpu.Buffer {
	b, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: len(d) * 4, Usage: gpu.BufferUniform, Data: deferredBytes(d)})
	return b
}

func deferredBytes(d []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}

// debugDeferredSelfCheck enables a pure-Go replica of the deferred kernel to be
// compared against the GPU output, isolating MSL-kernel bugs from marshaling
// bugs. Set by tests.
var debugDeferredSelfCheck bool

// deferredSelfCheck recomputes the kernel in Go and logs the first fragment
// where the GPU result diverges by more than 1 (per channel).
func deferredSelfCheck(n int, okMask, passthrough []bool, normals, worldpos, basecol, lights, matidx, materials, scene []float32, gpu []float32) {
	v4 := func(b []float32, i int) [4]float32 {
		return [4]float32{b[i*4], b[i*4+1], b[i*4+2], b[i*4+3]}
	}
	dot := func(a, b [4]float32) float32 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] + a[3]*b[3] }
	norm := func(a [4]float32) [4]float32 {
		l := float32(math.Sqrt(dot(a, a)))
		return [4]float32{a[0] / l, a[1] / l, a[2] / l, a[3] / l}
	}
	ambI, count := scene[4], int(scene[5])
	camPos := v4(scene, 0)
	for idx := 0; idx < n; idx++ {
		if !okMask[idx] || passthrough[idx] {
			continue
		}
		N, wpos, col := v4(normals, idx), v4(worldpos, idx), v4(basecol, idx)
		mi := int(matidx[idx])
		diff := [4]float32{materials[mi*9], materials[mi*9+1], materials[mi*9+2], materials[mi*9+3]}
		spec := [4]float32{materials[mi*9+4], materials[mi*9+5], materials[mi*9+6], materials[mi*9+7]}
		shin := materials[mi*9+8]
		acc := [3]float32{col[0] * ambI, col[1] * ambI, col[2] * ambI}
		for i := 0; i < count; i++ {
			lt := lights[i*10]
			lp := [4]float32{lights[i*10+1], lights[i*10+2], lights[i*10+3], lights[i*10+4]}
			lc := [4]float32{lights[i*10+5], lights[i*10+6], lights[i*10+7], lights[i*10+8]}
			li := lights[i*10+9]
			var L [4]float32
			var I float32
			if lt < 0.5 {
				Ldir := [4]float32{lp[0] - wpos[0], lp[1] - wpos[1], lp[2] - wpos[2], lp[3] - wpos[3]}
				L = norm(Ldir)
				I = li / float32(math.Sqrt(dot(Ldir, Ldir)))
			} else {
				L = [4]float32{-lp[0], -lp[1], -lp[2], 0}
				I = li
			}
			V := norm([4]float32{camPos[0] - wpos[0], camPos[1] - wpos[1], camPos[2] - wpos[2], camPos[3] - wpos[3]})
			H := norm([4]float32{L[0] + V[0], L[1] + V[1], L[2] + V[2], L[3] + V[3]})
			Ld := math.Clamp(dot(N, L), 0, 1)
			Ls := math.Pow(math.Clamp(dot(N, H), 0, 1), shin)
			for c := 0; c < 3; c++ {
				acc[c] += diff[c]*(col[c]*Ld*I)/255 + spec[c]*(lc[c]*Ls*I)/255
			}
		}
		for c := 0; c < 3; c++ {
			if d := acc[c] - gpu[idx*4+c]; d > 1 || d < -1 {
				println("deferredSelfCheck: idx", idx, "chan", c, "go", int(acc[c]), "gpu", int(gpu[idx*4+c]), "matidx", mi)
				return
			}
		}
	}
	println("deferredSelfCheck: GPU matches pure-Go replica for all shaded fragments")
}
