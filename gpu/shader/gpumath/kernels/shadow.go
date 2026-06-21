// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kernels

import . "poly.red/gpu/shader/gpumath"

// Shadow multiplies each shaded color by a shadow factor computed from N shadow
// maps, authored once: it runs as Go on the CPU and its source (ShadowSrc)
// compiles to the GPU. It runs in place after Shade over the shaded float buffer.
// su = [width, depthLen, n, _]; mats holds N column-major light view-projection
// matrices; depths holds N shadow maps of depthLen each.
func Shadow(gid uint, fragxyz []float32, recv []float32, depths []float32, mats []float32, color []float32, su []float32) {
	if recv[gid] < 0.5 {
		return
	}
	fx := fragxyz[gid*4]
	fy := fragxyz[gid*4+1]
	fz := fragxyz[gid*4+2]
	occ := float32(0)
	width := int(su[0])
	dl := int(su[1])
	n := int(su[2])
	for k := 0; k < n; k++ {
		M := M4(
			V4(mats[k*16], mats[k*16+1], mats[k*16+2], mats[k*16+3]),
			V4(mats[k*16+4], mats[k*16+5], mats[k*16+6], mats[k*16+7]),
			V4(mats[k*16+8], mats[k*16+9], mats[k*16+10], mats[k*16+11]),
			V4(mats[k*16+12], mats[k*16+13], mats[k*16+14], mats[k*16+15]),
		)
		clip := M.MulV(V4(fx, fy, fz, 1))
		sx := clip.X / clip.W
		sy := clip.Y / clip.W
		sz := clip.Z / clip.W
		idx := int(sx) + int(sy)*width
		if idx > 0 {
			if idx < dl {
				if sz < depths[k*dl+idx]-0.03 {
					occ = occ + 1
				}
			}
		}
	}
	wf := Pow(0.5, occ)
	// Match the engine: uint8(clamp(round(blinn),0,255) * w), truncated.
	color[gid*4] = Floor(Clampf(Round(color[gid*4]), 0.0, 255.0) * wf)
	color[gid*4+1] = Floor(Clampf(Round(color[gid*4+1]), 0.0, 255.0) * wf)
	color[gid*4+2] = Floor(Clampf(Round(color[gid*4+2]), 0.0, 255.0) * wf)
}
