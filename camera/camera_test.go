// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
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
	pos := math.NewVec3[float32](-550, 194, 734)
	lookAt := math.NewVec3[float32](-1000, 0, 0)
	up := math.NewVec3[float32](0, 1, 1)
	fov := float32(45.0)
	aspect := float32(1.6)
	near := float32(-100.0)
	far := float32(-600.0)

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

	oc.SetPosition(math.NewVec3[float32](0, 0, 0))
	if !oc.Position().Eq(math.NewVec3[float32](0, 0, 0)) {
		t.Fatalf("unexpected camera position, got %v, want %v", oc.Position(), math.NewVec3[float32](0, 0, 0))
	}

	target, gotup := oc.LookAt()
	if !target.Eq(lookAt) || !gotup.Eq(up) {
		t.Fatalf("unexpected target or up, want %v, %v, got %v, %v", lookAt, up, target, gotup)
	}
	oc.SetLookAt(math.NewVec3[float32](0, 0, 0), math.NewVec3[float32](0, 1, 0))
	target, gotup = oc.LookAt()
	if !target.Eq(math.NewVec3[float32](0, 0, 0)) || !gotup.Eq(math.NewVec3[float32](0, 1, 0)) {
		t.Fatalf("unexpected target or up, want %v, %v, got %v, %v", math.NewVec3[float32](0, 0, 0), math.NewVec3[float32](0, 1, 0), target, gotup)
	}

	oc = camera.NewPerspective(
		camera.Position(pos),
		camera.LookAt(lookAt, up),
		camera.ViewFrustum(fov, aspect, near, far),
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

	oc.SetPosition(math.NewVec3[float32](0, 0, 0))
	if !oc.Position().Eq(math.NewVec3[float32](0, 0, 0)) {
		t.Fatalf("unexpected camera position, got %v, want %v", oc.Position(), math.NewVec3[float32](0, 0, 0))
	}

	target, gotup = oc.LookAt()
	if !target.Eq(lookAt) || !gotup.Eq(up) {
		t.Fatalf("unexpected target or up, want %v, %v, got %v, %v", lookAt, up, target, gotup)
	}
	oc.SetLookAt(math.NewVec3[float32](0, 0, 0), math.NewVec3[float32](0, 1, 0))
	target, gotup = oc.LookAt()
	if !target.Eq(math.NewVec3[float32](0, 0, 0)) || !gotup.Eq(math.NewVec3[float32](0, 1, 0)) {
		t.Fatalf("unexpected target or up, want %v, %v, got %v, %v", math.NewVec3[float32](0, 0, 0), math.NewVec3[float32](0, 1, 0), target, gotup)
	}
}

func TestViewMatrix(t *testing.T) {
	pos := math.NewVec3[float32](-550, 194, 734)
	lookAt := math.NewVec3[float32](-1000, 0, 0)
	up := math.NewVec3[float32](0, 1, 1)
	fov := float32(45.0)
	aspect := float32(1.6)
	near := float32(-100.0)
	far := float32(-600.0)

	want := math.NewMat4[float32](
		0.6469966, 0.5391638, -0.5391638, 646.9966,
		-0.56693095, 0.8130083, 0.13269116, -566.93097,
		0.5098869, 0.21981792, 0.83168226, -372.66165,
		0, 0, 0, 1,
	)

	vm := camera.ViewMatrix(pos, lookAt, up)

	if !vm.Eq(want) {
		t.Errorf("view matrix is wrong,\nwant\n%v\ngot\n%v", want, vm)
	}

	cp := camera.NewPerspective(
		camera.Position(pos),
		camera.LookAt(lookAt, up),
		camera.ViewFrustum(fov, aspect, near, far),
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
	pos := math.NewVec3[float32](-550, 194, 734)
	lookAt := math.NewVec3[float32](-1000, 0, 0)
	up := math.NewVec3[float32](0, 1, 1)
	fov := float32(45.0)
	aspect := float32(1.6)
	near := float32(-100.0)
	far := float32(-600.0)

	want := math.NewMat4[float32](
		-1.5088835, 0, 0, 0,
		0, -2.4142134, 0, 0,
		0, 0, -1.4, 240,
		0, 0, 1, 0,
	)

	cp := camera.NewPerspective(
		camera.Position(pos),
		camera.LookAt(lookAt, up),
		camera.ViewFrustum(fov, aspect, near, far),
	)
	vm := cp.ProjMatrix()
	if !vm.Eq(want) {
		t.Errorf("perspective projection matrix is wrong,\nwant\n%v\ngot\n%v", want, vm)
	}

	left := float32(-0.5)
	right := float32(0.5)
	top := float32(1.0)
	bottom := float32(-0.5)
	near = float32(0.0)
	far = float32(-3.0)
	want = math.NewMat4[float32](
		2, 0, 0, 0,
		0, 1.3333334, 0, -0.33333334,
		0, 0, 0.6666667, 1,
		0, 0, 0, 1,
	)
	op := camera.NewOrthographic(
		camera.Position(pos),
		camera.LookAt(lookAt, up),
		camera.ViewFrustum(left, right, bottom, top, near, far),
	)
	vm = op.ProjMatrix()
	if !vm.Eq(want) {
		t.Errorf("view matrix is wrong,\nwant\n%v\ngot\n%v", want, vm)
	}
}

func BenchmarkCamera(b *testing.B) {
	w, h := float32(1920), float32(1080)
	c1 := camera.NewPerspective(
		camera.Position(math.NewVec3[float32](-0.5, 0.5, 0.5)),
		camera.LookAt(
			math.NewVec3[float32](0, 0, -0.5),
			math.NewVec3[float32](0, 1, 0),
		),
		camera.ViewFrustum(45, w/h, -0.1, -3),
	)

	b.Run("Perspective_ViewMatrix", func(b *testing.B) {
		b.ReportAllocs()
		var m math.Mat4[float32]
		for i := 0; i < b.N; i++ {
			m = c1.ViewMatrix()
		}
		_ = m
	})
	b.Run("Perspective_ProjMatrix", func(b *testing.B) {
		b.ReportAllocs()
		var m math.Mat4[float32]
		for i := 0; i < b.N; i++ {
			m = c1.ProjMatrix()
		}
		_ = m
	})

	c2 := camera.NewOrthographic(
		camera.Position(math.NewVec3[float32](-0.5, 0.5, 0.5)),
		camera.LookAt(
			math.NewVec3[float32](0, 0, -0.5),
			math.NewVec3[float32](0, 1, 0),
		),
		camera.ViewFrustum(
			-10, 10, -10, 10, 10, -10,
		),
	)

	b.Run("Orthographic_ViewMatrix", func(b *testing.B) {
		b.ReportAllocs()
		var m math.Mat4[float32]
		for i := 0; i < b.N; i++ {
			m = c2.ViewMatrix()
		}
		_ = m
	})
	b.Run("Orthographic_ProjMatrix", func(b *testing.B) {
		b.ReportAllocs()
		var m math.Mat4[float32]
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
			camera.ViewFrustum(1, 1, 1, 1, 1, 1),
		)
	})

	t.Run("ortho", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("camera misuse did not panic")
			}
		}()

		camera.NewOrthographic(
			camera.ViewFrustum(1, 1, 1, 1),
		)
	})
}
