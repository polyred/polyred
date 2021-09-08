package font

import (
	"image"
	"image/draw"
	"io"

	"poly.red/math"
)

// Drawer draws text on a destination image.
//
// A Drawer is not safe for concurrent use by multiple goroutines, since its
// Face is not.
type Drawer struct {
	// Dst is the destination image.
	Dst draw.Image
	// Src is the source image.
	Src image.Image
	// Face provides the glyph mask images.
	Face Face
	// Dot is the baseline location to draw the next glyph. The majority of the
	// affected pixels will be above and to the right of the dot, but some may
	// be below or to the left. For example, drawing a 'j' in an italic face
	// may affect pixels below and to the left of the dot.
	Dot math.Point26_6
}

// DrawString draws s at the dot and advances the dot's location.
func (d *Drawer) DrawString(s string) {
	prevC := rune(-1)
	for _, c := range s {
		if prevC >= 0 {
			d.Dot.X += d.Face.Kern(prevC, c)
		}
		dr, mask, maskp, advance, ok := d.Face.Glyph(d.Dot, c)
		if !ok {
			continue
		}

		draw.DrawMask(d.Dst, dr, d.Src, image.Point{}, mask, maskp, draw.Over)
		d.Dot.X += advance
		prevC = c
	}
}

// Face is a font face. Its glyphs are often derived from a font file, such as
// "Comic_Sans_MS.ttf", but a face has a specific size, style, weight and
// hinting. For example, the 12pt and 18pt versions of Comic Sans are two
// different faces, even if derived from the same font file.
//
// A Face is not safe for concurrent use by multiple goroutines, as its methods
// may re-use implementation-specific caches and mask image buffers.
//
// To create a Face, look to other packages that implement specific font file
// formats.
type Face interface {
	io.Closer

	// Glyph returns the draw.DrawMask parameters (dr, mask, maskp) to draw r's
	// glyph at the sub-pixel destination location dot, and that glyph's
	// advance width.
	//
	// It returns !ok if the face does not contain a glyph for r.
	//
	// The contents of the mask image returned by one Glyph call may change
	// after the next Glyph call. Callers that want to cache the mask must make
	// a copy.
	Glyph(dot math.Point26_6, r rune) (
		dr image.Rectangle, mask image.Image, maskp image.Point, advance math.Int26_6, ok bool)

	// GlyphBounds returns the bounding box of r's glyph, drawn at a dot equal
	// to the origin, and that glyph's advance width.
	//
	// It returns !ok if the face does not contain a glyph for r.
	//
	// The glyph's ascent and descent are equal to -bounds.Min.Y and
	// +bounds.Max.Y. The glyph's left-side and right-side bearings are equal
	// to bounds.Min.X and advance-bounds.Max.X. A visual depiction of what
	// these metrics are is at
	// https://developer.apple.com/library/archive/documentation/TextFonts/Conceptual/CocoaTextArchitecture/Art/glyphterms_2x.png
	GlyphBounds(r rune) (bounds math.Rectangle26_6, advance math.Int26_6, ok bool)

	// GlyphAdvance returns the advance width of r's glyph.
	//
	// It returns !ok if the face does not contain a glyph for r.
	GlyphAdvance(r rune) (advance math.Int26_6, ok bool)

	// Kern returns the horizontal adjustment for the kerning pair (r0, r1). A
	// positive kern means to move the glyphs further apart.
	Kern(r0, r1 rune) math.Int26_6

	// Metrics returns the metrics for this Face.
	Metrics() Metrics
}

// Metrics holds the metrics for a Face. A visual depiction is at
// https://developer.apple.com/library/mac/documentation/TextFonts/Conceptual/CocoaTextArchitecture/Art/glyph_metrics_2x.png
type Metrics struct {
	// Height is the recommended amount of vertical space between two lines of
	// text.
	Height math.Int26_6

	// Ascent is the distance from the top of a line to its baseline.
	Ascent math.Int26_6

	// Descent is the distance from the bottom of a line to its baseline. The
	// value is typically positive, even though a descender goes below the
	// baseline.
	Descent math.Int26_6

	// XHeight is the distance from the top of non-ascending lowercase letters
	// to the baseline.
	XHeight math.Int26_6

	// CapHeight is the distance from the top of uppercase letters to the
	// baseline.
	CapHeight math.Int26_6

	// CaretSlope is the slope of a caret as a vector with the Y axis pointing up.
	// The slope {0, 1} is the vertical caret.
	CaretSlope image.Point
}

// face is a basic font face whose glyphs all have the same metrics.
//
// It is safe to use concurrently.
type face struct {
	// Advance is the glyph advance, in pixels.
	Advance int
	// Width is the glyph width, in pixels.
	Width int
	// Height is the inter-line height, in pixels.
	Height int
	// Ascent is the glyph ascent, in pixels.
	Ascent int
	// Descent is the glyph descent, in pixels.
	Descent int
	// Left is the left side bearing, in pixels. A positive value means that
	// all of a glyph is to the right of the dot.
	Left int

	// Mask contains all of the glyph masks. Its width is typically the Face's
	// Width, and its height a multiple of the Face's Height.
	Mask image.Image
	// Ranges map runes to sub-images of Mask. The rune ranges must not
	// overlap, and must be in increasing rune order.
	Ranges []Range
}

func (f *face) Close() error                  { return nil }
func (f *face) Kern(r0, r1 rune) math.Int26_6 { return 0 }

func (f *face) Metrics() Metrics {
	return Metrics{
		Height:     math.I(f.Height),
		Ascent:     math.I(f.Ascent),
		Descent:    math.I(f.Descent),
		XHeight:    math.I(f.Ascent),
		CapHeight:  math.I(f.Ascent),
		CaretSlope: image.Point{X: 0, Y: 1},
	}
}

func (f *face) Glyph(dot math.Point26_6, r rune) (
	dr image.Rectangle, mask image.Image, maskp image.Point, advance math.Int26_6, ok bool) {

loop:
	for _, rr := range [2]rune{r, '\ufffd'} {
		for _, rng := range f.Ranges {
			if rr < rng.Low || rng.High <= rr {
				continue
			}
			maskp.Y = (int(rr-rng.Low) + rng.Offset) * (f.Ascent + f.Descent)
			ok = true
			break loop
		}
	}
	if !ok {
		return image.Rectangle{}, nil, image.Point{}, 0, false
	}

	x := int(dot.X+32)>>6 + f.Left
	y := int(dot.Y+32) >> 6
	dr = image.Rectangle{
		Min: image.Point{
			X: x,
			Y: y - f.Ascent,
		},
		Max: image.Point{
			X: x + f.Width,
			Y: y + f.Descent,
		},
	}

	return dr, f.Mask, maskp, math.I(f.Advance), true
}

func (f *face) GlyphBounds(r rune) (bounds math.Rectangle26_6, advance math.Int26_6, ok bool) {
	return math.R(0, -f.Ascent, f.Width, +f.Descent), math.I(f.Advance), true
}

func (f *face) GlyphAdvance(r rune) (advance math.Int26_6, ok bool) {
	return math.I(f.Advance), true
}

// Range maps a contiguous range of runes to vertically adjacent sub-images of
// a Face's Mask image. The rune range is inclusive on the low end and
// exclusive on the high end.
//
// If Low <= r && r < High, then the rune r is mapped to the sub-image of
// Face.Mask whose bounds are image.Rect(0, y*h, Face.Width, (y+1)*h),
// where y = (int(r-Low) + Offset) and h = (Face.Ascent + Face.Descent).
type Range struct {
	Low, High rune
	Offset    int
}

// Face7x13 is a Face derived from the public domain X11 misc-fixed font files.
//
// At the moment, it holds the printable characters in ASCII starting with
// space, and the Unicode replacement character U+FFFD.
//
// Its data is entirely self-contained and does not require loading from
// separate files.
var Face7x13 Face = &face{
	Advance: 7,
	Width:   6,
	Height:  13,
	Ascent:  11,
	Descent: 2,
	Mask:    mask7x13,
	Ranges: []Range{
		{'\u0020', '\u007f', 0},
		{'\ufffd', '\ufffe', 95},
	},
}
