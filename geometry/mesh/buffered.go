// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh

import (
	"image/color"

	"poly.red/geometry/primitive"
	"poly.red/material"
	"poly.red/math"
	"poly.red/object"
)

var _ Mesh = &BufferedMesh{}

// AttrName is a string-based name of a vertex attribute.
type AttrName string

var (
	AttrPos AttrName = "position"
	AttrNor AttrName = "normal"
	AttrUV  AttrName = "uv"
	AttrCol AttrName = "color"
)

// BufferAttr is a dense representation of a attribute buffer.
type BufferAttr struct {
	// Stride is the stride of the value buffer
	Stride int
	// Values contains all values of the attribute buffer
	Values []float64
}

// NewBufferAttr creates a new attribute buffer.
func NewBufferAttr(stride int, values []float64) *BufferAttr {
	return &BufferAttr{
		stride, values,
	}
}

// BufferedMesh is a dense representation of a surface geometry and
// implements the Mesh interface.
type BufferedMesh struct {
	vertIdx    []uint64
	attributes map[AttrName]*BufferAttr
	aabb       *primitive.AABB
	material   material.Material

	math.TransformContext
}

// NewBufferedMesh returns an empty BufferedMesh.
func NewBufferedMesh() *BufferedMesh {
	bm := &BufferedMesh{
		attributes: map[AttrName]*BufferAttr{
			AttrPos: nil,
			AttrNor: nil,
			AttrUV:  nil,
			AttrCol: nil,
		},
	}
	bm.ResetContext()
	return bm
}

// SetVertexIndex sets the vertex indices for the given buffered mesh.
func (bm *BufferedMesh) SetVertexIndex(vertIdx []uint64) {
	bm.vertIdx = vertIdx
}

// SetAttr sets a given attribute buffer to the given buffered mesh.
func (bm *BufferedMesh) SetAttr(name AttrName, attribute *BufferAttr) {
	bm.attributes[name] = attribute
}

// GetAttr gets the attribute buffer of a given name.
func (bm *BufferedMesh) GetAttr(name AttrName) *BufferAttr {
	return bm.attributes[name]
}

// Type returns object.TypeMesh.
func (bm *BufferedMesh) Type() object.Type {
	return object.TypeMesh
}

// AABB computes the axis-aligned bounding box of the given buffered mesh
// in 3D space.
func (bm *BufferedMesh) AABB() primitive.AABB {
	if bm.aabb == nil {
		min := math.NewVec3(math.MaxFloat64, math.MaxFloat64, math.MaxFloat64)
		max := math.NewVec3(-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64)
		attr := bm.GetAttr(AttrPos)
		for _, vIndex := range bm.vertIdx {
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
	min := bm.aabb.Min.ToVec4(1).Apply(bm.ModelMatrix()).ToVec3()
	max := bm.aabb.Max.ToVec4(1).Apply(bm.ModelMatrix()).ToVec3()
	return primitive.AABB{Min: min, Max: max}
}

// Normalize normalizes the given buffered mesh into a unit cube.
func (bm *BufferedMesh) Normalize() {
	aabb := bm.AABB()
	center := aabb.Min.Add(aabb.Max).Scale(0.5, 0.5, 0.5)
	radius := aabb.Max.Sub(aabb.Min).Len() / 2
	fac := 1 / radius

	// scale all vertices
	attr := bm.GetAttr(AttrPos)
	for _, vIndex := range bm.vertIdx {
		x := attr.Values[attr.Stride*int(vIndex)+0]
		y := attr.Values[attr.Stride*int(vIndex)+1]
		z := attr.Values[attr.Stride*int(vIndex)+2]
		v := math.NewVec4(x, y, z, 1).Apply(bm.ModelMatrix()).Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac, 1)
		attr.Values[attr.Stride*int(vIndex)+0] = v.X
		attr.Values[attr.Stride*int(vIndex)+1] = v.Y
		attr.Values[attr.Stride*int(vIndex)+2] = v.Z
	}

	// update AABB after scaling
	min := aabb.Min.Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac)
	max := aabb.Max.Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac)
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
	return uint64(len(bm.vertIdx) / 3)
}

func (bm *BufferedMesh) Faces(iter func(primitive.Face, material.Material) bool) {

	attrPos := bm.GetAttr(AttrPos)
	attrNor := bm.GetAttr(AttrNor)
	attrColor := bm.GetAttr(AttrCol)
	attrUV := bm.GetAttr(AttrUV)

	for i := 0; i < len(bm.vertIdx); i += 3 {
		var px, py, pz, nx, ny, nz, u, v float64
		var cr, cb, cg, ca uint8
		px = attrPos.Values[bm.vertIdx[i]+0]
		py = attrPos.Values[bm.vertIdx[i]+1]
		pz = attrPos.Values[bm.vertIdx[i]+2]
		if attrNor != nil {
			nx = attrNor.Values[bm.vertIdx[i]+0]
			ny = attrNor.Values[bm.vertIdx[i]+1]
			nz = attrNor.Values[bm.vertIdx[i]+2]
		}
		if attrColor != nil {
			cr = uint8(attrColor.Values[bm.vertIdx[i]+0])
			cb = uint8(attrColor.Values[bm.vertIdx[i]+1])
			cg = uint8(attrColor.Values[bm.vertIdx[i]+2])
			ca = uint8(attrColor.Values[bm.vertIdx[i]+3])
		}
		if attrUV != nil {
			u = attrUV.Values[bm.vertIdx[i]+0]
			v = attrUV.Values[bm.vertIdx[i]+1]
		}
		v1 := primitive.Vertex{
			Pos: math.NewVec4(px, py, pz, 1),
			Nor: math.NewVec3(nx, ny, nz),
			UV:  math.NewVec2(u, v),
			Col: color.RGBA{cr, cb, cg, ca},
		}

		px = attrPos.Values[bm.vertIdx[i+1]+0]
		py = attrPos.Values[bm.vertIdx[i+1]+1]
		pz = attrPos.Values[bm.vertIdx[i+1]+2]
		if attrNor != nil {
			nx = attrNor.Values[bm.vertIdx[i+1]+0]
			ny = attrNor.Values[bm.vertIdx[i+1]+1]
			nz = attrNor.Values[bm.vertIdx[i+1]+2]
		}
		if attrColor != nil {
			cr = uint8(attrColor.Values[bm.vertIdx[i+1]+0])
			cb = uint8(attrColor.Values[bm.vertIdx[i+1]+1])
			cg = uint8(attrColor.Values[bm.vertIdx[i+1]+2])
			ca = uint8(attrColor.Values[bm.vertIdx[i+1]+3])
		}
		if attrUV != nil {
			u = attrUV.Values[bm.vertIdx[i+1]+0]
			v = attrUV.Values[bm.vertIdx[i+1]+1]
		}
		v2 := primitive.Vertex{
			Pos: math.NewVec4(px, py, pz, 1),
			Nor: math.NewVec3(nx, ny, nz),
			UV:  math.NewVec2(u, v),
			Col: color.RGBA{cr, cb, cg, ca},
		}

		px = attrPos.Values[bm.vertIdx[i+2]+0]
		py = attrPos.Values[bm.vertIdx[i+2]+1]
		pz = attrPos.Values[bm.vertIdx[i+2]+2]
		if attrNor != nil {
			nx = attrNor.Values[bm.vertIdx[i+2]+0]
			ny = attrNor.Values[bm.vertIdx[i+2]+1]
			nz = attrNor.Values[bm.vertIdx[i+2]+2]
		}
		if attrColor != nil {
			cr = uint8(attrColor.Values[bm.vertIdx[i+2]+0])
			cb = uint8(attrColor.Values[bm.vertIdx[i+2]+1])
			cg = uint8(attrColor.Values[bm.vertIdx[i+2]+2])
			ca = uint8(attrColor.Values[bm.vertIdx[i+2]+3])
		}
		if attrUV != nil {
			u = attrUV.Values[bm.vertIdx[i+2]+0]
			v = attrUV.Values[bm.vertIdx[i+2]+1]
		}
		v3 := primitive.Vertex{
			Pos: math.NewVec4(px, py, pz, 1),
			Nor: math.NewVec3(nx, ny, nz),
			UV:  math.NewVec2(u, v),
			Col: color.RGBA{cr, cb, cg, ca},
		}

		if !iter(&primitive.Triangle{V: [3]primitive.Vertex{v1, v2, v3}},
			bm.material) {
			return
		}
	}
}

func (bm *BufferedMesh) GetVertexIndex() []uint64 {
	return bm.vertIdx
}

func (bm *BufferedMesh) GetVertexBuffer() []*primitive.Vertex {
	attrPos := bm.GetAttr(AttrPos)
	attrNor := bm.GetAttr(AttrNor)
	attrColor := bm.GetAttr(AttrCol)
	attrUV := bm.GetAttr(AttrUV)

	vs := make([]*primitive.Vertex, len(bm.vertIdx))
	for i := 0; i < len(bm.vertIdx); i += 3 {
		var px, py, pz, nx, ny, nz, u, v float64
		var cr, cb, cg, ca uint8
		px = attrPos.Values[bm.vertIdx[i]+0]
		py = attrPos.Values[bm.vertIdx[i]+1]
		pz = attrPos.Values[bm.vertIdx[i]+2]
		if attrNor != nil {
			nx = attrNor.Values[bm.vertIdx[i]+0]
			ny = attrNor.Values[bm.vertIdx[i]+1]
			nz = attrNor.Values[bm.vertIdx[i]+2]
		}
		if attrColor != nil {
			cr = uint8(attrColor.Values[bm.vertIdx[i]+0])
			cb = uint8(attrColor.Values[bm.vertIdx[i]+1])
			cg = uint8(attrColor.Values[bm.vertIdx[i]+2])
			ca = uint8(attrColor.Values[bm.vertIdx[i]+3])
		}
		if attrUV != nil {
			u = attrUV.Values[bm.vertIdx[i]+0]
			v = attrUV.Values[bm.vertIdx[i]+1]
		}
		vs[i] = &primitive.Vertex{
			Pos: math.NewVec4(px, py, pz, 1),
			Nor: math.NewVec3(nx, ny, nz),
			UV:  math.NewVec2(u, v),
			Col: color.RGBA{cr, cb, cg, ca},
		}

		px = attrPos.Values[bm.vertIdx[i+1]+0]
		py = attrPos.Values[bm.vertIdx[i+1]+1]
		pz = attrPos.Values[bm.vertIdx[i+1]+2]
		if attrNor != nil {
			nx = attrNor.Values[bm.vertIdx[i+1]+0]
			ny = attrNor.Values[bm.vertIdx[i+1]+1]
			nz = attrNor.Values[bm.vertIdx[i+1]+2]
		}
		if attrColor != nil {
			cr = uint8(attrColor.Values[bm.vertIdx[i+1]+0])
			cb = uint8(attrColor.Values[bm.vertIdx[i+1]+1])
			cg = uint8(attrColor.Values[bm.vertIdx[i+1]+2])
			ca = uint8(attrColor.Values[bm.vertIdx[i+1]+3])
		}
		if attrUV != nil {
			u = attrUV.Values[bm.vertIdx[i+1]+0]
			v = attrUV.Values[bm.vertIdx[i+1]+1]
		}
		vs[i+1] = &primitive.Vertex{
			Pos: math.NewVec4(px, py, pz, 1),
			Nor: math.NewVec3(nx, ny, nz),
			UV:  math.NewVec2(u, v),
			Col: color.RGBA{cr, cb, cg, ca},
		}

		px = attrPos.Values[bm.vertIdx[i+2]+0]
		py = attrPos.Values[bm.vertIdx[i+2]+1]
		pz = attrPos.Values[bm.vertIdx[i+2]+2]
		if attrNor != nil {
			nx = attrNor.Values[bm.vertIdx[i+2]+0]
			ny = attrNor.Values[bm.vertIdx[i+2]+1]
			nz = attrNor.Values[bm.vertIdx[i+2]+2]
		}
		if attrColor != nil {
			cr = uint8(attrColor.Values[bm.vertIdx[i+2]+0])
			cb = uint8(attrColor.Values[bm.vertIdx[i+2]+1])
			cg = uint8(attrColor.Values[bm.vertIdx[i+2]+2])
			ca = uint8(attrColor.Values[bm.vertIdx[i+2]+3])
		}
		if attrUV != nil {
			u = attrUV.Values[bm.vertIdx[i+2]+0]
			v = attrUV.Values[bm.vertIdx[i+2]+1]
		}
		vs[i+2] = &primitive.Vertex{
			Pos: math.NewVec4(px, py, pz, 1),
			Nor: math.NewVec3(nx, ny, nz),
			UV:  math.NewVec2(u, v),
			Col: color.RGBA{cr, cb, cg, ca},
		}
	}
	return vs
}
