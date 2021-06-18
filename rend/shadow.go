// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package rend

import (
	"fmt"
	"image"
	"image/color"
	"sync"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/geometry/primitive"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/utils"
)

type shadowInfo struct {
	active   bool
	settings *light.ShadowMap
	depths   []float64
	lock     []sync.Mutex
}

func (r *Renderer) initShadowMaps() {
	r.shadowBufs = make([]shadowInfo, len(r.scene.LightSources))
	for i := 0; i < len(r.scene.LightSources); i++ {
		if !r.scene.LightSources[i].CastShadow() {
			continue
		}

		// initialize scene camera
		tm := camera.ViewMatrix(
			r.scene.LightSources[i].Position(),
			r.scene.Center(),
			math.NewVector(0, 1, 0, 0),
		).
			MulM(r.scene.Camera.ViewMatrix().Inv()).
			MulM(r.scene.Camera.ProjMatrix().Inv())
		v1 := math.NewVector(1, 1, 1, 1).Apply(tm).Pos()
		v2 := math.NewVector(1, 1, -1, 1).Apply(tm).Pos()
		v3 := math.NewVector(1, -1, 1, 1).Apply(tm).Pos()
		v4 := math.NewVector(-1, 1, 1, 1).Apply(tm).Pos()
		v5 := math.NewVector(-1, -1, 1, 1).Apply(tm).Pos()
		v6 := math.NewVector(1, -1, -1, 1).Apply(tm).Pos()
		v7 := math.NewVector(-1, 1, -1, 1).Apply(tm).Pos()
		v8 := math.NewVector(-1, -1, -1, 1).Apply(tm).Pos()
		aabb := primitive.NewAABB(v1, v2, v3, v4, v5, v6, v7, v8)
		le := aabb.Min.X
		ri := aabb.Max.X
		bo := aabb.Min.Y
		to := aabb.Max.Y
		ne := aabb.Max.Z
		fa := aabb.Min.Z - 2
		// aspect := ri / to
		// fov := 2 * math.Atan(to/math.Abs(ne))

		li := r.scene.LightSources[i]
		var c camera.Interface
		switch l := li.(type) {
		case *light.Point:
			// TODO: why perspective camera does not work?
			// c = camera.NewPerspective(
			// 	li.Position(),
			// 	r.scene.Center(),
			// 	math.NewVector(0, 1, 0, 0),
			// 	fov, aspect, 0.001, 100,
			// )
			c = camera.NewOrthographic(
				l.Position(),
				r.scene.Center(),
				math.NewVector(0, 1, 0, 0),
				le, ri, bo, to, ne, fa,
			)
		default:
		}
		r.shadowBufs[i].active = true
		r.shadowBufs[i].settings = light.NewShadowMap(
			light.WithShadowMapCamera(c),
		)
		r.shadowBufs[i].depths = make([]float64, r.width*r.height)
		r.shadowBufs[i].lock = make([]sync.Mutex, r.width*r.height)
	}
}

func (r *Renderer) passShadows(index int) {
	if !r.scene.LightSources[index].CastShadow() {
		return
	}

	if r.debug {
		done := utils.Timed("forward pass (shadow)")
		defer done()
		defer func() {
			img := image.NewRGBA(image.Rect(0, 0, r.width, r.height))
			for i := 0; i < r.width; i++ {
				for j := 0; j < r.height; j++ {
					z := r.shadowBufs[index].depths[i+(r.height-j-1)*r.width]
					img.Set(i, j, color.RGBA{
						uint8(z * 255),
						uint8(z * 255),
						uint8(z * 255),
						255,
					})
				}
			}
			file := fmt.Sprintf("shadow-%d.png", index)
			fmt.Printf("saving (shadow map)... %s\n", file)
			utils.Save(img, file)
		}()
	}

	c := r.shadowBufs[index].settings.Camera()
	matView := c.ViewMatrix()
	matProj := c.ProjMatrix()
	matVP := math.ViewportMatrix(float64(r.width), float64(r.height))
	for m := range r.scene.Meshes {
		mesh := r.scene.Meshes[m]
		uniforms := map[string]interface{}{
			"matModel":  mesh.ModelMatrix(),
			"matView":   matView,
			"matProj":   matProj,
			"matVP":     matVP,
			"matNormal": mesh.NormalMatrix(),
		}

		length := len(mesh.Faces)
		for i := 0; i < length; i += int(r.concurrentSize) {
			ii := i
			r.limiter.Execute(func() {
				for k := int32(0); k < r.concurrentSize; k++ {
					if ii+int(k) >= length {
						return
					}
					r.drawDepth(index, uniforms, mesh.Faces[ii+int(k)], mesh.Material)
				}
			})
		}
	}
	r.limiter.Wait()
}

func (r *Renderer) drawDepth(index int, uniforms map[string]interface{}, tri *primitive.Triangle, mat material.Material) {
	var t1, t2, t3 primitive.Vertex
	if mat != nil {
		t1 = mat.VertexShader(tri.V1, uniforms)
		t2 = mat.VertexShader(tri.V2, uniforms)
		t3 = mat.VertexShader(tri.V3, uniforms)
	} else {
		t1 = defaultVertexShader(tri.V1, uniforms)
		t2 = defaultVertexShader(tri.V2, uniforms)
		t3 = defaultVertexShader(tri.V3, uniforms)
	}

	// Backface culling
	if t2.Pos.Sub(t1.Pos).Cross(t3.Pos.Sub(t1.Pos)).Z < 0 {
		return
	}

	// Viewfrustum culling
	if !r.inViewport(t1.Pos, t2.Pos, t3.Pos) {
		return
	}

	// Compute AABB make the AABB a little bigger that align with pixels
	// to contain the entire triangle
	aabb := primitive.NewAABB(t1.Pos, t2.Pos, t3.Pos)
	xmin := int(math.Round(aabb.Min.X) - 1)
	xmax := int(math.Round(aabb.Max.X) + 1)
	ymin := int(math.Round(aabb.Min.Y) - 1)
	ymax := int(math.Round(aabb.Max.Y) + 1)

	for x := xmin; x <= xmax; x++ {
		for y := ymin; y <= ymax; y++ {
			if x < 0 || x >= r.width || y < 0 || y >= r.height {
				continue
			}

			w1, w2, w3 := r.barycoord(x, y, t1.Pos, t2.Pos, t3.Pos)

			// Is inside triangle?
			if w1 < 0 || w2 < 0 || w3 < 0 {
				continue
			}

			// Z-test
			z := w1*t1.Pos.Z + w2*t2.Pos.Z + w3*t3.Pos.Z
			if !r.shadowDepthTest(index, x, y, z) {
				continue
			}

			// update shadow map
			idx := x + y*r.width
			r.shadowBufs[index].lock[idx].Lock()
			r.shadowBufs[index].depths[idx] = z
			r.shadowBufs[index].lock[idx].Unlock()
		}
	}
}

func (r *Renderer) shadowDepthTest(index int, x, y int, z float64) bool {
	idx := x + y*r.width
	buf := r.shadowBufs[index]

	buf.lock[idx].Lock()
	defer buf.lock[idx].Unlock()
	return !(z <= buf.depths[idx])
}

func (r *Renderer) shadingVisibility(
	x, y int,
	shadowIdx int,
	info *gInfo,
	uniforms map[string]interface{},
) bool {
	if !r.scene.LightSources[shadowIdx].CastShadow() {
		return true
	}

	matVP := uniforms["matVP"].(math.Matrix)
	matScreenToWorld := uniforms["matScreenToWorld"].(math.Matrix)
	shadowMap := &r.shadowBufs[shadowIdx]

	// transform scrren coordinate to light viewport
	screenCoord := math.NewVector(float64(x), float64(y), info.z, 1).
		Apply(matScreenToWorld).
		Apply(shadowMap.settings.Camera().ViewMatrix()).
		Apply(shadowMap.settings.Camera().ProjMatrix()).
		Apply(matVP).Pos()

	lightX, lightY := int(screenCoord.X), int(screenCoord.Y)
	bufIdx := lightX + lightY*r.width

	shadow := 0
	if bufIdx > 0 && bufIdx < len(shadowMap.depths) {
		shadowZ := shadowMap.depths[bufIdx]
		bias := shadowMap.settings.Bias()
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
