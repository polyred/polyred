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
	newSampler(desc SamplerDescriptor) backendSampler
	newRenderPipeline(vmod backendShaderModule, ventry string, fmod backendShaderModule, fentry string, color TextureFormat, extraColor []TextureFormat, depth TextureFormat) (backendRenderPipeline, error)
	newCommandBuffer() backendCommandBuffer
	newWindowSurface(display, window uintptr, w, h int) (backendWindowSurface, error)
	waitIdle()
	close() error
}

type backendTexture interface {
	readPixels() []byte
	write(pixels []byte, bytesPerRow int)
}

// backendWindowSurface is an on-screen swapchain bound to a native window.
type backendWindowSurface interface {
	acquire() backendTexture // render target for the next frame
	present() error          // blit the acquired texture to the window, swap buffers
	resize(w, h int) error
	readback() []byte // the presented pixels, top-down RGBA (for testing/screenshots)
	release()
}

type backendSampler interface{ isSampler() }

type backendRenderPipeline interface{ isRenderPipeline() }

// renderColorTarget is one extra color attachment (1..N) of a render pass.
type renderColorTarget struct {
	tex   backendTexture
	clear [4]float64
}

// renderPassInfo is the backend-facing description of a render pass.
type renderPassInfo struct {
	color      backendTexture
	load       LoadOp
	clearColor [4]float64
	extraColor []renderColorTarget // color attachments 1..N
	depth      backendTexture      // optional depth attachment
	clearDepth float64
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
	setComputeTexture(index int, t backendTexture)
	setComputeSampler(index int, s backendSampler)
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
