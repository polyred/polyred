## Features

- IO
  + [x] OBJ file loader
  + [ ] MTL file loader
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
  * [ ] LOD
  * [ ] Tesselation
- rendering facilities:
  + [x] perspective and orthographic camera
  + [x] orbit control 
  + [x] scene graph
  + [ ] BVH acceleration, morton coding, cache coherence optimization
  + [x] primitive pass
  + [x] shader programming
  + [x] deferred shading pass
  + [x] abstract concurrent screen pass
  + [x] depth test and z-buffer pass
  + [x] wireframe drawing (Bresenham algorithm)
  + [x] rasterization pass
    * [x] clipping
    * [x] backface culling
    * [x] viewfrustum culling
    * [x] occlusion culling
    * [x] perspective correct interpolation
  + [ ] ray tracing
  + [ ] tiled rendering
  + [ ] hybrid rendering
  + [ ] tiled and clustered forward rendering
  + [ ] visibility buffer
  + [ ] command buffer?
  + [ ] transparency
  + [ ] raster order group
  + anti-aliasing pass
    * [x] MSAA
    * [ ] TAA?
  + [ ] denoising
  + [ ] Tangent space normal mapping
  + [ ] Physically-based rendering (PBR)
  + [ ] Alpha testing
  + [x] Alpha blending
  + [ ] Skeletal animation
  + [ ] Rendering statistics (TODO: what should we do about this?)
  + [x] GUI window
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
  + [ ] texture formats (RGBA, BGRA, ...)
+ material
  + [x] basic material
  + [x] Blinn-Phong material
  + [x] Diffuse material
  + [ ] Lambertian material
  + [ ] Glossy material
  + [ ] Micofacet material
  + [ ] Physically-based materials
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