// Copyright 2022 The Polyred Authors. All rights reserved.
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
	allMats := map[string]material.ID{}
	for name, mat := range fi.Materials {
		var id material.ID
		// Convention: .obj loader will create a nil material if there is no material specified.
		// FIXME: review obj loader logic.
		if mat != nil {
			switch mat.Illum {
			case 0, 2:
				opts := []material.Option{
					material.Name(name),
					material.Diffuse(mat.Diffuse),
					material.Specular(mat.Specular),
					material.Shininess(mat.Shininess),
					material.FlatShading(false),
					material.AmbientOcclusion(false),
					material.ReceiveShadow(false),
				}
				if mat.MapKd == "" {
					opts = append(opts, material.Texture(buffer.NewUniformTexture(color.Blue)))
				} else {
					p := filepath.Join(filepath.Clean(fi.MtlDir), filepath.Clean(mat.MapKd))
					opts = append(opts, material.Texture(buffer.NewTexture(
						buffer.TextureImage(
							// FIXME: this might be problematic when the MapKd is a windows \ separated path.
							imageutil.MustLoadImage(p,
								// FIXME: make sure renderer know about this and do gamma correct if needed.
								imageutil.GammaCorrect(true),
							)),
						buffer.TextureIsoMipmap(true),
					)))
				}

				id = material.NewBlinnPhong(opts...)
			default:
				panic("unsupported illumination model")
			}
		}
		allMats[name] = id
	}

	// Create all mesh objects.
	usedMats := map[int64]string{}
	for i := range fi.Objs {
		var (
			tris     []*primitive.Triangle
			quads    []*primitive.Quad
			polygons []*primitive.Polygon

			// All used materials for the primitives of the current object.
			ids = []material.ID{}
		)

		for _, face := range fi.Objs[i].Faces {
			materialID := int64(0)
			if mat, ok := allMats[face.Material]; ok {
				materialID = int64(mat)
			}
			ids = append(ids, material.ID(materialID))
			usedMats[materialID] = face.Material

			switch len(face.Vertices) {
			case 3:
				tris = append(tris, newTrianglePrimitive(fi, &face, materialID))
			case 4:
				quads = append(quads, newQuadPrimitive(fi, &face, materialID))
			default:
				polygons = append(polygons, newPolygonPrimitive(fi, &face, materialID))
			}

		}

		var m mesh.Mesh
		switch {
		case len(tris) != 0 && len(quads) == 0: // only with triangles
			m = mesh.NewTriangleMesh(tris)
		case len(tris) == 0 && len(quads) != 0: // only with quads
			m = mesh.NewQuadMesh(quads)
		case len(tris) != 0 && len(quads) != 0: // hybrid
			faces := []primitive.Face{}
			for i := range tris {
				faces = append(faces, tris[i])
			}
			for i := range quads {
				faces = append(faces, quads[i])
			}
			for i := range polygons {
				faces = append(faces, polygons[i])
			}
			m = mesh.NewPolygonMesh(faces)
		}

		// All material ID for this current geometry.
		g.Add(geometry.New(m, ids...))
	}

	// Remove used materials from allMats, then free the remaining
	// materials from the material pool (because they are not used)
	// so that we wont run into memory leak.
	for _, name := range usedMats {
		delete(allMats, name)
	}
	for _, id := range allMats {
		material.Del(id)
	}

	return g, nil
}

func newTrianglePrimitive(f *obj.File, face *obj.Face, materialID int64) *primitive.Triangle {
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

func newQuadPrimitive(f *obj.File, face *obj.Face, materialID int64) *primitive.Quad {
	t := &primitive.Quad{
		V1:         primitive.NewVertex(),
		V2:         primitive.NewVertex(),
		V3:         primitive.NewVertex(),
		V4:         primitive.NewVertex(),
		MaterialID: materialID,
	}
	vs := face.Vertices
	t.V1.Pos = math.NewVec4(f.Vertices[3*vs[0]+0], f.Vertices[3*vs[0]+1], f.Vertices[3*vs[0]+2], 1)
	t.V2.Pos = math.NewVec4(f.Vertices[3*vs[1]+0], f.Vertices[3*vs[1]+1], f.Vertices[3*vs[1]+2], 1)
	t.V3.Pos = math.NewVec4(f.Vertices[3*vs[2]+0], f.Vertices[3*vs[2]+1], f.Vertices[3*vs[2]+2], 1)
	t.V4.Pos = math.NewVec4(f.Vertices[3*vs[3]+0], f.Vertices[3*vs[3]+1], f.Vertices[3*vs[3]+2], 1)

	vs = face.Normals
	if len(f.Normals) > 0 {
		t.V1.Nor = math.NewVec4(f.Normals[3*vs[0]], f.Normals[3*vs[0]+1], f.Normals[3*vs[0]+2], 0)
		t.V2.Nor = math.NewVec4(f.Normals[3*vs[1]], f.Normals[3*vs[1]+1], f.Normals[3*vs[1]+2], 0)
		t.V3.Nor = math.NewVec4(f.Normals[3*vs[2]], f.Normals[3*vs[2]+1], f.Normals[3*vs[2]+2], 0)
		t.V4.Nor = math.NewVec4(f.Normals[3*vs[3]], f.Normals[3*vs[3]+1], f.Normals[3*vs[3]+2], 0)
	}

	vs = face.Uvs
	if len(f.Uvs) > 0 {
		t.V1.UV = math.NewVec2(f.Uvs[2*vs[0]], f.Uvs[2*vs[0]+1])
		t.V2.UV = math.NewVec2(f.Uvs[2*vs[1]], f.Uvs[2*vs[1]+1])
		t.V3.UV = math.NewVec2(f.Uvs[2*vs[2]], f.Uvs[2*vs[2]+1])
		t.V4.UV = math.NewVec2(f.Uvs[2*vs[3]], f.Uvs[2*vs[3]+1])
	}

	t.V1.Col = color.FromValue[float32](1, 1, 1, 1)
	t.V2.Col = color.FromValue[float32](1, 1, 1, 1)
	t.V3.Col = color.FromValue[float32](1, 1, 1, 1)
	t.V4.Col = color.FromValue[float32](1, 1, 1, 1)
	return t
}

func newPolygonPrimitive(f *obj.File, face *obj.Face, materialID int64) *primitive.Polygon {
	t := &primitive.Polygon{
		Verts:      make([]*primitive.Vertex, len(face.Vertices)),
		MaterialID: materialID,
	}
	for i := range t.Verts {
		t.Verts[i] = primitive.NewVertex()

		vs := face.Vertices
		t.Verts[i].Pos = math.NewVec4(f.Vertices[3*vs[i%len(vs)]+0], f.Vertices[3*vs[i%len(vs)]+1], f.Vertices[3*vs[i%len(vs)]+2], 1)

		vs = face.Normals
		if len(f.Normals) > 0 {
			t.Verts[i].Nor = math.NewVec4(f.Normals[3*vs[i%len(vs)]], f.Normals[3*vs[i%len(vs)]+1], f.Normals[3*vs[i%len(vs)]+2], 0)
		}

		vs = face.Uvs
		if len(f.Uvs) > 0 {
			t.Verts[i].UV = math.NewVec2(f.Uvs[2*vs[i%len(vs)]], f.Uvs[2*vs[i%len(vs)]+1])
		}

		t.Verts[i].Col = color.FromValue[float32](1, 1, 1, 1)
	}

	return t
}
