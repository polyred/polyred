# ddd

A software hybrid renderer written in pure Go.

```go
import "changkun.de/x/ddd"
```

_Caution: experiment, expect it to break at any time. Use it at your own risk._

## Features

- IO
  + [x] OBJ file loader
  + [ ] OBJ file exporter
- geometry
  + [x] triangle mesh
  + [x] quad mesh
  + [ ] quad dominant mesh
  + [ ] half-edge mesh
  + [ ] built-in geometries
    * [x] plane
    * [ ] cube
- hybrid rendering:
  + [x] rasterization
    * [ ] clipping
    * [x] backface culling
    * [x] viewfrustum culling
    * [x] occlusion culling
  + [x] depth test 
  + [ ] alpha test
  + [x] deferred shading
  + [ ] ambient occlusion
  + [ ] ray tracing
  + anti-aliasing
    * [x] MSAA
- texturing
  + [x] isotropic mipmap
  + [ ] anisotropic mipmap
  + [ ] arbitrary texture size
+ material
  + [x] basic material
  + [x] Blinn-Phong material
  + [ ] Lambertian material
  + [ ] Diffuse material
  + [ ] Glossy material
  + [ ] Micofacet material
- lighting
  + [ ] poing light
  + [ ] shadow mapping
- general
  + [x] concurrent processing

<!-- ## Performance

ddd renders in parallel and utilizes all the CPUs.

| model | resolution | msaa | rendering (ms) |
|:------|:-----------|:-----|:---------------|
| stanford bunny |  -->

![](./examples/teaser.png)


## More Examples

| Example | Code |
|:-------:|:-----:|
|<img src="./examples/bunny/bunny.png" width="300px"/>|[bunny](./examples/bunny/bunny.go)|
|<img src="./examples/dragon/dragon.png" width="300px"/>|[dragon](./examples/dragon/dragon.go)|

## License

Copyright &copy; 2020-2021 [Changkun Ou](https://changkun.de). All rights reserved.