// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

// End-to-end cgo-free Vulkan compute dispatch: a GLSL compute shader (compiled
// to SPIR-V by glslang in CI) doubling a storage buffer, run through the full
// Vulkan path via purego — shader module, descriptor set, compute pipeline,
// command buffer, vkCmdDispatch, queue submit — with the result checked against
// the CPU. This proves Vulkan compute works cgo-free (item #2, "then Vulkan"),
// verified in CI on Mesa lavapipe. Gated on POLYRED_VK_PROBE=1; also skips if
// glslangValidator is unavailable.
package gpu_test

import (
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

type vkShaderModuleCreateInfo struct {
	sType    uint32
	pNext    uintptr
	flags    uint32
	codeSize uint64
	pCode    uintptr
}

type vkDescriptorSetLayoutBinding struct {
	binding            uint32
	descriptorType     uint32
	descriptorCount    uint32
	stageFlags         uint32
	pImmutableSamplers uintptr
}

type vkDescriptorSetLayoutCreateInfo struct {
	sType        uint32
	pNext        uintptr
	flags        uint32
	bindingCount uint32
	pBindings    uintptr
}

type vkPipelineLayoutCreateInfo struct {
	sType                  uint32
	pNext                  uintptr
	flags                  uint32
	setLayoutCount         uint32
	pSetLayouts            uintptr
	pushConstantRangeCount uint32
	pPushConstantRanges    uintptr
}

type vkPipelineShaderStageCreateInfo struct {
	sType               uint32
	pNext               uintptr
	flags               uint32
	stage               uint32
	module              uintptr
	pName               uintptr
	pSpecializationInfo uintptr
}

type vkComputePipelineCreateInfo struct {
	sType              uint32
	pNext              uintptr
	flags              uint32
	stage              vkPipelineShaderStageCreateInfo
	layout             uintptr
	basePipelineHandle uintptr
	basePipelineIndex  int32
}

type vkDescriptorPoolSize struct {
	typ             uint32
	descriptorCount uint32
}

type vkDescriptorPoolCreateInfo struct {
	sType         uint32
	pNext         uintptr
	flags         uint32
	maxSets       uint32
	poolSizeCount uint32
	pPoolSizes    uintptr
}

type vkDescriptorSetAllocateInfo struct {
	sType              uint32
	pNext              uintptr
	descriptorPool     uintptr
	descriptorSetCount uint32
	pSetLayouts        uintptr
}

type vkDescriptorBufferInfo struct {
	buffer uintptr
	offset uint64
	rng    uint64
}

type vkWriteDescriptorSet struct {
	sType            uint32
	pNext            uintptr
	dstSet           uintptr
	dstBinding       uint32
	dstArrayElement  uint32
	descriptorCount  uint32
	descriptorType   uint32
	pImageInfo       uintptr
	pBufferInfo      uintptr
	pTexelBufferView uintptr
}

type vkCommandPoolCreateInfo struct {
	sType            uint32
	pNext            uintptr
	flags            uint32
	queueFamilyIndex uint32
}

type vkCommandBufferAllocateInfo struct {
	sType              uint32
	pNext              uintptr
	commandPool        uintptr
	level              uint32
	commandBufferCount uint32
}

type vkCommandBufferBeginInfo struct {
	sType            uint32
	pNext            uintptr
	flags            uint32
	pInheritanceInfo uintptr
}

type vkSubmitInfo struct {
	sType                uint32
	pNext                uintptr
	waitSemaphoreCount   uint32
	pWaitSemaphores      uintptr
	pWaitDstStageMask    uintptr
	commandBufferCount   uint32
	pCommandBuffers      uintptr
	signalSemaphoreCount uint32
	pSignalSemaphores    uintptr
}

const (
	vkStructShaderModuleCreateInfo        = 16
	vkStructDescriptorSetLayoutCreateInfo = 32
	vkStructPipelineLayoutCreateInfo      = 30
	vkStructPipelineShaderStageCreateInfo = 18
	vkStructComputePipelineCreateInfo     = 29
	vkStructDescriptorPoolCreateInfo      = 33
	vkStructDescriptorSetAllocateInfo     = 34
	vkStructWriteDescriptorSet            = 35
	vkStructCommandPoolCreateInfo         = 39
	vkStructCommandBufferAllocateInfo     = 40
	vkStructCommandBufferBeginInfo        = 42
	vkStructSubmitInfo                    = 4

	vkDescriptorTypeStorageBuffer = 7
	vkShaderStageCompute          = 0x20
	vkPipelineBindPointCompute    = 1
)

const vkComputeShaderGLSL = `#version 450
layout(local_size_x = 1) in;
layout(std430, binding = 0) readonly buffer A { float a[]; };
layout(std430, binding = 1) buffer O { float o[]; };
void main() {
	uint i = gl_GlobalInvocationID.x;
	o[i] = a[i] * 2.0;
}`

func TestVulkanComputeDispatch(t *testing.T) {
	if os.Getenv("POLYRED_VK_PROBE") != "1" {
		t.Skip("set POLYRED_VK_PROBE=1 to run the headless Vulkan compute dispatch")
	}
	glslang, err := exec.LookPath("glslangValidator")
	if err != nil {
		t.Skipf("glslangValidator not found: %v", err)
	}

	// Compile GLSL -> SPIR-V.
	dir := t.TempDir()
	comp := filepath.Join(dir, "k.comp")
	spv := filepath.Join(dir, "k.spv")
	if err := os.WriteFile(comp, []byte(vkComputeShaderGLSL), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(glslang, "-V", "--target-env", "vulkan1.0", comp, "-o", spv).CombinedOutput(); err != nil {
		t.Fatalf("glslang failed: %v\n%s", err, out)
	}
	spvBytes, err := os.ReadFile(spv)
	if err != nil {
		t.Fatal(err)
	}
	words := make([]uint32, len(spvBytes)/4)
	for i := range words {
		words[i] = binary.LittleEndian.Uint32(spvBytes[i*4:])
	}

	lib, err := purego.Dlopen("libvulkan.so.1", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Skipf("libvulkan.so.1 not available: %v", err)
	}
	sym := func(name string) uintptr {
		p, e := purego.Dlsym(lib, name)
		if e != nil {
			t.Fatalf("dlsym %s: %v", name, e)
		}
		return p
	}
	fn := map[string]uintptr{}
	for _, name := range []string{
		"vkCreateInstance", "vkEnumeratePhysicalDevices", "vkGetPhysicalDeviceQueueFamilyProperties",
		"vkCreateDevice", "vkGetDeviceQueue", "vkGetPhysicalDeviceMemoryProperties",
		"vkCreateBuffer", "vkGetBufferMemoryRequirements", "vkAllocateMemory", "vkBindBufferMemory",
		"vkMapMemory", "vkCreateShaderModule", "vkCreateDescriptorSetLayout", "vkCreatePipelineLayout",
		"vkCreateComputePipelines", "vkCreateDescriptorPool", "vkAllocateDescriptorSets",
		"vkUpdateDescriptorSets", "vkCreateCommandPool", "vkAllocateCommandBuffers",
		"vkBeginCommandBuffer", "vkCmdBindPipeline", "vkCmdBindDescriptorSets", "vkCmdDispatch",
		"vkEndCommandBuffer", "vkQueueSubmit", "vkQueueWaitIdle",
	} {
		fn[name] = sym(name)
	}
	call := func(name string, args ...uintptr) {
		r, _, _ := purego.SyscallN(fn[name], args...)
		if int32(r) != 0 {
			t.Fatalf("%s failed: VkResult=%d", name, int32(r))
		}
	}

	// Instance + device + compute queue (validated by the device test).
	ci := vkInstanceCreateInfo{sType: vkStructureTypeInstanceCreateInfo}
	var instance uintptr
	call("vkCreateInstance", uintptr(unsafe.Pointer(&ci)), 0, uintptr(unsafe.Pointer(&instance)))
	var nd uint32
	purego.SyscallN(fn["vkEnumeratePhysicalDevices"], instance, uintptr(unsafe.Pointer(&nd)), 0)
	devs := make([]uintptr, nd)
	purego.SyscallN(fn["vkEnumeratePhysicalDevices"], instance, uintptr(unsafe.Pointer(&nd)), uintptr(unsafe.Pointer(&devs[0])))
	pd := devs[0]
	var qn uint32
	purego.SyscallN(fn["vkGetPhysicalDeviceQueueFamilyProperties"], pd, uintptr(unsafe.Pointer(&qn)), 0)
	qp := make([]byte, int(qn)*24)
	purego.SyscallN(fn["vkGetPhysicalDeviceQueueFamilyProperties"], pd, uintptr(unsafe.Pointer(&qn)), uintptr(unsafe.Pointer(&qp[0])))
	qfi := uint32(0)
	for q := 0; q < int(qn); q++ {
		if *(*uint32)(unsafe.Pointer(&qp[q*24]))&vkQueueComputeBit != 0 {
			qfi = uint32(q)
			break
		}
	}
	prio := float32(1)
	qci := vkDeviceQueueCreateInfo{sType: vkStructDeviceQueueCreateInfo, queueFamilyIndex: qfi, queueCount: 1, pQueuePriorities: uintptr(unsafe.Pointer(&prio))}
	dci := vkDeviceCreateInfo{sType: vkStructDeviceCreateInfo, queueCreateInfoCount: 1, pQueueCreateInfos: uintptr(unsafe.Pointer(&qci))}
	var device uintptr
	call("vkCreateDevice", pd, uintptr(unsafe.Pointer(&dci)), 0, uintptr(unsafe.Pointer(&device)))
	var queue uintptr
	purego.SyscallN(fn["vkGetDeviceQueue"], device, uintptr(qfi), 0, uintptr(unsafe.Pointer(&queue)))

	// Host-visible memory type.
	memProps := make([]byte, 1024)
	purego.SyscallN(fn["vkGetPhysicalDeviceMemoryProperties"], pd, uintptr(unsafe.Pointer(&memProps[0])))
	memTypeCount := *(*uint32)(unsafe.Pointer(&memProps[0]))
	hostType := func(bits uint32) uint32 {
		for i := 0; i < int(memTypeCount); i++ {
			if bits&(1<<uint(i)) == 0 {
				continue
			}
			f := *(*uint32)(unsafe.Pointer(&memProps[4+i*8]))
			if f&vkMemoryHostVisible != 0 && f&vkMemoryHostCoherent != 0 {
				return uint32(i)
			}
		}
		t.Fatal("no host-visible coherent memory type")
		return 0
	}

	const count = 256
	const size = count * 4
	type buf struct {
		buffer, memory uintptr
		ptr            unsafe.Pointer
	}
	mkBuf := func() buf {
		bci := vkBufferCreateInfo{sType: vkStructBufferCreateInfo, size: size, usage: vkBufferUsageStorageBuffer}
		var b uintptr
		call("vkCreateBuffer", device, uintptr(unsafe.Pointer(&bci)), 0, uintptr(unsafe.Pointer(&b)))
		var req vkMemoryRequirements
		purego.SyscallN(fn["vkGetBufferMemoryRequirements"], device, b, uintptr(unsafe.Pointer(&req)))
		mai := vkMemoryAllocateInfo{sType: vkStructMemoryAllocateInfo, allocationSize: req.size, memoryTypeIndex: hostType(req.memoryTypeBits)}
		var mem uintptr
		call("vkAllocateMemory", device, uintptr(unsafe.Pointer(&mai)), 0, uintptr(unsafe.Pointer(&mem)))
		call("vkBindBufferMemory", device, b, mem, 0)
		var p uintptr
		call("vkMapMemory", device, mem, 0, uintptr(size), 0, uintptr(unsafe.Pointer(&p)))
		return buf{b, mem, unsafe.Pointer(p)}
	}
	a := mkBuf()
	o := mkBuf()
	in := unsafe.Slice((*float32)(a.ptr), count)
	for i := range in {
		in[i] = float32(i)
	}

	// Shader module.
	smci := vkShaderModuleCreateInfo{sType: vkStructShaderModuleCreateInfo, codeSize: uint64(len(spvBytes)), pCode: uintptr(unsafe.Pointer(&words[0]))}
	var module uintptr
	call("vkCreateShaderModule", device, uintptr(unsafe.Pointer(&smci)), 0, uintptr(unsafe.Pointer(&module)))

	// Descriptor set layout: two storage buffers.
	bindings := []vkDescriptorSetLayoutBinding{
		{binding: 0, descriptorType: vkDescriptorTypeStorageBuffer, descriptorCount: 1, stageFlags: vkShaderStageCompute},
		{binding: 1, descriptorType: vkDescriptorTypeStorageBuffer, descriptorCount: 1, stageFlags: vkShaderStageCompute},
	}
	dslci := vkDescriptorSetLayoutCreateInfo{sType: vkStructDescriptorSetLayoutCreateInfo, bindingCount: 2, pBindings: uintptr(unsafe.Pointer(&bindings[0]))}
	var dsl uintptr
	call("vkCreateDescriptorSetLayout", device, uintptr(unsafe.Pointer(&dslci)), 0, uintptr(unsafe.Pointer(&dsl)))

	plci := vkPipelineLayoutCreateInfo{sType: vkStructPipelineLayoutCreateInfo, setLayoutCount: 1, pSetLayouts: uintptr(unsafe.Pointer(&dsl))}
	var pipeLayout uintptr
	call("vkCreatePipelineLayout", device, uintptr(unsafe.Pointer(&plci)), 0, uintptr(unsafe.Pointer(&pipeLayout)))

	entry := []byte("main\x00")
	cpci := vkComputePipelineCreateInfo{
		sType: vkStructComputePipelineCreateInfo,
		stage: vkPipelineShaderStageCreateInfo{
			sType: vkStructPipelineShaderStageCreateInfo, stage: vkShaderStageCompute,
			module: module, pName: uintptr(unsafe.Pointer(&entry[0])),
		},
		layout: pipeLayout,
	}
	var pipeline uintptr
	call("vkCreateComputePipelines", device, 0, 1, uintptr(unsafe.Pointer(&cpci)), 0, uintptr(unsafe.Pointer(&pipeline)))

	// Descriptor pool + set, pointing the two bindings at the buffers.
	poolSize := vkDescriptorPoolSize{typ: vkDescriptorTypeStorageBuffer, descriptorCount: 2}
	dpci := vkDescriptorPoolCreateInfo{sType: vkStructDescriptorPoolCreateInfo, maxSets: 1, poolSizeCount: 1, pPoolSizes: uintptr(unsafe.Pointer(&poolSize))}
	var pool uintptr
	call("vkCreateDescriptorPool", device, uintptr(unsafe.Pointer(&dpci)), 0, uintptr(unsafe.Pointer(&pool)))
	dsai := vkDescriptorSetAllocateInfo{sType: vkStructDescriptorSetAllocateInfo, descriptorPool: pool, descriptorSetCount: 1, pSetLayouts: uintptr(unsafe.Pointer(&dsl))}
	var set uintptr
	call("vkAllocateDescriptorSets", device, uintptr(unsafe.Pointer(&dsai)), uintptr(unsafe.Pointer(&set)))

	biA := vkDescriptorBufferInfo{buffer: a.buffer, rng: size}
	biO := vkDescriptorBufferInfo{buffer: o.buffer, rng: size}
	writes := []vkWriteDescriptorSet{
		{sType: vkStructWriteDescriptorSet, dstSet: set, dstBinding: 0, descriptorCount: 1, descriptorType: vkDescriptorTypeStorageBuffer, pBufferInfo: uintptr(unsafe.Pointer(&biA))},
		{sType: vkStructWriteDescriptorSet, dstSet: set, dstBinding: 1, descriptorCount: 1, descriptorType: vkDescriptorTypeStorageBuffer, pBufferInfo: uintptr(unsafe.Pointer(&biO))},
	}
	purego.SyscallN(fn["vkUpdateDescriptorSets"], device, 2, uintptr(unsafe.Pointer(&writes[0])), 0, 0)

	// Command buffer: bind pipeline + set, dispatch.
	cpoolci := vkCommandPoolCreateInfo{sType: vkStructCommandPoolCreateInfo, queueFamilyIndex: qfi}
	var cmdPool uintptr
	call("vkCreateCommandPool", device, uintptr(unsafe.Pointer(&cpoolci)), 0, uintptr(unsafe.Pointer(&cmdPool)))
	cbai := vkCommandBufferAllocateInfo{sType: vkStructCommandBufferAllocateInfo, commandPool: cmdPool, level: 0, commandBufferCount: 1}
	var cmd uintptr
	call("vkAllocateCommandBuffers", device, uintptr(unsafe.Pointer(&cbai)), uintptr(unsafe.Pointer(&cmd)))
	begin := vkCommandBufferBeginInfo{sType: vkStructCommandBufferBeginInfo}
	call("vkBeginCommandBuffer", cmd, uintptr(unsafe.Pointer(&begin)))
	purego.SyscallN(fn["vkCmdBindPipeline"], cmd, vkPipelineBindPointCompute, pipeline)
	purego.SyscallN(fn["vkCmdBindDescriptorSets"], cmd, vkPipelineBindPointCompute, pipeLayout, 0, 1, uintptr(unsafe.Pointer(&set)), 0, 0)
	purego.SyscallN(fn["vkCmdDispatch"], cmd, count, 1, 1)
	call("vkEndCommandBuffer", cmd)

	si := vkSubmitInfo{sType: vkStructSubmitInfo, commandBufferCount: 1, pCommandBuffers: uintptr(unsafe.Pointer(&cmd))}
	call("vkQueueSubmit", queue, 1, uintptr(unsafe.Pointer(&si)), 0)
	call("vkQueueWaitIdle", queue)

	// Verify out = in * 2.
	out := unsafe.Slice((*float32)(o.ptr), count)
	for i := range out {
		if out[i] != float32(i)*2 {
			t.Fatalf("Vulkan compute out[%d] = %v, want %v", i, out[i], float32(i)*2)
		}
	}
	t.Logf("cgo-free Vulkan compute dispatch: %d/%d elements doubled correctly", count, count)
}
