// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

// Render-pass support for the Metal backend (cgo-free via purego), companion to
// the compute support in mtl_darwin.go. Validated headless: clear + triangle +
// texture readback.
package mtl

import (
	"errors"
	"unsafe"

	"github.com/ebitengine/purego/objc"
)

var (
	selNewRenderPipeline   = objc.RegisterName("newRenderPipelineStateWithDescriptor:error:")
	selSetVertexFunction   = objc.RegisterName("setVertexFunction:")
	selSetFragmentFunction = objc.RegisterName("setFragmentFunction:")
	selColorAttachments    = objc.RegisterName("colorAttachments")
	selObjectAtIndexed     = objc.RegisterName("objectAtIndexedSubscript:")
	selRenderPassDesc      = objc.RegisterName("renderPassDescriptor")
	selSetTexture          = objc.RegisterName("setTexture:")
	selSetLoadAction       = objc.RegisterName("setLoadAction:")
	selSetStoreAction      = objc.RegisterName("setStoreAction:")
	selSetClearColor       = objc.RegisterName("setClearColor:")
	selRenderEncoder       = objc.RegisterName("renderCommandEncoderWithDescriptor:")
	selSetRenderPipeline   = objc.RegisterName("setRenderPipelineState:")
	selSetVertexBuffer     = objc.RegisterName("setVertexBuffer:offset:atIndex:")
	selSetFragmentBuffer   = objc.RegisterName("setFragmentBuffer:offset:atIndex:")
	selSetVertexBytes      = objc.RegisterName("setVertexBytes:length:atIndex:")
	selDrawPrimitives      = objc.RegisterName("drawPrimitives:vertexStart:vertexCount:")
	selSetUsage            = objc.RegisterName("setUsage:")
	selGetBytes            = objc.RegisterName("getBytes:bytesPerRow:fromRegion:mipmapLevel:")

	selNewSamplerState   = objc.RegisterName("newSamplerStateWithDescriptor:")
	selSetMinFilter      = objc.RegisterName("setMinFilter:")
	selSetMagFilter      = objc.RegisterName("setMagFilter:")
	selSetSAddressMode   = objc.RegisterName("setSAddressMode:")
	selSetTAddressMode   = objc.RegisterName("setTAddressMode:")
	selSetComputeTexture = objc.RegisterName("setTexture:atIndex:")
	selSetComputeSampler = objc.RegisterName("setSamplerState:atIndex:")

	// Depth-stencil support.
	selSetDepthAttachPixFmt = objc.RegisterName("setDepthAttachmentPixelFormat:")
	selDepthAttachment      = objc.RegisterName("depthAttachment")
	selSetClearDepth        = objc.RegisterName("setClearDepth:")
	selNewDepthStencilState = objc.RegisterName("newDepthStencilStateWithDescriptor:")
	selSetDepthCompareFunc  = objc.RegisterName("setDepthCompareFunction:")
	selSetDepthWriteEnabled = objc.RegisterName("setDepthWriteEnabled:")
	selSetDepthStencilState = objc.RegisterName("setDepthStencilState:")
)

// SamplerMinMagFilter selects nearest or linear filtering.
type SamplerMinMagFilter uint8

const (
	SamplerFilterNearest SamplerMinMagFilter = 0
	SamplerFilterLinear  SamplerMinMagFilter = 1
)

// SamplerAddressMode selects how out-of-range texture coordinates are handled.
type SamplerAddressMode uint8

const (
	SamplerAddressClampToEdge SamplerAddressMode = 0
	SamplerAddressRepeat      SamplerAddressMode = 2
)

// SamplerDescriptor configures a sampler.
type SamplerDescriptor struct {
	MinFilter    SamplerMinMagFilter
	MagFilter    SamplerMinMagFilter
	SAddressMode SamplerAddressMode
	TAddressMode SamplerAddressMode
}

// SamplerState is a compiled texture sampler.
type SamplerState struct {
	samplerState objc.ID
}

// MakeSamplerState creates a sampler state object.
func (d Device) MakeSamplerState(sd SamplerDescriptor) SamplerState {
	desc := objc.ID(objc.GetClass("MTLSamplerDescriptor")).Send(selAlloc).Send(selInit)
	desc.Send(selSetMinFilter, uint64(sd.MinFilter))
	desc.Send(selSetMagFilter, uint64(sd.MagFilter))
	desc.Send(selSetSAddressMode, uint64(sd.SAddressMode))
	desc.Send(selSetTAddressMode, uint64(sd.TAddressMode))
	s := d.device.Send(selNewSamplerState, desc)
	desc.Send(selRelease)
	return SamplerState{s}
}

// SetTexture binds a texture for the compute function.
func (cce ComputeCommandEncoder) SetTexture(t Texture, index int) {
	cce.commandEncoder.Send(selSetComputeTexture, t.texture, uint64(index))
}

// SetSamplerState binds a sampler for the compute function.
func (cce ComputeCommandEncoder) SetSamplerState(s SamplerState, index int) {
	cce.commandEncoder.Send(selSetComputeSampler, s.samplerState, uint64(index))
}

// mtlClearColor matches MTLClearColor (4 doubles).
type mtlClearColor struct{ red, green, blue, alpha float64 }

// TextureUsage describes how a texture may be used.
// https://developer.apple.com/documentation/metal/mtltextureusage.
type TextureUsage uint8

const (
	TextureUsageShaderRead   TextureUsage = 1 << 0
	TextureUsageShaderWrite  TextureUsage = 1 << 1
	TextureUsageRenderTarget TextureUsage = 1 << 2
)

// LoadAction is what a render pass does with an attachment at the start.
// https://developer.apple.com/documentation/metal/mtlloadaction.
type LoadAction uint8

const (
	LoadActionDontCare LoadAction = 0
	LoadActionLoad     LoadAction = 1
	LoadActionClear    LoadAction = 2
)

// StoreAction is what a render pass does with an attachment at the end.
// https://developer.apple.com/documentation/metal/mtlstoreaction.
type StoreAction uint8

const (
	StoreActionDontCare StoreAction = 0
	StoreActionStore    StoreAction = 1
)

// PrimitiveType is the geometry primitive a draw call assembles.
// https://developer.apple.com/documentation/metal/mtlprimitivetype.
type PrimitiveType uint8

const (
	PrimitiveTypePoint         PrimitiveType = 0
	PrimitiveTypeLine          PrimitiveType = 1
	PrimitiveTypeLineStrip     PrimitiveType = 2
	PrimitiveTypeTriangle      PrimitiveType = 3
	PrimitiveTypeTriangleStrip PrimitiveType = 4
)

// ClearColor is the value an attachment is cleared to.
type ClearColor struct{ Red, Green, Blue, Alpha float64 }

// RenderPipelineState is a compiled render pipeline.
// https://developer.apple.com/documentation/metal/mtlrenderpipelinestate.
type RenderPipelineState struct {
	renderPipelineState objc.ID
}

// CompareFunction is the depth comparison test.
// https://developer.apple.com/documentation/metal/mtlcomparefunction.
type CompareFunction uint8

const (
	CompareFunctionNever     CompareFunction = 0
	CompareFunctionLess      CompareFunction = 1
	CompareFunctionLessEqual CompareFunction = 3
	CompareFunctionGreater   CompareFunction = 4
	CompareFunctionAlways    CompareFunction = 7
)

// DepthStencilState is a compiled depth/stencil state.
// https://developer.apple.com/documentation/metal/mtldepthstencilstate.
type DepthStencilState struct {
	depthStencilState objc.ID
}

// DepthStencilDescriptor configures a depth/stencil state.
type DepthStencilDescriptor struct {
	DepthCompareFunction CompareFunction
	DepthWriteEnabled    bool
}

// MakeDepthStencilState creates a depth/stencil state object.
func (d Device) MakeDepthStencilState(desc DepthStencilDescriptor) DepthStencilState {
	dsd := objc.ID(objc.GetClass("MTLDepthStencilDescriptor")).Send(selAlloc).Send(selInit)
	defer dsd.Send(selRelease)
	dsd.Send(selSetDepthCompareFunc, uint64(desc.DepthCompareFunction))
	var write uint64
	if desc.DepthWriteEnabled {
		write = 1
	}
	dsd.Send(selSetDepthWriteEnabled, write)
	return DepthStencilState{d.device.Send(selNewDepthStencilState, dsd)}
}

// RenderPipelineDescriptor configures a render pipeline.
type RenderPipelineDescriptor struct {
	VertexFunction   Function
	FragmentFunction Function
	ColorPixelFormat PixelFormat
	// DepthPixelFormat is the depth attachment format, or 0 (Invalid) for none.
	DepthPixelFormat PixelFormat
}

// MakeRenderPipelineState creates a render pipeline state object.
// https://developer.apple.com/documentation/metal/mtldevice/1433369-makerenderpipelinestate.
func (d Device) MakeRenderPipelineState(desc RenderPipelineDescriptor) (RenderPipelineState, error) {
	rpd := objc.ID(objc.GetClass("MTLRenderPipelineDescriptor")).Send(selAlloc).Send(selInit)
	defer rpd.Send(selRelease)
	rpd.Send(selSetVertexFunction, desc.VertexFunction.function)
	rpd.Send(selSetFragmentFunction, desc.FragmentFunction.function)
	att := rpd.Send(selColorAttachments).Send(selObjectAtIndexed, uint64(0))
	att.Send(selSetPixelFormat, uint64(desc.ColorPixelFormat))
	rpd.Send(selSetDepthAttachPixFmt, uint64(desc.DepthPixelFormat))

	var err objc.ID
	pso := d.device.Send(selNewRenderPipeline, rpd, unsafe.Pointer(&err))
	if pso == 0 {
		return RenderPipelineState{}, errors.New(nsErrorString(err))
	}
	return RenderPipelineState{pso}, nil
}

// ColorAttachment configures a single render-pass color attachment.
type ColorAttachment struct {
	Texture     Texture
	LoadAction  LoadAction
	StoreAction StoreAction
	ClearColor  ClearColor
}

// DepthAttachment configures a render-pass depth attachment.
type DepthAttachment struct {
	Texture     Texture
	LoadAction  LoadAction
	StoreAction StoreAction
	ClearDepth  float64
}

// RenderPassDescriptor describes a render pass's attachments.
type RenderPassDescriptor struct {
	ColorAttachment0 ColorAttachment
	// Depth is the optional depth attachment. It is used when its Texture is set
	// (a non-zero texture id).
	Depth DepthAttachment
}

// objc builds the MTLRenderPassDescriptor.
func (rp RenderPassDescriptor) objc() objc.ID {
	d := objc.ID(objc.GetClass("MTLRenderPassDescriptor")).Send(selRenderPassDesc)
	att := d.Send(selColorAttachments).Send(selObjectAtIndexed, uint64(0))
	c := rp.ColorAttachment0
	att.Send(selSetTexture, c.Texture.texture)
	att.Send(selSetLoadAction, uint64(c.LoadAction))
	att.Send(selSetStoreAction, uint64(c.StoreAction))
	att.Send(selSetClearColor, mtlClearColor{c.ClearColor.Red, c.ClearColor.Green, c.ClearColor.Blue, c.ClearColor.Alpha})
	if rp.Depth.Texture.texture != 0 {
		da := d.Send(selDepthAttachment)
		da.Send(selSetTexture, rp.Depth.Texture.texture)
		da.Send(selSetLoadAction, uint64(rp.Depth.LoadAction))
		da.Send(selSetStoreAction, uint64(rp.Depth.StoreAction))
		da.Send(selSetClearDepth, rp.Depth.ClearDepth)
	}
	return d
}

// RenderCommandEncoder encodes a render pass.
// https://developer.apple.com/documentation/metal/mtlrendercommandencoder.
type RenderCommandEncoder struct {
	CommandEncoder
}

// MakeRenderCommandEncoder creates a render command encoder for the pass.
func (cb CommandBuffer) MakeRenderCommandEncoder(desc RenderPassDescriptor) RenderCommandEncoder {
	enc := cb.commandBuffer.Send(selRenderEncoder, desc.objc())
	return RenderCommandEncoder{CommandEncoder{enc}}
}

// SetRenderPipelineState sets the current render pipeline state.
func (rce RenderCommandEncoder) SetRenderPipelineState(rps RenderPipelineState) {
	rce.commandEncoder.Send(selSetRenderPipeline, rps.renderPipelineState)
}

// SetDepthStencilState sets the depth/stencil state for subsequent draws.
func (rce RenderCommandEncoder) SetDepthStencilState(s DepthStencilState) {
	rce.commandEncoder.Send(selSetDepthStencilState, s.depthStencilState)
}

// SetVertexBuffer binds a buffer for the vertex function.
func (rce RenderCommandEncoder) SetVertexBuffer(b Buffer, offset, index int) {
	rce.commandEncoder.Send(selSetVertexBuffer, b.buffer, uint64(offset), uint64(index))
}

// SetFragmentBuffer binds a buffer for the fragment function.
func (rce RenderCommandEncoder) SetFragmentBuffer(b Buffer, offset, index int) {
	rce.commandEncoder.Send(selSetFragmentBuffer, b.buffer, uint64(offset), uint64(index))
}

// SetVertexBytes sets inline data for the vertex function.
func (rce RenderCommandEncoder) SetVertexBytes(b []byte, index int) {
	rce.commandEncoder.Send(selSetVertexBytes, unsafe.Pointer(&b[0]), uint64(len(b)), uint64(index))
}

// DrawPrimitives draws vertexCount vertices starting at vertexStart.
func (rce RenderCommandEncoder) DrawPrimitives(typ PrimitiveType, vertexStart, vertexCount int) {
	rce.commandEncoder.Send(selDrawPrimitives, uint64(typ), uint64(vertexStart), uint64(vertexCount))
}

// GetBytes reads texture pixels back into dst (e.g. for headless readback).
func (t Texture) GetBytes(dst []byte, bytesPerRow int, region Region, level int) {
	r := mtlRegion{origin: region.Origin.c(), size: region.Size.c()}
	t.texture.Send(selGetBytes, unsafe.Pointer(&dst[0]), uint64(bytesPerRow), r, uint64(level))
}
