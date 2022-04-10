// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package model

import (
	"fmt"
	"path/filepath"

	"poly.red/buffer"
	"poly.red/color"
	"poly.red/geometry"
	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/internal/cache"
	"poly.red/internal/imageutil"
	"poly.red/material"
	"poly.red/math"
	"poly.red/model/obj"
	"poly.red/scene"
)

func MustLoad(path string) *scene.Group {
	m, err := Load(path)
	if err != nil {
		panic(fmt.Errorf("mesh: cannot load a given mesh: %w", err))
	}
	return m
}

func Load(path string) (*scene.Group, error) {
	switch filepath.Ext(path) {
	case ".obj":
		return LoadObj(path)
	default:
		panic("mesh: unsupported format")
	}
}

// LoadObjAs loads a .obj file to a Mesh object.
func LoadObj(path string) (*scene.Group, error) {
	f, err := obj.Load(path)
	if err != nil {
		return nil, fmt.Errorf("mesh: failed to open file %s: %w", path, err)
	}

	return loadObjScene(f)
}

func loadObjScene(f *obj.File) (*scene.Group, error) {
	g := scene.NewGroup()

	// Create all materials.
	ms := map[string]material.Material{}
	for name, m := range f.Materials {
		opts := []material.BlinnPhongOption{
			material.Kdiff(m.Diffuse),
			material.Kspec(m.Specular),
			material.Shininess(m.Shininess),
			material.FlatShading(false),
			material.AmbientOcclusion(false),
			material.ReceiveShadow(false),
		}
		if m.MapKd != "" {
			opts = append(opts, material.Texture(buffer.NewTexture(
				buffer.TextureImage(
					// Note: this might be problematic when the MapKd is a windows \ separated path.
					imageutil.MustLoadImage(filepath.Join(filepath.Clean(f.MtlDir), filepath.Clean(m.MapKd)),
						imageutil.GammaCorrect(true),
					)),
				buffer.TextureIsoMipmap(true),
			)))
		} else {
			opts = append(opts, material.Texture(buffer.NewUniformTexture(color.Blue)))
		}

		mat := material.NewBlinnPhong(opts...)
		cache.Set(mat.ID(), mat)
		ms[name] = mat
	}

	for i := range f.Objs {
		var tris []*primitive.Triangle
		for _, face := range f.Objs[i].Faces {
			materialId := uint64(0)
			if mat, ok := ms[face.Material]; ok {
				materialId = mat.ID()
			}
			t := &primitive.Triangle{
				V1:         primitive.NewVertex(),
				V2:         primitive.NewVertex(),
				V3:         primitive.NewVertex(),
				MaterialID: materialId,
			}
			vs := face.Vertices
			t.V1.Pos = math.NewVec4(f.Vertices[3*vs[0]+0], f.Vertices[3*vs[0]+1], f.Vertices[3*vs[0]+2], 1)
			t.V2.Pos = math.NewVec4(f.Vertices[3*vs[1]+0], f.Vertices[3*vs[1]+1], f.Vertices[3*vs[1]+2], 1)
			t.V3.Pos = math.NewVec4(f.Vertices[3*vs[2]+0], f.Vertices[3*vs[2]+1], f.Vertices[3*vs[2]+2], 1)

			vs = face.Normals
			if len(f.Normals) > 0 {
				t.V1.Nor = math.NewVec4(f.Normals[3*vs[0]], f.Normals[3*vs[0]+1], f.Normals[3*vs[0]+2], 0)
				t.V2.Nor = math.NewVec4(f.Normals[3*vs[1]], f.Normals[3*vs[1]+1], f.Normals[3*vs[1]+2], 0)
				t.V3.Nor = math.NewVec4(f.Normals[3*vs[2]], f.Normals[3*vs[2]+1], f.Normals[3*vs[2]+2], 0)
				if t.V1.Nor.IsZero() {
					t.V1.Nor = t.Normal()
				}
				if t.V2.Nor.IsZero() {
					t.V1.Nor = t.Normal()
				}
				if t.V3.Nor.IsZero() {
					t.V1.Nor = t.Normal()
				}
			}

			vs = face.Uvs
			if len(f.Uvs) > 0 {
				t.V1.UV = math.NewVec2(f.Uvs[2*vs[0]], f.Uvs[2*vs[0]+1])
				t.V2.UV = math.NewVec2(f.Uvs[2*vs[1]], f.Uvs[2*vs[1]+1])
				t.V3.UV = math.NewVec2(f.Uvs[2*vs[2]], f.Uvs[2*vs[2]+1])
			}

			t.V1.Col = color.FromValue[float32](1, 1, 1, 1)
			t.V2.Col = color.FromValue[float32](1, 1, 1, 1)
			t.V3.Col = color.FromValue[float32](1, 1, 1, 1)
			tris = append(tris, t)
		}

		geom := geometry.NewWith(mesh.NewTriangleMesh(tris), nil)
		g.Add(geom)
	}
	return g, nil
}
