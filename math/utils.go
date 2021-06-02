package math

import "math"

var (
	Cos = math.Cos
	Sin = math.Sin
	Pi  = math.Pi
)

// DefaultEpsilon is a default epsilon value for computation.
const DefaultEpsilon = 1e-7

// ApproxEq approximately compares v1 and v2.
func ApproxEq(v1, v2, epsilon float64) bool {
	return math.Abs(v1-v2) <= epsilon
}

// Clamp clamps a given value in [min, max].
func Clamp(n, min, max float64) float64 {
	return math.Min(math.Max(n, min), max)
}

// ClampV clamps a vector in [min, max].
func ClampV(v Vector, min, max float64) Vector {
	return Vector{
		Clamp(v.X, min, max),
		Clamp(v.Y, min, max),
		Clamp(v.Z, min, max),
		Clamp(v.W, min, max),
	}
}

// ViewportMatrix returns the viewport matrix.
func ViewportMatrix(w, h float64) Matrix {
	return Matrix{
		w / 2, 0, 0, w / 2,
		0, h / 2, 0, h / 2,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}
