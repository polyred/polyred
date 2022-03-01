// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

// Int26_6 is a signed 26.6 fixed-point number.
//
// The integer part ranges from -33554432 to 33554431, inclusive. The
// fractional part has 6 bits of precision.
//
// For example, the number one-and-a-quarter is Int26_6(1<<6 + 1<<4).
type Int26_6 int32

// Point26_6 is a 26.6 fixed-point coordinate pair.
//
// It is analogous to the image.Point type in the standard library.
type Point26_6 struct {
	X, Y Int26_6
}

// I returns the integer value i as an Int26_6.
//
// For example, passing the integer value 2 yields Int26_6(128).
func I(i int) Int26_6 {
	return Int26_6(i << 6)
}

// P returns the integer values x and y as a Point26_6.
//
// For example, passing the integer values (2, -3) yields Point26_6{128, -192}.
func P(x, y int) Point26_6 {
	return Point26_6{Int26_6(x << 6), Int26_6(y << 6)}
}

// Rectangle26_6 is a 26.6 fixed-point coordinate rectangle. The Min bound is
// inclusive and the Max bound is exclusive. It is well-formed if Min.X <=
// Max.X and likewise for Y.
//
// It is analogous to the image.Rectangle type in the standard library.
type Rectangle26_6 struct {
	Min, Max Point26_6
}

// R returns the integer values minX, minY, maxX, maxY as a Rectangle26_6.
//
// For example, passing the integer values (0, 1, 2, 3) yields
// Rectangle26_6{Point26_6{0, 64}, Point26_6{128, 192}}.
//
// Like the image.Rect function in the standard library, the returned rectangle
// has minimum and maximum coordinates swapped if necessary so that it is
// well-formed.
func R(minX, minY, maxX, maxY int) Rectangle26_6 {
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	if minY > maxY {
		minY, maxY = maxY, minY
	}
	return Rectangle26_6{
		Point26_6{
			Int26_6(minX << 6),
			Int26_6(minY << 6),
		},
		Point26_6{
			Int26_6(maxX << 6),
			Int26_6(maxY << 6),
		},
	}
}
