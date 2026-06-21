// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

// Second Vulkan step (after the instance/device probe): a cgo-free logical
// device + host-visible storage buffer memory roundtrip. It creates a device on
// a compute queue family, allocates a HOST_VISIBLE|HOST_COHERENT storage buffer,
// maps it, writes a pattern, and reads it back. This is the device + memory
// foundation a Vulkan compute backend (item #2) sits on; the remaining piece is
// descriptor sets + a compute pipeline + dispatch. Verified in CI on Mesa
// lavapipe. Gated on POLYRED_VK_PROBE=1.
package gpu_test

import (
	"os"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

type vkDeviceQueueCreateInfo struct {
	sType            uint32
	pNext            uintptr
	flags            uint32
	queueFamilyIndex uint32
	queueCount       uint32
	pQueuePriorities uintptr
}

type vkDeviceCreateInfo struct {
	sType                   uint32
	pNext                   uintptr
	flags                   uint32
	queueCreateInfoCount    uint32
	pQueueCreateInfos       uintptr
	enabledLayerCount       uint32
	ppEnabledLayerNames     uintptr
	enabledExtensionCount   uint32
	ppEnabledExtensionNames uintptr
	pEnabledFeatures        uintptr
}

type vkBufferCreateInfo struct {
	sType                 uint32
	pNext                 uintptr
	flags                 uint32
	size                  uint64
	usage                 uint32
	sharingMode           uint32
	queueFamilyIndexCount uint32
	pQueueFamilyIndices   uintptr
}

type vkMemoryRequirements struct {
	size           uint64
	alignment      uint64
	memoryTypeBits uint32
}

type vkMemoryAllocateInfo struct {
	sType           uint32
	pNext           uintptr
	allocationSize  uint64
	memoryTypeIndex uint32
}

const (
	vkStructDeviceQueueCreateInfo = 2
	vkStructDeviceCreateInfo      = 3
	vkStructMemoryAllocateInfo    = 5
	vkStructBufferCreateInfo      = 12

	vkBufferUsageStorageBuffer = 0x20
	vkMemoryHostVisible        = 0x2
	vkMemoryHostCoherent       = 0x4
)

func TestVulkanDeviceMemory(t *testing.T) {
	if os.Getenv("POLYRED_VK_PROBE") != "1" {
		t.Skip("set POLYRED_VK_PROBE=1 to run the headless Vulkan device test")
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
	var (
		vkCreateInstance      = sym("vkCreateInstance")
		vkEnumPhysDevices     = sym("vkEnumeratePhysicalDevices")
		vkGetQueueFamilyProps = sym("vkGetPhysicalDeviceQueueFamilyProperties")
		vkCreateDevice        = sym("vkCreateDevice")
		vkGetMemProps         = sym("vkGetPhysicalDeviceMemoryProperties")
		vkCreateBuffer        = sym("vkCreateBuffer")
		vkGetBufferMemReqs    = sym("vkGetBufferMemoryRequirements")
		vkAllocateMemory      = sym("vkAllocateMemory")
		vkBindBufferMemory    = sym("vkBindBufferMemory")
		vkMapMemory           = sym("vkMapMemory")
		vkUnmapMemory         = sym("vkUnmapMemory")
	)
	check := func(name string, r uintptr) {
		if int32(r) != 0 {
			t.Fatalf("%s failed: VkResult=%d", name, int32(r))
		}
	}

	// Instance.
	ci := vkInstanceCreateInfo{sType: vkStructureTypeInstanceCreateInfo}
	var instance uintptr
	r, _, _ := purego.SyscallN(vkCreateInstance, uintptr(unsafe.Pointer(&ci)), 0, uintptr(unsafe.Pointer(&instance)))
	check("vkCreateInstance", r)

	// Physical device 0.
	var n uint32
	purego.SyscallN(vkEnumPhysDevices, instance, uintptr(unsafe.Pointer(&n)), 0)
	if n == 0 {
		t.Fatal("no physical devices")
	}
	devs := make([]uintptr, n)
	purego.SyscallN(vkEnumPhysDevices, instance, uintptr(unsafe.Pointer(&n)), uintptr(unsafe.Pointer(&devs[0])))
	pd := devs[0]

	// Compute queue family index.
	var qn uint32
	purego.SyscallN(vkGetQueueFamilyProps, pd, uintptr(unsafe.Pointer(&qn)), 0)
	qprops := make([]byte, int(qn)*24)
	purego.SyscallN(vkGetQueueFamilyProps, pd, uintptr(unsafe.Pointer(&qn)), uintptr(unsafe.Pointer(&qprops[0])))
	qfi := -1
	for q := 0; q < int(qn); q++ {
		if *(*uint32)(unsafe.Pointer(&qprops[q*24]))&vkQueueComputeBit != 0 {
			qfi = q
			break
		}
	}
	if qfi < 0 {
		t.Fatal("no compute queue family")
	}

	// Logical device with one compute queue.
	prio := float32(1)
	qci := vkDeviceQueueCreateInfo{
		sType:            vkStructDeviceQueueCreateInfo,
		queueFamilyIndex: uint32(qfi),
		queueCount:       1,
		pQueuePriorities: uintptr(unsafe.Pointer(&prio)),
	}
	dci := vkDeviceCreateInfo{
		sType:                vkStructDeviceCreateInfo,
		queueCreateInfoCount: 1,
		pQueueCreateInfos:    uintptr(unsafe.Pointer(&qci)),
	}
	var device uintptr
	r, _, _ = purego.SyscallN(vkCreateDevice, pd, uintptr(unsafe.Pointer(&dci)), 0, uintptr(unsafe.Pointer(&device)))
	check("vkCreateDevice", r)

	// Storage buffer.
	const count = 256
	const size = count * 4
	bci := vkBufferCreateInfo{
		sType: vkStructBufferCreateInfo,
		size:  size,
		usage: vkBufferUsageStorageBuffer,
	}
	var buffer uintptr
	r, _, _ = purego.SyscallN(vkCreateBuffer, device, uintptr(unsafe.Pointer(&bci)), 0, uintptr(unsafe.Pointer(&buffer)))
	check("vkCreateBuffer", r)

	var req vkMemoryRequirements
	purego.SyscallN(vkGetBufferMemReqs, device, buffer, uintptr(unsafe.Pointer(&req)))

	// Find a HOST_VISIBLE|HOST_COHERENT memory type allowed by the buffer.
	memProps := make([]byte, 1024)
	purego.SyscallN(vkGetMemProps, pd, uintptr(unsafe.Pointer(&memProps[0])))
	typeCount := *(*uint32)(unsafe.Pointer(&memProps[0]))
	memTypeIdx := -1
	for i := 0; i < int(typeCount); i++ {
		if req.memoryTypeBits&(1<<uint(i)) == 0 {
			continue
		}
		flags := *(*uint32)(unsafe.Pointer(&memProps[4+i*8]))
		if flags&vkMemoryHostVisible != 0 && flags&vkMemoryHostCoherent != 0 {
			memTypeIdx = i
			break
		}
	}
	if memTypeIdx < 0 {
		t.Fatal("no host-visible coherent memory type")
	}

	mai := vkMemoryAllocateInfo{
		sType:           vkStructMemoryAllocateInfo,
		allocationSize:  req.size,
		memoryTypeIndex: uint32(memTypeIdx),
	}
	var memory uintptr
	r, _, _ = purego.SyscallN(vkAllocateMemory, device, uintptr(unsafe.Pointer(&mai)), 0, uintptr(unsafe.Pointer(&memory)))
	check("vkAllocateMemory", r)
	r, _, _ = purego.SyscallN(vkBindBufferMemory, device, buffer, memory, 0)
	check("vkBindBufferMemory", r)

	// Map, write a pattern, read it back.
	var ptr uintptr
	r, _, _ = purego.SyscallN(vkMapMemory, device, memory, 0, uintptr(size), 0, uintptr(unsafe.Pointer(&ptr)))
	check("vkMapMemory", r)
	mapped := unsafe.Slice((*float32)(unsafe.Pointer(ptr)), count)
	for i := range mapped {
		mapped[i] = float32(i) * 1.5
	}
	for i := range mapped {
		if mapped[i] != float32(i)*1.5 {
			t.Fatalf("host-visible memory roundtrip mismatch at %d: %v", i, mapped[i])
		}
	}
	purego.SyscallN(vkUnmapMemory, device, memory)
	t.Logf("cgo-free Vulkan device + host-visible storage buffer (%d floats) roundtrip OK", count)
}
