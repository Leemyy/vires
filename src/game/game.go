package game

import (
	"math"
	"time"
)

type Vires int

type Field struct {
	Players          []*Player
	Cells            []*Cell
	Movements        chan []*Movement
	MovementID       chan int
	Collisions       chan Timed
	CollisionUpdates chan<- Collision
	Size             Vec
}

func NewField(players []string, collisionUpdates chan<- Collision) *Field {
	// mapgen algorithm here ...
	mvs := make(chan []*Movement, 1)
	mvs <- []*Movement{}
	mvid := make(chan int, 1)
	mvid <- 0
	cols := make(chan Timed, 1)
	cols <- NewTimed()
	f := &Field{
		Players: []*Player{},
		// change to cells from mapgen algorithm later!
		Cells:            nil,
		Movements:        mvs,
		MovementID:       mvid,
		Collisions:       cols,
		CollisionUpdates: collisionUpdates,
		// change to size from mapgen algorithm later!
		Size: Vec{},
	}
	return f
}

func (f *Field) removeCollisions(m *Movement) {
	cols := <-f.Collisions
	for c := range cols {
		collision := c.(Collision)
		if MovementInCollision(m, collision) {
			cols.Remove(c)
		}
	}
	f.Collisions <- cols
}

func (f *Field) removeMovement(m *Movement) {
	movements := <-f.Movements
	for i, mv := range movements {
		if mv == m {
			// delete the movement
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

func (f *Field) collide(c Collision) {
	a := c.A
	b := c.B
	na := a.Moving
	nb := b.Moving
	// merge movements if two movements with the same owner and the same target collide
	if a.Owner == b.Owner {
		if a.Target == b.Target {
			// collision with friendly movement
			f.removeMovement(b)
			// mutates the movements
			f.updateVires(a, na+nb)
			f.CollisionUpdates <- c
		}
		// no collision, friendly movements cross each other
		return
	}
	// standard collision
	// mutates the movements
	f.updateVires(a, na-nb)
	f.updateVires(b, nb-na)
	f.CollisionUpdates <- c
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
	// Time the movement was started at
	Start time.Time
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
	A *Movement
	B *Movement
}

func MovementInCollision(m *Movement, c Collision) bool {
	return c.A == m || c.B == m
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
	for _, mv := range mov {
		// movements where the owner is the same
		// but the target isn't don't collide;
		// the movements just pass each other
		if !(m.Owner == mv.Owner && m.Target != mv.Target) {
			dt, collides := CollisionTime(mv, m)
			if collides {
				// t_d = t_0 + dt - t_n
				// t_d = time to delay by
				// t_0 = time the movement started at
				// dt = time of the collision - t_0
				// t_n = current time
				delay := m.Start.Add(time.Duration(int64(dt * float64(time.Second)))).Sub(time.Now())
				c := Collision{mv, m}
				cols := <-f.Collisions
				cols.Start(c, delay, func() {
					// collisions are treated immutably,
					// there is no need to lock c
					// because it is never mutated
					f.collide(c)
				})
				f.Collisions <- cols
			}
		}
	}
	f.Movements <- mov
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
		Start:     time.Now(),
	}
	id++
	f.MovementID <- id
	f.findCollisions(mov)
	movements := <-f.Movements
	movements = append(movements, mov)
	f.Movements <- movements
}
