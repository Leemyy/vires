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
	MovementID       chan int
	Collisions       chan CollisionHeap
	ReplaceFirstColl chan struct{}
	Size             Vec
	Quit             chan struct{}
	CollisionUpdates chan *Collision
}

func (f *Field) removeCollisions(m *Movement) {
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

	f.removeCollisions(m)
}

func (f *Field) updateVires(m *Movement, n Vires) {
	m.UpdateVires(n)
	if n > 0 {
		// amount of vires affects collisions, update collisions
		f.removeCollisions(m)
		f.findCollisions(m)
		return
	}
	// movement died
	f.removeMovement(m)
}

func (f *Field) collide(c *Collision) {
	a := c.A
	b := c.B
	na := a.Moving
	nb := b.Moving
	// merge movements if two movements with the same owner and the same target collide
	if a.Owner == b.Owner {
		if a.Target == b.Target {
			// collision with friendly movement
			f.removeMovement(b)
			// mutates the movements, CollisionUpdates receives updated values
			f.updateVires(a, na+nb)
			f.CollisionUpdates <- c
		}
		// no collision, friendly movements cross each other
		return
	}
	// standard collision
	// mutates the movements, CollisionUpdates receives updated values
	f.updateVires(a, na-nb)
	f.updateVires(b, nb-na)
	f.CollisionUpdates <- c
}

func (f *Field) runCollisions() chan<- struct{} {
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
	ID     int
	Owner  *Player
	Moving Vires
	Target *Cell
	Body   Circle
	// |Direction| = v, [v] = points/s
	Direction Vec
}

func Radius(n Vires) float64 {
	// placeholder, needs testing
	return float64(n)
}

func Speed(n Vires) float64 {
	// placeholder, needs testing
	if n == 0 {
		return 0
	}
	return 100 / float64(n)
}

func (m *Movement) UpdateVires(n Vires) {
	m.Moving = n
	m.Direction = Scale(m.Direction, Speed(n))
	m.Body.Radius = Radius(n)
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
	b1 := m1.Body
	b2 := m2.Body
	p := SubVec(b1.Location, b2.Location)
	v := SubVec(m1.Direction, m2.Direction)
	r := math.Max(b1.Radius, b2.Radius)
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

func (f *Field) findCollisions(m *Movement) {
	mov := <-f.Movements
	cols := <-f.Collisions
	first := cols[0]
	replaceFirst := false
	for _, mv := range mov {
		if !(m.Owner == mv.Owner && m.Target == mv.Target) {
			dt, collides := CollisionTime(mv, m)
			if collides {
				t := time.Now().Add(time.Duration(int64(dt * float64(time.Second))))
				heap.Push(&cols, Collision{mv, m, t})
				if t.Before(first.Time) {
					replaceFirst = true
				}
			}
		}
	}
	if replaceFirst {
		// notify runCollisions that it should stop the current timer
		f.ReplaceFirstColl <- struct{}{}
	}
	f.Movements <- mov
	f.Collisions <- cols
}

func (f *Field) Move(owner *Player, n Vires, start Vec, target *Cell) {
	id := <-f.MovementID
	mov := &Movement{
		ID:        id,
		Owner:     owner,
		Moving:    n,
		Target:    target,
		Body:      Circle{start, Radius(n)},
		Direction: Scale(SubVec(target.Body.Location, start), Speed(n)),
	}
	id++
	f.MovementID <- id
	f.findCollisions(mov)
	movements := <-f.Movements
	movements = append(movements, mov)
	f.Movements <- movements
}
