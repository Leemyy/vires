package game

import (
	"image"
	"math"
	"time"
)

type Vires int

type Field struct {
	Cells []*Cell
	Size  image.Point
}

type Circle struct {
	Location image.Point
	Radius   float64
}

func InCircle(p image.Point, c Circle) bool {
	vecPC := p.Sub(p, c.Location)
	// |vecPC|
	dist := math.Sqrt(math.Pow(vecPC.X, 2) + math.Pow(vecPC.Y, 2))
	return dist < c.Radius
}

func Colliding(c1, c2 Circle) bool {
	if c1.Radius > c2.Radius {
		return InCircle(c2.Location, c1)
	}
	return InCircle(c1.Location, c2)
}

type CircleTrajectory struct {
	Start  image.Point
	End    image.Point
	Radius float64
}

type Cell struct {
	Capacity            chan int
	ReplicationInterval time.Duration
	Stationed           Vires
	Owner               *Player
	Shape               Circle
}

func (c *Cell) Replicate() {
	cp := <-c.Capacity
	c.Capacity <- cp + 1
}

func (c *Cell) ReplicateContinuously(quit chan struct{}) {
	t := time.NewTicker(c.ReplicationInterval)
	go func() {
		select {
		case <-quit:
			t.Stop()
			return
		default:
			c.Replicate()
			<-t.C
		}
	}()
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
