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
		panic(err)
	}
	return m
}

func Load(path string) (*scene.Group, error) {
	switch v := filepath.Ext(path); v {
	case ".obj":
		return loadObj(path)
	default:
		return nil, fmt.Errorf("model: unsupported format %v", v)
	}
}

func loadObj(path string) (*scene.Group, error) {
	fi, err := obj.Load(path)
	if err != nil {
		return nil, fmt.Errorf("model: cannot load the given file %v: %w", path, err)
	}

	g := scene.NewGroup()

	// Create all materials.
	ms := map[string]material.Material{}
	for name, mat := range fi.Materials {
		var m material.Material
		switch mat.Illum {
		case 0, 2:
			opts := []material.Option[material.BlinnPhong]{
				material.Diffuse(mat.Diffuse),
				material.Specular(mat.Specular),
				material.Shininess(mat.Shininess),
				material.FlatShading[material.BlinnPhong](false),
				material.AmbientOcclusion[material.BlinnPhong](false),
				material.ReceiveShadow[material.BlinnPhong](false),
			}
			if mat.MapKd != "" {
				opts = append(opts, material.Texture[material.BlinnPhong](buffer.NewTexture(
					buffer.TextureImage(
						// Note: this might be problematic when the MapKd is a windows \ separated path.
						imageutil.MustLoadImage(filepath.Join(filepath.Clean(fi.MtlDir), filepath.Clean(mat.MapKd)),
							imageutil.GammaCorrect(true),
						)),
					buffer.TextureIsoMipmap(true),
				)))
			} else {
				opts = append(opts, material.Texture[material.BlinnPhong](buffer.NewUniformTexture(color.Blue)))
			}

			m = material.NewBlinnPhong(opts...)
		default:
			panic("unsupported illumination model")
		}

		cache.Set(m.ID(), m)
		ms[name] = m
	}

	// Create all mesh objects.
	for i := range fi.Objs {
		var faces []*primitive.Triangle
		for _, face := range fi.Objs[i].Faces {
			materialID := uint64(0)
			if mat, ok := ms[face.Material]; ok {
				materialID = mat.ID()
			}

			var f primitive.Face
			switch len(face.Vertices) {
			case 3:
				f = newTrianglePrimitive(fi, &face, materialID)
			default:
				panic("unsupported")
			}

			faces = append(faces, f.(*primitive.Triangle))
		}

		geom := geometry.NewWith(mesh.NewTriangleMesh(faces), nil)
		g.Add(geom)
	}
	return g, nil
}

func newTrianglePrimitive(f *obj.File, face *obj.Face, materialID uint64) primitive.Face {
	t := &primitive.Triangle{
		V1:         primitive.NewVertex(),
		V2:         primitive.NewVertex(),
		V3:         primitive.NewVertex(),
		MaterialID: materialID,
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
	return t
}
