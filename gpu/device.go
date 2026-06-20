// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gpu is a backend-agnostic abstraction for running compute (and,
// later, rendering) pipelines on the GPU. It mirrors the WebGPU object model
// — Device, Queue, Buffer, BindGroup, Pipeline, CommandEncoder — and selects a
// driver (Metal, and later Vulkan/DX12/GL) underneath. It is cgo-free.
//
// See docs/gpu-abstraction.md for the full design and locked decisions. This
// file is Phase 1: the compute subset, with the Metal backend.
package gpu

import "errors"

// Driver identifies a GPU backend.
type Driver int

const (
	// DriverAuto selects the best available driver for the platform.
	DriverAuto Driver = iota
	DriverMetal
	DriverVulkan
	DriverD3D12
	DriverGL
)

func (d Driver) String() string {
	switch d {
	case DriverMetal:
		return "metal"
	case DriverVulkan:
		return "vulkan"
	case DriverD3D12:
		return "d3d12"
	case DriverGL:
		return "gl"
	default:
		return "auto"
	}
}

// ErrUnsupported is returned by Open when no usable GPU driver is available.
var ErrUnsupported = errors.New("gpu: no supported GPU driver available")

// Option configures Open.
type Option func(*config)

type config struct {
	driver Driver
}

// WithDriver forces a specific driver instead of auto-selection.
func WithDriver(d Driver) Option {
	return func(c *config) { c.driver = d }
}

// Device is the root object: the factory for GPU resources and the owner of the
// command queue. Obtain one with Open.
type Device struct {
	b      backend
	driver Driver
	queue  *Queue
}

// Open negotiates a GPU device for the selected (or best available) driver.
func Open(opts ...Option) (*Device, error) {
	var c config
	for _, o := range opts {
		o(&c)
	}
	b, drv, err := openBackend(c.driver)
	if err != nil {
		return nil, err
	}
	d := &Device{b: b, driver: drv}
	d.queue = &Queue{d: d}
	return d, nil
}

// Driver reports the active driver.
func (d *Device) Driver() Driver { return d.driver }

// Queue returns the device's command queue.
func (d *Device) Queue() *Queue { return d.queue }

// Close releases the device and its backend resources.
func (d *Device) Close() error { return d.b.close() }

// BufferUsage is a bitmask describing how a buffer will be used; the backend
// maps it to the appropriate storage mode.
type BufferUsage uint32

const (
	BufferCopySrc BufferUsage = 1 << iota
	BufferCopyDst
	BufferStorage // read-write storage buffer (SSBO / Metal device buffer)
	BufferUniform // read-only uniform/constant buffer
	BufferMapRead
	BufferMapWrite
)

// BufferDescriptor describes a buffer to create.
type BufferDescriptor struct {
	Label string
	Size  int // size in bytes; if 0 and Data != nil, len(Data) is used
	Usage BufferUsage
	Data  []byte // optional initial contents
}

// Buffer is a GPU memory allocation.
type Buffer struct {
	b    backendBuffer
	size int
}

// Size returns the buffer size in bytes.
func (b *Buffer) Size() int { return b.size }

// Bytes returns a CPU-visible view of the buffer's contents (valid for
// shared/managed storage). The slice aliases GPU memory; copy out what you need.
func (b *Buffer) Bytes() []byte { return b.b.bytes() }

// Release frees the buffer.
func (b *Buffer) Release() { b.b.release() }

// NewBuffer allocates a buffer.
func (d *Device) NewBuffer(desc BufferDescriptor) (*Buffer, error) {
	size := desc.Size
	if size == 0 {
		size = len(desc.Data)
	}
	if size == 0 {
		return nil, errors.New("gpu: buffer size must be > 0")
	}
	bb, err := d.b.newBuffer(size, desc.Usage, desc.Data)
	if err != nil {
		return nil, err
	}
	return &Buffer{b: bb, size: size}, nil
}

// ShaderStage is a bitmask of pipeline stages a binding is visible to.
type ShaderStage uint32

const (
	StageVertex ShaderStage = 1 << iota
	StageFragment
	StageCompute
)

// ShaderSource carries per-backend shader text. Normally produced by the
// Go→shader compiler; per-language fields let one module hold all variants and
// allow hand-authored escape-hatch shaders. This phase uses MSL.
type ShaderSource struct {
	MSL   string
	GLSL  string
	HLSL  string
	SPIRV []byte
}

// ShaderModule is a compiled shader library for the active backend.
type ShaderModule struct {
	b backendShaderModule
}

// NewShaderModule compiles shader source for the active backend.
func (d *Device) NewShaderModule(src ShaderSource) (*ShaderModule, error) {
	bm, err := d.b.newShaderModule(src)
	if err != nil {
		return nil, err
	}
	return &ShaderModule{b: bm}, nil
}

// BindingKind is the resource type of a bind-group entry.
type BindingKind int

const (
	StorageBuffer BindingKind = iota
	UniformBuffer
)

// BindGroupLayoutEntry declares one binding in a layout.
type BindGroupLayoutEntry struct {
	Binding    int
	Visibility ShaderStage
	Kind       BindingKind
}

// BindGroupLayout declares the shape of a bind group.
type BindGroupLayout struct {
	entries []BindGroupLayoutEntry
}

// NewBindGroupLayout creates a bind-group layout.
func (d *Device) NewBindGroupLayout(entries ...BindGroupLayoutEntry) *BindGroupLayout {
	return &BindGroupLayout{entries: entries}
}

// PipelineLayout is the ordered list of bind-group layouts a pipeline expects.
type PipelineLayout struct {
	groups []*BindGroupLayout
}

// NewPipelineLayout creates a pipeline layout from ordered group layouts.
func (d *Device) NewPipelineLayout(groups ...*BindGroupLayout) *PipelineLayout {
	return &PipelineLayout{groups: groups}
}

// BindGroupEntry binds a concrete resource to a binding index.
type BindGroupEntry struct {
	Binding int
	Buffer  *Buffer
}

// BindGroup is a concrete set of resources matching a BindGroupLayout.
type BindGroup struct {
	layout  *BindGroupLayout
	entries []BindGroupEntry
}

// NewBindGroup creates a bind group for the given layout.
func (d *Device) NewBindGroup(layout *BindGroupLayout, entries ...BindGroupEntry) *BindGroup {
	return &BindGroup{layout: layout, entries: entries}
}

// ComputePipelineDescriptor describes a compute pipeline.
type ComputePipelineDescriptor struct {
	Label  string
	Layout *PipelineLayout
	Module *ShaderModule
	Entry  string
}

// ComputePipeline is a compiled compute pipeline.
type ComputePipeline struct {
	b      backendComputePipeline
	layout *PipelineLayout
}

// NewComputePipeline creates a compute pipeline.
func (d *Device) NewComputePipeline(desc ComputePipelineDescriptor) (*ComputePipeline, error) {
	if desc.Module == nil {
		return nil, errors.New("gpu: compute pipeline requires a shader module")
	}
	bp, err := d.b.newComputePipeline(desc.Module.b, desc.Entry)
	if err != nil {
		return nil, err
	}
	return &ComputePipeline{b: bp, layout: desc.Layout}, nil
}

// CommandEncoder records GPU commands into a CommandBuffer.
type CommandEncoder struct {
	d   *Device
	cmd backendCommandBuffer
}

// NewCommandEncoder begins recording a command buffer.
func (d *Device) NewCommandEncoder() *CommandEncoder {
	return &CommandEncoder{d: d, cmd: d.b.newCommandBuffer()}
}

// BeginComputePass starts a compute pass.
func (e *CommandEncoder) BeginComputePass() *ComputePass {
	e.cmd.beginCompute()
	return &ComputePass{e: e}
}

// Finish ends recording and returns the command buffer for submission.
func (e *CommandEncoder) Finish() *CommandBuffer {
	return &CommandBuffer{cmd: e.cmd}
}

// ComputePass encodes commands in a compute pass.
type ComputePass struct {
	e *CommandEncoder
}

// SetPipeline binds the compute pipeline for subsequent dispatches.
func (p *ComputePass) SetPipeline(cp *ComputePipeline) {
	p.e.cmd.setComputePipeline(cp.b)
}

// SetBindGroup binds a bind group at the given group index. For the flat Metal
// mapping, bindings translate directly to buffer indices.
func (p *ComputePass) SetBindGroup(group int, bg *BindGroup) {
	for _, e := range bg.entries {
		p.e.cmd.setBuffer(e.Buffer.b, 0, e.Binding)
	}
}

// Dispatch runs the pipeline over a grid of the given number of threads (x*y*z).
// Workgroup sizing is chosen by the backend from the pipeline limits.
func (p *ComputePass) Dispatch(x, y, z int) {
	p.e.cmd.dispatch(x, y, z)
}

// End finishes the compute pass.
func (p *ComputePass) End() {
	p.e.cmd.endCompute()
}

// CommandBuffer is a finished, submittable set of recorded commands.
type CommandBuffer struct {
	cmd backendCommandBuffer
}

// Queue submits command buffers to the GPU.
type Queue struct {
	d *Device
}

// Submit submits command buffers for execution.
func (q *Queue) Submit(cbs ...*CommandBuffer) {
	for _, cb := range cbs {
		cb.cmd.commit()
	}
}

// WaitIdle blocks until all submitted work has completed.
func (q *Queue) WaitIdle() {
	q.d.b.waitIdle()
}
