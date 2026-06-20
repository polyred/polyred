// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gpu

import "errors"

// TextureFormat is a texture pixel format.
type TextureFormat int

const (
	// RGBA8Unorm is 8-bit normalized unsigned RGBA.
	RGBA8Unorm TextureFormat = iota
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
	bp, err := d.b.newRenderPipeline(desc.VertexModule.b, desc.VertexEntry, desc.FragmentModule.b, desc.FragmentEntry, desc.ColorFormat)
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

// RenderPassDescriptor describes a render pass (single color attachment).
type RenderPassDescriptor struct {
	ColorTexture *Texture
	Load         LoadOp
	ClearColor   [4]float64 // RGBA, used when Load == LoadClear
}

// RenderPass encodes draw commands.
type RenderPass struct {
	e *CommandEncoder
}

// BeginRenderPass starts a render pass.
func (e *CommandEncoder) BeginRenderPass(desc RenderPassDescriptor) *RenderPass {
	e.cmd.beginRender(renderPassInfo{
		color:      desc.ColorTexture.b,
		load:       desc.Load,
		clearColor: desc.ClearColor,
	})
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
