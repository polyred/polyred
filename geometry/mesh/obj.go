// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"poly.red/color"
	"poly.red/geometry/primitive"
	"poly.red/math"
)

type rawdata struct {
	vIdx []int
	tIdx []int
	nIdx []int
	v    []float64
	t    []float64
	n    []float64
}

func LoadOBJ(path string) (Mesh, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("mesh: failed to load obj: %w", err)
	}
	defer f.Close()

	r := rawdata{
		vIdx: make([]int, 0, 1024),
		tIdx: make([]int, 0, 1024),
		nIdx: make([]int, 0, 1024),
		v:    make([]float64, 0, 1024*3),
		t:    make([]float64, 0, 1024*2),
		n:    make([]float64, 0, 1024*3),
	}

	s := bufio.NewScanner(f)
	fIdx := 0
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
			r.v = append(r.v, parseFloats(args)...)
		case "vt":
			r.t = append(r.t, parseFloats(args)...)
		case "vn":
			r.n = append(r.n, parseFloats(args)...)
		case "f":
			fvs := make([]int, len(args))
			fvts := make([]int, len(args))
			fvns := make([]int, len(args))
			for i, arg := range args {
				v := strings.Split(arg+"//", "/")
				fvs[i] = parseIndex(v[0], len(r.v))
				fvts[i] = parseIndex(v[1], len(r.t))
				fvns[i] = parseIndex(v[2], len(r.n))
			}

			// Prase the object into triangle mesh.
			for i := 1; i < len(fvs)-1; i++ {
				i1, i2, i3 := 0, i, i+1
				t := primitive.Triangle{}
				r.vIdx[fIdx+0] = fvs[i1]
				r.vIdx[fIdx+1] = fvs[i2]
				r.vIdx[fIdx+2] = fvs[i3]
				r.nIdx[fIdx+0] = fvns[i1]
				r.nIdx[fIdx+1] = fvns[i2]
				r.nIdx[fIdx+2] = fvns[i3]
				if math.NewVec3(
					r.n[fvns[i1]],
					r.n[fvns[i2]],
					r.n[fvns[i3]],
				).IsZero() {
					r.n[fvns[i1]] = 0
					r.n[fvns[i2]] = 0
					r.n[fvns[i3]] = 0
				}
				if t.V[0].Nor.IsZero() {
					t.V[0].Nor = t.Normal()
				}
				if t.V[1].Nor.IsZero() {
					t.V[1].Nor = t.Normal()
				}
				if t.V[2].Nor.IsZero() {
					t.V[2].Nor = t.Normal()
				}
				t.V[0].UV = vertTex[fvts[i1]]
				t.V[1].UV = vertTex[fvts[i2]]
				t.V[2].UV = vertTex[fvts[i3]]
				t.V[0].Col = color.FromHex("#ffffff")
				t.V[1].Col = color.FromHex("#ffffff")
				t.V[2].Col = color.FromHex("#ffffff")
				tris = append(tris, &t)
			}
		}
	}

	m := NewBufferedMesh()
	m.SetVertexIndex(vertIdx)
	m.SetAttr(AttrPos, NewBufferAttr(3, vertPos))
	m.SetAttr(AttrUV, NewBufferAttr(2, vertTex))
	m.SetAttr(AttrNor, NewBufferAttr(3, vertNor))
	return m, s.Err()
}

func parseFloats(items []string) []float64 {
	result := make([]float32, len(items))
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
