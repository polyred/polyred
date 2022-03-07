// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh_test

import (
	"fmt"
	"testing"

	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
)

func BenchmarkNewTriangleMesh(b *testing.B) {

	for n := 1024; n < 1024*8; n *= 2 {
		ts := make([]*primitive.Triangle, n)
		for i := 0; i < n; i++ {
			t := &primitive.Triangle{
				V1: primitive.NewRandomVertex(),
				V2: primitive.NewRandomVertex(),
				V3: primitive.NewRandomVertex(),
			}
			ts[i] = t
		}

		b.Run(fmt.Sprintf("tri-%d", n), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				mesh.NewTriangleSoup(ts)
			}
		})
	}

}
