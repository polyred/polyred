// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh

import (
	"image/color"

	"poly.red/buffer"
	"poly.red/geometry/primitive"
	"poly.red/math"
)

var _ Mesh[float32] = &BufferedMesh{}

type AttribType int

const (
	AttribUndefined AttribType = iota
	AttribPosition
	AttribNormal
	AttriTexcoord
	AttribColor
)

var attribNames = map[AttribType]string{
	AttribUndefined: "undefined",
	AttribPosition:  "position",
	AttribNormal:    "normal",
	AttriTexcoord:   "texcoord",
	AttribColor:     "color",
}

func (a AttribType) String() string {
	return attribNames[a]
}

type BufferAttribute struct {
	Stride int
	Values []float32
}

func NewBufferAttrib(stride int, values []float32) *BufferAttribute {
	return &BufferAttribute{stride, values}
}

// BufferedMesh is a dense representation of a surface geometry and
// implements the Mesh interface.
type BufferedMesh struct {
	ibo   buffer.IndexBuffer
	vbo   buffer.VertexBuffer
	attrs map[AttribType]*BufferAttribute

	tris []*primitive.Triangle
	aabb *primitive.AABB
}

func NewBufferedMesh() *BufferedMesh {
	bm := &BufferedMesh{
		attrs: map[AttribType]*BufferAttribute{
			AttribPosition: nil,
			AttribNormal:   nil,
			AttriTexcoord:  nil,
			AttribColor:    nil,
		},
	}
	return bm
}

func (bm *BufferedMesh) SetIndexBuffer(ibo buffer.IndexBuffer) { bm.ibo = ibo }
func (bm *BufferedMesh) SetAttribute(name AttribType, attribute *BufferAttribute) {
	bm.attrs[name] = attribute
}
func (bm *BufferedMesh) GetAttribute(name AttribType) *BufferAttribute { return bm.attrs[name] }

func (bm *BufferedMesh) AABB() primitive.AABB {
	if bm.aabb == nil {
		min := math.NewVec3[float32](math.MaxFloat32, math.MaxFloat32, math.MaxFloat32)
		max := math.NewVec3[float32](-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32)
		attr := bm.GetAttribute(AttribPosition)
		for _, idx := range bm.ibo {
			x := attr.Values[attr.Stride*int(idx)+0]
			y := attr.Values[attr.Stride*int(idx)+1]
			z := attr.Values[attr.Stride*int(idx)+2]
			min.X = math.Min(min.X, x)
			min.Y = math.Min(min.Y, y)
			min.Z = math.Min(min.Z, z)
			max.X = math.Max(max.X, x)
			max.Y = math.Max(max.Y, y)
			max.Z = math.Max(max.Z, z)
		}
		bm.aabb = &primitive.AABB{Min: min, Max: max}
	}
	return primitive.AABB{Min: bm.aabb.Min, Max: bm.aabb.Max}
}

func (bm *BufferedMesh) Normalize() {
	aabb := bm.AABB()
	center := aabb.Min.Add(aabb.Max).Scale(0.5, 0.5, 0.5)
	radius := aabb.Max.Sub(aabb.Min).Len() / 2
	fac := 1 / radius

	// scale all vertices
	attr := bm.GetAttribute(AttribPosition)
	for _, vIndex := range bm.ibo {
		x := attr.Values[attr.Stride*int(vIndex)+0]
		y := attr.Values[attr.Stride*int(vIndex)+1]
		z := attr.Values[attr.Stride*int(vIndex)+2]
		v := math.NewVec4(x, y, z, 1).Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac, 1)
		attr.Values[attr.Stride*int(vIndex)+0] = v.X
		attr.Values[attr.Stride*int(vIndex)+1] = v.Y
		attr.Values[attr.Stride*int(vIndex)+2] = v.Z
	}

	// update AABB after scaling
	min := aabb.Min.Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac)
	max := aabb.Max.Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac)
	bm.aabb = &primitive.AABB{Min: min, Max: max}
}

func (bm *BufferedMesh) Triangles() []*primitive.Triangle {
	if bm.tris != nil {
		return bm.tris
	}

	attrPos := bm.GetAttribute(AttribPosition)
	attrNor := bm.GetAttribute(AttribNormal)
	attrColor := bm.GetAttribute(AttribColor)
	attrUV := bm.GetAttribute(AttriTexcoord)
	tris := []*primitive.Triangle{}

	for i := 0; i < len(bm.ibo); i += 3 {
		var px, py, pz, nx, ny, nz, u, v float32
		var cr, cb, cg, ca uint8
		px = attrPos.Values[attrPos.Stride*bm.ibo[i]+0]
		py = attrPos.Values[attrPos.Stride*bm.ibo[i]+1]
		pz = attrPos.Values[attrPos.Stride*bm.ibo[i]+2]
		if attrNor != nil {
			nx = attrNor.Values[attrNor.Stride*bm.ibo[i]+0]
			ny = attrNor.Values[attrNor.Stride*bm.ibo[i]+1]
			nz = attrNor.Values[attrNor.Stride*bm.ibo[i]+2]
		}
		if attrColor != nil {
			cr = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i]+0] * 0xff)
			cb = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i]+1] * 0xff)
			cg = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i]+2] * 0xff)
			ca = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i]+3] * 0xff)
		}
		if attrUV != nil {
			u = attrUV.Values[attrUV.Stride*bm.ibo[i]+0]
			v = attrUV.Values[attrUV.Stride*bm.ibo[i]+1]
		}
		v1 := primitive.NewVertex(
			primitive.Pos(math.NewVec4(px, py, pz, 1)),
			primitive.Nor(math.NewVec4(nx, ny, nz, 0)),
			primitive.Col(color.RGBA{cr, cb, cg, ca}),
			primitive.UV(math.NewVec2(u, v)),
		)

		px = attrPos.Values[attrPos.Stride*bm.ibo[i+1]+0]
		py = attrPos.Values[attrPos.Stride*bm.ibo[i+1]+1]
		pz = attrPos.Values[attrPos.Stride*bm.ibo[i+1]+2]
		if attrNor != nil {
			nx = attrNor.Values[attrNor.Stride*bm.ibo[i+1]+0]
			ny = attrNor.Values[attrNor.Stride*bm.ibo[i+1]+1]
			nz = attrNor.Values[attrNor.Stride*bm.ibo[i+1]+2]
		}
		if attrColor != nil {
			cr = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i+1]+0] * 0xff)
			cb = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i+1]+1] * 0xff)
			cg = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i+1]+2] * 0xff)
			ca = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i+1]+3] * 0xff)
		}
		if attrUV != nil {
			u = attrUV.Values[attrUV.Stride*bm.ibo[i+1]+0]
			v = attrUV.Values[attrUV.Stride*bm.ibo[i+1]+1]
		}
		v2 := primitive.NewVertex(
			primitive.Pos(math.NewVec4(px, py, pz, 1)),
			primitive.Nor(math.NewVec4(nx, ny, nz, 0)),
			primitive.Col(color.RGBA{cr, cb, cg, ca}),
			primitive.UV(math.NewVec2(u, v)),
		)

		px = attrPos.Values[attrPos.Stride*bm.ibo[i+2]+0]
		py = attrPos.Values[attrPos.Stride*bm.ibo[i+2]+1]
		pz = attrPos.Values[attrPos.Stride*bm.ibo[i+2]+2]
		if attrNor != nil {
			nx = attrNor.Values[attrNor.Stride*bm.ibo[i+2]+0]
			ny = attrNor.Values[attrNor.Stride*bm.ibo[i+2]+1]
			nz = attrNor.Values[attrNor.Stride*bm.ibo[i+2]+2]
		}
		if attrColor != nil {
			cr = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i+2]+0] * 0xff)
			cb = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i+2]+1] * 0xff)
			cg = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i+2]+2] * 0xff)
			ca = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i+2]+3] * 0xff)
		}
		if attrUV != nil {
			u = attrUV.Values[attrUV.Stride*bm.ibo[i+2]+0]
			v = attrUV.Values[attrUV.Stride*bm.ibo[i+2]+1]
		}
		v3 := primitive.NewVertex(
			primitive.Pos(math.NewVec4(px, py, pz, 1)),
			primitive.Nor(math.NewVec4(nx, ny, nz, 0)),
			primitive.Col(color.RGBA{cr, cb, cg, ca}),
			primitive.UV(math.NewVec2(u, v)),
		)

		tris = append(tris, &primitive.Triangle{V1: v1, V2: v2, V3: v3})
	}
	bm.tris = tris
	return tris
}

func (bm *BufferedMesh) IndexBuffer() buffer.IndexBuffer { return bm.ibo }
func (bm *BufferedMesh) VertexBuffer() buffer.VertexBuffer {
	if bm.vbo != nil {
		return bm.vbo
	}

	attrPos := bm.GetAttribute(AttribPosition)
	attrNor := bm.GetAttribute(AttribNormal)
	attrColor := bm.GetAttribute(AttribColor)
	attrUV := bm.GetAttribute(AttriTexcoord)

	var px, py, pz, nx, ny, nz, u, v float32
	var cr, cb, cg, ca uint8

	bm.vbo = make([]*primitive.Vertex, len(bm.ibo))
	for i := 0; i < len(bm.ibo); i++ {
		px = attrPos.Values[attrPos.Stride*bm.ibo[i]+0]
		py = attrPos.Values[attrPos.Stride*bm.ibo[i]+1]
		pz = attrPos.Values[attrPos.Stride*bm.ibo[i]+2]
		if attrNor != nil {
			nx = attrNor.Values[attrNor.Stride*bm.ibo[i]+0]
			ny = attrNor.Values[attrNor.Stride*bm.ibo[i]+1]
			nz = attrNor.Values[attrNor.Stride*bm.ibo[i]+2]
		}
		if attrColor != nil {
			cr = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i]+0] * 0xff)
			cb = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i]+1] * 0xff)
			cg = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i]+2] * 0xff)
			ca = uint8(attrColor.Values[attrColor.Stride*bm.ibo[i]+3] * 0xff)
		}
		if attrUV != nil {
			u = attrUV.Values[attrUV.Stride*bm.ibo[i]+0]
			v = attrUV.Values[attrUV.Stride*bm.ibo[i]+1]
		}
		bm.vbo[i] = primitive.NewVertex(
			primitive.Pos(math.NewVec4(px, py, pz, 1)),
			primitive.Nor(math.NewVec4(nx, ny, nz, 0)),
			primitive.Col(color.RGBA{cr, cb, cg, ca}),
			primitive.UV(math.NewVec2(u, v)),
		)
	}
	return bm.vbo
}
