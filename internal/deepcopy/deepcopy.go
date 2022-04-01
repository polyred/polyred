// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package deepcopy implements the proposal https://go.dev/issue/51520.
package deepcopy

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

// Value copies src to dst recursively.
//
// Two values of identical type are deeply copied if one of the following
// cases apply.
//
// Numbers, bools, strings are deeply copied and have different underlying
// memory address.
//
// Slice and Array values are deeply copied, including its elements.
//
// Map values are deeply copied for all of its key and corresponding
// values.
//
// Pointer values are deeply copied for their pointed value, and the
// pointer points to the deeply copied value.
//
// Struct values are deeply copied for all fields, including exported
// and unexported.
//
// Interface values are deeply copied if the underlying type can be
// deeply copied.
//
// There are a few exceptions that may result in a deeply copied value not
// deeply equal (asserted by DeepEqual(dst, src)) to the source value:
//
// 1) Func values are still refer to the same function
// 2) Chan values are replaced by newly created channels
// 3) One-way Chan values (receive or read-only) values are still refer
//    to the same channel
//
// Note that while correct uses of Value do exist, they are not rare.
// The use of Value often indicates the copying object does not contain
// a singleton or is never meant to be copied, such as sync.Mutex, os.File,
// net.Conn, js.Value, etc. In these cases, the copied value retains the
// memory representations of the source value but may result in unexpected
// consequences in follow-up usage, the caller should clear these values
// depending on their usage context.
func Value[T any](src T) (dst T) {
	ptrs := map[uintptr]any{}
	ret := copyAny(src, ptrs)
	var ok bool
	if dst, ok = ret.(T); ok || ret == nil {
		return
	}
	panic(fmt.Sprintf("reflect: internal error: copied value is not typed in %T, got %T", src, ret))
}

func copyAny(src any, ptrs map[uintptr]any) (dst any) {
	v := reflect.ValueOf(src)
	if !v.IsValid() {
		return src
	}

	// Look up the corresponding copy function.
	switch v.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128, reflect.Func:
		dst = copyPremitive(src, ptrs)
	case reflect.String:
		dst = strings.Clone(src.(string))
	case reflect.Slice:
		dst = copySlice(src, ptrs)
	case reflect.Array:
		dst = copyArray(src, ptrs)
	case reflect.Map:
		dst = copyMap(src, ptrs)
	case reflect.Ptr, reflect.UnsafePointer:
		dst = copyPointer(src, ptrs)
	case reflect.Struct:
		dst = copyStruct(src, ptrs)
	case reflect.Interface:
		dst = copyAny(src, ptrs)
	case reflect.Chan:
		dst = copyChan(src, ptrs)
	default:
		panic(fmt.Sprintf("reflect: internal error: unknown type %v", v.Kind()))
	}
	return
}

func copyPremitive(src any, ptr map[uintptr]any) (dst any) {
	kind := reflect.ValueOf(src).Kind()
	switch kind {
	case reflect.Array, reflect.Chan, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.Struct, reflect.UnsafePointer:
		panic(fmt.Sprintf("reflect: internal error: type %v is not a primitive", kind))
	}
	dst = src
	return
}

func copySlice(x any, ptrs map[uintptr]any) any {
	v := reflect.ValueOf(x)
	kind := v.Kind()
	if kind != reflect.Slice {
		panic(fmt.Sprintf("reflect: internal error: type %v is not a slice", kind))
	}

	size := v.Len()
	t := reflect.TypeOf(x)
	dc := reflect.MakeSlice(t, size, size)
	for i := 0; i < size; i++ {
		iv := reflect.ValueOf(copyAny(v.Index(i).Interface(), ptrs))
		if iv.IsValid() {
			dc.Index(i).Set(iv)
		}
	}
	return dc.Interface()
}

func copyArray(x any, ptrs map[uintptr]any) any {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Array {
		panic(fmt.Errorf("reflect: internal error: must be an Array; got %v", v.Kind()))
	}
	t := reflect.TypeOf(x)
	size := t.Len()
	dc := reflect.New(reflect.ArrayOf(size, t.Elem())).Elem()
	for i := 0; i < size; i++ {
		item := copyAny(v.Index(i).Interface(), ptrs)
		dc.Index(i).Set(reflect.ValueOf(item))
	}
	return dc.Interface()
}

func copyMap(x any, ptrs map[uintptr]any) any {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Map {
		panic(fmt.Errorf("reflect: internal error: must be a Map; got %v", v.Kind()))
	}
	t := reflect.TypeOf(x)
	dc := reflect.MakeMapWithSize(t, v.Len())
	iter := v.MapRange()
	for iter.Next() {
		item := copyAny(iter.Value().Interface(), ptrs)
		k := copyAny(iter.Key().Interface(), ptrs)
		dc.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(item))
	}
	return dc.Interface()
}

func copyPointer(x any, ptrs map[uintptr]any) any {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Pointer {
		panic(fmt.Errorf("reflect: internal error: must be a Pointer or Ptr; got %v", v.Kind()))
	}
	addr := uintptr(v.UnsafePointer())
	if dc, ok := ptrs[addr]; ok {
		return dc
	}
	t := reflect.TypeOf(x)
	dc := reflect.New(t.Elem())
	ptrs[addr] = dc.Interface()
	if !v.IsNil() {
		item := copyAny(v.Elem().Interface(), ptrs)
		iv := reflect.ValueOf(item)
		if iv.IsValid() {
			dc.Elem().Set(reflect.ValueOf(item))
		}
	}
	return dc.Interface()
}

func copyStruct(x any, ptrs map[uintptr]any) any {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Struct {
		panic(fmt.Errorf("reflect: internal error: must be a Struct; got %v", v.Kind()))
	}
	t := reflect.TypeOf(x)
	dc := reflect.New(t)
	for i := 0; i < t.NumField(); i++ {
		item := copyAny(valueInterfaceUnsafe(v.Field(i)), ptrs)
		setField(dc.Elem().Field(i), reflect.ValueOf(item))
	}
	return dc.Elem().Interface()
}

func copyChan(x any, ptrs map[uintptr]any) any {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Chan {
		panic(fmt.Errorf("reflect: internal error: must be a Chan; got %v", v.Kind()))
	}
	t := reflect.TypeOf(x)
	dir := t.ChanDir()
	var dc any
	switch dir {
	case reflect.BothDir:
		dc = reflect.MakeChan(t, v.Cap()).Interface()
	case reflect.SendDir, reflect.RecvDir:
		dc = x
	}
	return dc
}

// valueInterfaceUnsafe overpasses the reflect package check regarding
// unexported methods.
func valueInterfaceUnsafe(v reflect.Value) any {
	return reflect_valueInterface(v, false)
}

// setField sets the given value to the field value, regardless whether
// the filed is exported or not.
func setField(field reflect.Value, value reflect.Value) {
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(value)
}

//go:linkname reflect_valueInterface reflect.valueInterface
func reflect_valueInterface(v reflect.Value, safe bool) any
