// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math_test

import (
	"math/rand"
	"testing"

	"changkun.de/x/polyred/math"
)

var (
	v    float64
	vv   math.Vec2
	vvv  math.Vec3
	vvvv math.Vec4
)

func BenchmarkVec_Eq(b *testing.B) {
	b.Run("Vec2", func(b *testing.B) {
		v1 := math.NewVec2(rand.Float64(), rand.Float64())
		v2 := math.NewVec2(rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v1.Eq(v2)
		}
	})
	b.Run("Vec3", func(b *testing.B) {
		v1 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())
		v2 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v1.Eq(v2)
		}
	})
	b.Run("Vec4", func(b *testing.B) {
		v1 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())
		v2 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v1.Eq(v2)
		}
	})
}

func BenchmarkVec_Add(b *testing.B) {
	b.Run("Vec2", func(b *testing.B) {
		v1 := math.NewVec2(rand.Float64(), rand.Float64())
		v2 := math.NewVec2(rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vv = v1.Add(v2)
		}
	})
	b.Run("Vec3", func(b *testing.B) {
		v1 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())
		v2 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vvv = v1.Add(v2)
		}
	})
	b.Run("Vec4", func(b *testing.B) {
		v1 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())
		v2 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vvvv = v1.Add(v2)
		}
	})
}

func BenchmarkVec_Sub(b *testing.B) {
	b.Run("Vec2", func(b *testing.B) {
		v1 := math.NewVec2(rand.Float64(), rand.Float64())
		v2 := math.NewVec2(rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vv = v1.Sub(v2)
		}
	})
	b.Run("Vec3", func(b *testing.B) {
		v1 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())
		v2 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vvv = v1.Sub(v2)
		}
	})
	b.Run("Vec4", func(b *testing.B) {
		v1 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())
		v2 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vvvv = v1.Sub(v2)
		}
	})
}
func BenchmarkVec_IsZero(b *testing.B) {
	b.Run("Vec2", func(b *testing.B) {
		v1 := math.NewVec2(rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v1.IsZero()
		}
	})
	b.Run("Vec3", func(b *testing.B) {
		v1 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v1.IsZero()
		}
	})
	b.Run("Vec4", func(b *testing.B) {
		v1 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v1.IsZero()
		}
	})
}

func BenchmarkVec_Scale(b *testing.B) {
	b.Run("Vec2", func(b *testing.B) {
		v1 := math.NewVec2(rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vv = v1.Scale(2, 2)
		}
	})
	b.Run("Vec3", func(b *testing.B) {
		v1 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vvv = v1.Scale(2, 2, 2)
		}
	})
	b.Run("Vec4", func(b *testing.B) {
		v1 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vvvv = v1.Scale(2, 2, 2, 2)
		}
	})
}

func BenchmarkVec_Translate(b *testing.B) {
	b.Run("Vec2", func(b *testing.B) {
		v1 := math.NewVec2(rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vv = v1.Translate(2, 2)
		}
	})
	b.Run("Vec3", func(b *testing.B) {
		v1 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vvv = v1.Translate(2, 2, 2)
		}
	})
	b.Run("Vec4", func(b *testing.B) {
		v1 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vvvv = v1.Translate(2, 2, 2)
		}
	})
}

func BenchmarkVec_Dot(b *testing.B) {
	b.Run("Vec2", func(b *testing.B) {
		v1 := math.NewVec2(rand.Float64(), rand.Float64())
		v2 := math.NewVec2(rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v = v1.Dot(v2)
		}
	})
	b.Run("Vec3", func(b *testing.B) {
		v1 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())
		v2 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v = v1.Dot(v2)
		}
	})
	b.Run("Vec4", func(b *testing.B) {
		v1 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())
		v2 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v = v1.Dot(v2)
		}
	})
}

func BenchmarkVec_Len(b *testing.B) {
	b.Run("Vec2", func(b *testing.B) {
		v1 := math.NewVec2(rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v = v1.Len()
		}
	})
	b.Run("Vec3", func(b *testing.B) {
		v1 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v = v1.Len()
		}
	})
	b.Run("Vec4", func(b *testing.B) {
		v1 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v = v1.Len()
		}
	})
}

func BenchmarkVec_Unit(b *testing.B) {
	b.Run("Vec2", func(b *testing.B) {
		v1 := math.NewVec2(rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vv = v1.Unit()
		}
	})
	b.Run("Vec3", func(b *testing.B) {
		v1 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vvv = v1.Unit()
		}
	})
	b.Run("Vec4", func(b *testing.B) {
		v1 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vvvv = v1.Unit()
		}
	})
}
func BenchmarkVec_Apply(b *testing.B) {
	b.Run("Vec2", func(b *testing.B) {
		v1 := math.NewVec2(rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vv = v1.Unit()
		}
	})
	b.Run("Vec3", func(b *testing.B) {
		v1 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vvv = v1.Unit()
		}
	})
	b.Run("Vec4", func(b *testing.B) {
		v1 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vvvv = v1.Unit()
		}
	})
}

func BenchmarkVec_Cross(b *testing.B) {
	b.Run("Vec3", func(b *testing.B) {
		v1 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())
		v2 := math.NewVec3(rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vvv = v1.Cross(v2)
		}
	})
	b.Run("Vec4", func(b *testing.B) {
		v1 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())
		v2 := math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vvvv = v1.Cross(v2)
		}
	})
}
