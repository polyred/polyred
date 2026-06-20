# GPU Abstraction — Design

Status: draft / for discussion
Author: The Polyred Authors
Last updated: 2026-06-20

## 1. Goal and scope

A dedicated Go abstraction that lets Go code run **compute** and **rendering**
pipelines on the GPU, with **swappable drivers** underneath (Metal, Vulkan,
DirectX 12, OpenGL/GLES). The renderer in `polyred` should be able to target it
without knowing which driver is live.

This document is a design, not an implementation. It exists to pin down the one
genuinely hard decision (the backend shape clash, §3), sketch the core types
(§4), prove them against one compute and one render slice (§5), settle the
cgo-free question (§6), and describe how the abstraction folds into `polyred`
and is consumed by the renderer (§7), followed by a phased roadmap (§8).

### Current reality (what already exists)

- `poly.red/x/gpu` (the sibling `../gpu` repo, never pushed to GitHub) holds raw
  driver bindings: **Metal** (`mtl/`, ~590 lines, cgo) and **OpenGL/GLES**
  (`gl/`, ~2400 lines, cgo) plus EGL/X11/CGL context creation (`ctx/`).
- `device.go` — the unifying abstraction — is **empty** (a package doc comment).
  `vk/` and `dx12/` are **empty stubs**. So the keystone was never laid.
- `internal/dl/` is a nascent **cgo-free** loader (purego-style `dlopen`/`dlsym`
  via assembly + `//go:cgo_import_dynamic`). It is the broken piece today
  (truncated `sys_darwin_arm64.s`).
- The only working demo is matrix `Add/Sub/Sqrt/Mul`, which talks to each
  backend **directly** — there is no API in between.
- `polyred`'s in-repo `gpu/` is a copy of that same toy.

Conclusion: the hard, low-value plumbing (driver bindings) is ~40% done; the
high-value part (a `Device` API the renderer programs against) is 0% done.

### Prior art we should not reinvent

- **WebGPU** is effectively the spec for "compute + render, swappable drivers."
  Its model (Adapter → Device → Queue; Buffer/Texture/ShaderModule;
  Render/ComputePipeline; CommandEncoder → pass encoder → CommandBuffer →
  Submit) is what we adopt conceptually. Our existing Metal backend already maps
  ~1:1 onto it.
- **gio** (`gioui.org`) already abstracts a `Device` over GL/D3D11/Metal/Vulkan
  — and our GL/EGL code is literally "Modified from gioui/gio." Its `driver`
  package is the closest structural template; read it before finalizing types.
- **Ebitengine** (`graphicsdriver`) solves the same multi-backend-in-Go problem
  and has done the cgo-free `purego` work. Relevant to both the API and §6.

We mirror WebGPU's conceptual model and simplify (defer bind-group *layouts*,
defer the adapter-feature negotiation). We do **not** bind `wgpu-native`: it is a
C library and would defeat the cgo-free goal (§6).

## 2. Non-goals

- A general-purpose, spec-complete WebGPU implementation. We implement the
  subset polyred needs.
- Replacing the CPU rasterizer. It stays as the reference / fallback path; the
  GPU path is additive.
- Window/surface management beyond what `ctx/` already provides.

## 3. The core decision: which backend shape does the abstraction mirror?

This drives every type below, so it is section 1, not a footnote.

The backends have **clashing shapes**:

| | Metal / Vulkan / DX12 / WebGPU | OpenGL / GLES |
|---|---|---|
| Model | explicit: queue + command buffers + pipeline-state objects | stateful global context |
| Threading | command buffers buildable off-thread | context must be **current on one OS thread** |
| Our evidence | `mtl`: `Device.MakeCommandQueue → MakeCommandBuffer → MakeComputeCommandEncoder → SetComputePipelineState → DispatchThreads → Commit` | `gl` demo: `runtime.LockOSThread()` + `ctx.MakeCurrent()`, then `UseProgram` / `DispatchCompute` / `MemoryBarrier` / `Finish` against a global `*Functions` |

**Decision: mirror the explicit command-buffer model (WebGPU/Metal).** The three
modern backends map onto it directly. The GL backend **emulates** it: it owns a
dedicated goroutine pinned with `runtime.LockOSThread`, keeps the EGL/CGL context
current there, and serializes recorded commands onto it. A `CommandBuffer`
recorded by the abstraction becomes, for GL, a queued list of calls replayed on
that goroutine; `Queue.Submit` is a channel send + completion wait.

Mirroring GL instead (stateful, implicit) would cripple Metal/Vulkan/DX12 and
throw away their multi-threaded command encoding. So GL pays the emulation cost,
not the other three.

Implication: the GL backend gets a small internal "context thread" runtime
(record → enqueue → replay-on-locked-thread). This is the main new code the
abstraction adds beyond what bindings already provide.

## 4. Core types (Go sketches, anchored on WebGPU)

These are illustrative signatures to react to, not final. Package `gpu`.

```go
// Driver selects a backend. The abstraction picks the best available by
// default; callers may force one (tests, debugging, headless CI).
type Driver int

const (
	DriverAuto Driver = iota
	DriverMetal
	DriverVulkan
	DriverD3D12
	DriverGL
)

// Open negotiates an adapter and returns a Device for the chosen driver.
// Equivalent to WebGPU requestAdapter + requestDevice.
func Open(opts ...Option) (*Device, error)

// Device is the root object. It owns the queue and is the factory for all
// GPU resources. Maps to mtl.Device / VkDevice / ID3D12Device / a GL context.
type Device struct{ /* backend-private */ }

func (d *Device) Driver() Driver
func (d *Device) Queue() *Queue
func (d *Device) NewBuffer(desc BufferDescriptor) (*Buffer, error)
func (d *Device) NewTexture(desc TextureDescriptor) (*Texture, error)
func (d *Device) NewShaderModule(src ShaderSource) (*ShaderModule, error)
func (d *Device) NewComputePipeline(desc ComputePipelineDescriptor) (*ComputePipeline, error)
func (d *Device) NewRenderPipeline(desc RenderPipelineDescriptor) (*RenderPipeline, error)
func (d *Device) NewCommandEncoder() *CommandEncoder
func (d *Device) Close() error

// BufferUsage / TextureUsage are bitflags (WebGPU-style) that let each backend
// pick the right storage mode (e.g. mtl.ResourceStorageModeShared).
type BufferUsage uint32

const (
	BufferCopySrc BufferUsage = 1 << iota
	BufferCopyDst
	BufferStorage // SSBO / Metal buffer / UAV
	BufferUniform
	BufferVertex
	BufferIndex
	BufferMapRead
	BufferMapWrite
)

type BufferDescriptor struct {
	Label string
	Size  int
	Usage BufferUsage
	Data  []byte // optional initial contents
}

type Buffer struct{ /* ... */ }

func (b *Buffer) Size() int
// Map returns a CPU-visible view for readback/upload (shared/managed memory).
func (b *Buffer) Map() ([]byte, error)
func (b *Buffer) Unmap()
func (b *Buffer) Release()

// ShaderSource carries per-backend shader text. The abstraction does NOT
// translate shading languages; callers (or a future transpiler) provide the
// variant for the live driver. Keyed so one module can hold all variants.
type ShaderSource struct {
	MSL    string // Metal Shading Language
	GLSL   string // GLSL compute/vertex/fragment (#version 310 es ...)
	HLSL   string // DX12
	SPIRV  []byte // Vulkan (and DX12 via translation)
	Entry  string // entry point name
	Stage  ShaderStage
}

type ComputePipelineDescriptor struct {
	Label  string
	Module *ShaderModule
	Entry  string
}

// Binding model: start with a flat "bindings = ordered resources" scheme
// (matches the demo's SetBuffer(index) and GL BindBufferBase(index)).
// Bind-group *layouts* are deferred until a backend needs them (Vulkan/DX12).
type Binding struct {
	Index  int
	Buffer *Buffer // exactly one of Buffer/Texture set
	Texture *Texture
}

// CommandEncoder records GPU work. Encodes into pass encoders, then Finish()
// produces a CommandBuffer handed to Queue.Submit. Mirrors WebGPU/Metal; the
// GL backend records these and replays them on its context thread (§3).
type CommandEncoder struct{ /* ... */ }

func (e *CommandEncoder) BeginComputePass() *ComputePass
func (e *CommandEncoder) BeginRenderPass(desc RenderPassDescriptor) *RenderPass
func (e *CommandEncoder) CopyBufferToBuffer(src, dst *Buffer, size int)
func (e *CommandEncoder) Finish() *CommandBuffer

type ComputePass struct{ /* ... */ }

func (p *ComputePass) SetPipeline(cp *ComputePipeline)
func (p *ComputePass) SetBindings(b ...Binding)
func (p *ComputePass) Dispatch(x, y, z int) // workgroup counts
func (p *ComputePass) End()

type RenderPassDescriptor struct {
	ColorAttachments []ColorAttachment // target Texture + load/clear/store
	DepthAttachment  *DepthAttachment
}

type RenderPass struct{ /* ... */ }

func (p *RenderPass) SetPipeline(rp *RenderPipeline)
func (p *RenderPass) SetBindings(b ...Binding)
func (p *RenderPass) SetVertexBuffer(slot int, b *Buffer)
func (p *RenderPass) SetIndexBuffer(b *Buffer, fmt IndexFormat)
func (p *RenderPass) DrawIndexed(indexCount, instanceCount int)
func (p *RenderPass) End()

type Queue struct{ /* ... */ }

func (q *Queue) Submit(cb ...*CommandBuffer)
func (q *Queue) WaitIdle() // blocks until submitted work completes
```

Notes:
- This is the WebGPU object graph with the negotiation machinery trimmed. Every
  method above has a near-literal Metal counterpart in today's `mtl` package.
- `ShaderSource` deliberately punts cross-language shader translation. Polyred's
  shaders are Go funcs today (§7); a later milestone can generate MSL/GLSL from a
  restricted Go subset or a small DSL. Not in scope here.

## 5. Two slices that validate the types

### 5a. Compute slice — port the matrix demo to the `Device` API

The existing `tests/math_darwin.go` does, in raw Metal: make 3 buffers → command
queue → command buffer → compute encoder → set pipeline + buffers → dispatch →
commit → wait → read back. Re-expressed through the abstraction:

```go
dev, _ := gpu.Open()                      // picks Metal on darwin, GL elsewhere
mod, _ := dev.NewShaderModule(addSource)  // addSource.MSL + addSource.GLSL
pipe, _ := dev.NewComputePipeline(gpu.ComputePipelineDescriptor{Module: mod, Entry: "add"})

a := must(dev.NewBuffer(gpu.BufferDescriptor{Size: n*4, Usage: gpu.BufferStorage | gpu.BufferCopyDst, Data: bytesOf(m1)}))
b := must(dev.NewBuffer(gpu.BufferDescriptor{Size: n*4, Usage: gpu.BufferStorage | gpu.BufferCopyDst, Data: bytesOf(m2)}))
out := must(dev.NewBuffer(gpu.BufferDescriptor{Size: n*4, Usage: gpu.BufferStorage | gpu.BufferMapRead}))

enc := dev.NewCommandEncoder()
cp := enc.BeginComputePass()
cp.SetPipeline(pipe)
cp.SetBindings(gpu.Binding{Index: 0, Buffer: a}, gpu.Binding{Index: 1, Buffer: b}, gpu.Binding{Index: 2, Buffer: out})
cp.Dispatch(ceilDiv(n, 256), 1, 1)
cp.End()
dev.Queue().Submit(enc.Finish())
dev.Queue().WaitIdle()

result := unsafe.Slice((*float32)(...), n) // from out.Map()
```

- **Metal** satisfies this 1:1 with the current bindings.
- **GL** satisfies it by recording the 8 calls and replaying them on its context
  thread; `SetBindings` → `BindBufferBase(index)`, `Dispatch` → `DispatchCompute`,
  `WaitIdle` → `MemoryBarrier` + `Finish`. Where GL strains: it needs the locked
  thread and an explicit `MemoryBarrier` the others express as submit semantics —
  exactly the emulation cost §3 assigns to GL.

This slice is the first implementation milestone: it deletes the duplicated
`add/sub/sqrt/mul` per-backend code in favor of one path.

### 5b. Render slice — polyred's deferred pass as a RenderPipeline

Polyred renders via `render.NewRenderer(opts).Render() *image.RGBA`
(`render/raster.go`), with passes shadow → forward → deferred → AA, shaders as a
`shader.Program` interface, uniforms in an `MVP` struct, and a `FragmentBuffer`
G-buffer. The GPU mapping:

| polyred concept | GPU abstraction |
|---|---|
| `shader.Program` (`Vertex`/`Fragment` funcs) | a `RenderPipeline` (vertex+fragment `ShaderModule`) |
| `MVP` uniform struct (`shader/mvp.go`) | a `BufferUniform` buffer bound via `SetBindings` |
| mesh vertices/indices | `BufferVertex` / `BufferIndex`, `DrawIndexed` |
| `FragmentBuffer` color+depth (`buffer/buffer.go`) | a color `Texture` + depth `Texture` as `RenderPassDescriptor` attachments |
| forward pass rasterization | one `RenderPass` per object batch |
| deferred shading pass | a `ComputePass` (or full-screen `RenderPass`) over the G-buffer textures |
| final `*image.RGBA` | `CopyTextureToBuffer` + `Map()` readback (headless) or present via `ctx` drawable |

The renderer keeps scene traversal, culling, and `MVP` assembly on the CPU
(cheap, already `sched.Pool`-parallel); only rasterization + shading move to the
GPU. The `Program` interface is the seam: a `GPUProgram` implementation provides
shader modules instead of Go callbacks.

## 6. cgo-free: decision needed

The "aimed to be cgo free" goal dominates the architecture, so make it explicit.

State today: Metal **and** GL both use cgo; only `internal/dl` (hand-rolled asm
`dlopen`/`dlsym` trampolines) is cgo-free, and it is broken. Hand-rolling
per-arch assembly trampolines is reinventing **purego**.

Two framings:

- **(A) Hard requirement.** Every backend routes its symbol loading through a
  purego-style loader; no `import "C"`. Cost: a per-platform/arch dynamic-call
  effort. Recommendation if chosen: **adopt `github.com/ebitengine/purego`**
  rather than maintain `internal/dl`'s assembly. Metal then loads
  Objective-C/`objc_msgSend` via purego (as Ebitengine does), GL via
  `dlopen`/`GetProcAddress`. Highest payoff (pure-Go builds, trivial
  cross-compile) but most up-front work and the riskiest surface (objc_msgSend
  ABI per arch).
- **(B) Aspiration / phased.** Define every backend behind a `loader` interface;
  ship **cgo-first** now (reuse the working `mtl`/`gl` cgo bindings), and slot a
  purego loader in later without changing the `Device` API. Lower risk, unblocks
  the abstraction immediately, keeps the cgo-free option open.

**Recommendation: (B) now, (A) as a tracked milestone.** The valuable,
reviewable work is the `Device` API and the renderer consuming it (§5, §7); that
work is identical regardless of loader. Gate the abstraction on cgo-first
bindings, and make "swap Metal+GL onto purego" a later milestone that adopts
purego instead of repairing `internal/dl`. This also lets CI stay green on day
one. (User decision: confirm A-vs-B and whether `internal/dl` is retired in
favor of purego.)

## 7. Folding into polyred + renderer consumption

Per the chosen direction, the abstraction lives **inside the polyred repo** and
replaces the in-repo `gpu/` toy.

Mechanics:
1. Move `poly.red/x/gpu`'s real code (`mtl`, `gl`, `ctx`, `syscall`,
   `internal/dl`) into `polyred` under `gpu/` (e.g. `gpu/mtl`, `gpu/gl`,
   `gpu/ctx`), replacing the current toy. One module, one `go.mod`, no
   cross-module `go.sum` friction, no `go.work` needed.
2. Delete the empty `vk/`, `dx12/`, `device.go` stubs; recreate them as real
   packages only when implemented.
3. Retire the separate `github.com/polyred/gpu` remote (never published) — fold
   its history or archive it; the code becomes part of polyred.
4. Add the new top-level `gpu.Device` API (§4) as `gpu/device.go` (real, not a
   doc stub), with `gpu/metal_*.go`, `gpu/gl_*.go` backend adapters wrapping the
   moved low-level bindings.

Renderer consumption (the directive's real test — the renderer must *use* it):
- Add a `render.Backend` seam. `NewRenderer` gains an option `Backend(b)` where
  `b` is `CPU` (today's path) or `GPU(dev *gpu.Device)`.
- Implement `shader.Program` for GPU via a `GPUProgram` that carries shader
  modules; `FragmentBuffer` gains a GPU-texture-backed variant; `MVP` is uploaded
  to a uniform buffer each frame.
- First integration target: the **deferred shading pass** (`passDeferred`,
  `render/raster.go`) as a compute pass over G-buffer textures — it is the most
  self-contained, embarrassingly-parallel stage and needs no rasterizer changes.

## 8. Phased roadmap

1. **Unblock + fold in.** Fix the `../gpu` build (truncated asm / go.sum), move
   the real backends into `polyred/gpu`, delete stubs, retire the side repo.
   Keep cgo-first (§6 option B). Exit: `go build ./... && go vet ./...` green,
   matrix demo still passes.
2. **Compute slice (§5a).** Land the `Device`/`Buffer`/`ComputePipeline`/
   `CommandEncoder` API; reimplement matrix `Add/Sub/Sqrt/Mul` through it on
   Metal **and** GL (GL via the context-thread emulation, §3). Delete the
   duplicated per-backend math. Exit: one code path, both backends green.
3. **Render slice (§5b).** Add `RenderPipeline`/`RenderPass`; port `passDeferred`
   to the GPU behind `render.Backend(GPU(...))`. Exit: a scene renders to an
   identical (within tolerance) `*image.RGBA` via CPU and GPU.
4. **cgo-free (§6 option A).** Adopt purego; move Metal + GL symbol loading off
   cgo behind the loader interface. Exit: pure-Go build on darwin+linux.
5. **New backends.** Implement `vk/` then `dx12/` against the now-proven
   `Device` API (SPIR-V path already in `ShaderSource`).
6. **Shader story.** Decide MSL/GLSL/HLSL authoring vs. generating from a Go
   subset; until then, ship hand-written per-backend shader variants.

## 9. Open questions (for the reviewer)

- cgo-free **A vs B** and the fate of `internal/dl` vs adopting purego (§6).
- Binding model: flat ordered bindings now, or invest in bind-group layouts up
  front for Vulkan/DX12 (§4)?
- Shader authoring: hand-written per-backend variants vs. a Go→shader path (§4,
  §8.6) — acceptable to require MSL/GLSL by hand initially?
- Headless vs. windowed: prioritize offscreen render-to-image (CI-testable) over
  swapchain/present for the first render slice?
