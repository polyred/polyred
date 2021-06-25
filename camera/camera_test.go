// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package camera_test

import (
	"testing"

	"changkun.de/x/polyred/camera"
	"changkun.de/x/polyred/math"
)

func TestViewMatrix(t *testing.T) {
	pos := math.NewVector(-550, 194, 734, 1)
	lookAt := math.NewVector(-1000, 0, 0, 1)
	up := math.NewVector(0, 1, 1, 0)
	fov := 45.0
	aspect := 1.6
	near := -100.0
	far := -600.0

	want := math.NewMatrix(
		0.6469966392206304, 0.5391638660171921, -0.5391638660171921, 646.9966392206305,
		-0.5669309063966456, 0.8130082437851895, 0.13269115610921495, -566.9309063966456,
		0.5098869445372056, 0.21981792720048418, 0.8316822606451308, -372.6616376949569,
		0, 0, 0, 1,
	)

	vm := camera.ViewMatrix(pos, lookAt, up)

	if !vm.Eq(want) {
		t.Errorf("view matrix is wrong, want %v got %v", want, vm)
	}

	cp := camera.NewPerspective(pos, lookAt, up, fov, aspect, near, far)
	vm = cp.ViewMatrix()
	if !vm.Eq(want) {
		t.Errorf("view matrix is wrong, want %v got %v", want, vm)
	}

	op := camera.NewOrthographic(pos, lookAt, up, -1, 1, -1, 1, 1, -1)
	vm = op.ViewMatrix()
	if !vm.Eq(want) {
		t.Errorf("view matrix is wrong, want %v got %v", want, vm)
	}
}

func TestProjMatrix(t *testing.T) {
	pos := math.NewVector(-550, 194, 734, 1)
	lookAt := math.NewVector(-1000, 0, 0, 1)
	up := math.NewVector(0, 1, 1, 0)
	fov := 45.0
	aspect := 1.6
	near := -100.0
	far := -600.0

	want := math.NewMatrix(
		-1.5088834764831844, 0, 0, 0,
		0, -2.414213562373095, 0, 0,
		0, 0, -1.4, 240,
		0, 0, 1, 0,
	)

	cp := camera.NewPerspective(pos, lookAt, up, fov, aspect, near, far)
	vm := cp.ProjMatrix()
	if !vm.Eq(want) {
		t.Errorf("perspective projection matrix is wrong, want %v got %v", want, vm)
	}

	left := -0.5
	right := 0.5
	top := 1.0
	bottom := -0.5
	near = 0.0
	far = -3.0
	want = math.NewMatrix(
		2, 0, 0, 0,
		0, 1.3333333, 0, -0.3333333,
		0, 0, 0.6666666, 1,
		0, 0, 0, 1,
	)
	op := camera.NewOrthographic(pos, lookAt, up, left, right, bottom, top, near, far)
	vm = op.ProjMatrix()
	if !vm.Eq(want) {
		t.Errorf("view matrix is wrong, want %v got %v", want, vm)
	}
}

func BenchmarkCamera(b *testing.B) {
	w, h := 1920, 1080
	c1 := camera.NewPerspective(
		math.NewVector(-0.5, 0.5, 0.5, 1),
		math.NewVector(0, 0, -0.5, 1),
		math.NewVector(0, 1, 0, 0),
		45,
		float64(w)/float64(h),
		-0.1,
		-3,
	)

	b.Run("Perspective_ViewMatrix", func(b *testing.B) {
		b.ReportAllocs()
		var m math.Matrix
		for i := 0; i < b.N; i++ {
			m = c1.ViewMatrix()
		}
		_ = m
	})
	b.Run("Perspective_ProjMatrix", func(b *testing.B) {
		b.ReportAllocs()
		var m math.Matrix
		for i := 0; i < b.N; i++ {
			m = c1.ProjMatrix()
		}
		_ = m
	})

	c2 := camera.NewOrthographic(
		math.NewVector(-0.5, 0.5, 0.5, 1),
		math.NewVector(0, 0, -0.5, 1),
		math.NewVector(0, 1, 0, 0),
		-10, 10, -10, 10, 10, -10,
	)

	b.Run("Orthographic_ViewMatrix", func(b *testing.B) {
		b.ReportAllocs()
		var m math.Matrix
		for i := 0; i < b.N; i++ {
			m = c2.ViewMatrix()
		}
		_ = m
	})
	b.Run("Orthographic_ProjMatrix", func(b *testing.B) {
		b.ReportAllocs()
		var m math.Matrix
		for i := 0; i < b.N; i++ {
			m = c2.ProjMatrix()
		}
		_ = m
	})
}
