// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package io

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"changkun.de/x/ddd/color"
	"changkun.de/x/ddd/geometry"
	"changkun.de/x/ddd/geometry/primitive"
	"changkun.de/x/ddd/math"
)

// LoadOBJ loads a .obj file to a TriangleMesh object
func LoadOBJ(data io.Reader) (geometry.Mesh, error) {

	vs := make([]math.Vector, 1)
	vts := make([]math.Vector, 1, 1024)
	vns := make([]math.Vector, 1, 1024)

	var tris []*primitive.Triangle

	s := bufio.NewScanner(data)
	for s.Scan() {
		l := s.Text()
		fields := strings.Fields(l)
		if len(fields) == 0 { // nothing to read
			continue
		}
		k := fields[0]
		args := fields[1:]
		switch k {
		case "v":
			coord := parseFloats(args)
			vs = append(vs, math.NewVector(coord[0], coord[1], coord[2], 1))
		case "vt":
			coord := parseFloats(args)
			vts = append(vts, math.NewVector(coord[0], coord[1], 0, 1))
		case "vn":
			coord := parseFloats(args)
			vns = append(vns, math.NewVector(coord[0], coord[1], coord[2], 0))
		case "f":
			fvs := make([]int, len(args))
			fvts := make([]int, len(args))
			fvns := make([]int, len(args))
			for i, arg := range args {
				v := strings.Split(arg+"//", "/")
				fvs[i] = parseIndex(v[0], len(vs))
				fvts[i] = parseIndex(v[1], len(vts))
				fvns[i] = parseIndex(v[2], len(vns))
			}
			for i := 1; i < len(fvs)-1; i++ {
				i1, i2, i3 := 0, i, i+1
				t := primitive.Triangle{}
				t.V1.Pos = vs[fvs[i1]]
				t.V2.Pos = vs[fvs[i2]]
				t.V3.Pos = vs[fvs[i3]]
				t.V1.Nor = vns[fvns[i1]]
				t.V2.Nor = vns[fvns[i2]]
				t.V3.Nor = vns[fvns[i3]]
				if t.V1.Nor.IsZero() {
					t.V1.Nor = t.FaceNormal()
				}
				if t.V2.Nor.IsZero() {
					t.V1.Nor = t.FaceNormal()
				}
				if t.V3.Nor.IsZero() {
					t.V1.Nor = t.FaceNormal()
				}
				t.V1.UV = vts[fvts[i1]]
				t.V2.UV = vts[fvts[i2]]
				t.V3.UV = vts[fvts[i3]]
				t.V1.Col = color.FromHex("#ffffff")
				t.V2.Col = color.FromHex("#ffffff")
				t.V3.Col = color.FromHex("#ffffff")
				tris = append(tris, &t)
			}
		}
	}
	return geometry.NewTriangleSoup(tris), s.Err()
}

func parseFloats(items []string) []float64 {
	result := make([]float64, len(items))
	for i, item := range items {
		f, _ := strconv.ParseFloat(item, 64)
		result[i] = f
	}
	return result
}

func parseIndex(value string, length int) int {
	parsed, _ := strconv.ParseInt(value, 0, 0)
	n := int(parsed)
	if n < 0 {
		n += length
	}
	return n
}
