// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package ddd

import (
	"fmt"
	"image"
	_ "image/jpeg" // for jpg encoding
	"os"
)

// Vertex is a vertex that contains the necessary information for
// describing a mesh.
type Vertex struct {
	Position Vector
	Color    Vector
	UV       Vector
	Normal   Vector
}

// Triangle is a triangle that contains three vertices.
type Triangle struct {
	v1, v2, v3 Vertex
}

// TriangleMesh implements a triangular mesh.
type TriangleMesh struct {
	triangles       []*Triangle
	scaleMatrix     Matrix
	translateMatrix Matrix
	texture         *Texture
	modelMatrix     Matrix
	normalMatrix    Matrix
}

// NewTriangleMesh returns a triangular mesh.
func NewTriangleMesh(ts []*Triangle) *TriangleMesh {
	return &TriangleMesh{
		triangles:       ts,
		scaleMatrix:     IdentityMatrix,
		translateMatrix: IdentityMatrix,
		modelMatrix:     IdentityMatrix,
		normalMatrix:    IdentityMatrix,
	}
}

// SetScale sets the scale matrix.
func (m *TriangleMesh) SetScale(v Vector) {
	m.scaleMatrix = Matrix{
		v.X, 0, 0, 0,
		0, v.Y, 0, 0,
		0, 0, v.Z, 0,
		0, 0, 0, 1,
	}
}

// SetTranslate sets the translate matrix.
func (m *TriangleMesh) SetTranslate(v Vector) {
	m.translateMatrix = Matrix{
		1, 0, 0, v.X,
		0, 1, 0, v.Y,
		0, 0, 1, v.Z,
		0, 0, 0, 1,
	}
}

// Texture is a texture
type Texture struct {
	Shininess float64
	data      image.Image
	width     int
	height    int
}

// SetTexture sets the texture of a given mesh.
func (m *TriangleMesh) SetTexture(path string, shininess float64) error {
	tex, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot read texture: %v", err)
	}
	defer tex.Close()

	data, _, err := image.Decode(tex)
	if err != nil {
		return fmt.Errorf("decode texture error: %v", err)
	}

	m.texture = &Texture{
		Shininess: shininess,
		data:      data,
		width:     data.Bounds().Max.X,
		height:    data.Bounds().Max.Y,
	}

	return nil
}
