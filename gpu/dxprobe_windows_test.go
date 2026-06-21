// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

// Bounded probe proving the cgo-free DirectX 12 path: it creates a D3D12 device
// via d3d12.dll through the syscall package (cgo-free, native on Windows). On the
// GitHub windows-latest runner (no GPU) this is served by WARP / the Microsoft
// Basic Render Driver software rasterizer, so a future cgo-free DX12 compute
// backend (item #2, "and DX12") is CI-verifiable in software, the same way GL
// (Mesa llvmpipe) and Vulkan (lavapipe) are. It does not implement a backend; it
// de-risks the next step. Gated on POLYRED_DX_PROBE=1.
package gpu

import (
	"os"
	"syscall"
	"testing"
	"unsafe"
)

type dxGUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// IID_ID3D12Device.
var iidID3D12Device = dxGUID{0x189819f1, 0x1db6, 0x4b57, [8]byte{0xbe, 0x54, 0x18, 0x21, 0x33, 0x9b, 0x85, 0xf7}}

// IID_IDXGIFactory4 and IID_IDXGIAdapter, for the WARP fallback.
var iidIDXGIFactory4 = dxGUID{0x1bc6ea02, 0xef36, 0x464f, [8]byte{0xbf, 0x0c, 0x21, 0xca, 0x39, 0xe5, 0x16, 0x8a}}
var iidIDXGIAdapter = dxGUID{0x2411e7e1, 0x12ac, 0x4ccf, [8]byte{0x8d, 0x2c, 0x59, 0x76, 0x9e, 0x84, 0xd2, 0xc7}}

// IID_ID3D12Resource.
var iidID3D12Resource = dxGUID{0x696442be, 0xa72e, 0x4059, [8]byte{0xbc, 0x79, 0x5b, 0x5c, 0x98, 0x04, 0x0f, 0xad}}

const dxFeatureLevel11_0 = 0xb000

// comCall invokes COM method `index` on `obj` (obj->vtbl[index](obj, args...)).
func comCall(obj uintptr, index int, args ...uintptr) uintptr {
	vtbl := *(*uintptr)(unsafe.Pointer(obj))
	method := *(*uintptr)(unsafe.Pointer(vtbl + uintptr(index)*unsafe.Sizeof(uintptr(0))))
	r, _, _ := syscall.SyscallN(method, append([]uintptr{obj}, args...)...)
	return r
}

func dx12CreateDevice(t *testing.T) uintptr {
	t.Helper()
	d3d12 := syscall.NewLazyDLL("d3d12.dll")
	if err := d3d12.Load(); err != nil {
		t.Skipf("d3d12.dll not available: %v", err)
	}
	var device uintptr
	hr, _, _ := d3d12.NewProc("D3D12CreateDevice").Call(0, dxFeatureLevel11_0, uintptr(unsafe.Pointer(&iidID3D12Device)), uintptr(unsafe.Pointer(&device)))
	if int32(hr) != 0 || device == 0 {
		t.Skipf("no D3D12 device (HRESULT=0x%x)", uint32(hr))
	}
	return device
}

type d3d12HeapProperties struct {
	Type                 uint32
	CPUPageProperty      uint32
	MemoryPoolPreference uint32
	CreationNodeMask     uint32
	VisibleNodeMask      uint32
}

type d3d12ResourceDesc struct {
	Dimension         uint32
	Alignment         uint64
	Width             uint64
	Height            uint32
	DepthOrArraySize  uint16
	MipLevels         uint16
	Format            uint32
	SampleDescCount   uint32
	SampleDescQuality uint32
	Layout            uint32
	Flags             uint32
}

// TestDX12ResourceMap validates the cgo-free COM calling convention against the
// software D3D12 device: create an UPLOAD-heap buffer resource via the device's
// CreateCommittedResource vtable method, map it, and roundtrip a pattern. This is
// the device + memory foundation a DX12 compute backend sits on; the rest is the
// command queue/list + compute PSO + dispatch.
func TestDX12ResourceMap(t *testing.T) {
	if os.Getenv("POLYRED_DX_PROBE") != "1" {
		t.Skip("set POLYRED_DX_PROBE=1 to run the headless D3D12 resource test")
	}
	device := dx12CreateDevice(t)

	const count = 256
	const size = count * 4
	// UPLOAD heap (CPU-writable, GPU-readable); required initial state is
	// GENERIC_READ (0xAC3).
	hp := d3d12HeapProperties{Type: 2 /*UPLOAD*/}
	desc := d3d12ResourceDesc{
		Dimension: 1 /*BUFFER*/, Width: size, Height: 1, DepthOrArraySize: 1,
		MipLevels: 1, Format: 0, SampleDescCount: 1, Layout: 1, /*ROW_MAJOR*/
	}
	var res uintptr
	// ID3D12Device::CreateCommittedResource is vtable index 25.
	hr := comCall(device, 25,
		uintptr(unsafe.Pointer(&hp)), 0, uintptr(unsafe.Pointer(&desc)),
		0xAC3 /*GENERIC_READ*/, 0,
		uintptr(unsafe.Pointer(&iidID3D12Resource)), uintptr(unsafe.Pointer(&res)))
	if int32(hr) != 0 || res == 0 {
		t.Fatalf("CreateCommittedResource failed: HRESULT=0x%x", uint32(hr))
	}

	// ID3D12Resource::Map is vtable index 8.
	var ptr uintptr
	hr = comCall(res, 8, 0 /*subresource*/, 0 /*pReadRange=nil*/, uintptr(unsafe.Pointer(&ptr)))
	if int32(hr) != 0 || ptr == 0 {
		t.Fatalf("ID3D12Resource::Map failed: HRESULT=0x%x", uint32(hr))
	}
	mapped := unsafe.Slice((*float32)(unsafe.Pointer(ptr)), count)
	for i := range mapped {
		mapped[i] = float32(i) * 1.5
	}
	for i := range mapped {
		if mapped[i] != float32(i)*1.5 {
			t.Fatalf("D3D12 UPLOAD buffer roundtrip mismatch at %d: %v", i, mapped[i])
		}
	}
	comCall(res, 9 /*Unmap*/, 0, 0)
	t.Logf("cgo-free D3D12 device + UPLOAD buffer (%d floats) COM roundtrip OK", count)
}

func TestDX12DeviceProbe(t *testing.T) {
	if os.Getenv("POLYRED_DX_PROBE") != "1" {
		t.Skip("set POLYRED_DX_PROBE=1 to run the headless D3D12 (WARP) probe")
	}
	d3d12 := syscall.NewLazyDLL("d3d12.dll")
	if err := d3d12.Load(); err != nil {
		t.Skipf("d3d12.dll not available: %v", err)
	}
	createDevice := d3d12.NewProc("D3D12CreateDevice")

	// First try the default adapter (nil): on a runner with the Basic Render
	// Driver this gives a software D3D12 device directly.
	var device uintptr
	hr, _, _ := createDevice.Call(0, dxFeatureLevel11_0, uintptr(unsafe.Pointer(&iidID3D12Device)), uintptr(unsafe.Pointer(&device)))
	if int32(hr) == 0 && device != 0 {
		t.Logf("D3D12 device created on the default adapter (HRESULT=0x%x)", uint32(hr))
		return
	}
	t.Logf("default-adapter D3D12CreateDevice HRESULT=0x%x; trying WARP via DXGI", uint32(hr))

	// Fallback: enumerate the WARP adapter through DXGI and create a device on it.
	dxgi := syscall.NewLazyDLL("dxgi.dll")
	if err := dxgi.Load(); err != nil {
		t.Fatalf("dxgi.dll not available: %v", err)
	}
	createFactory := dxgi.NewProc("CreateDXGIFactory1")
	var factory uintptr
	hr, _, _ = createFactory.Call(uintptr(unsafe.Pointer(&iidIDXGIFactory4)), uintptr(unsafe.Pointer(&factory)))
	if int32(hr) != 0 || factory == 0 {
		t.Fatalf("CreateDXGIFactory1 failed: HRESULT=0x%x", uint32(hr))
	}
	// IDXGIFactory4::EnumWarpAdapter is vtable slot 28 (IUnknown 0-2, IDXGIObject
	// 3-6, IDXGIFactory 7-12, IDXGIFactory1 13-14, IDXGIFactory2 15-24,
	// IDXGIFactory3 25, IDXGIFactory4 26-27 ... EnumWarpAdapter is the second
	// IDXGIFactory4 method): index resolved from the vtable below.
	vtbl := *(*uintptr)(unsafe.Pointer(factory))
	enumWarp := *(*uintptr)(unsafe.Pointer(vtbl + 28*unsafe.Sizeof(uintptr(0))))
	var warp uintptr
	hr, _, _ = syscall.SyscallN(enumWarp, factory, uintptr(unsafe.Pointer(&iidIDXGIAdapter)), uintptr(unsafe.Pointer(&warp)))
	if int32(hr) != 0 || warp == 0 {
		t.Fatalf("EnumWarpAdapter failed: HRESULT=0x%x", uint32(hr))
	}
	hr, _, _ = createDevice.Call(warp, dxFeatureLevel11_0, uintptr(unsafe.Pointer(&iidID3D12Device)), uintptr(unsafe.Pointer(&device)))
	if int32(hr) != 0 || device == 0 {
		t.Fatalf("D3D12CreateDevice on WARP failed: HRESULT=0x%x", uint32(hr))
	}
	t.Logf("D3D12 device created on the WARP software adapter (HRESULT=0x%x)", uint32(hr))
}
