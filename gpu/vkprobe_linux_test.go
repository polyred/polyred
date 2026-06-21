// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

// Bounded probe proving the cgo-free Vulkan path: it loads libvulkan via purego,
// creates an instance, enumerates physical devices, and confirms a
// compute-capable queue family exists. On the CI runner this is satisfied by
// Mesa's lavapipe software Vulkan ICD, so a future cgo-free Vulkan compute
// backend (item #2, "then Vulkan") is CI-verifiable without GPU hardware, the
// same way the GL backend is. It does not implement a backend; it de-risks the
// next step. Gated on POLYRED_VK_PROBE=1 so it only runs in the vk-probe job.
package gpu_test

import (
	"os"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

// vkInstanceCreateInfo mirrors the C struct layout (Go's natural alignment of a
// uint32 followed by a pointer inserts the same 4-byte padding as C on 64-bit).
type vkInstanceCreateInfo struct {
	sType                   uint32
	pNext                   uintptr
	flags                   uint32
	pApplicationInfo        uintptr
	enabledLayerCount       uint32
	ppEnabledLayerNames     uintptr
	enabledExtensionCount   uint32
	ppEnabledExtensionNames uintptr
}

const (
	vkStructureTypeInstanceCreateInfo = 1
	vkQueueComputeBit                 = 0x2
)

func TestVulkanComputeProbe(t *testing.T) {
	if os.Getenv("POLYRED_VK_PROBE") != "1" {
		t.Skip("set POLYRED_VK_PROBE=1 to run the headless Vulkan probe")
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
	vkCreateInstance := sym("vkCreateInstance")
	vkEnumeratePhysicalDevices := sym("vkEnumeratePhysicalDevices")
	vkGetPhysicalDeviceProperties := sym("vkGetPhysicalDeviceProperties")
	vkGetQueueFamilyProps := sym("vkGetPhysicalDeviceQueueFamilyProperties")

	ci := vkInstanceCreateInfo{sType: vkStructureTypeInstanceCreateInfo}
	var instance uintptr
	if r, _, _ := purego.SyscallN(vkCreateInstance, uintptr(unsafe.Pointer(&ci)), 0, uintptr(unsafe.Pointer(&instance))); int32(r) != 0 {
		t.Fatalf("vkCreateInstance failed: VkResult=%d", int32(r))
	}

	var count uint32
	purego.SyscallN(vkEnumeratePhysicalDevices, instance, uintptr(unsafe.Pointer(&count)), 0)
	if count == 0 {
		t.Fatal("no Vulkan physical devices")
	}
	devices := make([]uintptr, count)
	purego.SyscallN(vkEnumeratePhysicalDevices, instance, uintptr(unsafe.Pointer(&count)), uintptr(unsafe.Pointer(&devices[0])))

	foundCompute := false
	for _, pd := range devices {
		// VkPhysicalDeviceProperties: deviceType (uint32) at offset 16, then
		// deviceName[256] at offset 20. Over-allocate; we only read the head.
		props := make([]byte, 1024)
		purego.SyscallN(vkGetPhysicalDeviceProperties, pd, uintptr(unsafe.Pointer(&props[0])))
		deviceType := *(*uint32)(unsafe.Pointer(&props[16]))
		t.Logf("Vulkan device: %q (type=%d)", goCStr(props[20:]), deviceType)

		var qn uint32
		purego.SyscallN(vkGetQueueFamilyProps, pd, uintptr(unsafe.Pointer(&qn)), 0)
		if qn == 0 {
			continue
		}
		// VkQueueFamilyProperties is 24 bytes; queueFlags (uint32) is at offset 0.
		qprops := make([]byte, int(qn)*24)
		purego.SyscallN(vkGetQueueFamilyProps, pd, uintptr(unsafe.Pointer(&qn)), uintptr(unsafe.Pointer(&qprops[0])))
		for q := 0; q < int(qn); q++ {
			flags := *(*uint32)(unsafe.Pointer(&qprops[q*24]))
			if flags&vkQueueComputeBit != 0 {
				foundCompute = true
			}
		}
	}
	if !foundCompute {
		t.Fatal("no Vulkan compute-capable queue family found")
	}
	t.Log("cgo-free Vulkan instance + compute-capable device confirmed")
}

func goCStr(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
