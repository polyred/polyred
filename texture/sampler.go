// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package texture

type Format int

const (
	FormatRGBA Format = iota
	FormatBGRA
	FormatDepth
)

type Encode int

const (
	EncodeLinear Encode = iota
	EncodeSRGB
	EncodeGamma
)

// Filter is a texture filter.
type Filter int

const (
	// FilterNearest gets the nearest value of the texture element
	// (in Manhattan distance) to the specified texture coordinates.
	FilterNearest Filter = iota
	// FilterNearest gets the weighted average of the four closest
	// texture elements, and can include items wrapped or repeated from
	// other parts of a texture, depending on wrapping approach, and
	// on the exact mapping.
	FilterLinear

	// Magfilter
	FilterNearestMipmapNearest
	FilterNearestMipmapLinear
	FilterLinearMipmapNearest
	FilterLinearMipmapLinear
)

// Wrap define the texture's horizontal and vertical wrapping.
type Wrap int

const (
	// WrapRepeat allows texture repeat to infinity.
	WrapRepeat Wrap = iota
	// WrapEdge uses the last pixel of the texture stretches to the edge of the mesh.
	WrapEdge
	// WrapMirror repeats to infinity, but mirroring on each repeat.
	WrapMirror
)
