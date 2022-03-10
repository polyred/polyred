// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package modifiers

import (
	"container/heap"
	"log"

	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/material"
	"poly.red/math"
)

type vert struct {
	v *primitive.Vertex
	q math.Mat4[float32]
}

type face struct {
	f       *primitive.Triangle
	removed bool
}

// Simplify simplifies the given triangle soup down to the target faces.
// If the targetFaces is greater than the given face, the function returns
// the original mesh. Otherwise, it returns the newly created triangle soup.
func Simplify(m *mesh.TriangleSoup, targetFaces int) *mesh.TriangleSoup {
	numFaces := m.NumTriangles()
	if numFaces <= uint64(targetFaces) {
		return m
	}

	vectorVertex := make(map[math.Vec4[float32]]*primitive.Vertex)
	m.Faces(func(f primitive.Face[float32], _ material.Material) bool {
		f.Triangles(func(t *primitive.Triangle) bool {
			vectorVertex[t.V1.Pos] = t.V1
			vectorVertex[t.V2.Pos] = t.V2
			vectorVertex[t.V3.Pos] = t.V3
			return true
		})
		return true
	})

	// accumlate quadric matrices for each vertex based on its faces
	m.Faces(func(f primitive.Face[float32], _ material.Material) bool {
		f.Triangles(func(t *primitive.Triangle) bool {
			// Face Quadric
			n := t.Normal()
			x, y, z := t.V1.Pos.X, t.V1.Pos.Y, t.V1.Pos.Z
			a, b, c := n.X, n.Y, n.Z
			d := -a*x - b*y - c*z
			q := math.NewMat4(
				a*a, a*b, a*c, a*d,
				a*b, b*b, b*c, b*d,
				a*c, b*c, c*c, c*d,
				a*d, b*d, c*d, d*d,
			)

			v1 := vectorVertex[t.V1.Pos]
			v2 := vectorVertex[t.V2.Pos]
			v3 := vectorVertex[t.V3.Pos]
			v1.AttrFlat = map[primitive.Attribute]any{}
			v2.AttrFlat = map[primitive.Attribute]any{}
			v3.AttrFlat = map[primitive.Attribute]any{}

			if _, ok := v1.AttrFlat["quadric"]; !ok {
				v1.AttrFlat["quadric"] = math.Mat4[float32]{}
			} else {
				v1.AttrFlat["quadric"] = v1.AttrFlat["quadric"].(math.Mat4[float32]).Add(q)
			}

			if _, ok := v2.AttrFlat["quadric"]; !ok {
				v2.AttrFlat["quadric"] = math.Mat4[float32]{}
			} else {
				v2.AttrFlat["quadric"] = v2.AttrFlat["quadric"].(math.Mat4[float32]).Add(q)
			}

			if _, ok := v3.AttrFlat["quadric"]; !ok {
				v3.AttrFlat["quadric"] = math.Mat4[float32]{}
			} else {
				v3.AttrFlat["quadric"] = v3.AttrFlat["quadric"].(math.Mat4[float32]).Add(q)
			}
			return true
		})
		return true
	})

	// create faces and map vertex => faces
	vertexFaces := make(map[*primitive.Vertex][]*face)
	m.Faces(func(f primitive.Face[float32], _ material.Material) bool {
		f.Triangles(func(t *primitive.Triangle) bool {
			v1 := vectorVertex[t.V1.Pos]
			v2 := vectorVertex[t.V2.Pos]
			v3 := vectorVertex[t.V3.Pos]
			f := &face{primitive.NewTriangle(v1, v2, v3), false}
			vertexFaces[v1] = append(vertexFaces[v1], f)
			vertexFaces[v2] = append(vertexFaces[v2], f)
			vertexFaces[v3] = append(vertexFaces[v3], f)
			return true
		})
		return true
	})

	// find distinct pairs
	// TODO: pair vertices within a threshold distance of each other
	pairs := make(map[edgeKey]*edgePair)
	m.Faces(func(f primitive.Face[float32], m material.Material) bool {
		f.Triangles(func(t *primitive.Triangle) bool {
			v1 := vectorVertex[t.V1.Pos]
			v2 := vectorVertex[t.V2.Pos]
			v3 := vectorVertex[t.V3.Pos]
			pairs[edgeKey{v1.Pos, v2.Pos}] = newEdgePair(v1, v2)
			pairs[edgeKey{v2.Pos, v3.Pos}] = newEdgePair(v2, v3)
			pairs[edgeKey{v3.Pos, v1.Pos}] = newEdgePair(v3, v1)
			return true
		})
		return true
	})

	// enqueue pairs and map vertex => pairs
	var queue edgeQueue
	vertexPairs := make(map[*primitive.Vertex][]*edgePair)
	for _, p := range pairs {
		heap.Push(&queue, p)
		vertexPairs[p.A] = append(vertexPairs[p.A], p)
		vertexPairs[p.B] = append(vertexPairs[p.B], p)
	}

	// simplify
	for numFaces > uint64(targetFaces) {
		// pop best pair
		p := heap.Pop(&queue).(*edgePair)
		log.Println("done: ", numFaces, p.ErrorCache, p.Removed)

		if p.Removed {
			continue
		}
		p.Removed = true

		// get related faces
		distinctFaces := make(map[*face]bool)
		for _, f := range vertexFaces[p.A] {
			if !f.removed {
				distinctFaces[f] = true
			}
		}
		for _, f := range vertexFaces[p.B] {
			if !f.removed {
				distinctFaces[f] = true
			}
		}

		// get related pairs
		distinctPairs := make(map[*edgePair]bool)
		for _, q := range vertexPairs[p.A] {
			if !q.Removed {
				distinctPairs[q] = true
			}
		}
		for _, q := range vertexPairs[p.B] {
			if !q.Removed {
				distinctPairs[q] = true
			}
		}

		// create the new vertex
		v := &primitive.Vertex{Pos: p.estimateV(), AttrFlat: map[primitive.Attribute]any{}}
		v.AttrFlat["quadric"] = p.Quadric()

		// update faces
		newFaces := make([]*primitive.Triangle, 0, len(distinctFaces))
		valid := true
		for f := range distinctFaces {
			v1, v2, v3 := f.f.V1, f.f.V2, f.f.V3
			if v1.Pos == p.A.Pos || v1.Pos == p.B.Pos {
				v1 = v
			}
			if v2.Pos == p.A.Pos || v2.Pos == p.B.Pos {
				v2 = v
			}
			if v3.Pos == p.A.Pos || v3.Pos == p.B.Pos {
				v3 = v
			}
			f := primitive.NewTriangle(v1, v2, v3)
			if f.V1.Pos == f.V2.Pos || f.V1.Pos == f.V3.Pos || f.V2.Pos == f.V3.Pos {
				continue
			}
			if f.Normal().Dot(f.Normal()) < 1e-3 {
				valid = false
				break
			}
			newFaces = append(newFaces, f)
		}
		if !valid {
			continue
		}
		delete(vertexFaces, p.A)
		delete(vertexFaces, p.B)
		for f := range distinctFaces {
			f.removed = true
			numFaces--
		}
		for _, f := range newFaces {
			numFaces++
			vertexFaces[f.V1] = append(vertexFaces[f.V1], &face{f, false})
			vertexFaces[f.V2] = append(vertexFaces[f.V2], &face{f, false})
			vertexFaces[f.V3] = append(vertexFaces[f.V3], &face{f, false})
		}

		// update pairs and prune current pair
		delete(vertexPairs, p.A)
		delete(vertexPairs, p.B)
		seen := make(map[math.Vec4[float32]]bool)
		for q := range distinctPairs {
			q.Removed = true
			heap.Remove(&queue, q.Index)
			a, b := q.A, q.B
			if a == p.A || a == p.B {
				a = v
			}
			if b == p.A || b == p.B {
				b = v
			}
			if b == v {
				// swap so that a == v
				a, b = b, a
			}
			if _, ok := seen[b.Pos]; ok {
				// only want distinct neighbors
				continue
			}
			seen[b.Pos] = true
			q = newEdgePair(a, b)
			heap.Push(&queue, q)
			vertexPairs[a] = append(vertexPairs[a], q)
			vertexPairs[b] = append(vertexPairs[b], q)
		}
	}

	// find distinct faces
	distinctFaces := make(map[*face]bool)
	for _, faces := range vertexFaces {
		for _, f := range faces {
			if !f.removed {
				distinctFaces[f] = true
			}
		}
	}

	// construct resulting mesh
	triangles := make([]*primitive.Triangle, len(distinctFaces))
	i := 0
	for f := range distinctFaces {
		triangles[i] = primitive.NewTriangle(f.f.V1, f.f.V2, f.f.V3)
		i++
	}
	return mesh.NewTriangleSoup(triangles)
}

type edgeKey struct{ A, B math.Vec4[float32] }

type edgePair struct {
	A, B       *primitive.Vertex
	Index      int
	Removed    bool
	ErrorCache float32
}

func newEdgePair(a, b *primitive.Vertex) *edgePair {
	return &edgePair{A: a, B: b, Index: -1, Removed: false, ErrorCache: -1}
}

func (e *edgePair) Error() float32 {
	if e.ErrorCache < 0 {
		e.ErrorCache = quadricError(e.Quadric(), e.estimateV())
	}
	return e.ErrorCache
}

func (p *edgePair) Quadric() math.Mat4[float32] {
	q1 := p.A.AttrFlat["quadric"].(math.Mat4[float32])
	q2 := p.B.AttrFlat["quadric"].(math.Mat4[float32])
	return q1.Add(q2)
}

func (p *edgePair) estimateV() math.Vec4[float32] {
	q := p.Quadric()
	if math.Abs(q.Det()) > 1e-3 {
		b := math.NewMat4(
			q.X00, q.X01, q.X02, q.X03,
			q.X10, q.X11, q.X12, q.X13,
			q.X20, q.X21, q.X22, q.X23,
			0, 0, 0, 1,
		)
		v := b.Inv().MulV(math.NewVec4[float32](0, 0, 0, 1))
		if !math.IsNaN(float64(v.X)) && !math.IsNaN(float64(v.Y)) && !math.IsNaN(float64(v.Z)) {
			return v
		}
	}
	// cannot compute best vector with matrix
	// look for best vector along edge
	const n = 32
	a := p.A.Pos
	b := p.B.Pos
	d := b.Sub(a)
	bestE := float32(-1.0)
	bestV := math.Vec4[float32]{}
	for i := 0; i <= n; i++ {
		t := float32(i) / n
		v := a.Add(d.Scale(t, t, t, 1))
		e := quadricError(q, v)
		if bestE < 0 || e < bestE {
			bestE = e
			bestV = v
		}
	}
	return bestV
}

func quadricError(a math.Mat4[float32], v math.Vec4[float32]) float32 {
	return (v.X*a.X00*v.X + v.Y*a.X10*v.X + v.Z*a.X20*v.X + a.X30*v.X +
		v.X*a.X01*v.Y + v.Y*a.X11*v.Y + v.Z*a.X21*v.Y + a.X31*v.Y +
		v.X*a.X02*v.Z + v.Y*a.X12*v.Z + v.Z*a.X22*v.Z + a.X32*v.Z +
		v.X*a.X03 + v.Y*a.X13 + v.Z*a.X23 + a.X33)
}

type edgeQueue []*edgePair

func (pq edgeQueue) Len() int {
	return len(pq)
}

func (pq edgeQueue) Less(i, j int) bool {
	return pq[i].Error() < pq[j].Error()
}

func (pq edgeQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *edgeQueue) Push(x interface{}) {
	item := x.(*edgePair)
	item.Index = len(*pq)
	*pq = append(*pq, item)
}

func (pq *edgeQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.Index = -1
	*pq = old[:n-1]
	return item
}
