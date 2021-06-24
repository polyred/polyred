// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package geometry

import (
	"image/color"

	"changkun.de/x/polyred/geometry/primitive"
	"changkun.de/x/polyred/material"
	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/object"
)

var _ Mesh = &BufferedMesh{}

type AttributeName string

var (
	AttributePos   AttributeName = "position"
	AttributeNor   AttributeName = "normal"
	AttributeUV    AttributeName = "uv"
	AttributeColor AttributeName = "color"
)

type BufferAttribute struct {
	Stride int
	Values []float64
}

// BufferedMesh is a dense representation of a surface geometry and
// implements the Mesh interface.
type BufferedMesh struct {
	vbo        []int64
	attributes map[AttributeName]*BufferAttribute
	aabb       *primitive.AABB
	material   material.Material

	math.TransformContext
}

func NewBufferedMesh() *BufferedMesh {
	bm := &BufferedMesh{
		attributes: map[AttributeName]*BufferAttribute{
			AttributePos:   nil,
			AttributeNor:   nil,
			AttributeUV:    nil,
			AttributeColor: nil,
		},
	}
	bm.ResetContext()
	return bm
}

func (bm *BufferedMesh) SetVertexBuffer(vbo []int64) {
	bm.vbo = vbo
}

func (bm *BufferedMesh) SetAttribute(name AttributeName, attribute *BufferAttribute) {
	bm.attributes[name] = attribute
}

func (bm *BufferedMesh) GetAttribute(name AttributeName) *BufferAttribute {
	return bm.attributes[name]
}

func (bm *BufferedMesh) Type() object.Type {
	return object.TypeMesh
}

func (bm *BufferedMesh) AABB() primitive.AABB {
	if bm.aabb == nil {
		min := math.NewVector(math.MaxFloat64, math.MaxFloat64, math.MaxFloat64, 1)
		max := math.NewVector(-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64, 1)
		attr := bm.GetAttribute(AttributePos)
		for _, vIndex := range bm.vbo {
			x := attr.Values[attr.Stride*int(vIndex)+0]
			y := attr.Values[attr.Stride*int(vIndex)+1]
			z := attr.Values[attr.Stride*int(vIndex)+2]
			min.X = math.Min(min.X, x)
			min.Y = math.Min(min.Y, y)
			min.Z = math.Min(min.Z, z)
			max.X = math.Max(max.X, x)
			max.Y = math.Max(max.Y, y)
			max.Z = math.Max(max.Z, z)
		}
		bm.aabb = &primitive.AABB{Min: min, Max: max}
	}
	min := bm.aabb.Min.Apply(bm.ModelMatrix())
	max := bm.aabb.Max.Apply(bm.ModelMatrix())
	return primitive.AABB{Min: min, Max: max}
}

func (bm *BufferedMesh) Normalize() {
	aabb := bm.AABB()
	center := aabb.Min.Add(aabb.Max).Pos()
	radius := aabb.Max.Sub(aabb.Min).Len() / 2
	fac := 1 / radius

	// scale all vertices
	attr := bm.GetAttribute(AttributePos)
	for _, vIndex := range bm.vbo {
		x := attr.Values[attr.Stride*int(vIndex)+0]
		y := attr.Values[attr.Stride*int(vIndex)+1]
		z := attr.Values[attr.Stride*int(vIndex)+2]
		v := math.NewVector(x, y, z, 1).Apply(bm.ModelMatrix()).Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac, 1)
		attr.Values[attr.Stride*int(vIndex)+0] = v.X
		attr.Values[attr.Stride*int(vIndex)+1] = v.Y
		attr.Values[attr.Stride*int(vIndex)+2] = v.Z
	}

	// update AABB after scaling
	min := aabb.Min.Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac, 1)
	max := aabb.Max.Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac, 1)
	bm.aabb = &primitive.AABB{Min: min, Max: max}
	bm.ResetContext()
}

func (bm *BufferedMesh) GetMaterial() material.Material {
	return bm.material
}

func (bm *BufferedMesh) SetMaterial(mat material.Material) {
	bm.material = mat
}

func (bm *BufferedMesh) NumTriangles() uint64 {
	return uint64(len(bm.vbo) / 3)
}

func (bm *BufferedMesh) Faces(iter func(primitive.Face, material.Material) bool) {

	attrPos := bm.GetAttribute(AttributePos)
	attrNor := bm.GetAttribute(AttributeNor)
	attrColor := bm.GetAttribute(AttributeColor)
	attrUV := bm.GetAttribute(AttributeUV)

	for i := 0; i < len(bm.vbo); i += 3 {
		var px, py, pz, nx, ny, nz, u, v float64
		var cr, cb, cg, ca uint8
		px = attrPos.Values[bm.vbo[i]+0]
		py = attrPos.Values[bm.vbo[i]+1]
		pz = attrPos.Values[bm.vbo[i]+2]
		if attrNor != nil {
			nx = attrNor.Values[bm.vbo[i]+0]
			ny = attrNor.Values[bm.vbo[i]+1]
			nz = attrNor.Values[bm.vbo[i]+2]
		}
		if attrColor != nil {
			cr = uint8(attrColor.Values[bm.vbo[i]+0])
			cb = uint8(attrColor.Values[bm.vbo[i]+1])
			cg = uint8(attrColor.Values[bm.vbo[i]+2])
			ca = uint8(attrColor.Values[bm.vbo[i]+3])
		}
		if attrUV != nil {
			u = attrUV.Values[bm.vbo[i]+0]
			v = attrUV.Values[bm.vbo[i]+1]
		}
		v1 := primitive.Vertex{
			Pos: math.NewVector(px, py, pz, 1),
			Nor: math.NewVector(nx, ny, nz, 0),
			UV:  math.NewVector(u, v, 0, 1),
			Col: color.RGBA{cr, cb, cg, ca},
		}

		px = attrPos.Values[bm.vbo[i+1]+0]
		py = attrPos.Values[bm.vbo[i+1]+1]
		pz = attrPos.Values[bm.vbo[i+1]+2]
		if attrNor != nil {
			nx = attrNor.Values[bm.vbo[i+1]+0]
			ny = attrNor.Values[bm.vbo[i+1]+1]
			nz = attrNor.Values[bm.vbo[i+1]+2]
		}
		if attrColor != nil {
			cr = uint8(attrColor.Values[bm.vbo[i+1]+0])
			cb = uint8(attrColor.Values[bm.vbo[i+1]+1])
			cg = uint8(attrColor.Values[bm.vbo[i+1]+2])
			ca = uint8(attrColor.Values[bm.vbo[i+1]+3])
		}
		if attrUV != nil {
			u = attrUV.Values[bm.vbo[i+1]+0]
			v = attrUV.Values[bm.vbo[i+1]+1]
		}
		v2 := primitive.Vertex{
			Pos: math.NewVector(px, py, pz, 1),
			Nor: math.NewVector(nx, ny, nz, 0),
			UV:  math.NewVector(u, v, 0, 1),
			Col: color.RGBA{cr, cb, cg, ca},
		}

		px = attrPos.Values[bm.vbo[i+2]+0]
		py = attrPos.Values[bm.vbo[i+2]+1]
		pz = attrPos.Values[bm.vbo[i+2]+2]
		if attrNor != nil {
			nx = attrNor.Values[bm.vbo[i+2]+0]
			ny = attrNor.Values[bm.vbo[i+2]+1]
			nz = attrNor.Values[bm.vbo[i+2]+2]
		}
		if attrColor != nil {
			cr = uint8(attrColor.Values[bm.vbo[i+2]+0])
			cb = uint8(attrColor.Values[bm.vbo[i+2]+1])
			cg = uint8(attrColor.Values[bm.vbo[i+2]+2])
			ca = uint8(attrColor.Values[bm.vbo[i+2]+3])
		}
		if attrUV != nil {
			u = attrUV.Values[bm.vbo[i+2]+0]
			v = attrUV.Values[bm.vbo[i+2]+1]
		}
		v3 := primitive.Vertex{
			Pos: math.NewVector(px, py, pz, 1),
			Nor: math.NewVector(nx, ny, nz, 0),
			UV:  math.NewVector(u, v, 0, 1),
			Col: color.RGBA{cr, cb, cg, ca},
		}

		if !iter(&primitive.Triangle{
			V1: v1, V2: v2, V3: v3,
		}, bm.material) {
			return
		}
	}
}
