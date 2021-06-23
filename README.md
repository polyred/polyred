<img src="./examples/teaser.png" alt="logo" height="255" align="right" />


# ddd [![Go Reference](https://pkg.go.dev/badge/github.com/changkun/ddd.svg)](https://pkg.go.dev/changkun.de/x/ddd) [![Latest relsease](https://img.shields.io/github/v/tag/changkun/ddd?label=latest)](https://github.com/changkun/ddd/releases)

Software hybrid rendering facilities written in pure Go.

```go
import "changkun.de/x/ddd"
```

_Caution: experiment, expect it to break at any time. Use it at your own risk._

## Features

- IO
  + [x] OBJ file loader
  + [ ] OBJ file exporter
  + [x] Gamma correction
- geometry
  + [x] buffered mesh
  + [x] triangle soup
  + [ ] triangle mesh
  + [ ] quad mesh
  + [ ] quad dominant mesh
  + [ ] half-edge mesh
  + [ ] built-in geometries
    * [x] plane
    * [ ] cube
  + [ ] geometry processing algorithms
    * [ ] smooth normals
- rendering facilities:
  + [x] scene graph
  + [x] rasterization pass
    * [ ] clipping
    * [x] backface culling
    * [x] viewfrustum culling
    * [x] occlusion culling
  + [x] depth test and z-buffer pass
  + [ ] alpha test
  + [x] deferred shading pass
  + ambient occlusion
    + [x] screen-space ambient occlusion (SSAO)
    + [ ] horizon-based ambient occlusion (HBAO)
  + [ ] ray tracing
  + anti-aliasing pass
    * [x] MSAA
  + [ ] denoising
- texturing
  + filters
    + [x] linear
    + [x] bilinear
    + [x] trilinear
    + [x] barycentric
    + [ ] cubic
    + [ ] custom
  + [x] isotropic mipmap
  + [ ] anisotropic mipmap
  + [x] arbitrary texture size
+ material
  + [x] basic material
  + [x] Blinn-Phong material
  + [ ] Lambertian material
  + [ ] Diffuse material
  + [ ] Glossy material
  + [ ] Micofacet material
- lighting
  + [x] point light
  + [ ] directional light
  + [x] shadow mapping
- general
  + [x] concurrent processing

## License

Copyright &copy; 2020-2021 [Changkun Ou](https://changkun.de). All rights reserved.