// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build !darwin
// +build !darwin

package app

// TODO: for non-darwin systems

// import (
// 	"image"

// 	"poly.red/app/internal/gl"
// )

// type glContext struct {
// }

// func newGLContext(cfg *config) (*glContext, error) {
// 	gl.MakeCurrentContext()
// 	gl.DrawBuffer(gl.FRONT)
// 	gl.PixelZoom(1, -1)
// 	return glContext, nil
// }

// // flush flushes the containing pixel buffer of the given image to the
// // hardware frame buffer for display prupose. The given image is assumed
// // to be non-nil pointer.
// func (w *Window) flush(buf *frameBuf) error {
// 	dx, dy := int32(buf.img.Bounds().Dx()), int32(buf.img.Bounds().Dy())
// 	gl.RasterPos2d(-1, 1)
// 	gl.Viewport(0, 0, dx, dy)
// 	gl.DrawPixels(dx, dy, gl.RGBA, gl.UNSIGNED_BYTE, buf.img.Pix)

// 	// We need a synchornization here. Similar to commandBuffer.WaitUntilCompleted.
// 	// See a general discussion about CPU, GPU and display synchornization here:
// 	//
// 	// Working with Metal: Fundamentals, 21:28
// 	// https://developer.apple.com/videos/play/wwdc2014/604/
// 	//
// 	// The difference of gl.Finish and gl.Flush can be found here:
// 	// https://www.khronos.org/registry/OpenGL-Refpages/gl2.1/xhtml/glFlush.xml
// 	// https://www.khronos.org/registry/OpenGL-Refpages/gl2.1/xhtml/glFinish.xml
// 	//
// 	// We may not need such an wait, if we are doing perfect timing.
// 	// See: https://golang.design/research/ultimate-channel/
// 	gl.Finish()
// 	return nil
// }

// // resetBuffers assign new buffers to the caches window buffers (w.bufs)
// // Note: with Metal, we always use BGRA pixel format.
// func (w *Window) resetBufs(r image.Rectangle) {
// 	// Nothing on OpenGL.
// }
