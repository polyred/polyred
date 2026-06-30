// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gpu

import "errors"

// TextureFormat is a texture pixel format.
type TextureFormat int

const (
	// FormatNone is the zero value: no/unspecified format (e.g. a render pipeline
	// with no depth attachment).
	FormatNone TextureFormat = iota
	// RGBA8Unorm is 8-bit normalized unsigned RGBA.
	RGBA8Unorm
	// Depth32Float is a 32-bit float depth format, for a depth render target.
	Depth32Float
	// RGBA32Float is 32-bit float RGBA, for a float render target such as a
	// G-buffer attachment that stores world positions or normals at full precision.
	RGBA32Float
)

// TextureDescriptor describes a texture to create.
type TextureDescriptor struct {
	Label        string
	Format       TextureFormat
	Width        int
	Height       int
	RenderTarget bool // usable as a render-pass color attachment
}

// Texture is a GPU image, usable as a render target and/or sampled resource.
type Texture struct {
	b backendTexture
	w int
	h int
}

// Width returns the texture width in pixels.
func (t *Texture) Width() int { return t.w }

// Height returns the texture height in pixels.
func (t *Texture) Height() int { return t.h }

// ReadPixels copies the texture's pixels back to CPU memory (tightly packed,
// 4 bytes/pixel for RGBA8Unorm). Used for headless render-to-image.
func (t *Texture) ReadPixels() []byte { return t.b.readPixels() }

// Write uploads tightly-packed pixel data (4 bytes/pixel for RGBA8Unorm) into
// the texture.
func (t *Texture) Write(pixels []byte) { t.b.write(pixels, t.w*4) }

// FilterMode selects texture filtering.
type FilterMode int

const (
	FilterNearest FilterMode = iota
	FilterLinear
)

// AddressMode selects out-of-range texture-coordinate handling.
type AddressMode int

const (
	AddressClampToEdge AddressMode = iota
	AddressRepeat
)

// SamplerDescriptor configures a sampler.
type SamplerDescriptor struct {
	MinFilter FilterMode
	MagFilter FilterMode
	AddressU  AddressMode
	AddressV  AddressMode
}

// Sampler describes how a shader reads a texture.
type Sampler struct {
	b backendSampler
}

// NewSampler creates a sampler.
func (d *Device) NewSampler(desc SamplerDescriptor) *Sampler {
	return &Sampler{b: d.b.newSampler(desc)}
}

// SetTexture binds a texture for sampling at the given texture index.
func (p *ComputePass) SetTexture(index int, t *Texture) {
	p.e.cmd.setComputeTexture(index, t.b)
}

// SetSampler binds a sampler at the given sampler index.
func (p *ComputePass) SetSampler(index int, s *Sampler) {
	p.e.cmd.setComputeSampler(index, s.b)
}

// NewTexture allocates a texture.
func (d *Device) NewTexture(desc TextureDescriptor) (*Texture, error) {
	if desc.Width <= 0 || desc.Height <= 0 {
		return nil, errors.New("gpu: texture dimensions must be > 0")
	}
	bt, err := d.b.newTexture(desc.Format, desc.Width, desc.Height, desc.RenderTarget)
	if err != nil {
		return nil, err
	}
	return &Texture{b: bt, w: desc.Width, h: desc.Height}, nil
}

// RenderPipelineDescriptor describes a render pipeline. The vertex and fragment
// stages may come from the same or different shader modules.
type RenderPipelineDescriptor struct {
	Label          string
	Layout         *PipelineLayout
	VertexModule   *ShaderModule
	VertexEntry    string
	FragmentModule *ShaderModule
	FragmentEntry  string
	ColorFormat    TextureFormat
	// ExtraColorFormats are the formats of color attachments 1..N (attachment 0 is
	// ColorFormat). Empty for a single color target. Used for a G-buffer (MRT).
	ExtraColorFormats []TextureFormat
	// DepthFormat is the depth attachment format (FormatNone for no depth test).
	// When set, the pipeline depth-tests with "less" and writes depth.
	DepthFormat TextureFormat
}

// RenderPipeline is a compiled render pipeline.
type RenderPipeline struct {
	b backendRenderPipeline
}

// NewRenderPipeline creates a render pipeline.
func (d *Device) NewRenderPipeline(desc RenderPipelineDescriptor) (*RenderPipeline, error) {
	if desc.VertexModule == nil || desc.FragmentModule == nil {
		return nil, errors.New("gpu: render pipeline requires vertex and fragment modules")
	}
	bp, err := d.b.newRenderPipeline(desc.VertexModule.b, desc.VertexEntry, desc.FragmentModule.b, desc.FragmentEntry, desc.ColorFormat, desc.ExtraColorFormats, desc.DepthFormat)
	if err != nil {
		return nil, err
	}
	return &RenderPipeline{b: bp}, nil
}

// LoadOp is what a render pass does with the target at the start.
type LoadOp int

const (
	// LoadClear clears the target to ClearColor.
	LoadClear LoadOp = iota
	// LoadLoad preserves existing contents.
	LoadLoad
)

// Primitive is the geometry primitive a draw assembles.
type Primitive int

const (
	TriangleList Primitive = iota
	TriangleStrip
	LineList
	PointList
)

// RenderPassDescriptor describes a render pass (single color attachment, optional
// depth attachment).
type RenderPassDescriptor struct {
	ColorTexture *Texture
	Load         LoadOp
	ClearColor   [4]float64 // RGBA, used when Load == LoadClear
	// ExtraColorTargets are color attachments 1..N (attachment 0 is ColorTexture).
	// Each is cleared to its ClearColor when Load == LoadClear. Used for a G-buffer.
	ExtraColorTargets []ColorTarget
	// DepthTexture is the optional depth attachment (cleared to ClearDepth, which
	// defaults to 1.0 when zero). Nil for no depth.
	DepthTexture *Texture
	ClearDepth   float64
}

// ColorTarget is one color attachment of a render pass.
type ColorTarget struct {
	Texture    *Texture
	ClearColor [4]float64
}

// RenderPass encodes draw commands.
type RenderPass struct {
	e *CommandEncoder
}

// BeginRenderPass starts a render pass.
func (e *CommandEncoder) BeginRenderPass(desc RenderPassDescriptor) *RenderPass {
	info := renderPassInfo{
		color:      desc.ColorTexture.b,
		load:       desc.Load,
		clearColor: desc.ClearColor,
	}
	for _, t := range desc.ExtraColorTargets {
		info.extraColor = append(info.extraColor, renderColorTarget{tex: t.Texture.b, clear: t.ClearColor})
	}
	if desc.DepthTexture != nil {
		info.depth = desc.DepthTexture.b
		info.clearDepth = desc.ClearDepth
		if info.clearDepth == 0 {
			info.clearDepth = 1
		}
	}
	e.cmd.beginRender(info)
	return &RenderPass{e: e}
}

// SetPipeline binds the render pipeline.
func (p *RenderPass) SetPipeline(rp *RenderPipeline) {
	p.e.cmd.setRenderPipeline(rp.b)
}

// SetBindGroup binds resources for the render stages.
func (p *RenderPass) SetBindGroup(group int, bg *BindGroup) {
	for _, e := range bg.entries {
		p.e.cmd.setRenderBuffer(e.Buffer.b, 0, e.Binding)
	}
}

// SetVertexBuffer binds a vertex buffer at the given index.
func (p *RenderPass) SetVertexBuffer(index int, b *Buffer) {
	p.e.cmd.setVertexBuffer(b.b, index)
}

// Draw draws count vertices starting at start.
func (p *RenderPass) Draw(prim Primitive, start, count int) {
	p.e.cmd.draw(prim, start, count)
}

// End finishes the render pass.
func (p *RenderPass) End() {
	p.e.cmd.endRender()
}
