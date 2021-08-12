// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package camera_test

import (
	"testing"

	"poly.red/camera"
	"poly.red/math"
	"poly.red/object"
)

func TestCameraProperties(t *testing.T) {
	pos := math.NewVec3(-550, 194, 734)
	lookAt := math.NewVec3(-1000, 0, 0)
	up := math.NewVec3(0, 1, 1)
	fov := 45.0
	aspect := 1.6
	near := -100.0
	far := -600.0

	oc := camera.NewOrthographic(
		camera.Position(pos),
		camera.LookAt(lookAt, up),
	)
	if oc.Type() != object.TypeCamera {
		t.Fatalf("camera type does not return a type of camera, got %v", oc.Type())
	}

	if oc.Fov() != 1.5707963267948966 {
		t.Fatalf("unexpected camera fov, got %v, want %v", oc.Fov(), 1.5707963267948966)
	}

	if oc.Aspect() != 1 {
		t.Fatalf("unexpected camera aspect, got %v, want %v", oc.Aspect(), 1)
	}

	oc.SetAspect(1, 2)
	if oc.Aspect() != 0.5 {
		t.Fatalf("unexpected camera apsect, got %v, want %v", oc.Aspect(), 0.5)
	}

	if !oc.Position().Eq(pos) {
		t.Fatalf("unexpected camera position, got %v, want %v", oc.Position(), pos)
	}

	oc.SetPosition(math.NewVec3(0, 0, 0))
	if !oc.Position().Eq(math.NewVec3(0, 0, 0)) {
		t.Fatalf("unexpected camera position, got %v, want %v", oc.Position(), math.NewVec3(0, 0, 0))
	}

	target, gotup := oc.LookAt()
	if !target.Eq(lookAt) || !gotup.Eq(up) {
		t.Fatalf("unexpected target or up, want %v, %v, got %v, %v", lookAt, up, target, gotup)
	}
	oc.SetLookAt(math.NewVec3(0, 0, 0), math.NewVec3(0, 1, 0))
	target, gotup = oc.LookAt()
	if !target.Eq(math.NewVec3(0, 0, 0)) || !gotup.Eq(math.NewVec3(0, 1, 0)) {
		t.Fatalf("unexpected target or up, want %v, %v, got %v, %v", math.NewVec3(0, 0, 0), math.NewVec3(0, 1, 0), target, gotup)
	}

	oc = camera.NewPerspective(
		camera.Position(pos),
		camera.LookAt(lookAt, up),
		camera.PerspFrustum(fov, aspect, near, far),
	)
	if oc.Type() != object.TypeCamera {
		t.Fatalf("camera type does not return a type of camera, got %v", oc.Type())
	}

	if oc.Fov() != 45 {
		t.Fatalf("unexpected camera fov, got %v, want %v", oc.Fov(), 45)
	}

	if oc.Aspect() != 1.6 {
		t.Fatalf("unexpected camera aspect, got %v, want %v", oc.Aspect(), 1.6)
	}

	oc.SetAspect(1, 2)
	if oc.Aspect() != 0.5 {
		t.Fatalf("unexpected camera apsect, got %v, want %v", oc.Aspect(), 0.5)
	}

	if !oc.Position().Eq(pos) {
		t.Fatalf("unexpected camera position, got %v, want %v", oc.Position(), pos)
	}

	oc.SetPosition(math.NewVec3(0, 0, 0))
	if !oc.Position().Eq(math.NewVec3(0, 0, 0)) {
		t.Fatalf("unexpected camera position, got %v, want %v", oc.Position(), math.NewVec3(0, 0, 0))
	}

	target, gotup = oc.LookAt()
	if !target.Eq(lookAt) || !gotup.Eq(up) {
		t.Fatalf("unexpected target or up, want %v, %v, got %v, %v", lookAt, up, target, gotup)
	}
	oc.SetLookAt(math.NewVec3(0, 0, 0), math.NewVec3(0, 1, 0))
	target, gotup = oc.LookAt()
	if !target.Eq(math.NewVec3(0, 0, 0)) || !gotup.Eq(math.NewVec3(0, 1, 0)) {
		t.Fatalf("unexpected target or up, want %v, %v, got %v, %v", math.NewVec3(0, 0, 0), math.NewVec3(0, 1, 0), target, gotup)
	}
}

func TestViewMatrix(t *testing.T) {
	pos := math.NewVec3(-550, 194, 734)
	lookAt := math.NewVec3(-1000, 0, 0)
	up := math.NewVec3(0, 1, 1)
	fov := 45.0
	aspect := 1.6
	near := -100.0
	far := -600.0

	want := math.NewMat4(
		0.6469966392206304, 0.5391638660171921, -0.5391638660171921, 646.9966392206305,
		-0.5669309063966456, 0.8130082437851895, 0.13269115610921495, -566.9309063966456,
		0.5098869445372056, 0.21981792720048418, 0.8316822606451308, -372.6616376949569,
		0, 0, 0, 1,
	)

	vm := camera.ViewMatrix(pos, lookAt, up)

	if !vm.Eq(want) {
		t.Errorf("view matrix is wrong, want %v got %v", want, vm)
	}

	cp := camera.NewPerspective(
		camera.Position(pos),
		camera.LookAt(lookAt, up),
		camera.PerspFrustum(fov, aspect, near, far),
	)
	vm = cp.ViewMatrix()
	if !vm.Eq(want) {
		t.Errorf("view matrix is wrong, want %v got %v", want, vm)
	}

	op := camera.NewOrthographic(
		camera.Position(pos),
		camera.LookAt(lookAt, up),
	)
	vm = op.ViewMatrix()
	if !vm.Eq(want) {
		t.Errorf("view matrix is wrong, want %v got %v", want, vm)
	}
}

func TestProjMatrix(t *testing.T) {
	pos := math.NewVec3(-550, 194, 734)
	lookAt := math.NewVec3(-1000, 0, 0)
	up := math.NewVec3(0, 1, 1)
	fov := 45.0
	aspect := 1.6
	near := -100.0
	far := -600.0

	want := math.NewMat4(
		-1.5088834764831844, 0, 0, 0,
		0, -2.414213562373095, 0, 0,
		0, 0, -1.4, 240,
		0, 0, 1, 0,
	)

	cp := camera.NewPerspective(
		camera.Position(pos),
		camera.LookAt(lookAt, up),
		camera.PerspFrustum(fov, aspect, near, far),
	)
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
	want = math.NewMat4(
		2, 0, 0, 0,
		0, 1.3333333, 0, -0.3333333,
		0, 0, 0.6666666, 1,
		0, 0, 0, 1,
	)
	op := camera.NewOrthographic(
		camera.Position(pos),
		camera.LookAt(lookAt, up),
		camera.OrthoFrustum(left, right, bottom, top, near, far),
	)
	vm = op.ProjMatrix()
	if !vm.Eq(want) {
		t.Errorf("view matrix is wrong, want %v got %v", want, vm)
	}
}

func BenchmarkCamera(b *testing.B) {
	w, h := 1920, 1080
	c1 := camera.NewPerspective(
		camera.Position(math.NewVec3(-0.5, 0.5, 0.5)),
		camera.LookAt(
			math.NewVec3(0, 0, -0.5),
			math.NewVec3(0, 1, 0),
		),
		camera.PerspFrustum(
			45,
			float64(w)/float64(h),
			-0.1,
			-3,
		),
	)

	b.Run("Perspective_ViewMatrix", func(b *testing.B) {
		b.ReportAllocs()
		var m math.Mat4
		for i := 0; i < b.N; i++ {
			m = c1.ViewMatrix()
		}
		_ = m
	})
	b.Run("Perspective_ProjMatrix", func(b *testing.B) {
		b.ReportAllocs()
		var m math.Mat4
		for i := 0; i < b.N; i++ {
			m = c1.ProjMatrix()
		}
		_ = m
	})

	c2 := camera.NewOrthographic(
		camera.Position(math.NewVec3(-0.5, 0.5, 0.5)),
		camera.LookAt(
			math.NewVec3(0, 0, -0.5),
			math.NewVec3(0, 1, 0),
		),
		camera.OrthoFrustum(
			-10, 10, -10, 10, 10, -10,
		),
	)

	b.Run("Orthographic_ViewMatrix", func(b *testing.B) {
		b.ReportAllocs()
		var m math.Mat4
		for i := 0; i < b.N; i++ {
			m = c2.ViewMatrix()
		}
		_ = m
	})
	b.Run("Orthographic_ProjMatrix", func(b *testing.B) {
		b.ReportAllocs()
		var m math.Mat4
		for i := 0; i < b.N; i++ {
			m = c2.ProjMatrix()
		}
		_ = m
	})
}

func TestCmaeraMisuse(t *testing.T) {

	t.Run("persp", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("camera misuse did not panic")
			}
		}()

		camera.NewPerspective(
			camera.OrthoFrustum(1, 1, 1, 1, 1, 1),
		)
	})

	t.Run("ortho", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("camera misuse did not panic")
			}
		}()

		camera.NewOrthographic(
			camera.PerspFrustum(1, 1, 1, 1),
		)
	})
}
