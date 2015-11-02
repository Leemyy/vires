package game

import (
	"image"
	"time"
)

type Vires int

type Field struct {
	Cells []*Cell
	Size  image.Point
}

type Cell struct {
	Capacity            int
	ReplicationInterval time.Duration
	Stationed           Vires
	Owner               *Player
	Location            image.Point
}

type Player struct {
	Name  string
	Cells []*Cell
}

type Movement struct {
	Moving   Vires
	Target   Cell
	Location image.Point
}
