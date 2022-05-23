// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package deepcopy_test

import (
	"fmt"
	"reflect"
	"testing"

	"poly.red/internal/deepcopy"
)

func ExampleValue() {
	tests := []any{
		`"Now cut that out!"`,
		39,
		true,
		false,
		2.14,
		[]string{
			"Phil Harris",
			"Rochester van Jones",
			"Mary Livingstone",
			"Dennis Day",
		},
		[2]string{
			"Jell-O",
			"Grape-Nuts",
		},
	}

	for _, expected := range tests {
		actual := deepcopy.Value(expected)
		fmt.Println(actual)
	}
	// Output:
	// "Now cut that out!"
	// 39
	// true
	// false
	// 2.14
	// [Phil Harris Rochester van Jones Mary Livingstone Dennis Day]
	// [Jell-O Grape-Nuts]
}

type Foo struct {
	Foo *Foo
	Bar int
}

func ExampleMap() {
	x := map[string]*Foo{
		"foo": {Bar: 1},
		"bar": {Bar: 2},
	}
	y := deepcopy.Value(x)
	for _, k := range []string{"foo", "bar"} { // to ensure consistent order
		fmt.Printf("x[\"%v\"] = y[\"%v\"]: %v\n", k, k, x[k] == y[k])
		fmt.Printf("x[\"%v\"].Foo = y[\"%v\"].Foo: %v\n", k, k, x[k].Foo == y[k].Foo)
		fmt.Printf("x[\"%v\"].Bar = y[\"%v\"].Bar: %v\n", k, k, x[k].Bar == y[k].Bar)
	}
	// Output:
	// x["foo"] = y["foo"]: false
	// x["foo"].Foo = y["foo"].Foo: false
	// x["foo"].Bar = y["foo"].Bar: true
	// x["bar"] = y["bar"]: false
	// x["bar"].Foo = y["bar"].Foo: false
	// x["bar"].Bar = y["bar"].Bar: true
}

func TestInterface(t *testing.T) {
	x := []any{nil}
	y := deepcopy.Value(x)
	if !reflect.DeepEqual(x, y) || len(y) != 1 {
		t.Errorf("expect %v == %v; y had length %v (expected 1)", x, y, len(y))
	}
	var a any
	b := deepcopy.Value(a)
	if a != b {
		t.Errorf("expected %v == %v", a, b)
	}
}

func TestAvoidInfiniteLoops(t *testing.T) {
	x := &Foo{
		Bar: 4,
	}
	x.Foo = x
	y := deepcopy.Value(x)
	if x == y {
		t.Fatalf("expect x==y returns false")
	}
	if x != x.Foo {
		t.Fatalf("expect x==x.Foo returns true")
	}
	if y != y.Foo {
		t.Fatalf("expect y==y.Foo returns true")
	}
}

func TestSpecialKind(t *testing.T) {
	x := func() int { return 42 }
	y := deepcopy.Value(x)
	if y() != 42 || reflect.DeepEqual(x, y) {
		t.Fatalf("copied value is equal")
	}

	a := map[bool]any{true: x}
	b := deepcopy.Value(a)
	if reflect.DeepEqual(a, b) {
		t.Fatalf("copied value is equal")
	}

	c := []any{x}
	d := deepcopy.Value(c)
	if reflect.DeepEqual(c, d) {
		t.Fatalf("copied value is equal")
	}
}
