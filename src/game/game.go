package game

import "image"

type Vires int

type Match struct {
	Field   *Field
	Players []*Players
}

type Field struct {
	Cells []*Cell
	Size  image.Point
}

type Circle struct {
	Location image.Point
	Radius   float64
}

type CircleTrajectory struct {
	Start  image.Point
	End    image.Point
	Radius float64
}

type Cell struct {
	Capacity chan int
	// [ReplicationSpeed] = vires/s
	ReplicationSpeed float64
	Stationed        Vires
	Owner            *Player
	Shape            Circle
}

type Player struct {
	Name  string
	Cells []*Cell
}

type Movement struct {
	Moving Vires
	Target *Cell
	Shape  Circle
	// [Speed] = dots/s
	Speed float64
}
