package vec

import "math"

// V represents a 2D vector.
type V struct {
	X float64
	Y float64
}

var zvec = V{0, 0}

// AddV adds two vectors.
func AddV(v1, v2 V) V {
	return V{v1.X + v2.X, v1.Y + v2.Y}
}

// Add adds a scalar to a vector.
func Add(v V, n float64) V {
	return V{v.X + n, v.Y + n}
}

// SubV subtracs the second vector
// from the first vector.
func SubV(v1, v2 V) V {
	return V{v1.X - v2.X, v1.Y - v2.Y}
}

// Sub subtracts a scalar from a vector.
func Sub(v V, n float64) V {
	return V{v.X - n, v.Y - n}
}

// Mul multiplies a vector with a scalar.
func Mul(v V, n float64) V {
	return V{n * v.X, n * v.Y}
}

// Div divides a vector by a scalar.
func Div(v V, n float64) V {
	return V{v.X / n, v.Y / n}
}

// Dot calculates the dot-product
// of two vectors.
func Dot(v1 V, v2 V) float64 {
	return v1.X*v2.X + v1.Y*v2.Y
}

// Abs calculates the length
// of a vector.
func Abs(v V) float64 {
	return math.Sqrt(Dot(v, v))
}

// Unit calculates a
// vector with a length
// of 1 in the direction
// of the specified vector.
func Unit(v V) V {
	if v == zvec {
		return v
	}
	return Div(v, Abs(v))
}

// Scale scales a vector
// to the new length.
func Scale(v V, n float64) V {
	return Mul(Unit(v), n)
}

// Dist calculates the distance
// between two vectors.
func Dist(v1 V, v2 V) float64 {
	return Abs(SubV(v1, v2))
}
