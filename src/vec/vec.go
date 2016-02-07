package vec

import "math"

type V struct {
	X float64
	Y float64
}

var zvec = V{0, 0}

func AddV(v1, v2 V) V {
	return V{v1.X + v2.X, v1.X + v2.X}
}

func Add(v V, n float64) V {
	return V{v.X + n, v.Y + n}
}

func SubV(v1, v2 V) V {
	return V{v1.X - v2.X, v1.Y - v2.Y}
}

func Sub(v V, n float64) V {
	return V{v.X - n, v.Y - n}
}

func Mul(v V, n float64) V {
	return V{n * v.X, n * v.Y}
}

func Div(v V, n float64) V {
	return V{v.X / n, v.Y / n}
}

func Dot(v1 V, v2 V) float64 {
	return v1.X*v2.X + v1.Y*v2.Y
}

func Abs(v V) float64 {
	return math.Sqrt(Dot(v, v))
}

func Unit(v V) V {
	if v == zvec {
		return v
	}
	return Div(v, Abs(v))
}

func Scale(v V, n float64) V {
	return Mul(Unit(v), n)
}
