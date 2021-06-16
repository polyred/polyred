// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package camera_test

import (
	"testing"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/math"
)

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
