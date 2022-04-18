// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"fmt"
	"image"
	"image/color"

	"poly.red/buffer"
	"poly.red/camera"
	"poly.red/geometry"
	"poly.red/geometry/primitive"
	"poly.red/internal/imageutil"
	"poly.red/light"
	"poly.red/math"
	"poly.red/scene"
	"poly.red/shader"

	"poly.red/internal/profiling"
	"poly.red/internal/spinlock"
)

type shadowInfo struct {
	active bool
	camera camera.Interface
	depths []float32
	lock   []spinlock.SpinLock
}

func (r *Renderer) initShadowMaps() {
	lightSources, _ := r.cfg.Scene.Lights()
	r.shadowBufs = make([]shadowInfo, len(lightSources))
	for i := 0; i < len(lightSources); i++ {
		if !lightSources[i].CastShadow() {
			continue
		}

		// initialize scene camera
		tm := camera.ViewMatrix(
			lightSources[i].Position(),
			r.cfg.Scene.Center(),
			math.NewVec3[float32](0, 1, 0),
		).MulM(r.cfg.Camera.ViewMatrix().Inv()).MulM(r.cfg.Camera.ProjMatrix().Inv())
		v1 := math.NewVec4[float32](1, 1, 1, 1).Apply(tm).Pos().ToVec3()
		v2 := math.NewVec4[float32](1, 1, -1, 1).Apply(tm).Pos().ToVec3()
		v3 := math.NewVec4[float32](1, -1, 1, 1).Apply(tm).Pos().ToVec3()
		v4 := math.NewVec4[float32](-1, 1, 1, 1).Apply(tm).Pos().ToVec3()
		v5 := math.NewVec4[float32](-1, -1, 1, 1).Apply(tm).Pos().ToVec3()
		v6 := math.NewVec4[float32](1, -1, -1, 1).Apply(tm).Pos().ToVec3()
		v7 := math.NewVec4[float32](-1, 1, -1, 1).Apply(tm).Pos().ToVec3()
		v8 := math.NewVec4[float32](-1, -1, -1, 1).Apply(tm).Pos().ToVec3()
		aabb := primitive.NewAABB(v1, v2, v3, v4, v5, v6, v7, v8)
		le := aabb.Min.X
		ri := aabb.Max.X
		bo := aabb.Min.Y
		to := aabb.Max.Y
		ne := aabb.Max.Z
		fa := aabb.Min.Z - 2
		// aspect := ri / to
		// fov := 2 * math.Atan(to/math.Abs(ne))

		li := lightSources[i]
		var c camera.Interface
		switch l := li.(type) {
		case *light.Point:
			// TODO: why perspective camera does not work?
			// c = camera.NewPerspective(
			// 	li.Position(),
			// 	r.scene.Center(),
			// 	math.NewVec4(0, 1, 0, 0),
			// 	fov, aspect, 0.001, 100,
			// )
			// TODO: use cube shadow map for point light
			c = camera.NewOrthographic(
				camera.Position(l.Position()),
				camera.LookAt(r.cfg.Scene.Center(),
					math.NewVec3[float32](0, 1, 0)),
				camera.ViewFrustum(le, ri, bo, to, ne, fa),
			)
		default:
		}
		r.shadowBufs[i].active = true
		r.shadowBufs[i].camera = c
		r.shadowBufs[i].depths = make([]float32, r.bufs[0].Bounds().Dx()*r.bufs[0].Bounds().Dy())
		r.shadowBufs[i].lock = make([]spinlock.SpinLock, r.bufs[0].Bounds().Dx()*r.bufs[0].Bounds().Dy())
	}
}

func (r *Renderer) passShadows(index int) {
	lightSources, _ := r.cfg.Scene.Lights()
	if !lightSources[index].CastShadow() {
		return
	}

	if r.cfg.Debug {
		done := profiling.Timed("forward pass (shadow)")
		defer done()
		defer func() {
			img := image.NewRGBA(image.Rect(0, 0, r.bufs[0].Bounds().Dx(), r.bufs[0].Bounds().Dy()))
			for i := 0; i < r.bufs[0].Bounds().Dx(); i++ {
				for j := 0; j < r.bufs[0].Bounds().Dy(); j++ {
					z := r.shadowBufs[index].depths[i+(r.bufs[0].Bounds().Dy()-j-1)*r.bufs[0].Bounds().Dx()]
					img.SetRGBA(i, j, color.RGBA{
						uint8(z * 255),
						uint8(z * 255),
						uint8(z * 255),
						255,
					})
				}
			}
			file := fmt.Sprintf("shadow-%d.png", index)
			fmt.Printf("saving (shadow map)... %s\n", file)
			imageutil.Save(img, file)
		}()
	}

	scene.IterObjects(r.cfg.Scene, func(g *geometry.Geometry, modelMatrix math.Mat4[float32]) bool {
		mvp := shader.MVP{
			Model:    modelMatrix.MulM(g.ModelMatrix()),
			View:     r.shadowBufs[index].camera.ViewMatrix(),
			Proj:     r.shadowBufs[index].camera.ProjMatrix(),
			Viewport: math.ViewportMatrix(float32(r.bufs[0].Bounds().Dx()), float32(r.bufs[0].Bounds().Dy())),
		}

		// NormalMatrix can be ((Tcamera * Tmodel)^(-1))^T or ((Tmodel)^(-1))^T
		// depending on which transformation space. Here we use the 2nd form,
		// i.e. model space normal matrix to save some computation of camera
		// transforamtion in the shading process.
		// The reason we need normal matrix is that normals are transformed
		// incorrectly using MVP matrices. However, a normal matrix helps us
		// to fix the problem.
		mvp.Normal = mvp.Model.Inv().T()

		tris := g.Triangles()
		r.sched.Add(len(tris))
		for i := range tris {
			t := tris[i]
			r.sched.Run(func() {
				if !t.IsValid() {
					return
				}
				r.drawDepth(index, t, mvp)
			})
		}
		return true
	})
	r.sched.Wait()
}

func (r *Renderer) drawDepth(index int, t *primitive.Triangle, mvp shader.MVP) {
	var t1, t2, t3 *primitive.Vertex
	t1 = &primitive.Vertex{
		Pos: mvp.Proj.MulM(mvp.View).MulM(mvp.Model).MulV(t.V1.Pos),
		Col: t.V1.Col,
		UV:  t.V1.UV,
		Nor: t.V1.Nor.Apply(mvp.Normal),
	}
	t2 = &primitive.Vertex{
		Pos: mvp.Proj.MulM(mvp.View).MulM(mvp.Model).MulV(t.V2.Pos),
		Col: t.V2.Col,
		UV:  t.V2.UV,
		Nor: t.V2.Nor.Apply(mvp.Normal),
	}
	t3 = &primitive.Vertex{
		Pos: mvp.Proj.MulM(mvp.View).MulM(mvp.Model).MulV(t.V3.Pos),
		Col: t.V3.Col,
		UV:  t.V3.UV,
		Nor: t.V3.Nor.Apply(mvp.Normal),
	}
	t1.Pos = t1.Pos.Apply(mvp.Viewport).Pos()
	t2.Pos = t2.Pos.Apply(mvp.Viewport).Pos()
	t3.Pos = t3.Pos.Apply(mvp.Viewport).Pos()

	// Backface culling
	if t2.Pos.Sub(t1.Pos).Cross(t3.Pos.Sub(t1.Pos)).Z < 0 {
		return
	}

	// Viewfrustum culling
	if r.cullViewFrustum(r.bufs[0], t1.Pos, t2.Pos, t3.Pos) {
		return
	}

	// Compute AABB make the AABB a little bigger that align with pixels
	// to contain the entire triangle
	aabb := primitive.NewAABB(t1.Pos.ToVec3(), t2.Pos.ToVec3(), t3.Pos.ToVec3())
	xmin := int(math.Round(aabb.Min.X) - 1)
	xmax := int(math.Round(aabb.Max.X) + 1)
	ymin := int(math.Round(aabb.Min.Y) - 1)
	ymax := int(math.Round(aabb.Max.Y) + 1)
	for x := xmin; x <= xmax; x++ {
		for y := ymin; y <= ymax; y++ {
			if !r.bufs[0].In(x, y) {
				continue
			}
			p := math.NewVec2(float32(x)+0.5, float32(y)+0.5)
			bc := math.Barycoord(p, t1.Pos.ToVec2(), t2.Pos.ToVec2(), t3.Pos.ToVec2())

			// Is inside triangle?
			if bc[0] < -math.Epsilon || bc[1] < -math.Epsilon || bc[2] < -math.Epsilon {
				continue
			}

			// Z-test
			z := bc[0]*t1.Pos.Z + bc[1]*t2.Pos.Z + bc[2]*t3.Pos.Z
			if !r.shadowDepthTest(index, x, y, z) {
				continue
			}

			// update shadow map
			idx := x + y*r.bufs[0].Bounds().Dx()
			r.shadowBufs[index].lock[idx].Lock()
			r.shadowBufs[index].depths[idx] = z
			r.shadowBufs[index].lock[idx].Unlock()
		}
	}
}

func (r *Renderer) shadowDepthTest(index int, x, y int, z float32) bool {
	idx := x + y*r.bufs[0].Bounds().Dx()
	buf := r.shadowBufs[index]

	buf.lock[idx].Lock()
	defer buf.lock[idx].Unlock()
	return !(z <= buf.depths[idx])
}

func (r *Renderer) shadingVisibility(shadowIdx int,
	info buffer.Fragment, uniforms *shader.MVP,
) bool {
	lightSources, _ := r.cfg.Scene.Lights()
	if !lightSources[shadowIdx].CastShadow() {
		return true
	}

	matVP := uniforms.Viewport
	matScreenToWorld := uniforms.ViewportToWorld
	shadowMap := &r.shadowBufs[shadowIdx]

	// transform scrren coordinate to light viewport
	screenCoord := math.NewVec4(float32(info.X), float32(info.Y), info.Depth, 1).
		Apply(matScreenToWorld).
		Apply(shadowMap.camera.ViewMatrix()).
		Apply(shadowMap.camera.ProjMatrix()).
		Apply(matVP).Pos()

	lightX, lightY := int(screenCoord.X), int(screenCoord.Y)
	bufIdx := lightX + lightY*r.bufs[0].Bounds().Dx()

	shadow := float32(0)
	if bufIdx > 0 && bufIdx < len(shadowMap.depths) {
		shadowZ := shadowMap.depths[bufIdx]
		const bias = 0.03
		if screenCoord.Z < shadowZ-bias {
			shadow++
		}

		// bufIdx2 := lightX + 1 + lightY*r.width
		// bufIdx3 := lightX + (lightY+1)*r.width
		// bufIdx4 := lightX + 1 + (lightY+1)*r.width
		// if (bufIdx2 > 0 && bufIdx2 < len(shadowMap.depths)) &&
		// 	(bufIdx3 > 0 && bufIdx3 < len(shadowMap.depths)) &&
		// 	(bufIdx4 > 0 && bufIdx4 < len(shadowMap.depths)) {

		// 	shadowZ2 := shadowMap.depths[bufIdx2]
		// 	if screenCoord.Z < shadowZ2-bias {
		// 		shadow++
		// 	}
		// 	shadowZ3 := shadowMap.depths[bufIdx3]
		// 	if screenCoord.Z < shadowZ3-bias {
		// 		shadow++
		// 	}
		// 	shadowZ4 := shadowMap.depths[bufIdx4]
		// 	if screenCoord.Z < shadowZ4-bias {
		// 		shadow++
		// 	}
		// }
	}

	return shadow > 0
}
