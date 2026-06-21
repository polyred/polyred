// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kernels

import . "poly.red/gpu/shader/gpumath"

// SRGB is the analytic linear-to-sRGB transfer (color/srgb.go), authored once: it
// runs as Go on the CPU (call it per element) and its source (SRGBSrc) compiles to
// the GPU. It is the renderer's gamma pass; the CPU default uses a LUT
// approximation of the same curve.
func SRGB(gid uint, in []float32, out []float32) {
	v := in[gid]
	if v <= 0.0031308 {
		out[gid] = v * 12.92
	} else {
		out[gid] = 1.055*Pow(v, 0.4166666666) - 0.055
	}
}
