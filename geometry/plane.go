package geometry

import (
	"image/color"

	"changkun.de/x/ddd/geometry/primitive"
	"changkun.de/x/ddd/math"
)

func NewPlane(width, height float64) *TriangleMesh {
	v1 := primitive.Vertex{
		Pos: math.NewVector(-0.5*width, 0, -0.5*height, 1),
		UV:  math.NewVector(0, 1, 0, 1),
		Nor: math.NewVector(0, 1, 0, 0),
		Col: color.RGBA{255, 0, 0, 255},
	}
	v2 := primitive.Vertex{
		Pos: math.NewVector(-0.5*width, 0, 0.5*height, 1),
		UV:  math.NewVector(0, 0, 0, 1),
		Nor: math.NewVector(0, 1, 0, 0),
		Col: color.RGBA{0, 255, 0, 255},
	}
	v3 := primitive.Vertex{
		Pos: math.NewVector(0.5*width, 0, 0.5*height, 1),
		UV:  math.NewVector(1, 0, 0, 1),
		Nor: math.NewVector(0, 1, 0, 0),
		Col: color.RGBA{0, 0, 255, 255},
	}
	v4 := primitive.Vertex{
		Pos: math.NewVector(0.5*width, 0, -0.5*height, 1),
		UV:  math.NewVector(1, 1, 0, 1),
		Nor: math.NewVector(0, 1, 0, 0),
		Col: color.RGBA{0, 0, 0, 255},
	}
	return NewTriangleMesh([]*primitive.Triangle{
		{V1: v1, V2: v2, V3: v3},
		{V1: v1, V2: v3, V3: v4},
	})
}
