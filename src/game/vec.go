package game

import "math"

func Sq(n float64) float64 {
	return n * n
}

type Vec struct {
	X float64
	Y float64
}

var zvec = Vec{0, 0}

func AddVec(v1, v2 Vec) Vec {
	return Vec{v1.X + v2.X, v1.X + v2.X}
}

func Add(v Vec, n float64) Vec {
	return Vec{v.X + n, v.Y + n}
}

func SubVec(v1, v2 Vec) Vec {
	return Vec{v1.X - v2.X, v1.Y - v2.Y}
}

func Sub(v Vec, n float64) Vec {
	return Vec{v.Y - n, v.Y - n}
}

func Mul(v Vec, n float64) Vec {
	return Vec{n * v.X, n * v.Y}
}

func Div(v Vec, n float64) Vec {
	return Vec{v.X / n, v.Y / n}
}

func Dot(v1 Vec, v2 Vec) float64 {
	return v1.X*v2.X + v1.Y*v2.Y
}

func Abs(v Vec) float64 {
	return math.Sqrt(Dot(v, v))
}

func Unit(v Vec) Vec {
	if v == zvec {
		return v
	}
	return Div(v, Abs(v))
}

func Scale(v Vec, n float64) Vec {
	if v == zvec {
		return v
	}
	return Mul(Unit(v), n)
}
