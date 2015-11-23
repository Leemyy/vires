package game

import (
	"container/heap"
	"math"
	"time"
)

type Vires int

type Match struct {
	Field   *Field
	Players []*Player
}

type Field struct {
	Cells            []*Cell
	Movements        chan []*Movement
	Collisions       chan CollisionHeap
	ReplaceFirstColl chan struct{}
	Size             Vec
	Quit             chan struct{}
}

func winner(a *Movement, b *Movement) (*Movement, *Movement) {
	switch {
	case a.Moving > b.Moving:
		return a, b
	case a.Moving < b.Moving:
		return b, a
	}
	return a, b
}

func (f *Field) removeMovement(m *Movement) {
	// remove movement
	movements := <-f.Movements
	for i, mv := range movements {
		if mv == m {
			movements = append(movements[:i], movements[i+1:]...)
			break
		}
	}
	f.Movements <- movements

	// remove collisions
	cols := <-f.Collisions
	toRemove := []int{}
	for i, c := range cols {
		if c.A == m || c.B == m {
			toRemove = append(toRemove, i)
		}
	}
	for _, r := range toRemove {
		heap.Remove(&cols, r)
	}
	f.Collisions <- cols
}

func (f *Field) collide(c *Collision) {
	w, l := winner(c.A, c.B)
	// safe, only the goroutine from runCollisions() is accessing these
	w.Moving -= l.Moving
	f.removeMovement(l)
	if w.Moving == 0 {
		f.removeMovement(w)
	}
}

func (f *Field) runCollisions() chan struct{} {
	updateFirst := make(chan struct{}, 1)
	go func() {
		for {
			cols := <-f.Collisions
			first := cols[0]
			f.Collisions <- cols
			t := time.NewTimer(first.Time.Sub(time.Now()))
			select {
			case <-f.Quit:
				return
			case <-updateFirst:
				t.Stop()
			case <-t.C:
				f.collide(first)
			}
		}
	}()
	return updateFirst

}

type Circle struct {
	Location Vec
	Radius   float64
}

type Cell struct {
	Capacity int
	// [ReplicationSpeed] = vires/s
	ReplicationSpeed float64
	Stationed        Vires
	Owner            *Player
	Body             Circle
}

type Player struct {
	Name  string
	Cells []*Cell
}

type Movement struct {
	Moving Vires
	Target *Cell
	Body   Circle
	// |Direction| = v, [v] = points/s
	Direction Vec
}

type Collision struct {
	A    *Movement
	B    *Movement
	Time time.Time
}

type CollisionHeap []*Collision

func (h CollisionHeap) Len() int           { return len(h) }
func (h CollisionHeap) Less(i, j int) bool { return h[i].Time.Before(h[j].Time) }
func (h CollisionHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *CollisionHeap) Push(x interface{}) {
	*h = append(*h, x.(*Collision))
}

func (h *CollisionHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func CollisionTime(m1 *Movement, m2 *Movement) (float64, bool) {
	p := SubVec(m1.Body.Location, m2.Body.Location)
	v := SubVec(m1.Direction, m2.Direction)
	r := math.Max(m1.Body.Radius, m2.Body.Radius)
	d := Unit(v)
	tempP := Dot(p, d)
	tempR := Sq(tempP) - Sq(p.X) - Sq(p.Y) + Sq(r)
	if tempR < 0 {
		// no collision
		return math.NaN(), false
	}
	tr := math.Sqrt(tempR)
	t1 := tr - tempP
	t2 := -tr - tempP
	switch {
	case t1 > 0 && t2 > 0:
		// collision is at t
		t := math.Min(t1, t2) / Abs(v)
		return t, true
	case t1 <= 0 && t2 <= 0:
		// no collision
		return math.NaN(), false
	default:
		// collision is now
		return 0, true
	}
}

func Radius(n Vires) float64 {
	// placeholder, needs testing
	return float64(n)
}

func Speed(n Vires) float64 {
	// placeholder, needs testing
	return 100 / float64(n)
}

func (f *Field) Move(n Vires, start Vec, target *Cell) {
	mov := &Movement{
		Moving:    n,
		Target:    target,
		Body:      Circle{start, Radius(n)},
		Direction: Scale(SubVec(target.Body.Location, start), Speed(n)),
	}
	movements := <-f.Movements
	cols := <-f.Collisions
	first := cols[0]
	replaceFirst := false
	for _, m := range movements {
		dt, collides := CollisionTime(mov, m)
		if collides {
			t := time.Now().Add(time.Duration(int64(dt * float64(time.Second))))
			heap.Push(&cols, Collision{mov, m, t})
			if t.Before(first.Time) {
				replaceFirst = true
			}
		}
	}
	if replaceFirst {
		// notify runCollisions that it should stop the current timer
		f.ReplaceFirstColl <- struct{}{}
	}
	movements = append(movements, mov)
	f.Movements <- movements
	f.Collisions <- cols
}
