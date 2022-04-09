// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package model

import (
	"fmt"
	"log"
	"path/filepath"

	"poly.red/buffer"
	"poly.red/color"
	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/internal/cache"
	"poly.red/internal/imageutil"
	"poly.red/material"
	"poly.red/math"
	"poly.red/model/obj"
	"poly.red/scene"
	"poly.red/scene/object"
)

func MustLoadAs[T mesh.Mesh[float32]](path string) *scene.Group {
	m, err := LoadAs[T](path)
	if err != nil {
		panic(fmt.Errorf("mesh: cannot load a given mesh: %w", err))
	}
	return m
}

func LoadAs[T mesh.Mesh[float32]](path string) (*scene.Group, error) {
	switch filepath.Ext(path) {
	case ".obj":
		return LoadObjAs[T](path)
	default:
		panic("mesh: unsupported format")
	}
}

// LoadObjAs loads a .obj file to a Mesh object.
func LoadObjAs[T mesh.Mesh[float32]](path string) (*scene.Group, error) {
	f, err := obj.Load(path)
	if err != nil {
		return nil, fmt.Errorf("mesh: failed to open file %s: %w", path, err)
	}

	var x T
	switch any(x).(type) {
	case *mesh.TriangleMesh:
		return loadObjScene(f)
	default:
		panic("unsupported")
	}
}

func loadObjScene(f *obj.File) (*scene.Group, error) {
	objs := []object.Object[float32]{}
	for _, obj := range f.Objs {
		fmat := f.Materials[obj.Faces[0].Material]
		var mat *material.BlinnPhong
		if fmat.MapKd != "" {
			log.Println("loading: ", filepath.Join(f.MtlDir, fmat.MapKd))
			mat = material.NewBlinnPhong(
				material.Texture(buffer.NewTexture(
					buffer.TextureImage(
						imageutil.MustLoadImage(filepath.Join(f.MtlDir, fmat.MapKd),
							imageutil.GammaCorrect(true),
						)),
					buffer.TextureIsoMipmap(true),
				)),
				material.Kdiff(fmat.Diffuse), material.Kspec(fmat.Specular),
				material.Shininess(fmat.Shininess),
				material.FlatShading(true),
				material.AmbientOcclusion(false),
				material.ReceiveShadow(false),
			)
			cache.Set(mat.Material.ID, mat)
		}

		var tris []*primitive.Triangle
		for _, face := range obj.Faces {
			t := primitive.Triangle{
				V1: primitive.NewVertex(),
				V2: primitive.NewVertex(),
				V3: primitive.NewVertex(),
			}

			vs := face.Vertices
			t.V1.Pos = math.NewVec4(f.Vertices[vs[0]], f.Vertices[vs[0]+1], f.Vertices[vs[0]+2], 1)
			t.V2.Pos = math.NewVec4(f.Vertices[vs[1]], f.Vertices[vs[1]+1], f.Vertices[vs[1]+2], 1)
			t.V3.Pos = math.NewVec4(f.Vertices[vs[2]], f.Vertices[vs[2]+1], f.Vertices[vs[2]+2], 1)

			vs = face.Normals
			t.V1.Nor = math.NewVec4(f.Normals[vs[0]], f.Normals[vs[0]+1], f.Normals[vs[0]+2], 0)
			t.V2.Nor = math.NewVec4(f.Normals[vs[1]], f.Normals[vs[1]+1], f.Normals[vs[1]+2], 0)
			t.V3.Nor = math.NewVec4(f.Normals[vs[2]], f.Normals[vs[2]+1], f.Normals[vs[2]+2], 0)
			if t.V1.Nor.IsZero() {
				t.V1.Nor = t.Normal()
			}
			if t.V2.Nor.IsZero() {
				t.V1.Nor = t.Normal()
			}
			if t.V3.Nor.IsZero() {
				t.V1.Nor = t.Normal()
			}

			vs = face.Uvs
			t.V1.UV = math.NewVec2(f.Vertices[vs[0]], f.Uvs[vs[0]+1])
			t.V2.UV = math.NewVec2(f.Vertices[vs[1]], f.Uvs[vs[1]+1])
			t.V3.UV = math.NewVec2(f.Vertices[vs[2]], f.Uvs[vs[2]+1])

			t.V1.Col = color.FromValue[float32](1, 1, 1, 1)
			t.V2.Col = color.FromValue[float32](1, 1, 1, 1)
			t.V3.Col = color.FromValue[float32](1, 1, 1, 1)

			if mat != nil {
				t.MaterialId = mat.Material.ID
			}
			tris = append(tris, &t)
		}
		objs = append(objs, mesh.NewTriangleMesh(tris))
	}
	return scene.NewScene().Add(objs...), nil
}
