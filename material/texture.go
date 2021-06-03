package material

import (
	"image"
	"image/color"

	"changkun.de/x/ddd/math"
	"golang.org/x/image/draw"
)

// Texture represents a power-of-two 2D texture. The power-of-two means
// that the texture width and height must be a power of two. e.g. 1024x1024.
type Texture struct {
	Size   int
	mipmap []*image.RGBA
}

func NewTexture(data *image.RGBA) *Texture {
	if data.Bounds().Dx() != data.Bounds().Dy() {
		panic("image data width and height is not equal!")
	}
	t := &Texture{Size: data.Bounds().Dx()}
	if t.Size == 0 || t.Size == 1 {
		t.mipmap = []*image.RGBA{data}
		return t
	}

	if t.Size%2 != 0 || t.Size < 0 {
		panic("invalid texture size!")
	}

	L := int(math.Log2(float64(t.Size)) + 1)
	t.mipmap = make([]*image.RGBA, L)
	t.mipmap[0] = data

	for i := 1; i < L; i++ {
		size := t.Size / int(math.Pow(2, float64(i)))
		t.mipmap[i] = image.NewRGBA(image.Rect(0, 0, size, size))
		draw.CatmullRom.Scale(
			t.mipmap[i], t.mipmap[i].Bounds(),
			data, image.Rectangle{
				image.Point{0, 0},
				image.Point{size, size},
			}, draw.Over, nil)
	}

	return t
}

// Query fetches the color of at pixel (u, v). This function is a naive
// mipmap implementation that does magnification and minification.
func (t *Texture) Query(u, v float64, lod float64) color.RGBA {
	// Early error checking.
	if u < 0 || u > 1 || v < 0 || v > 1 {
		panic("out of UV query range")
	}

	// Make sure LOD is sitting on a valid range before proceed.
	if lod < 0 {
		lod = 0
	} else if lod >= float64(len(t.mipmap)-1) {
		lod = float64(len(t.mipmap) - 2)
	}

	// if lod < 1 {
	// 	return t.queryL0(u, v)
	// }
	// lod -= 1

	// Figure out two different mipmap levels, then compute
	// tri-linear interpolation between the two discrete mipmap levels.
	highLOD := int(math.Floor(lod))
	lowLOD := int(math.Floor(lod)) + 1
	if lowLOD > len(t.mipmap)-1 {
		lowLOD -= 1
	}
	return t.queryTrilinear(highLOD, lowLOD, lod-math.Floor(lod), u, v)
}

func (t *Texture) queryL0(u, v float64) color.RGBA {
	tex := t.mipmap[0]
	x := int(math.Floor(u * float64(t.Size-1))) // very coarse approximation.
	y := int(math.Floor(v * float64(t.Size-1))) // very coarse approximation.
	return tex.At(x, y).(color.RGBA)
}

func (t *Texture) queryTrilinear(h, l int, p, u, v float64) color.RGBA {
	size := float64(t.Size)
	L1 := t.queryBilinear(
		h,
		(u*size)/math.Pow(2, float64(h)),
		(v*size)/math.Pow(2, float64(h)),
	)
	L2 := t.queryBilinear(
		l,
		(u*size)/math.Pow(2, float64(l)),
		(v*size)/math.Pow(2, float64(l)),
	)
	return math.LerpC(L2, L1, p)
}

func (t *Texture) queryBilinear(lod int, x, y float64) color.RGBA {
	buf := t.mipmap[lod]
	size := buf.Bounds().Dx()
	if size == 1 {
		return buf.At(0, 0).(color.RGBA)
	}

	x = math.Floor(x)
	if x+1 >= float64(size) {
		x -= 1
	}
	y = math.Floor(y)
	if y+1 >= float64(size) {
		y -= 1
	}

	i := int(x)
	j := int(y)

	p1 := buf.At(i, j).(color.RGBA)
	p2 := buf.At(i+1, j).(color.RGBA)
	interpo1 := math.LerpC(p1, p2, x-math.Floor(x))
	p3 := buf.At(i, j+1).(color.RGBA)
	p4 := buf.At(i+1, j+1).(color.RGBA)
	interpo2 := math.LerpC(p3, p4, x-math.Floor(x))

	return math.LerpC(interpo1, interpo2, y-math.Floor(y))
}
