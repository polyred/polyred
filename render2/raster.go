// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image"
	"runtime"

	"poly.red/camera"
	"poly.red/color"
	"poly.red/geometry"
	"poly.red/geometry/primitive"
	"poly.red/math"
	"poly.red/scene"
	"poly.red/shader"
	"poly.red/texture/buffer"

	"poly.red/internal/sched"
)

// 1 second / 60fps = 16.6666 ms/frame
// 1 second / 30fps = 33.3333 ms/frame

// Renderer is a hybrid software renderer that implements
// rasterization and ray tracing.
type Renderer struct {
	// rendering options
	width      int
	height     int
	msaa       int
	scene      *scene.Scene
	background color.RGBA
	batchSize  int32
	workers    int

	// scheduling, use for hard interruption.
	running uint32 // atomic
	stop    uint32 // atomic

	// rendering caches
	sched          *sched.Pool
	buflen         int
	bufcurrent     int
	format         buffer.PixelFormat
	bufs           []*buffer.Buffer
	renderCamera   camera.Interface
	renderPerspect bool
	viewportMatrix math.Mat4
}

// NewRenderer creates a new renderer.
//
// The returned renderer implements a rasterization rendering pipeline.
func NewRenderer(opts ...Opt) *Renderer {
	r := &Renderer{ // default settings
		bufs:       nil,
		width:      800,
		height:     500,
		msaa:       1,
		scene:      nil,
		workers:    runtime.NumCPU(),
		batchSize:  32, // heuristic
		buflen:     2,  // default use two buffers.
		bufcurrent: 0,
	}
	for _, opt := range opts {
		opt(r)
	}

	r.bufs = make([]*buffer.Buffer, r.buflen)
	r.resetBufs()

	r.viewportMatrix = math.ViewportMatrix(
		float64(r.bufs[0].Bounds().Dx()),
		float64(r.bufs[0].Bounds().Dy()))
	r.sched = sched.New(sched.Workers(r.workers))
	runtime.SetFinalizer(r, func(r *Renderer) {
		r.sched.Release()
	})

	return r
}

func (r *Renderer) resetBufs() {
	for i := 0; i < r.buflen; i++ {
		r.bufs[i] = buffer.NewBuffer(
			image.Rect(0, 0, r.width*r.msaa, r.height*r.msaa),
			buffer.Format(r.format))
	}
}

func (r *Renderer) CurrentBuffer() *buffer.Buffer {
	return r.bufs[r.bufcurrent]
}

func (r *Renderer) NextBuffer() *buffer.Buffer {
	r.bufcurrent = (r.bufcurrent + 1) % r.buflen
	r.bufs[r.bufcurrent].Clear()
	return r.bufs[r.bufcurrent]
}

func (r *Renderer) DrawImage(buf *buffer.Buffer, img *image.RGBA) {
	for i := 0; i < buf.Bounds().Dx(); i++ {
		for j := 0; j < buf.Bounds().Dy(); j++ {
			buf.Set(i, j, buffer.Fragment{
				Ok: true,
				Fragment: primitive.Fragment{
					X: i, Y: j, Col: img.RGBAAt(i, img.Bounds().Dy()),
				},
			})
		}
	}
}

// DrawPrimitives is a pass that executes Draw call concurrently on all
// given triangle primitives, and draws all geometric and rendering
// information on the given buffer. This primitive uses supplied shader
// programs (i.e. currently supports vertex shader and fragment shader)
//
// See shader.Program for more information regarding shader programming.
func (r *Renderer) DrawPrimitives(buf *buffer.Buffer, m geometry.Renderable, p shader.VertexProgram) {
	r.startRunning()
	defer r.stopRunning()

	idx := m.IndexBuffer()
	verts := m.VertexBuffer()

	len := len(idx)
	if len%3 != 0 {
		panic("index buffer must be a 3 multiple")
	}

	r.sched.Add(uint64(len / 3))
	for i := 0; i < len; i += 3 {
		v1 := verts[i]
		v2 := verts[i+1]
		v3 := verts[i+2]
		r.sched.Run(func() {
			if primitive.IsValidTriangle(v1.Pos.ToVec3(), v2.Pos.ToVec3(), v3.Pos.ToVec3()) {
				r.DrawPrimitive(buf, p, v1, v2, v3)
			}
		})
	}
	r.sched.Wait()
}

// DrawPrimitive implements a triangle draw call of the rasteriation graphics pipeline.
func (r *Renderer) DrawPrimitive(buf *buffer.Buffer, p shader.VertexProgram, p1, p2, p3 *primitive.Vertex) {
	v1 := p(*p1)
	v2 := p(*p2)
	v3 := p(*p3)

	// For perspective corrected interpolation
	recipw := [3]float64{1, 1, 1}
	if r.renderPerspect {
		recipw[0] = -1 / v1.Pos.W
		recipw[1] = -1 / v2.Pos.W
		recipw[2] = -1 / v3.Pos.W
	}

	v1.Pos = v1.Pos.Apply(r.viewportMatrix).Pos()
	v2.Pos = v2.Pos.Apply(r.viewportMatrix).Pos()
	v3.Pos = v3.Pos.Apply(r.viewportMatrix).Pos()

	// TODO: which should be the first?

	// Back-face culling
	if v2.Pos.Sub(v1.Pos).Cross(v3.Pos.Sub(v1.Pos)).Z < 0 {
		return
	}

	// View frustum culling
	viewportAABB := primitive.NewAABB(
		math.NewVec3(float64(buf.Bounds().Dx()), float64(buf.Bounds().Dy()), 1),
		math.NewVec3(0, 0, 0),
		math.NewVec3(0, 0, -1),
	)
	if !viewportAABB.Intersect(primitive.NewAABB(v1.Pos.ToVec3(), v2.Pos.ToVec3(), v3.Pos.ToVec3())) {
		return
	}

	// Compute AABB make the AABB a little bigger that align with
	// pixels to contain the entire triangle
	aabb := primitive.NewAABB(v1.Pos.ToVec3(), v2.Pos.ToVec3(), v3.Pos.ToVec3())
	xmin := int(math.Round(aabb.Min.X) - 1)
	xmax := int(math.Round(aabb.Max.X) + 1)
	ymin := int(math.Round(aabb.Min.Y) - 1)
	ymax := int(math.Round(aabb.Max.Y) + 1)

	// TODO: parallize this loop, smarter scheduling to minimize lock
	// contention
	for x := xmin; x <= xmax; x++ {
		for y := ymin; y <= ymax; y++ {
			if !buf.In(x, y) {
				continue
			}

			p := math.NewVec2(float64(x)+0.5, float64(y)+0.5)

			// Compute barycentric coordinates of a triangle in screen
			// space and check if the fragment is inside triangle.
			bc := math.Barycoord(p, v1.Pos.ToVec2(), v2.Pos.ToVec2(), v3.Pos.ToVec2())
			if bc[0] < -math.Epsilon ||
				bc[1] < -math.Epsilon ||
				bc[2] < -math.Epsilon {
				continue
			}

			// Early Z-test. We normalize depth values to [0, 1], such that
			// the smallest depth value is 0. This collaborate with the buffer
			// clearing.
			z := ((bc[0]*v1.Pos.Z + bc[1]*v2.Pos.Z + bc[2]*v3.Pos.Z) + 1) / 2
			if !buf.DepthTest(x, y, z) {
				continue
			}

			frag := primitive.Fragment{
				X:     x,
				Y:     y,
				Depth: z,
			}

			// Interpolating UV
			uvX := r.interpolate([3]float64{v1.UV.X, v2.UV.X, v3.UV.X}, recipw, bc)
			uvY := r.interpolate([3]float64{v1.UV.Y, v2.UV.Y, v3.UV.Y}, recipw, bc)
			frag.UV = math.NewVec2(uvX, uvY)

			p1 := math.NewVec2(p.X+1, p.Y)
			p2 := math.NewVec2(p.X, p.Y+1)
			bcx := math.Barycoord(p1, v1.Pos.ToVec2(), v2.Pos.ToVec2(), v3.Pos.ToVec2())
			bcy := math.Barycoord(p2, v1.Pos.ToVec2(), v2.Pos.ToVec2(), v3.Pos.ToVec2())

			uvdU := r.interpolate([3]float64{v1.UV.X, v2.UV.X, v3.UV.X}, recipw, bcx)
			uvdX := r.interpolate([3]float64{v1.UV.Y, v2.UV.Y, v3.UV.Y}, recipw, bcx)
			uvdV := r.interpolate([3]float64{v1.UV.Y, v2.UV.Y, v3.UV.Y}, recipw, bcy)
			uvdY := r.interpolate([3]float64{v1.UV.Y, v2.UV.Y, v3.UV.Y}, recipw, bcy)
			frag.Du = math.Sqrt((uvdU-uvX)*(uvdU-uvX) + (uvdX-uvY)*(uvdX-uvY))
			frag.Dv = math.Sqrt((uvdV-uvX)*(uvdV-uvX) + (uvdY-uvY)*(uvdY-uvY))

			// Interpolate normal
			if !v1.Nor.IsZero() && !v2.Nor.IsZero() && !v3.Nor.IsZero() {
				nx := r.interpolate([3]float64{v1.Nor.X, v2.Nor.X, v3.Nor.X}, recipw, bc)
				ny := r.interpolate([3]float64{v1.Nor.Y, v2.Nor.Y, v3.Nor.Y}, recipw, bc)
				nz := r.interpolate([3]float64{v1.Nor.Z, v2.Nor.Z, v3.Nor.Z}, recipw, bc)
				frag.Nor = math.NewVec4(nx, ny, nz, 0)
			}

			// Interpolate colors
			if v1.Col != color.Discard || v2.Col != color.Discard || v3.Col != color.Discard {
				cr := r.interpolate([3]float64{float64(v1.Col.R), float64(v2.Col.R), float64(v3.Col.R)}, recipw, bc)
				cg := r.interpolate([3]float64{float64(v1.Col.G), float64(v2.Col.G), float64(v3.Col.G)}, recipw, bc)
				cb := r.interpolate([3]float64{float64(v1.Col.B), float64(v2.Col.B), float64(v3.Col.B)}, recipw, bc)
				ca := r.interpolate([3]float64{float64(v1.Col.A), float64(v2.Col.A), float64(v3.Col.A)}, recipw, bc)
				frag.Col = color.RGBA{
					R: uint8(math.Clamp(cr, 0, 0xff)),
					G: uint8(math.Clamp(cg, 0, 0xff)),
					B: uint8(math.Clamp(cb, 0, 0xff)),
					A: uint8(math.Clamp(ca, 0, 0xff)),
				}
			}

			// Interpolate custom varying
			if len(v1.AttrSmooth) > 0 {
				r.interpoVaryings(v1.AttrSmooth, v2.AttrSmooth, v3.AttrSmooth, frag.AttrSmooth, recipw, bc)
			}

			buf.Set(x, y, buffer.Fragment{
				Ok:       true,
				Fragment: frag,
			})
		}
	}
}

// interpoVaryings perspective correct interpolates
func (r *Renderer) interpoVaryings(v1, v2, v3, frag map[string]interface{},
	recipw, bc [3]float64) {
	l := len(frag)
	for name := range v1 {
		switch val1 := v1[name].(type) {
		case float64:
			if l > 0 {
				frag[name] = r.interpolate([3]float64{
					val1, v2[name].(float64), v3[name].(float64),
				}, recipw, bc)
			}
		case math.Vec2:
			x := r.interpolate([3]float64{
				val1.X,
				v2[name].(math.Vec4).X,
				v3[name].(math.Vec4).X,
			}, recipw, bc)
			y := r.interpolate([3]float64{
				val1.Y,
				v2[name].(math.Vec4).Y,
				v3[name].(math.Vec4).Y,
			}, recipw, bc)
			if l > 0 {
				frag[name] = math.NewVec2(x, y)
			}
		case math.Vec3:
			x := r.interpolate([3]float64{
				val1.X,
				v2[name].(math.Vec4).X,
				v3[name].(math.Vec4).X,
			}, recipw, bc)
			y := r.interpolate([3]float64{
				val1.Y,
				v2[name].(math.Vec4).Y,
				v3[name].(math.Vec4).Y,
			}, recipw, bc)
			z := r.interpolate([3]float64{
				val1.Z,
				v2[name].(math.Vec4).Z,
				v3[name].(math.Vec4).Z,
			}, recipw, bc)
			if l > 0 {
				frag[name] = math.NewVec3(x, y, z)
			}
		case math.Vec4:
			x := r.interpolate([3]float64{
				val1.X,
				v2[name].(math.Vec4).X,
				v3[name].(math.Vec4).X,
			}, recipw, bc)
			y := r.interpolate([3]float64{
				val1.Y,
				v2[name].(math.Vec4).Y,
				v3[name].(math.Vec4).Y,
			}, recipw, bc)
			z := r.interpolate([3]float64{
				val1.Z,
				v2[name].(math.Vec4).Z,
				v3[name].(math.Vec4).Z,
			}, recipw, bc)
			w := r.interpolate([3]float64{
				val1.W,
				v2[name].(math.Vec4).W,
				v3[name].(math.Vec4).W,
			}, recipw, bc)
			if l > 0 {
				frag[name] = math.NewVec4(x, y, z, w)
			}
		}
	}
}

// interpolate does perspective-correct interpolation for the given varying.
//
// See: Low, Kok-Lim. "Perspective-correct interpolation." Technical writing,
// Department of Computer Science, University of North Carolina at
// Chapel Hill (2002).
func (r *Renderer) interpolate(varying, recipw, barycoord [3]float64) float64 {
	recipw[0] *= barycoord[0]
	recipw[1] *= barycoord[1]
	recipw[2] *= barycoord[2]
	norm := recipw[0]*varying[0] + recipw[1]*varying[1] + recipw[2]*varying[2]
	if r.renderPerspect {
		norm *= 1 / (recipw[0] + recipw[1] + recipw[2])
	}
	return norm
}

// DrawFragments is a concurrent executor of the given shader that travel
// through all fragments. Each fragment executes the given shaders exactly once.
//
// One should not manipulate the given image buffer in the shader.
// Instead, return the resulting color in the shader can avoid data race.
func (r *Renderer) DrawFragments(buf *buffer.Buffer, funcs ...shader.FragmentProgram) {
	if funcs == nil {
		return
	}

	r.startRunning()
	defer r.stopRunning()

	// Because the shader executes exactly on each pixel once, there is
	// no need to lock the pixel while reading or writing.

	w := buf.Bounds().Dx()
	h := buf.Bounds().Dy()
	n := w * h

	batchSize := int(r.batchSize)
	wsteps := w / batchSize
	hsteps := h / batchSize

	defer r.sched.Wait()

	if wsteps == 0 && hsteps == 0 {
		r.sched.Add(1)

		// Note: sadly that the executing function will escape to the
		// heap which increases the memory allocation. No workaround.
		r.sched.Run(func() {
			for i := 0; i < n; i++ {
				r.DrawFragment(buf, i%w, i/w, funcs...)
			}
		})
		return
	}

	numTasks := n / batchSize
	r.sched.Add(uint64(numTasks))
	for i := 0; i < numTasks; i++ {
		ii := i
		r.sched.Run(func() {
			x0 := ii * batchSize
			x1 := x0 + batchSize
			for j := x0; j < x1; j++ {
				x, y := j%w, j/w
				r.DrawFragment(buf, x, y, funcs...)
			}
		})
	}

	if n%batchSize != 0 {
		r.sched.Add(1)
		r.sched.Run(func() {
			for j := numTasks * batchSize; j < n; j++ {
				x, y := j%w, j/w
				r.DrawFragment(buf, x, y, funcs...)
			}
		})
	}
}

// DrawFragment executes the given shaders on a specific fragment.
//
// Note that it is caller's responsibility to protect the safty of fragment
// coordinates, as well as data race of the given buffer.
func (r *Renderer) DrawFragment(buf *buffer.Buffer, x, y int, shaders ...shader.FragmentProgram) {
	info := buf.UnsafeAt(x, y)
	for i := 0; i < len(shaders); i++ {
		info.Col = shaders[i](info.Fragment)
		if info.Col == color.Discard {
			return
		}
	}
	buf.UnsafeSet(x, y, info)
}
