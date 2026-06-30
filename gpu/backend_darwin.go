// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

package gpu

import (
	"errors"
	"unsafe"

	"poly.red/gpu/mtl"
)

// openBackend selects the Metal backend on darwin.
func openBackend(c config) (backend, Driver, error) {
	switch c.driver {
	case DriverAuto, DriverMetal:
		dev, err := mtl.CreateSystemDefaultDevice()
		if err != nil || !dev.Available() {
			return nil, DriverAuto, ErrUnsupported
		}
		return &metalBackend{dev: dev, queue: dev.MakeCommandQueue()}, DriverMetal, nil
	default:
		return nil, DriverAuto, ErrUnsupported
	}
}

type metalBackend struct {
	dev     mtl.Device
	queue   mtl.CommandQueue
	last    mtl.CommandBuffer // last committed buffer, for waitIdle
	hasLast bool
}

func (m *metalBackend) newBuffer(size int, usage BufferUsage, data []byte) (backendBuffer, error) {
	var buf mtl.Buffer
	if len(data) > 0 {
		buf = m.dev.MakeBuffer(unsafe.Pointer(&data[0]), uintptr(size), mtl.ResourceStorageModeShared)
	} else {
		buf = m.dev.MakeBuffer(nil, uintptr(size), mtl.ResourceStorageModeShared)
	}
	return &metalBuffer{buf: buf, size: size}, nil
}

func (m *metalBackend) newShaderModule(src ShaderSource) (backendShaderModule, error) {
	if src.MSL == "" {
		return nil, errors.New("gpu: metal backend requires ShaderSource.MSL")
	}
	lib, err := m.dev.MakeLibrary(src.MSL, mtl.CompileOptions{})
	if err != nil {
		return nil, err
	}
	return &metalModule{lib: lib}, nil
}

func (m *metalBackend) newComputePipeline(mod backendShaderModule, entry string) (backendComputePipeline, error) {
	mm, ok := mod.(*metalModule)
	if !ok {
		return nil, errors.New("gpu: shader module is not a metal module")
	}
	fn, err := mm.lib.MakeFunction(entry)
	if err != nil {
		return nil, err
	}
	cps, err := m.dev.MakeComputePipelineState(fn)
	if err != nil {
		return nil, err
	}
	return &metalPipeline{cps: cps, max: cps.MaxTotalThreadsPerThreadgroup()}, nil
}

func (m *metalBackend) newCommandBuffer() backendCommandBuffer {
	return &metalCmd{m: m, cb: m.queue.MakeCommandBuffer()}
}

// newWindowSurface is not implemented on the Metal backend yet; an on-screen
// CAMetalLayer drawable lands in a later phase.
func (m *metalBackend) newWindowSurface(display, window uintptr, w, h int) (backendWindowSurface, error) {
	return nil, ErrUnsupported
}

func (m *metalBackend) windowVisualID() uint32 { return 0 }

func (m *metalBackend) waitIdle() {
	if m.hasLast {
		m.last.WaitUntilCompleted()
	}
}

func (m *metalBackend) close() error {
	m.queue.Release()
	return nil
}

type metalBuffer struct {
	buf  mtl.Buffer
	size int
}

func (b *metalBuffer) bytes() []byte {
	return unsafe.Slice((*byte)(b.buf.Content()), b.size)
}
func (b *metalBuffer) release() { b.buf.Release() }

type metalModule struct{ lib mtl.Library }

func (*metalModule) isShaderModule() {}

type metalPipeline struct {
	cps mtl.ComputePipelineState
	max int
}

func (p *metalPipeline) maxThreads() int { return p.max }

type metalCmd struct {
	m    *metalBackend
	cb   mtl.CommandBuffer
	enc  mtl.ComputeCommandEncoder
	renc mtl.RenderCommandEncoder
	cur  *metalPipeline
}

func (c *metalCmd) beginCompute() { c.enc = c.cb.MakeComputeCommandEncoder() }

func (c *metalCmd) setComputePipeline(p backendComputePipeline) {
	mp := p.(*metalPipeline)
	c.cur = mp
	c.enc.SetComputePipelineState(mp.cps)
}

func (c *metalCmd) setBuffer(b backendBuffer, offset, index int) {
	c.enc.SetBuffer(b.(*metalBuffer).buf, offset, index)
}

func (c *metalCmd) dispatch(x, y, z int) {
	total := x
	if y > 1 {
		total *= y
	}
	if z > 1 {
		total *= z
	}
	tg := c.cur.max
	if tg > total {
		tg = total
	}
	if tg < 1 {
		tg = 1
	}
	c.enc.DispatchThreads(
		mtl.Size{Width: x, Height: y, Depth: z},
		mtl.Size{Width: tg, Height: 1, Depth: 1},
	)
}

func (c *metalCmd) endCompute() { c.enc.EndEncoding() }

func (c *metalCmd) commit() {
	c.cb.Commit()
	c.m.last = c.cb
	c.m.hasLast = true
}

// --- render support ---

func mtlFormat(f TextureFormat) mtl.PixelFormat {
	switch f {
	case RGBA8Unorm:
		return mtl.PixelFormatRGBA8UNorm
	case RGBA32Float:
		return mtl.PixelFormatRGBA32Float
	case Depth32Float:
		return mtl.PixelFormatDepth32Float
	default:
		return mtl.PixelFormatRGBA8UNorm
	}
}

func mtlPrim(p Primitive) mtl.PrimitiveType {
	switch p {
	case TriangleStrip:
		return mtl.PrimitiveTypeTriangleStrip
	case LineList:
		return mtl.PrimitiveTypeLine
	case PointList:
		return mtl.PrimitiveTypePoint
	default:
		return mtl.PrimitiveTypeTriangle
	}
}

func (m *metalBackend) newTexture(format TextureFormat, w, h int, renderTarget bool) (backendTexture, error) {
	usage := mtl.TextureUsageShaderRead
	if renderTarget {
		usage |= mtl.TextureUsageRenderTarget
	}
	// A depth texture cannot use Shared storage on macOS; it is a private
	// render-target attachment that is never read back to the CPU.
	storage := mtl.StorageModeShared
	if format == Depth32Float {
		storage = mtl.StorageModePrivate
		usage = mtl.TextureUsageRenderTarget
	}
	tex := m.dev.MakeTexture(mtl.TextureDescriptor{
		PixelFormat: mtlFormat(format),
		Width:       w,
		Height:      h,
		StorageMode: storage,
		Usage:       usage,
	})
	return &metalTexture{tex: tex, w: w, h: h}, nil
}

func (m *metalBackend) newRenderPipeline(vmod backendShaderModule, ventry string, fmod backendShaderModule, fentry string, color TextureFormat, extraColor []TextureFormat, depth TextureFormat) (backendRenderPipeline, error) {
	vfn, err := vmod.(*metalModule).lib.MakeFunction(ventry)
	if err != nil {
		return nil, err
	}
	ffn, err := fmod.(*metalModule).lib.MakeFunction(fentry)
	if err != nil {
		return nil, err
	}
	pdesc := mtl.RenderPipelineDescriptor{
		VertexFunction:   vfn,
		FragmentFunction: ffn,
		ColorPixelFormat: mtlFormat(color),
	}
	for _, f := range extraColor {
		pdesc.ExtraColorPixelFormats = append(pdesc.ExtraColorPixelFormats, mtlFormat(f))
	}
	p := &metalRenderPipeline{}
	if depth != FormatNone {
		pdesc.DepthPixelFormat = mtlFormat(depth)
		// Standard 3D depth test: keep the nearer fragment and write its depth.
		p.depthState = m.dev.MakeDepthStencilState(mtl.DepthStencilDescriptor{
			DepthCompareFunction: mtl.CompareFunctionLess,
			DepthWriteEnabled:    true,
		})
		p.hasDepth = true
	}
	rps, err := m.dev.MakeRenderPipelineState(pdesc)
	if err != nil {
		return nil, err
	}
	p.rps = rps
	return p, nil
}

type metalTexture struct {
	tex  mtl.Texture
	w, h int
}

func (t *metalTexture) readPixels() []byte {
	dst := make([]byte, t.w*t.h*4)
	t.tex.GetBytes(dst, t.w*4, mtl.RegionMake2D(0, 0, t.w, t.h), 0)
	return dst
}

func (t *metalTexture) write(pixels []byte, bytesPerRow int) {
	t.tex.ReplaceRegion(mtl.RegionMake2D(0, 0, t.w, t.h), 0, pixels, uintptr(bytesPerRow))
}

type metalSampler struct{ s mtl.SamplerState }

func (*metalSampler) isSampler() {}

func mtlFilter(f FilterMode) mtl.SamplerMinMagFilter {
	if f == FilterLinear {
		return mtl.SamplerFilterLinear
	}
	return mtl.SamplerFilterNearest
}

func mtlAddress(a AddressMode) mtl.SamplerAddressMode {
	if a == AddressRepeat {
		return mtl.SamplerAddressRepeat
	}
	return mtl.SamplerAddressClampToEdge
}

func (m *metalBackend) newSampler(desc SamplerDescriptor) backendSampler {
	return &metalSampler{s: m.dev.MakeSamplerState(mtl.SamplerDescriptor{
		MinFilter:    mtlFilter(desc.MinFilter),
		MagFilter:    mtlFilter(desc.MagFilter),
		SAddressMode: mtlAddress(desc.AddressU),
		TAddressMode: mtlAddress(desc.AddressV),
	})}
}

func (c *metalCmd) setComputeTexture(index int, t backendTexture) {
	c.enc.SetTexture(t.(*metalTexture).tex, index)
}

func (c *metalCmd) setComputeSampler(index int, s backendSampler) {
	c.enc.SetSamplerState(s.(*metalSampler).s, index)
}

type metalRenderPipeline struct {
	rps        mtl.RenderPipelineState
	depthState mtl.DepthStencilState
	hasDepth   bool
}

func (*metalRenderPipeline) isRenderPipeline() {}

func (c *metalCmd) beginRender(info renderPassInfo) {
	load := mtl.LoadActionLoad
	if info.load == LoadClear {
		load = mtl.LoadActionClear
	}
	desc := mtl.RenderPassDescriptor{
		ColorAttachment0: mtl.ColorAttachment{
			Texture:     info.color.(*metalTexture).tex,
			LoadAction:  load,
			StoreAction: mtl.StoreActionStore,
			ClearColor:  mtl.ClearColor{Red: info.clearColor[0], Green: info.clearColor[1], Blue: info.clearColor[2], Alpha: info.clearColor[3]},
		},
	}
	for _, t := range info.extraColor {
		desc.ExtraColorAttachments = append(desc.ExtraColorAttachments, mtl.ColorAttachment{
			Texture:     t.tex.(*metalTexture).tex,
			LoadAction:  load,
			StoreAction: mtl.StoreActionStore,
			ClearColor:  mtl.ClearColor{Red: t.clear[0], Green: t.clear[1], Blue: t.clear[2], Alpha: t.clear[3]},
		})
	}
	if info.depth != nil {
		desc.Depth = mtl.DepthAttachment{
			Texture:     info.depth.(*metalTexture).tex,
			LoadAction:  mtl.LoadActionClear,
			StoreAction: mtl.StoreActionDontCare,
			ClearDepth:  info.clearDepth,
		}
	}
	c.renc = c.cb.MakeRenderCommandEncoder(desc)
}

func (c *metalCmd) setRenderPipeline(p backendRenderPipeline) {
	mp := p.(*metalRenderPipeline)
	c.renc.SetRenderPipelineState(mp.rps)
	if mp.hasDepth {
		c.renc.SetDepthStencilState(mp.depthState)
	}
}

func (c *metalCmd) setRenderBuffer(b backendBuffer, offset, index int) {
	c.renc.SetFragmentBuffer(b.(*metalBuffer).buf, offset, index)
}

func (c *metalCmd) setVertexBuffer(b backendBuffer, index int) {
	c.renc.SetVertexBuffer(b.(*metalBuffer).buf, 0, index)
}

func (c *metalCmd) draw(prim Primitive, start, count int) {
	c.renc.DrawPrimitives(mtlPrim(prim), start, count)
}

func (c *metalCmd) endRender() { c.renc.EndEncoding() }
