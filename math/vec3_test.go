package math_test

import (
	"testing"

	"poly.red/math"
)

func TestVec3_Less(t *testing.T) {
	v1 := math.Vec3[float32]{1, 2, 3}
	v2 := math.Vec3[float32]{1, 2, 3}

	if v1.Less(v2) {
		t.Fatalf("unexpected Less results, expect equal, but got: %v", v1.Less(v2))
	}

	v3 := math.Vec3[float32]{1.000001, 2.00000001, 3.0000000001}
	if v1.Less(v3) {
		t.Fatalf("unexpected Less results, expect equal, but got: %v", v1.Less(v2))
	}
}
