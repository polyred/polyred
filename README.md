# polyred [![Go Reference](https://pkg.go.dev/badge/poly.red.svg)](https://pkg.go.dev/poly.red) [![polyred](https://github.com/polyred/polyred/actions/workflows/polyred.yml/badge.svg?branch=main)](https://github.com/polyred/polyred/actions/workflows/polyred.yml) ![](https://changkun.de/urlstat?mode=github&repo=polyred/polyred)

3D graphics in Go.

```go
import "poly.red"
```

_Warning: under experiment, expect to break at anytime._

## GPU compute and rendering, in Go

`poly.red/gpu` is a backend-agnostic GPU abstraction — a WebGPU-style
`Device`/`Queue`/`Buffer`/`Pipeline`/`CommandEncoder` API for running **compute**
and **rendering** pipelines, with a driver (Metal today) underneath. It is
**cgo-free**: the Metal/Objective-C runtime is reached through
[ebitengine/purego](https://github.com/ebitengine/purego), so it builds with
`CGO_ENABLED=0`.

Shaders are written **in Go** and compiled to the backend's shading language by
`poly.red/gpu/shader` — compute, vertex, and fragment kernels, with varyings,
uniforms, vector math, and control flow.

```go
dev, _ := gpu.Open()                 // Metal on darwin
defer dev.Close()

// Shaders authored in Go, compiled to MSL.
ks, _ := shader.Compile(`
package kernels
type Vec4 struct{ X, Y, Z, W float32 }
type VOut struct {
    Pos   Vec4 ` + "`gpu:\"position\"`" + `
    Color Vec4
}
//gpu:vertex
func VMain(vid uint, pos []float32, col []float32) VOut {
    return VOut{Vec4{pos[vid*2], pos[vid*2+1], 0, 1}, Vec4{col[vid*3], col[vid*3+1], col[vid*3+2], 1}}
}
//gpu:fragment
func FMain(in VOut) Vec4 { return in.Color }
`)
// ... build a render pipeline, render to a texture, read pixels back.
```

See the runnable end-to-end example:

```sh
go run ./cmd/gpudemo -o triangle.png    # Go shaders -> Metal -> PNG, cgo-free
```

The design, decisions, and roadmap live in
[`docs/gpu-abstraction.md`](docs/gpu-abstraction.md); implementation specs are in
[`specs/`](specs/README.md).

### Status

| Capability | State |
| --- | --- |
| `Device` API (buffers, bind groups, compute + render pipelines, passes) | working |
| Metal backend (compute + render), cgo-free via purego | working |
| Go→shader compiler (compute + vertex/fragment, varyings, uniforms, vector math) | working |
| GPU compute (matrix ops), headless render, GPU lighting math | proven by tests |
| Renderer GPU offload via render.GPU(dev): gamma + full deferred Blinn-Phong pass (bit-identical) | working |
| Windowed present; GL/Vulkan/DX12 backends | planned |
| OpenGL / Vulkan / DirectX 12 backends | planned |

cgo-free build/test of the Metal GPU stack on darwin:

```sh
CGO_ENABLED=0 go test ./gpu ./gpu/mtl ./gpu/shader ./gpu/tests
```
