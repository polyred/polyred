// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !darwin

package tests

import "poly.red/math"

// The original non-darwin GPU compute demo used a GL/GLES path that imported
// packages removed in the 2022 restructure (poly.red/internal/driver/egl and
// .../gles), which broke cross-platform builds. GPU compute is now provided
// through the cgo-free poly.red/gpu abstraction (Metal today; a cgo-free GL
// backend is a later phase). Until that backend lands, this package reports no
// device on non-darwin platforms, so its GPU tests skip.

type computeFunc struct{}

type shaderFn struct {
	funcAdd, funcSub, funcSqrt, funcMul computeFunc
}

type unavailableDevice struct{}

func (unavailableDevice) Available() bool { return false }

var device = unavailableDevice{}

const noBackend = "gpu/tests: no GPU compute backend on this platform yet (see poly.red/gpu)"

func add[T DataType](m1, m2 math.Mat[T]) math.Mat[T] { panic(noBackend) }
func sub[T DataType](m1, m2 math.Mat[T]) math.Mat[T] { panic(noBackend) }
func sqrt[T DataType](m math.Mat[T]) math.Mat[T]     { panic(noBackend) }
func mul[T DataType](m1, m2 math.Mat[T]) math.Mat[T] { panic(noBackend) }
