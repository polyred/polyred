// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gpu

// backend is the private driver interface the public Device API dispatches to.
// One implementation per driver (Metal now; GL/Vulkan/DX12 later) lets the
// public surface stay backend-agnostic. openBackend is provided per platform.
type backend interface {
	newBuffer(size int, usage BufferUsage, data []byte) (backendBuffer, error)
	newShaderModule(src ShaderSource) (backendShaderModule, error)
	newComputePipeline(mod backendShaderModule, entry string) (backendComputePipeline, error)
	newTexture(format TextureFormat, w, h int, renderTarget bool) (backendTexture, error)
	newRenderPipeline(vmod backendShaderModule, ventry string, fmod backendShaderModule, fentry string, color TextureFormat) (backendRenderPipeline, error)
	newCommandBuffer() backendCommandBuffer
	waitIdle()
	close() error
}

type backendTexture interface {
	readPixels() []byte
}

type backendRenderPipeline interface{ isRenderPipeline() }

// renderPassInfo is the backend-facing description of a render pass.
type renderPassInfo struct {
	color      backendTexture
	load       LoadOp
	clearColor [4]float64
}

type backendBuffer interface {
	bytes() []byte
	release()
}

type backendShaderModule interface{ isShaderModule() }

type backendComputePipeline interface{ maxThreads() int }

// backendCommandBuffer records and submits one command buffer. The Metal
// backend encodes eagerly (Metal is itself the explicit command-buffer model);
// the future GL backend records and replays on its context thread.
type backendCommandBuffer interface {
	beginCompute()
	setComputePipeline(backendComputePipeline)
	setBuffer(b backendBuffer, offset, index int)
	dispatch(x, y, z int)
	endCompute()

	beginRender(info renderPassInfo)
	setRenderPipeline(backendRenderPipeline)
	setRenderBuffer(b backendBuffer, offset, index int)
	setVertexBuffer(b backendBuffer, index int)
	draw(prim Primitive, start, count int)
	endRender()

	commit()
}
