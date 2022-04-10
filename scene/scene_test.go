// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package scene_test

import (
	"fmt"
	"math/rand"
	"testing"

	"poly.red/geometry"
	"poly.red/math"
	"poly.red/model"
	"poly.red/scene"
	"poly.red/scene/object"
)

func TestScene(t *testing.T) {
	s := scene.NewScene()
	p1 := geometry.NewWith(model.NewPlane(1, 1), nil)
	g := s.Add(p1)

	iterCount := 0
	scene.IterObjects(s, func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		iterCount++
		want := math.NewMat4[float32](
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1,
		)
		if !want.Eq(modelMatrix) {
			t.Fatalf("unexpected model matrix. want %v got %v", want, modelMatrix)
		}
		return true
	})
	if iterCount != 1 {
		t.Fatalf("unexpected iteration, want %v, got %v", 1, iterCount)
	}

	g.Scale(2, 2, 2)
	iterCount = 0
	scene.IterObjects(s, func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		iterCount++
		want := math.NewMat4[float32](
			2, 0, 0, 0,
			0, 2, 0, 0,
			0, 0, 2, 0,
			0, 0, 0, 1,
		)
		if !want.Eq(modelMatrix) {
			t.Fatalf("unexpected model matrix. want %v got %v", want, modelMatrix)
		}
		return true
	})
	if iterCount != 1 {
		t.Fatalf("unexpected iteration, want %v, got %v", 1, iterCount)
	}

	g.Translate(1, 2, 3)
	iterCount = 0
	scene.IterObjects(s, func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		iterCount++
		want := math.NewMat4[float32](
			2, 0, 0, 1,
			0, 2, 0, 2,
			0, 0, 2, 3,
			0, 0, 0, 1,
		)
		if !want.Eq(modelMatrix) {
			t.Fatalf("unexpected model matrix. want %v got %v", want, modelMatrix)
		}
		return true
	})
	if iterCount != 1 {
		t.Fatalf("unexpected iteration, want %v, got %v", 1, iterCount)
	}

	p2 := geometry.NewWith(model.NewPlane(1, 1), nil)
	g2 := scene.NewGroup()
	g2.SetName("another")
	g2.Add(p2)
	g.Add(g2)
	g2.Scale(2, 2, 2)
	g2.Translate(1, 1, 1)

	iterCount = 0
	s.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		iterCount++

		if o == p1 {
			t.Log("p1")
			want := math.NewMat4[float32](
				2, 0, 0, 1,
				0, 2, 0, 2,
				0, 0, 2, 3,
				0, 0, 0, 1,
			)
			if !want.Eq(modelMatrix) {
				t.Fatalf("unexpected model matrix. want %v got %v", want, modelMatrix)
			}
			return true
		}

		if o == p2 {
			t.Log("p2")
			want := math.NewMat4[float32](
				4, 0, 0, 3,
				0, 4, 0, 4,
				0, 0, 4, 5,
				0, 0, 0, 1,
			)
			if !want.Eq(modelMatrix) {
				t.Fatalf("unexpected model matrix. want %v got %v", want, modelMatrix)
			}
			return true
		}

		panic("unknown object")
	})
	if iterCount != 2 {
		t.Fatalf("unexpected iteration, want %v, got %v", 2, iterCount)
	}

	p3 := geometry.NewWith(model.NewPlane(1, 1), nil)
	g3 := scene.NewGroup()
	g3.SetName("another")
	g3.Add(p3)
	s.Add(g3)
	iterCount = 0
	s.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		iterCount++
		if o == p1 || o == p2 {
			return true
		}

		if o == p3 {
			t.Log("p3")
			want := math.NewMat4[float32](
				2, 0, 0, 1,
				0, 2, 0, 2,
				0, 0, 2, 3,
				0, 0, 0, 1,
			)
			if !want.Eq(modelMatrix) {
				t.Fatalf("unexpected model matrix. want %v got %v", want, modelMatrix)
			}
			return true
		}

		panic("unknown object")
	})
	if iterCount != 3 {
		t.Fatalf("unexpected iteration, want %v, got %v", 3, iterCount)
	}

	p4 := geometry.NewWith(model.NewPlane(1, 1), nil)
	s.Add(p4)
	iterCount = 0
	s.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		iterCount++
		if o == p1 || o == p2 || o == p3 {
			return true
		}

		if o == p4 {
			t.Log("p4")
			want := math.NewMat4[float32](
				2, 0, 0, 1,
				0, 2, 0, 2,
				0, 0, 2, 3,
				0, 0, 0, 1,
			)
			if !want.Eq(modelMatrix) {
				t.Fatalf("unexpected model matrix. want %v got %v", want, modelMatrix)
			}
			return true
		}

		panic("unknown object")
	})
	if iterCount != 4 {
		t.Fatalf("unexpected iteration, want %v, got %v", 4, iterCount)
	}

	g3.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		if o != p3 {
			t.Fatalf("unexpected iterated object: want %v got %v", p3, o)
		}
		return true
	})
}

func TestSceneGraphStop(t *testing.T) {
	s := scene.NewScene()
	p1 := geometry.NewWith(model.NewPlane(1, 1), nil)
	g := s.Add(p1)

	iterCount := 0

	g.Scale(2, 2, 2)
	g.Translate(1, 2, 3)
	p2 := geometry.NewWith(model.NewPlane(1, 1), nil)
	g2 := scene.NewGroup()
	g2.Add(p2)
	g.Add(g2)
	g2.Scale(2, 2, 2)
	g2.Translate(1, 1, 1)
	p3 := geometry.NewWith(model.NewPlane(1, 1), nil)
	g3 := scene.NewGroup()
	g3.Add(p3)
	s.Add(g3)
	p4 := geometry.NewWith(model.NewPlane(1, 1), nil)
	s.Add(p4)

	s.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		iterCount++
		if iterCount == 3 {
			return false
		}
		return true
	})
	if iterCount != 3 {
		t.Fatalf("iter objects does not stop properly, want %v got %v", 3, iterCount)
	}
}

func TestSceneGraphIterPanic(t *testing.T) {
	s := scene.NewScene()
	p1 := geometry.NewWith(model.NewPlane(1, 1), nil)
	s.Add(p1)

	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Fatalf("unexpected panic did not happen.")
	}()

	s.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		panic("unexpected panic")
	})
}

var all int

func createGroup(n int) *scene.Group {
	g := scene.NewGroup()
	for i := 0; i < n; i++ {
		p := geometry.NewWith(model.NewPlane(1, 1), nil)
		all++
		g.Add(p)
	}
	return g
}

func createDepthGroup(n, depth int) *scene.Group {
	if depth < 0 {
		return nil
	}
	g := scene.NewGroup()
	for i := 0; i <= n; i++ {
		var gg *scene.Group
		if rand.Intn(2) < 1 { // 50% chance to create a new random group.
			gg = createGroup(n)
			if gg == nil {
				continue
			}
		} else { // or enter to the next level.
			gg = createDepthGroup(n, depth-1)
			if gg == nil {
				continue
			}
		}
		g.Add(gg)
	}
	return g
}

func createRandomSceneGraph(n, depth int) *scene.Scene {
	s := scene.NewScene(createDepthGroup(n, depth))
	return s
}

func TestCreateDepthGroup(t *testing.T) {
	s := createRandomSceneGraph(1, 2)
	t.Log(all)

	n := 0
	s.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		n++
		return true
	})
	if all != n {
		t.Fatalf("unexpected created objects: want %v got %v", all, n)
	}
}

func TestGenericIterator(t *testing.T) {
	s := scene.NewScene()

	s.Add(geometry.NewWith(model.NewPlane(1, 1), nil))
	s.Add(geometry.NewWith(model.NewPlane(1, 1), nil))
	s.Add(geometry.NewWith(model.NewPlane(1, 1), nil))
	s.Add(geometry.NewWith(model.NewPlane(1, 1), nil))
	s.Add(geometry.NewWith(model.NewPlane(1, 1), nil))

	n := 0
	scene.IterObjects(s, func(o *geometry.Geometry, modelMatrix math.Mat4[float32]) bool {
		n++
		return true
	})
	if n != 5 {
		t.Fatalf("unexpected iteration, expect %v got %v", 5, n)
	}
}

func BenchmarkIterator(b *testing.B) {
	for i := 5; i < 8; i++ {
		for j := 5; j < 8; j++ {
			s := createRandomSceneGraph(i, j)
			b.Run(fmt.Sprintf("%v-%v", i, j), func(b *testing.B) {
				for k := 0; k < b.N; k++ {
					scene.IterObjects(s, func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
						return true
					})
				}
			})
		}
	}
}

func TestAddMultiple(t *testing.T) {
	p := geometry.NewWith(model.NewPlane(1, 1), nil)

	s := scene.NewScene()
	s.Add(p)
	s.Add(p)
	s.Add(p)
	s.Add(p)
	s.Add(p)
	s.Add(p)

	n := 0
	scene.IterObjects(s, func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		n++
		return true
	})
	if n != 6 {
		t.Fatalf("unexpected objects, want %v got %v", 6, n)
	}

	g := scene.NewGroup()
	g.Add(p)
	s.Add(g)
	n = 0
	scene.IterObjects(s, func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		n++
		t.Log(o.Name())
		return true
	})
	if n != 7 {
		t.Fatalf("unexpected objects, want %v got %v", 6, n)
	}
}

func TestSceneTransformation(t *testing.T) {
	g := model.MustLoad("../internal/testdata/bunny.obj")

	g.Scale(1500, 1500, 1500)
	g.Translate(-700, -5, 350)

	s := scene.NewScene()
	s.Add(g)
	scene.IterObjects(s, func(o *geometry.Geometry, modelMatrix math.Mat4[float32]) bool {
		t.Log(modelMatrix)
		return true
	})
}
