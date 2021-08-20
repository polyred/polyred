package bbb

import (
	"reflect"
	"testing"
	"unsafe"

	"poly.red/texture/buffer"
)

const size = unsafe.Sizeof(buffer.Fragment{})

var v *buffer.Fragment

func BenchmarkAccess(b *testing.B) {
	buf := []buffer.Fragment{
		{Ok: true},
		{Ok: true},
		{Ok: true},
		{Ok: true},
		{Ok: true},
		{Ok: true},
		{Ok: true},
		{Ok: true},
		{Ok: true},
		{Ok: true},
	}
	l := len(buf)
	b.Run("safe", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v = index(buf, i%l)
		}
	})
	b.Run("unsafe", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v = (*buffer.Fragment)(unsafe.Pointer(
				(*reflect.SliceHeader)(unsafe.Pointer(&buf)).Data + (uintptr(i%l) * size),
			))
		}
	})
}
