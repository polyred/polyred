// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package bytes

import (
	"reflect"
	"unsafe"
)

func FromStruct(s any) []byte {
	v := reflect.ValueOf(s)
	sz := int(v.Elem().Type().Size())
	return unsafe.Slice((*byte)(unsafe.Pointer(v.Pointer())), sz)
}

func FromSlice(s any) []byte {
	v := reflect.ValueOf(s)
	first := v.Index(0)
	sz := int(first.Type().Size())
	res := unsafe.Slice((*byte)(unsafe.Pointer(v.Pointer())), sz*v.Cap())
	return res[:sz*v.Len()]
}

func Convert[To, From any](s []From) []To {
	v := reflect.ValueOf(s)
	first := v.Index(0)
	sz := int(first.Type().Size())
	res := unsafe.Slice((*To)(unsafe.Pointer(v.Pointer())), sz*v.Cap())
	return res[:sz*v.Len()]
}
