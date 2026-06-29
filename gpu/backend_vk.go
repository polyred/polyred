// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

// This file wires the proven cgo-free Vulkan compute path (see the
// gpu/vk*_linux_test.go probes) behind the private backend interface, so
// gpu.Open(WithDriver(DriverVulkan)) is a first-class driver alongside Metal and
// GL. Vulkan is reached through purego (no cgo). Shader modules consume SPIR-V
// (ShaderSource.SPIRV); a Go->SPIR-V emitter is a follow-up, so callers compile
// their GLSL to SPIR-V (e.g. with glslang) today. Compute only; render is a
// follow-up. Verified in CI on Mesa lavapipe.
package gpu

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Vulkan struct types (C layout; Go's natural alignment matches on amd64).
type (
	vkInstanceCreateInfoB struct {
		sType, _                 uint32
		pNext                    uintptr
		flags, _                 uint32
		pApplicationInfo         uintptr
		enabledLayerCount, _     uint32
		ppEnabledLayerNames      uintptr
		enabledExtensionCount, _ uint32
		ppEnabledExtensionNames  uintptr
	}
	vkDeviceQueueCreateInfoB struct {
		sType                             uint32
		pNext                             uintptr
		flags, queueFamilyIndex, queueCnt uint32
		pQueuePriorities                  uintptr
	}
	vkDeviceCreateInfoB struct {
		sType                                     uint32
		pNext                                     uintptr
		flags, queueCreateInfoCount               uint32
		pQueueCreateInfos                         uintptr
		enabledLayerCount                         uint32
		ppEnabledLayerNames                       uintptr
		enabledExtensionCount                     uint32
		ppEnabledExtensionNames, pEnabledFeatures uintptr
	}
	vkBufferCreateInfoB struct {
		sType       uint32
		pNext       uintptr
		flags       uint32
		size        uint64
		usage, mode uint32
		qfiCount    uint32
		pQFI        uintptr
	}
	vkMemoryRequirementsB struct {
		size, alignment uint64
		memoryTypeBits  uint32
	}
	vkMemoryAllocateInfoB struct {
		sType          uint32
		pNext          uintptr
		allocationSize uint64
		memoryTypeIdx  uint32
	}
	vkShaderModuleCreateInfoB struct {
		sType    uint32
		pNext    uintptr
		flags    uint32
		codeSize uint64
		pCode    uintptr
	}
	vkDSLBindingB struct {
		binding, descriptorType, descriptorCount, stageFlags uint32
		pImmutableSamplers                                   uintptr
	}
	vkDSLCreateInfoB struct {
		sType        uint32
		pNext        uintptr
		flags        uint32
		bindingCount uint32
		pBindings    uintptr
	}
	vkPipelineLayoutCreateInfoB struct {
		sType                  uint32
		pNext                  uintptr
		flags, setLayoutCount  uint32
		pSetLayouts            uintptr
		pushConstantRangeCount uint32
		pPushConstantRanges    uintptr
	}
	vkShaderStageCreateInfoB struct {
		sType         uint32
		pNext         uintptr
		flags, stage  uint32
		module, pName uintptr
		pSpec         uintptr
	}
	vkComputePipelineCreateInfoB struct {
		sType        uint32
		pNext        uintptr
		flags        uint32
		stage        vkShaderStageCreateInfoB
		layout       uintptr
		basePipeline uintptr
		baseIndex    int32
	}
	vkDescriptorPoolSizeB       struct{ typ, descriptorCount uint32 }
	vkDescriptorPoolCreateInfoB struct {
		sType                         uint32
		pNext                         uintptr
		flags, maxSets, poolSizeCount uint32
		pPoolSizes                    uintptr
	}
	vkDSAllocateInfoB struct {
		sType          uint32
		pNext          uintptr
		descriptorPool uintptr
		count          uint32
		pSetLayouts    uintptr
	}
	vkDescriptorBufferInfoB struct {
		buffer      uintptr
		offset, rng uint64
	}
	vkWriteDescriptorSetB struct {
		sType                                                  uint32
		pNext                                                  uintptr
		dstSet                                                 uintptr
		dstBinding, dstArrayElement, descriptorCount, descType uint32
		pImageInfo, pBufferInfo, pTexelBufferView              uintptr
	}
	vkCommandPoolCreateInfoB struct {
		sType           uint32
		pNext           uintptr
		flags, queueFam uint32
	}
	vkCommandBufferAllocateInfoB struct {
		sType       uint32
		pNext       uintptr
		commandPool uintptr
		level, cnt  uint32
	}
	vkCommandBufferBeginInfoB struct {
		sType uint32
		pNext uintptr
		flags uint32
		pInh  uintptr
	}
	vkSubmitInfoB struct {
		sType             uint32
		pNext             uintptr
		waitCount         uint32
		pWait, pWaitStage uintptr
		cmdCount          uint32
		pCmd              uintptr
		signalCount       uint32
		pSignal           uintptr
	}
)

const (
	vksInstance    = 1
	vksDevQueue    = 2
	vksDevice      = 3
	vksSubmit      = 4
	vksMemAlloc    = 5
	vksBuffer      = 12
	vksShaderMod   = 16
	vksShaderStage = 18
	vksComputePipe = 29
	vksPipeLayout  = 30
	vksDSL         = 32
	vksDescPool    = 33
	vksDSAlloc     = 34
	vksWriteDS     = 35
	vksCmdPool     = 39
	vksCmdBufAlloc = 40
	vksCmdBegin    = 42

	vkUsageStorage      = 0x20
	vkMemHostVisibleB   = 0x2
	vkMemHostCoherentB  = 0x4
	vkDescStorageBuffer = 7
	vkStageComputeB     = 0x20
	vkBindCompute       = 1
	vkQueueComputeBitB  = 0x2
)

type vkBackend struct {
	lib     uintptr
	fn      map[string]uintptr
	pd      uintptr
	device  uintptr
	queue   uintptr
	qfi     uint32
	cmdPool uintptr
	memProp []byte
	memN    uint32
	mu      sync.Mutex
}

func (b *vkBackend) c(name string, args ...uintptr) {
	r, _, _ := purego.SyscallN(b.fn[name], args...)
	if int32(r) != 0 {
		panic(fmt.Sprintf("gpu/vk: %s failed: VkResult=%d", name, int32(r)))
	}
}

// openVKBackend opens the Vulkan backend; it is the linux entry point dispatched
// from the GL backend's openBackend when DriverVulkan is requested. Non-linux
// builds use the stub in backend_vk_other.go.
func openVKBackend(c config) (backend, Driver, error) {
	vb, err := newVKBackend()
	if err != nil {
		return nil, DriverAuto, err
	}
	return vb, DriverVulkan, nil
}

func newVKBackend() (b *vkBackend, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	lib, e := purego.Dlopen("libvulkan.so.1", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if e != nil {
		return nil, fmt.Errorf("gpu/vk: %w", e)
	}
	b = &vkBackend{lib: lib, fn: map[string]uintptr{}}
	for _, name := range []string{
		"vkCreateInstance", "vkEnumeratePhysicalDevices", "vkGetPhysicalDeviceQueueFamilyProperties",
		"vkCreateDevice", "vkGetDeviceQueue", "vkGetPhysicalDeviceMemoryProperties",
		"vkCreateBuffer", "vkGetBufferMemoryRequirements", "vkAllocateMemory", "vkBindBufferMemory",
		"vkMapMemory", "vkDestroyBuffer", "vkFreeMemory", "vkCreateShaderModule",
		"vkCreateDescriptorSetLayout", "vkCreatePipelineLayout", "vkCreateComputePipelines",
		"vkCreateDescriptorPool", "vkResetDescriptorPool", "vkAllocateDescriptorSets", "vkUpdateDescriptorSets",
		"vkCreateCommandPool", "vkResetCommandPool", "vkAllocateCommandBuffers", "vkBeginCommandBuffer",
		"vkCmdBindPipeline", "vkCmdBindDescriptorSets", "vkCmdDispatch", "vkEndCommandBuffer",
		"vkQueueSubmit", "vkDeviceWaitIdle", "vkDestroyDevice",
	} {
		p, e := purego.Dlsym(lib, name)
		if e != nil {
			return nil, fmt.Errorf("gpu/vk: dlsym %s: %w", name, e)
		}
		b.fn[name] = p
	}

	ici := vkInstanceCreateInfoB{sType: vksInstance}
	var instance uintptr
	b.c("vkCreateInstance", uintptr(unsafe.Pointer(&ici)), 0, uintptr(unsafe.Pointer(&instance)))

	var nd uint32
	purego.SyscallN(b.fn["vkEnumeratePhysicalDevices"], instance, uintptr(unsafe.Pointer(&nd)), 0)
	if nd == 0 {
		return nil, fmt.Errorf("gpu/vk: no physical devices")
	}
	devs := make([]uintptr, nd)
	purego.SyscallN(b.fn["vkEnumeratePhysicalDevices"], instance, uintptr(unsafe.Pointer(&nd)), uintptr(unsafe.Pointer(&devs[0])))
	b.pd = devs[0]

	var qn uint32
	purego.SyscallN(b.fn["vkGetPhysicalDeviceQueueFamilyProperties"], b.pd, uintptr(unsafe.Pointer(&qn)), 0)
	qp := make([]byte, int(qn)*24)
	purego.SyscallN(b.fn["vkGetPhysicalDeviceQueueFamilyProperties"], b.pd, uintptr(unsafe.Pointer(&qn)), uintptr(unsafe.Pointer(&qp[0])))
	found := false
	for q := 0; q < int(qn); q++ {
		if *(*uint32)(unsafe.Pointer(&qp[q*24]))&vkQueueComputeBitB != 0 {
			b.qfi = uint32(q)
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("gpu/vk: no compute queue family")
	}

	prio := float32(1)
	qci := vkDeviceQueueCreateInfoB{sType: vksDevQueue, queueFamilyIndex: b.qfi, queueCnt: 1, pQueuePriorities: uintptr(unsafe.Pointer(&prio))}
	dci := vkDeviceCreateInfoB{sType: vksDevice, queueCreateInfoCount: 1, pQueueCreateInfos: uintptr(unsafe.Pointer(&qci))}
	b.c("vkCreateDevice", b.pd, uintptr(unsafe.Pointer(&dci)), 0, uintptr(unsafe.Pointer(&b.device)))
	purego.SyscallN(b.fn["vkGetDeviceQueue"], b.device, uintptr(b.qfi), 0, uintptr(unsafe.Pointer(&b.queue)))

	b.memProp = make([]byte, 1024)
	purego.SyscallN(b.fn["vkGetPhysicalDeviceMemoryProperties"], b.pd, uintptr(unsafe.Pointer(&b.memProp[0])))
	b.memN = *(*uint32)(unsafe.Pointer(&b.memProp[0]))

	cpci := vkCommandPoolCreateInfoB{sType: vksCmdPool, queueFam: b.qfi}
	b.c("vkCreateCommandPool", b.device, uintptr(unsafe.Pointer(&cpci)), 0, uintptr(unsafe.Pointer(&b.cmdPool)))
	return b, nil
}

func (b *vkBackend) hostMemType(bits uint32) uint32 {
	for i := 0; i < int(b.memN); i++ {
		if bits&(1<<uint(i)) == 0 {
			continue
		}
		f := *(*uint32)(unsafe.Pointer(&b.memProp[4+i*8]))
		if f&vkMemHostVisibleB != 0 && f&vkMemHostCoherentB != 0 {
			return uint32(i)
		}
	}
	panic("gpu/vk: no host-visible coherent memory type")
}

type vkBuffer struct {
	b              *vkBackend
	buffer, memory uintptr
	ptr            unsafe.Pointer
	size           int
}

func (b *vkBackend) newBuffer(size int, usage BufferUsage, data []byte) (bb backendBuffer, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	b.mu.Lock()
	defer b.mu.Unlock()
	buf := &vkBuffer{b: b, size: size}
	bci := vkBufferCreateInfoB{sType: vksBuffer, size: uint64(size), usage: vkUsageStorage}
	b.c("vkCreateBuffer", b.device, uintptr(unsafe.Pointer(&bci)), 0, uintptr(unsafe.Pointer(&buf.buffer)))
	var req vkMemoryRequirementsB
	purego.SyscallN(b.fn["vkGetBufferMemoryRequirements"], b.device, buf.buffer, uintptr(unsafe.Pointer(&req)))
	mai := vkMemoryAllocateInfoB{sType: vksMemAlloc, allocationSize: req.size, memoryTypeIdx: b.hostMemType(req.memoryTypeBits)}
	b.c("vkAllocateMemory", b.device, uintptr(unsafe.Pointer(&mai)), 0, uintptr(unsafe.Pointer(&buf.memory)))
	b.c("vkBindBufferMemory", b.device, buf.buffer, buf.memory, 0)
	var p uintptr
	b.c("vkMapMemory", b.device, buf.memory, 0, uintptr(size), 0, uintptr(unsafe.Pointer(&p)))
	buf.ptr = unsafe.Pointer(p)
	if len(data) > 0 {
		copy(unsafe.Slice((*byte)(buf.ptr), size), data)
	}
	return buf, nil
}

func (b *vkBuffer) bytes() []byte {
	out := make([]byte, b.size)
	copy(out, unsafe.Slice((*byte)(b.ptr), b.size))
	return out
}

func (b *vkBuffer) release() {
	b.b.mu.Lock()
	defer b.b.mu.Unlock()
	purego.SyscallN(b.b.fn["vkDestroyBuffer"], b.b.device, b.buffer, 0)
	purego.SyscallN(b.b.fn["vkFreeMemory"], b.b.device, b.memory, 0)
}

type vkModule struct {
	module uintptr
}

func (vkModule) isShaderModule() {}

func (b *vkBackend) newShaderModule(src ShaderSource) (m backendShaderModule, err error) {
	if len(src.SPIRV) == 0 {
		return nil, fmt.Errorf("gpu/vk: ShaderSource.SPIRV is empty (the Vulkan backend needs SPIR-V; compile GLSL with glslang)")
	}
	if len(src.SPIRV)%4 != 0 {
		return nil, fmt.Errorf("gpu/vk: SPIR-V length %d is not a multiple of 4", len(src.SPIRV))
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	b.mu.Lock()
	defer b.mu.Unlock()
	words := make([]uint32, len(src.SPIRV)/4)
	for i := range words {
		words[i] = uint32(src.SPIRV[i*4]) | uint32(src.SPIRV[i*4+1])<<8 | uint32(src.SPIRV[i*4+2])<<16 | uint32(src.SPIRV[i*4+3])<<24
	}
	smci := vkShaderModuleCreateInfoB{sType: vksShaderMod, codeSize: uint64(len(src.SPIRV)), pCode: uintptr(unsafe.Pointer(&words[0]))}
	var mod uintptr
	b.c("vkCreateShaderModule", b.device, uintptr(unsafe.Pointer(&smci)), 0, uintptr(unsafe.Pointer(&mod)))
	return vkModule{module: mod}, nil
}

type vkPipeline struct {
	b      *vkBackend
	module uintptr
	entry  []byte

	built    bool
	nbind    int
	dsl      uintptr
	layout   uintptr
	pipeline uintptr
}

func (p *vkPipeline) maxThreads() int { return 1024 }

func (b *vkBackend) newComputePipeline(mod backendShaderModule, entry string) (backendComputePipeline, error) {
	// SPIR-V from our GLSL pipeline (glslang) always names the compute entry
	// point "main" (from GLSL's void main()), regardless of the Device-API entry
	// name (which is the Go kernel name, used by the Metal/MSL backend). Use the
	// SPIR-V convention here.
	return &vkPipeline{b: b, module: mod.(vkModule).module, entry: append([]byte("main"), 0)}, nil
}

// build lazily creates the descriptor-set layout, pipeline layout and compute
// pipeline once the binding count is known (at first dispatch).
func (p *vkPipeline) build(nbind int) {
	b := p.b
	binds := make([]vkDSLBindingB, nbind)
	for i := range binds {
		binds[i] = vkDSLBindingB{binding: uint32(i), descriptorType: vkDescStorageBuffer, descriptorCount: 1, stageFlags: vkStageComputeB}
	}
	dslci := vkDSLCreateInfoB{sType: vksDSL, bindingCount: uint32(nbind), pBindings: uintptr(unsafe.Pointer(&binds[0]))}
	b.c("vkCreateDescriptorSetLayout", b.device, uintptr(unsafe.Pointer(&dslci)), 0, uintptr(unsafe.Pointer(&p.dsl)))
	plci := vkPipelineLayoutCreateInfoB{sType: vksPipeLayout, setLayoutCount: 1, pSetLayouts: uintptr(unsafe.Pointer(&p.dsl))}
	b.c("vkCreatePipelineLayout", b.device, uintptr(unsafe.Pointer(&plci)), 0, uintptr(unsafe.Pointer(&p.layout)))
	cpci := vkComputePipelineCreateInfoB{
		sType:  vksComputePipe,
		stage:  vkShaderStageCreateInfoB{sType: vksShaderStage, stage: vkStageComputeB, module: p.module, pName: uintptr(unsafe.Pointer(&p.entry[0]))},
		layout: p.layout,
	}
	b.c("vkCreateComputePipelines", b.device, 0, 1, uintptr(unsafe.Pointer(&cpci)), 0, uintptr(unsafe.Pointer(&p.pipeline)))
	p.nbind = nbind
	p.built = true
}

type vkBufBind struct {
	buf   *vkBuffer
	index int
}

type vkCmd struct {
	b     *vkBackend
	pipe  *vkPipeline
	binds []vkBufBind
	gx    int
}

func (b *vkBackend) newCommandBuffer() backendCommandBuffer { return &vkCmd{b: b} }

// newWindowSurface is not implemented on the Vulkan backend yet (no swapchain /
// WSI wiring); an on-screen present lands in a later phase.
func (b *vkBackend) newWindowSurface(display, window uintptr, w, h int) (backendWindowSurface, error) {
	return nil, ErrUnsupported
}

func (b *vkBackend) windowVisualID() uint32 { return 0 }

func (c *vkCmd) beginCompute() {}
func (c *vkCmd) setComputePipeline(p backendComputePipeline) {
	c.pipe = p.(*vkPipeline)
}
func (c *vkCmd) setBuffer(buf backendBuffer, offset, index int) {
	c.binds = append(c.binds, vkBufBind{buf: buf.(*vkBuffer), index: index})
}
func (c *vkCmd) dispatch(x, y, z int) { c.gx = x }
func (c *vkCmd) endCompute()          {}

func (c *vkCmd) commit() {
	b := c.b
	b.mu.Lock()
	defer b.mu.Unlock()

	nbind := 0
	for _, bd := range c.binds {
		if bd.index+1 > nbind {
			nbind = bd.index + 1
		}
	}
	if !c.pipe.built {
		c.pipe.build(nbind)
	}

	// Fresh descriptor pool per submit (simplest correct lifetime).
	poolSize := vkDescriptorPoolSizeB{typ: vkDescStorageBuffer, descriptorCount: uint32(nbind)}
	dpci := vkDescriptorPoolCreateInfoB{sType: vksDescPool, maxSets: 1, poolSizeCount: 1, pPoolSizes: uintptr(unsafe.Pointer(&poolSize))}
	var pool uintptr
	b.c("vkCreateDescriptorPool", b.device, uintptr(unsafe.Pointer(&dpci)), 0, uintptr(unsafe.Pointer(&pool)))
	dsai := vkDSAllocateInfoB{sType: vksDSAlloc, descriptorPool: pool, count: 1, pSetLayouts: uintptr(unsafe.Pointer(&c.pipe.dsl))}
	var set uintptr
	b.c("vkAllocateDescriptorSets", b.device, uintptr(unsafe.Pointer(&dsai)), uintptr(unsafe.Pointer(&set)))

	infos := make([]vkDescriptorBufferInfoB, len(c.binds))
	writes := make([]vkWriteDescriptorSetB, len(c.binds))
	for i, bd := range c.binds {
		infos[i] = vkDescriptorBufferInfoB{buffer: bd.buf.buffer, rng: uint64(bd.buf.size)}
		writes[i] = vkWriteDescriptorSetB{
			sType: vksWriteDS, dstSet: set, dstBinding: uint32(bd.index), descriptorCount: 1,
			descType: vkDescStorageBuffer, pBufferInfo: uintptr(unsafe.Pointer(&infos[i])),
		}
	}
	purego.SyscallN(b.fn["vkUpdateDescriptorSets"], b.device, uintptr(len(writes)), uintptr(unsafe.Pointer(&writes[0])), 0, 0)

	purego.SyscallN(b.fn["vkResetCommandPool"], b.device, b.cmdPool, 0)
	cbai := vkCommandBufferAllocateInfoB{sType: vksCmdBufAlloc, commandPool: b.cmdPool, level: 0, cnt: 1}
	var cmd uintptr
	b.c("vkAllocateCommandBuffers", b.device, uintptr(unsafe.Pointer(&cbai)), uintptr(unsafe.Pointer(&cmd)))
	begin := vkCommandBufferBeginInfoB{sType: vksCmdBegin}
	b.c("vkBeginCommandBuffer", cmd, uintptr(unsafe.Pointer(&begin)))
	purego.SyscallN(b.fn["vkCmdBindPipeline"], cmd, vkBindCompute, c.pipe.pipeline)
	purego.SyscallN(b.fn["vkCmdBindDescriptorSets"], cmd, vkBindCompute, c.pipe.layout, 0, 1, uintptr(unsafe.Pointer(&set)), 0, 0)
	purego.SyscallN(b.fn["vkCmdDispatch"], cmd, uintptr(c.gx), 1, 1)
	b.c("vkEndCommandBuffer", cmd)

	si := vkSubmitInfoB{sType: vksSubmit, cmdCount: 1, pCmd: uintptr(unsafe.Pointer(&cmd))}
	b.c("vkQueueSubmit", b.queue, 1, uintptr(unsafe.Pointer(&si)), 0)
	purego.SyscallN(b.fn["vkDeviceWaitIdle"], b.device)
	purego.SyscallN(b.fn["vkResetDescriptorPool"], b.device, pool, 0)
}

func (b *vkBackend) waitIdle() {
	b.mu.Lock()
	defer b.mu.Unlock()
	purego.SyscallN(b.fn["vkDeviceWaitIdle"], b.device)
}

func (b *vkBackend) close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	purego.SyscallN(b.fn["vkDeviceWaitIdle"], b.device)
	purego.SyscallN(b.fn["vkDestroyDevice"], b.device, 0)
	return nil
}

// Render / texture / sampler are not implemented on the Vulkan backend yet.
func (b *vkBackend) newTexture(format TextureFormat, w, h int, renderTarget bool) (backendTexture, error) {
	return nil, fmt.Errorf("gpu/vk: textures not yet implemented")
}
func (b *vkBackend) newSampler(desc SamplerDescriptor) backendSampler { return nil }
func (b *vkBackend) newRenderPipeline(vmod backendShaderModule, ventry string, fmod backendShaderModule, fentry string, color TextureFormat, extraColor []TextureFormat, depth TextureFormat) (backendRenderPipeline, error) {
	return nil, fmt.Errorf("gpu/vk: render pipelines not yet implemented")
}

func (c *vkCmd) beginRender(info renderPassInfo)               {}
func (c *vkCmd) setRenderPipeline(backendRenderPipeline)       {}
func (c *vkCmd) setRenderBuffer(backendBuffer, int, int)       {}
func (c *vkCmd) setVertexBuffer(backendBuffer, int)            {}
func (c *vkCmd) draw(prim Primitive, start, count int)         {}
func (c *vkCmd) endRender()                                    {}
func (c *vkCmd) setComputeTexture(index int, t backendTexture) {}
func (c *vkCmd) setComputeSampler(index int, s backendSampler) {}
