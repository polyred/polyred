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
    * [ ] curvature
    * [ ] quadric error simplification
    * [ ] melax simplification
    * [ ] uv parameterization
- rendering facilities:
  + [x] scene graph
  + [ ] primitive pass
  + [x] deferred shading pass
  + [x] abstract concurrent screen pass
  + [x] depth test and z-buffer pass
  + [x] rasterization pass
    * [x] clipping
    * [x] backface culling
    * [x] viewfrustum culling
    * [x] occlusion culling
    * [x] perspective correct interpolation
  + [ ] alpha test
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
  + [x] Diffuse material
  + [ ] Lambertian material
  + [ ] Glossy material
  + [ ] Micofacet material
- lighting
  + [x] point light
  + [x] directional light
  + [ ] area light
  + [ ] spot light
- local shading
  + [x] shadow mapping
- global illumination
  + ambient occlusion
    + [x] screen-space ambient occlusion (SSAO)
    + [ ] horizon-based ambient occlusion (HBAO)


## License

Copyright &copy; 2020-2021 [Changkun Ou](https://changkun.de). All rights reserved.