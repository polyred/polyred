// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package buffer_test

import (
	"image"
	"reflect"
	"testing"

	"poly.red/buffer"
	"poly.red/color"
	"poly.red/geometry/primitive"
)

func TestNewBuffer(t *testing.T) {
	t.Run("RGBA", func(t *testing.T) {
		buf := buffer.NewBuffer(
			image.Rect(0, 0, 10, 10), buffer.Format(buffer.PixelFormatRGBA))
		if buf.Format() != buffer.PixelFormatRGBA {
			t.Fatalf("set buffer.Format option failed")
		}
	})
	t.Run("BGRA", func(t *testing.T) {
		buf := buffer.NewBuffer(
			image.Rect(0, 0, 10, 10), buffer.Format(buffer.PixelFormatBGRA))
		if buf.Format() != buffer.PixelFormatBGRA {
			t.Fatalf("set buffer.Format option failed")
		}
	})
}

func TestBuffer_FragmentOffset(t *testing.T) {
	t.Run("RGBA", func(t *testing.T) {
		buf := buffer.NewBuffer(image.Rect(0, 0, 10, 10), buffer.Format(buffer.PixelFormatRGBA))
		testBufferFragmentOffset(t, buf)
	})
	t.Run("BGRA", func(t *testing.T) {
		buf := buffer.NewBuffer(image.Rect(0, 0, 10, 10), buffer.Format(buffer.PixelFormatBGRA))
		testBufferFragmentOffset(t, buf)
	})
}

func testBufferFragmentOffset(t *testing.T, buf *buffer.FragmentBuffer) {
	w, h := 10, 10
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			idx := buf.FragmentOffset(i, j)
			if idx < 0 || idx > w*h {
				t.Fatalf("invalid fragment offset")
			}
		}
	}

	counter := 0
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			idx := buf.PixelOffset(i, h-j-1)
			counter++
			if idx < 0 || idx > w*h*4 {
				t.Fatalf("invalid fragment offset")
			}
		}
	}
	if counter != w*h {
		t.Fatalf("missing pixel offset")
	}
}

func newBuf(w, h int) []*buffer.FragmentBuffer {
	buf1 := buffer.NewBuffer(image.Rect(0, 0, w, h),
		buffer.Format(buffer.PixelFormatRGBA))
	buf2 := buffer.NewBuffer(image.Rect(0, 0, w, h),
		buffer.Format(buffer.PixelFormatBGRA))
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			buf1.Set(i, j, buffer.Fragment{
				Ok: true,
				Fragment: primitive.Fragment{
					X: i, Y: j, Depth: 1, Col: color.White,
				},
			})
			buf2.Set(i, j, buffer.Fragment{
				Ok: true,
				Fragment: primitive.Fragment{
					X: i, Y: j, Depth: 1, Col: color.White,
				},
			})
		}
	}
	return []*buffer.FragmentBuffer{buf1, buf2}
}

func TestBuffer_Clear(t *testing.T) {
	bufs := newBuf(10, 10)

	for _, buf := range bufs {
		buf.Clear()
		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				if !reflect.DeepEqual(buf.Get(i, j), buffer.Fragment{}) {
					t.Fatalf("cleared buffer still have non-zero value at (%d,%d), got %+v", i, j, buf.Get(i, j))
				}
			}
		}
	}
}

func TestBuffer_Image(t *testing.T) {
	bufs := newBuf(10, 10)
	for _, buf := range bufs {
		pix := buf.Image()

		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				if !reflect.DeepEqual(pix.RGBAAt(i, j), color.White) {
					t.Fatalf("returned frame buffer is not a white image at (%d,%d), got %+v", i, j, pix.RGBAAt(i, j))
				}
			}
		}
	}
}

func TestBuffer_Depth(t *testing.T) {
	bufs := newBuf(10, 10)
	for _, buf := range bufs {
		pix := buf.Depth()

		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				if !reflect.DeepEqual(pix.RGBAAt(i, j), color.White) {
					t.Fatalf("returned depth buffer is not white at (%d,%d), got %+v", i, j, pix.RGBAAt(i, j))
				}
			}
		}
	}
}

func TestBuffer_Access(t *testing.T) {
	bufs := newBuf(11, 12)

	for _, buf := range bufs {
		if buf.Bounds().Dx() != 11 || buf.Bounds().Dy() != 12 {
			t.Fatalf("unexpected bound of the returned buffer, got (%d, %d)", buf.Bounds().Dx(), buf.Bounds().Dy())
		}

		for i := -10; i < 0; i++ {
			for j := -10; j < 0; j++ {
				if buf.In(i, j) {
					t.Fatalf("invalid pixel access returned success at (%d,%d)", i, j)
				}
				if !reflect.DeepEqual(buf.Get(i, j), buffer.Fragment{}) {
					t.Fatalf("unexpected fragment access at (%d,%d), got %+v, want zero value", i, j, buf.Get(i, j))
				}
				buf.Set(i, j, buffer.Fragment{})
			}
		}

		pix := buf.Image()
		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				if !reflect.DeepEqual(pix.RGBAAt(i, j), color.White) {
					t.Fatalf("returned depth buffer is not white at (%d,%d), got %+v", i, j, pix.RGBAAt(i, j))
				}
			}
		}
	}
}

func BenchmarkBuffer_Clear(b *testing.B) {
	buf := buffer.NewBuffer(image.Rect(0, 0, 800, 800))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear()
	}
}
