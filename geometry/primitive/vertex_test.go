package primitive_test

import (
	"reflect"
	"testing"

	"poly.red/geometry/primitive"
)

func TestVertex_Copy(t *testing.T) {
	v := primitive.NewRandomVertex()
	u := v.Copy()

	if !reflect.DeepEqual(v, u) {
		t.Fatalf("copied content is not deeply equal to source")
	}
	if u == v {
		t.Fatalf("copied objects have same address")
	}
}

func TestVertex_AABB(t *testing.T) {
	v := primitive.NewRandomVertex()
	pos := v.Pos.ToVec3()
	aabb := v.AABB()

	if !aabb.Min.Eq(pos) {
		t.Errorf("vertex aabb min is not euqal to the position")
	}
	if !aabb.Max.Eq(pos) {
		t.Errorf("vertex aabb max is not euqal to the position")
	}
}
