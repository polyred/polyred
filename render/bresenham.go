package render

import (
	"image/color"

	"poly.red/math"
	"poly.red/texture/buffer"
)

// drawLine implements the Bresenham algorithm that draws a line
// segment starting from p1 and ends at p2. The drawn pixels are
// stored in a given buffer.
func DrawLine(buf *buffer.Buffer, p1 math.Vec4, p2 math.Vec4, color color.RGBA) {
	// TODO: test it with a demo
	if math.Abs(p2.Y-p1.Y) < math.Abs(p2.X-p1.X) {
		if p1.X > p2.X {
			p1, p2 = p2, p1
		}
		drawLineLow(buf, p1, p2, color)
	} else {
		if p1.Y > p2.Y {
			p1, p2 = p2, p1
		}
		drawLineHigh(buf, p1, p2, color)
	}
}

func drawLineLow(buf *buffer.Buffer, p1 math.Vec4, p2 math.Vec4, color color.RGBA) {
	x0 := math.Round(p1.X)
	y0 := math.Round(p1.Y)
	z0 := p1.Z
	x1 := math.Round(p2.X)
	y1 := math.Round(p2.Y)
	z1 := p2.Z

	dx := x1 - x0
	dy := y1 - y0
	yi := 1.0
	if dy < 0 {
		yi = -1
		dy = -dy
	}
	D := 2*dy - dx
	y := y0
	for x := x0; x <= x1; x++ {
		z := ((z1-z0)*(x-x0))/(x1-x0) + z0
		// Dealing with numeric issues. The interpolated z value above
		// might be quite different than the z value computed via
		// barycentric interpolation numerically. We use an approximate
		// z value to draw the wireframe, instead of using a depth test.
		// This approach may fail when the object is further away from
		// the camera. The caller does not have to worry about this.
		// Use it as-is. A known better approach is to cooperate with
		// barycentric coordinates using gradient but requires more
		// computation.
		//   if (approxEqual(depthBuf[x + y * this.width], z, epsilon)) {
		//     fragmentProcessing(frameBuf, depthBuf, x, y, z, color);
		//   }
		if buf.DepthTest(int(x), int(y), z) {
			buf.Set(int(x), int(y), buffer.Fragment{})
		}
		if D > 0 {
			y += yi
			D -= 2 * dx
		}
		D += 2 * dy
	}
}
func drawLineHigh(buf *buffer.Buffer, p1 math.Vec4, p2 math.Vec4, color color.RGBA) {
	x0 := math.Round(p1.X)
	y0 := math.Round(p1.Y)
	z0 := p1.Z
	x1 := math.Round(p2.X)
	y1 := math.Round(p2.Y)
	z1 := p2.Z

	dx := x1 - x0
	dy := y1 - y0
	xi := 1.0
	if dx < 0 {
		xi = -1
		dx = -dx
	}
	D := 2*dx - dy
	x := x0
	for y := y0; y <= y1; y++ {
		z := ((z1-z0)*(y-y0))/(y1-y0) + z0
		// Dealing with numeric issues. The interpolated z value above
		// might be quite different than the z value computed via
		// barycentric interpolation numerically. We use an approximate
		// z value to draw the wireframe, instead of using a depth test.
		// This approach may fail when the object is further away from
		// the camera. The caller does not have to worry about this.
		// Use it as-is. A known better approach is to cooperate with
		// barycentric coordinates using gradient but requires more
		// computation.
		// if approxEqual(depthBuf[x+y*this.width], z, epsilon) {
		// 	this.fragmentProcessing(frameBuf, depthBuf, x, y, z, color)
		// }
		if buf.DepthTest(int(x), int(y), z) {
			buf.Set(int(x), int(y), buffer.Fragment{})
		}
		if D > 0 {
			x += xi
			D -= 2 * dy
		}
		D += 2 * dx
	}
}
