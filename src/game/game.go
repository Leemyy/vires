package game

import (
	"math"
	"time"

	"github.com/mhuisi/vires/src/timed"
	"github.com/mhuisi/vires/src/vec"
)

const (
	ReplicationInterval = 1 * time.Second
)

type Vires int

type Field struct {
	Players         map[*Player]struct{}
	Cells           map[*Cell]struct{}
	Movements       map[*Movement]func() bool
	MovementID      int
	Collisions      map[Collision]func() bool
	Ops             *timed.Timed
	StopReplication func() bool
	Notifier        StateNotifier
	Size            vec.V
}

type StateNotifier struct {
	Collisions      chan<- Collision
	Conflicts       chan<- *Movement
	EliminatePlayer chan<- *Player
	Victory         chan<- *Player
}

func NewField(players []string, notifier StateNotifier) *Field {
	// mapgen algorithm here ...

	ps := make(map[*Player]struct{}, len(players))
	for i, p := range players {
		ps[&Player{
			ID:    i,
			Name:  p,
			Cells: 1,
		}] = struct{}{}
	}
	f := &Field{
		Players: ps,
		// change to cells from mapgen algorithm later!
		Cells:      map[*Cell]struct{}{},
		Movements:  map[*Movement]func() bool{},
		MovementID: 0,
		Collisions: map[Collision]func() bool{},
		Ops:        timed.New(),
		Notifier:   notifier,
		// change to size from mapgen algorithm later!
		Size: vec.V{},
	}
	f.startReplication()
	return f
}

func (f *Field) Close() {
	f.Ops.Close()
	f.StopReplication()
}

func (f *Field) startReplication() {
	var replicate func(time.Time)
	replicate = func(time.Time) {
		for c := range f.Cells {
			c.Merge(c.Replication)
		}
		f.StopReplication = f.Ops.Start(time.Now().Add(ReplicationInterval), replicate)
	}
	f.StopReplication = f.Ops.Start(time.Now().Add(ReplicationInterval), replicate)
}

func (f *Field) checkDominationVictory() {
	if len(f.Players) > 1 {
		return
	}
	var winner *Player
	for winner = range f.Players {
		break
	}
	f.Notifier.Victory <- winner
	f.Close()
}

func (f *Field) removeCollisions(m *Movement) {
	for c, stop := range f.Collisions {
		if c.A == m || c.B == m {
			delete(f.Collisions, c)
			stop()
		}
	}
}

func (f *Field) removeMovement(m *Movement) {
	stop := f.Movements[m]
	delete(f.Movements, m)
	stop()
}

func (f *Field) updateCollisions(m *Movement) {
	f.removeCollisions(m)
	f.findCollisions(m)
}

func (f *Field) updateVires(m *Movement, n Vires) {
	m.UpdateVires(n)
	if n > 0 {
		// amount of vires affects collisions, update collisions
		f.updateCollisions(m)
		return
	}
	// movement died
	f.removeMovement(m)
	f.removeCollisions(m)
}

func (f *Field) mergeMovements(a, b *Movement) {
	f.updateVires(b, 0)
	f.updateVires(a, a.Moving+b.Moving)
}

func (f *Field) collide(c Collision) {
	a := c.A
	b := c.B
	// merge movements if two movements with the same owner and the same target collide
	if a.Owner == b.Owner {
		if a.Target == b.Target {
			// collision with friendly movement
			f.mergeMovements(a, b)
			f.Notifier.Collisions <- c
		}
		// no collision, friendly movements cross each other
		return
	}
	// standard collision
	// mutates the movements
	na := a.Moving
	nb := b.Moving
	f.updateVires(a, na-nb)
	f.updateVires(b, nb-na)
	f.Notifier.Collisions <- c
}

type Circle struct {
	Location vec.V
	Radius   float64
}

type Cell struct {
	ID       int
	Capacity Vires
	// [Replication] = vires/cycle
	Replication Vires
	Stationed   Vires
	Owner       *Player
	Body        Circle
}

func Capacity(force float64) Vires {
	// placeholder, needs testing
	return Vires(10 * force)
}

func Replication(force float64) Vires {
	// placeholder, needs testing
	return Vires(force)
}

func CellRadius(force float64) float64 {
	// placeholder, needs testing
	return force
}

func NewCell(id int, force float64, owner *Player, loc vec.V) *Cell {
	return &Cell{
		ID:          id,
		Capacity:    Capacity(force),
		Replication: Replication(force),
		Stationed:   0,
		Owner:       owner,
		Body:        Circle{loc, CellRadius(force)},
	}
}

func (c *Cell) Merge(n Vires) {
	newStationed := c.Stationed + n
	switch {
	case newStationed < 0:
		newStationed = 0
	case newStationed > c.Capacity:
		newStationed = c.Capacity
	}
	c.Stationed = newStationed
}

type Player struct {
	ID    int
	Name  string
	Cells int
}

type Movement struct {
	ID     int
	Owner  *Player
	Moving Vires
	Target *Cell
	Body   Circle
	// |Direction| = v, [v] = points/s
	Direction vec.V
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
	m.Direction = vec.Scale(m.Direction, Speed(n))
	m.Body.Radius = Radius(n)
}

type Collision struct {
	A *Movement
	B *Movement
}

func sq(v float64) float64 {
	return v * v
}

func CollisionTime(m1 *Movement, m2 *Movement) (float64, bool) {
	// concept:
	// we treat one movement relative to the other movement
	// and then calculate the times at which the path of the smaller
	// movement intersects the circle bounds of the larger movement.
	// because we treat both movements relative to each other,
	// the center of the larger movement is at (0, 0).
	b1 := m1.Body
	b2 := m2.Body
	p := vec.SubV(b1.Location, b2.Location)
	v := vec.SubV(m1.Direction, m2.Direction)
	r := math.Max(b1.Radius, b2.Radius)
	d := vec.Unit(v)
	tempP := vec.Dot(p, d)
	tempR := sq(tempP) - sq(p.X) - sq(p.Y) + sq(r)
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
		t := math.Min(t1, t2) / vec.Abs(v)
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
	for mv := range f.Movements {
		// movements where the owner is the same
		// but the target isn't don't collide;
		// the movements just pass each other
		if !(m.Owner == mv.Owner && m.Target != mv.Target) {
			dt, collides := CollisionTime(mv, m)
			if collides {
				collideAt := m.Start.Add(time.Duration(dt * float64(time.Second)))
				c := Collision{mv, m}
				f.Collisions[c] = f.Ops.Start(collideAt, func(time.Time) {
					f.collide(c)
				})
			}
		}
	}
}

func (f *Field) conflict(mv *Movement) {
	tgt := mv.Target
	atkowner := mv.Owner
	tgtowner := tgt.Owner
	// same player, friendly units are merged into the cell
	if atkowner == tgtowner {
		tgt.Merge(mv.Moving)
	} else {
		cellVires := tgt.Stationed - mv.Moving
		// cell died, change owner
		if cellVires < 0 {
			tgtowner.Cells -= 1
			atkowner.Cells += 1
			// target has no cells left, eliminate owner from game
			if tgtowner.Cells == 0 {
				delete(f.Players, tgtowner)
				f.checkDominationVictory()
			}
			tgt.Owner = atkowner
			tgt.Merge(-cellVires)
		}
	}
	f.updateVires(mv, 0)
	f.Notifier.Conflicts <- mv
}

func (f *Field) addCellConflict(attacker *Movement) {
	target := attacker.Target
	speed := vec.Abs(attacker.Direction)
	dist := vec.Abs(vec.SubV(target.Body.Location, attacker.Body.Location))
	delay := dist / speed
	conflictAt := time.Now().Add(time.Duration(delay * float64(time.Second)))
	f.Movements[attacker] = f.Ops.Start(conflictAt, func(time.Time) {
		f.conflict(attacker)
	})
}

func (f *Field) Move(owner *Player, n Vires, start vec.V, target *Cell) {
	f.Ops.Start(time.Now(), func(time.Time) {
		mov := &Movement{
			ID:        f.MovementID,
			Owner:     owner,
			Moving:    n,
			Target:    target,
			Body:      Circle{start, Radius(n)},
			Direction: vec.Scale(vec.SubV(target.Body.Location, start), Speed(n)),
			Start:     time.Now(),
		}
		f.MovementID++
		f.findCollisions(mov)
		f.addCellConflict(mov)
	})

}
