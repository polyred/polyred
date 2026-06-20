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
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
)

// errGPUDeferredUnsupported signals the GPU deferred path cannot handle this
// scene; the caller falls back to the CPU shader.
var errGPUDeferredUnsupported = errors.New("render: scene not supported by GPU deferred path")

// gpuDeferredUsed is set when the GPU deferred path actually ran; tests use it
// to confirm the path was exercised rather than silently falling back.
var gpuDeferredUsed bool

// deferredKernel re-expresses shade()/shader.FragmentShader's Blinn-Phong path
// (ambient + point-light diffuse + specular) as a Go GPU kernel, generalized
// over a per-fragment G-buffer and N point lights. The base colour (texture
// query) and per-fragment normal are computed on the CPU and supplied; the
// per-light shading runs on the GPU. Verified against the engine in
// gpu/shader/blinnphong_parity_darwin_test.go.
const deferredKernel = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

type Scene struct {
	CamPos    Vec4
	Diffuse   Vec4
	Specular  Vec4
	Shininess float32
	AmbientI  float32
	NumLights float32
	Pad       float32
}

func Shade(gid uint, normals []float32, worldpos []float32, basecol []float32, lights []float32, s Scene, out []float32) {
	N := Vec4{normals[gid*4], normals[gid*4+1], normals[gid*4+2], normals[gid*4+3]}
	wpos := Vec4{worldpos[gid*4], worldpos[gid*4+1], worldpos[gid*4+2], worldpos[gid*4+3]}
	col := Vec4{basecol[gid*4], basecol[gid*4+1], basecol[gid*4+2], basecol[gid*4+3]}

	acc := col * s.AmbientI
	count := int(s.NumLights)
	for i := 0; i < count; i++ {
		lp := Vec4{lights[i*9], lights[i*9+1], lights[i*9+2], lights[i*9+3]}
		lc := Vec4{lights[i*9+4], lights[i*9+5], lights[i*9+6], lights[i*9+7]}
		li := lights[i*9+8]
		Ldir := lp - wpos
		L := normalize(Ldir)
		I := li / length(Ldir)
		V := normalize(s.CamPos - wpos)
		H := normalize(L + V)
		Ld := clamp(dot(N, L), 0.0, 1.0)
		Ls := pow(clamp(dot(N, H), 0.0, 1.0), s.Shininess)
		acc = acc + s.Diffuse*(col*(Ld*I))/255.0 + s.Specular*(lc*(Ls*I))/255.0
	}
	out[gid*4] = acc.X
	out[gid*4+1] = acc.Y
	out[gid*4+2] = acc.Z
	out[gid*4+3] = col.W
}
`

// gpuDeferredShade runs the deferred Blinn-Phong shading on the GPU and writes
// the shaded colours back into buf. It supports the common case (point lights +
// ambient, a single Blinn-Phong material with ambient-occlusion off);
// otherwise it returns errGPUDeferredUnsupported and the caller uses the CPU.
func gpuDeferredShade(dev *gpu.Device, buf *buffer.FragmentBuffer, ls []light.Source, es []light.Environment, camPos math.Vec3[float32], bg color.RGBA) error {
	// Validate the light set: point lights only (plus ambient environments).
	var lightData []float32
	for _, l := range ls {
		p, ok := l.(*light.Point)
		if !ok {
			return errGPUDeferredUnsupported
		}
		pos := p.Position()
		c := p.Color()
		lightData = append(lightData,
			pos.X, pos.Y, pos.Z, 1,
			float32(c.R), float32(c.G), float32(c.B), float32(c.A),
			p.Intensity())
	}
	if len(ls) == 0 {
		return errGPUDeferredUnsupported // engine returns base colour; not handled here
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
	okMask := make([]bool, n)

	var mat *material.BlinnPhong
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*w + x
			info := buf.UnsafeGet(x, y)
			if !info.Ok {
				continue
			}
			m := material.Get(material.ID(info.MaterialID))
			if m == nil {
				return errGPUDeferredUnsupported
			}
			bp, ok := m.(*material.BlinnPhong)
			if !ok || bp.AmbientOcclusion {
				return errGPUDeferredUnsupported
			}
			if mat == nil {
				mat = bp
			} else if mat != bp {
				return errGPUDeferredUnsupported // multiple materials not supported
			}

			okMask[idx] = true
			nor := info.Nor
			if bp.FlatShading {
				nor = info.FaceNor
			}
			normals[idx*4], normals[idx*4+1], normals[idx*4+2], normals[idx*4+3] = nor.X, nor.Y, nor.Z, 0
			worldpos[idx*4], worldpos[idx*4+1], worldpos[idx*4+2], worldpos[idx*4+3] = info.WordPos.X, info.WordPos.Y, info.WordPos.Z, 1

			// Base colour: same texture query as shader.FragmentShader.
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
	if mat == nil {
		return errGPUDeferredUnsupported // nothing shaded
	}

	scene := []float32{
		camPos.X, camPos.Y, camPos.Z, 1,
		float32(mat.Diffuse.R), float32(mat.Diffuse.G), float32(mat.Diffuse.B), float32(mat.Diffuse.A),
		float32(mat.Specular.R), float32(mat.Specular.G), float32(mat.Specular.B), float32(mat.Specular.A),
		mat.Shininess, ambientI, float32(len(ls)), 0,
	}

	shaded, err := runDeferredKernel(dev, n, normals, worldpos, basecol, lightData, scene)
	if err != nil {
		return err
	}

	// Write back: shaded colour for Ok fragments, background otherwise.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*w + x
			info := buf.UnsafeGet(x, y)
			if okMask[idx] {
				info.Col = color.RGBA{
					R: toByte(shaded[idx*4]),
					G: toByte(shaded[idx*4+1]),
					B: toByte(shaded[idx*4+2]),
					A: toByte(shaded[idx*4+3]),
				}
			} else {
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

func runDeferredKernel(dev *gpu.Device, n int, normals, worldpos, basecol, lights, scene []float32) ([]float32, error) {
	ks, err := gpushader.Compile(deferredKernel)
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
		sb(0), sb(1), sb(2), sb(3),
		gpu.BindGroupLayoutEntry{Binding: 4, Visibility: gpu.StageCompute, Kind: gpu.UniformBuffer},
		sb(5),
	)
	pipe, err := dev.NewComputePipeline(gpu.ComputePipelineDescriptor{Layout: dev.NewPipelineLayout(layout), Module: mod, Entry: "Shade"})
	if err != nil {
		return nil, err
	}

	if len(lights) == 0 {
		lights = []float32{0} // non-empty buffer
	}
	nb := storageBuf(dev, normals)
	wb := storageBuf(dev, worldpos)
	cb := storageBuf(dev, basecol)
	lb := storageBuf(dev, lights)
	scb := uniformBuf(dev, scene)
	out, err := dev.NewBuffer(gpu.BufferDescriptor{Size: n * 4 * 4, Usage: gpu.BufferStorage | gpu.BufferMapRead})
	if err != nil {
		return nil, err
	}
	defer func() {
		nb.Release()
		wb.Release()
		cb.Release()
		lb.Release()
		scb.Release()
		out.Release()
	}()

	bg := dev.NewBindGroup(layout,
		gpu.BindGroupEntry{Binding: 0, Buffer: nb},
		gpu.BindGroupEntry{Binding: 1, Buffer: wb},
		gpu.BindGroupEntry{Binding: 2, Buffer: cb},
		gpu.BindGroupEntry{Binding: 3, Buffer: lb},
		gpu.BindGroupEntry{Binding: 4, Buffer: scb},
		gpu.BindGroupEntry{Binding: 5, Buffer: out},
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

func deferredBytes(d []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}

func storageBuf(dev *gpu.Device, d []float32) *gpu.Buffer {
	b, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: len(d) * 4, Usage: gpu.BufferStorage | gpu.BufferCopyDst, Data: deferredBytes(d)})
	return b
}

func uniformBuf(dev *gpu.Device, d []float32) *gpu.Buffer {
	b, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: len(d) * 4, Usage: gpu.BufferUniform, Data: deferredBytes(d)})
	return b
}
