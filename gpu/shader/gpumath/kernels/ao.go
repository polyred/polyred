// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kernels

import . "poly.red/gpu/shader/gpumath"

// AO is the screen-space ambient occlusion pass (material/ao.go), authored once:
// it runs as Go on the CPU and its source (AOSrc) compiles to the GPU. For 8
// directions it marches the depth buffer, accumulates the max elevation angle,
// and darkens by pow(total, 10000). It runs in place after Shade/Shadow over the
// shaded float buffer. au = [width, height, _, _].
func AO(gid uint, fragxyz []float32, aoflag []float32, depthbuf []float32, color []float32, au []float32) {
	if aoflag[gid] < 0.5 {
		return
	}
	px := fragxyz[gid*4]
	py := fragxyz[gid*4+1]
	traceDepth := fragxyz[gid*4+2]
	width := int(au[0])
	height := int(au[1])
	total := float32(0)
	for d := 0; d < 8; d++ {
		ang := float32(d) * 0.78539816339744830961
		dirX := Cos(ang)
		dirY := Sin(ang)
		maxangle := float32(0)
		for t := 0; t < 100; t++ {
			ft := float32(t)
			dx := dirX * ft
			dy := dirY * ft
			distance := Sqrt(dx*dx + dy*dy)
			if distance >= 1.0 {
				ix := int(px + dx)
				iy := int(py + dy)
				if ix >= 0 {
					if ix < width {
						if iy >= 0 {
							if iy < height {
								elevation := depthbuf[iy*width+ix] - traceDepth
								maxangle = Maxf(maxangle, Atan(elevation/distance))
							}
						}
					}
				}
			}
		}
		total = total + (1.57079632679489661923 - maxangle)
	}
	total = total / (1.57079632679489661923 * 8.0)
	total = Pow(total, 10000.0)
	color[gid*4] = Floor(Clampf(Round(color[gid*4]), 0.0, 255.0) * total)
	color[gid*4+1] = Floor(Clampf(Round(color[gid*4+1]), 0.0, 255.0) * total)
	color[gid*4+2] = Floor(Clampf(Round(color[gid*4+2]), 0.0, 255.0) * total)
}
