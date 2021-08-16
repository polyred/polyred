// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package term

import "fmt"

var (
	before   = []byte("\033[")
	after    = []byte("m")
	reset    = []byte("\033[0;00m")
	fgcolors = fgTermRGB[16:232]
	bgcolors = bgTermRGB[16:232]
)

// fgBytes colorizes the foreground with the terminal color that matches
// the closest the RGB color.
func fgBytes(in []byte, r, g, b uint8) []byte {
	return colorize(col(r, g, b, true), in)
}

// bgBytes colorizes the background with the terminal color that matches
// the closest the RGB color.
func bgBytes(in []byte, r, g, b uint8) []byte {
	return colorize(col(r, g, b, false), in)
}

func colorize(color, in []byte) []byte {
	return append(append(append(append(before, color...), after...), in...), reset...)
}

func col(r, g, b uint8, foreground bool) []byte {
	// if all colors are equal, it might be in the grayscale range
	if r == g && g == b {
		color, ok := grayscale(r, foreground)
		if ok {
			return color
		}
	}

	// the general case approximates RGB by using the closest color.
	r6 := ((uint16(r) * 5) / 255)
	g6 := ((uint16(g) * 5) / 255)
	b6 := ((uint16(b) * 5) / 255)
	i := 36*r6 + 6*g6 + b6
	if foreground {
		return fgcolors[i]
	}
	return bgcolors[i]
}

func grayscale(scale uint8, foreground bool) ([]byte, bool) {
	var source [256][]byte

	if foreground {
		source = fgTermRGB
	} else {
		source = bgTermRGB
	}

	switch scale {
	case 0x08:
		return source[232], true
	case 0x12:
		return source[233], true
	case 0x1c:
		return source[234], true
	case 0x26:
		return source[235], true
	case 0x30:
		return source[236], true
	case 0x3a:
		return source[237], true
	case 0x44:
		return source[238], true
	case 0x4e:
		return source[239], true
	case 0x58:
		return source[240], true
	case 0x62:
		return source[241], true
	case 0x6c:
		return source[242], true
	case 0x76:
		return source[243], true
	case 0x80:
		return source[244], true
	case 0x8a:
		return source[245], true
	case 0x94:
		return source[246], true
	case 0x9e:
		return source[247], true
	case 0xa8:
		return source[248], true
	case 0xb2:
		return source[249], true
	case 0xbc:
		return source[250], true
	case 0xc6:
		return source[251], true
	case 0xd0:
		return source[252], true
	case 0xda:
		return source[253], true
	case 0xe4:
		return source[254], true
	case 0xee:
		return source[255], true
	}
	return nil, false
}

// \033[

var (
	fgTermRGB = [256][]byte{}
	bgTermRGB = [256][]byte{}
)

func init() {
	s1 := "38;5;%d"
	s2 := "48;5;%d"

	for i := 0; i < 256; i++ {
		fgTermRGB[i] = []byte(fmt.Sprintf(s1, i))
		bgTermRGB[i] = []byte(fmt.Sprintf(s2, i))
	}
}
