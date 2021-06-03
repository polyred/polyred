// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package geometry

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"changkun.de/x/ddd/math"
)

// LoadOBJ loads a .obj file to a TriangleMesh object
func LoadOBJ(data io.Reader) (*TriangleMesh, error) {

	vs := make([]math.Vector, 1, 1024)
	vts := make([]math.Vector, 1, 1024)
	vns := make([]math.Vector, 1, 1024)

	var tris []*Triangle

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
		case "v": // vertices
			coord := parseFloats(args)
			vs = append(vs, math.NewVector(coord[0], coord[1], coord[2], 1))
		case "vt": // uv texture coords
			coord := parseFloats(args)
			vts = append(vts, math.NewVector(coord[0], coord[1], 0, 1))
		case "vn": // vertex normals
			coord := parseFloats(args)
			vns = append(vns, math.NewVector(coord[0], coord[1], coord[2], 0))
		case "f": // faces
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
				t := Triangle{}
				t.V1.Position = vs[fvs[i1]]
				t.V2.Position = vs[fvs[i2]]
				t.V3.Position = vs[fvs[i3]]
				t.V1.Normal = vns[fvns[i1]]
				t.V2.Normal = vns[fvns[i2]]
				t.V3.Normal = vns[fvns[i3]]
				t.V1.UV = vts[fvts[i1]]
				t.V2.UV = vts[fvts[i2]]
				t.V3.UV = vts[fvts[i3]]
				tris = append(tris, &t)
			}
		}
	}
	return NewTriangleMesh(tris), s.Err()
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
