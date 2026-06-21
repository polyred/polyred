// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kernels

import _ "embed"

// ShadeSrc is the source of shade.go, exposed so the Go->shader compiler can
// compile the very same kernel that runs as Go on the CPU. The compiler ignores
// the package/import lines and lowers the gpumath calls to shader builtins.
//
//go:embed shade.go
var ShadeSrc string

// SRGBSrc is the source of srgb.go (the gamma pass).
//
//go:embed srgb.go
var SRGBSrc string

// ShadowSrc is the source of shadow.go (the shadow-map pass).
//
//go:embed shadow.go
var ShadowSrc string

// AOSrc is the source of ao.go (the screen-space ambient occlusion pass).
//
//go:embed ao.go
var AOSrc string
