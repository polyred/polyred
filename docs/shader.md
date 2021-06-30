# The Rasterization Rendering Pipeline

The `polyred` package offers customized rendering pipeline:

```
                Vertex Generation                  
                       |                           
                       v                           
                 Vertex Shading                    
                       |                           
                       v                           
               Fragment Generation                 
                       |                           
                       v                           
                Fragment Shading                   
```

with the following basic passes:

```go
// PrimitivePass is a pass that executes Draw call concurrently on all
// given triangle primitives, and draws all geometric and rendering
// information on the given buffer. This primitive uses supplied shader
// programs (i.e. currently supports vertex shader and fragment shader)
//
// See shader.Program for more information regarding shader programming.
func (r *Renderer) PrimitivePass(
	buf *Buffer,
	prog shader.Program,
	indexBuf []uint64,
	vertBuf []primitive.Vertex,
)

// ScreenPass is a concurrent executor of the given shader that travel
// through all pixels. Each pixel executes the given shader exactly once.
// One should not manipulate the given image buffer in the shader.
// Instead, return the resulting color in the shader can avoid data race.
func (r *Renderer) ScreenPass(buf *image.RGBA, shade shader.FragmentProgram)
```

# Primitives

TODO: vertex and fragment

# Shader Basics

Shader is an interface that implements the `VertexShader` and
`FragmentShader` methods:

```go
package shader

type Program interface {
	VertexShader(primitive.Vertex) primitive.Vertex
	FragmentShader(primitive.Fragment) color.RGBA
}
```

A `VertexShader` consumes a vertex and returns a transformed vertex.
The input and output vertex may convey different information.
The information stored in a vertex primitive will be interpolated depending
on the specified camera target. A orthographic camera will interpolate
vertex attributes linearly and the interpolation of a perspective camera
is, of course, perspective corrected.

TODO: more about varying in vertex and fragment, smooth and flat.

## License

Copyright &copy; 2020-2021 [Changkun Ou](https://changkun.de). All rights reserved.