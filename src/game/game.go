package game

import (
	"math"
	"time"

	"github.com/mhuisi/vires/src/timed"
)

type Vires int

type Field struct {
	Players    []*Player
	PlayerID   int
	Cells      []*Cell
	Movements  chan []*Movement
	MovementID chan int
	Ops        timed.Timed
	Notifier   StateNotifier
	Size       Vec
}

type CellConflict struct {
	Target   *Cell
	Attacker *Movement
}

type StateNotifier struct {
	Collisions      chan<- Collision
	CellConflicts   chan<- CellConflict
	EliminatePlayer chan<- *Player
	Victory         chan<- *Player
}

func NewField(players []string, notifier StateNotifier) *Field {
	// mapgen algorithm here ...
	mvs := make(chan []*Movement, 1)
	mvs <- []*Movement{}
	mvid := make(chan int, 1)
	mvid <- 0
	ops := make(chan timed.Timed, 1)
	ops <- timed.New()
	f := &Field{
		Players:  []*Player{},
		PlayerID: 0,
		// change to cells from mapgen algorithm later!
		Cells:      []*Cell{},
		Movements:  mvs,
		MovementID: mvid,
		Ops:        timed.New(),
		Notifier:   notifier,
		// change to size from mapgen algorithm later!
		Size: Vec{},
	}
	return f
}

func (f *Field) Close() {
	f.Ops.Close()
}

func (f *Field) checkDominationVictory() {
	if len(f.Players) != 1 {
		return
	}
	winner := f.Players[0]
	f.Notifier.Victory <- winner
	f.Close()
}

func (f *Field) removePlayer(pp *Player) {
	ps := f.Players
	for i, p := range ps {
		if p == pp {
			last := len(ps) - 1
			ps[i] = ps[last]
			f.Players = ps[:last]
			f.Notifier.EliminatePlayer <- pp
			f.checkDominationVictory()
		}
	}
}

func (f *Field) removeCollisions(m *Movement) {
	ops := f.Ops
	ops.Lock()
	defer ops.Unlock()
	for i, e := range ops.Entries {
		collision, ok := e.Key.(Collision)
		if ok && MovementInCollision(m, collision) {
			ops.Remove(i)
		}
	}
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
	Location Vec
	Radius   float64
}

type Cell struct {
	ID       int
	Capacity int
	// [ReplicationSpeed] = vires/s
	ReplicationSpeed float64
	Stationed        Vires
	Owner            *Player
	Body             Circle
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
	// concept:
	// we treat one movement relative to the other movement
	// and then calculate the times at which the path of the smaller
	// movement intersects the circle bounds of the larger movement.
	// because we treat both movements relative to each other,
	// the center of the larger movement is at (0, 0).
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
				f.Ops.Lock()
				f.Ops.Start(c, delay, func() {
					// collisions are treated immutably,
					// there is no need to lock c
					// because it is never mutated
					f.collide(c)
				})
				f.Ops.Unlock()
			}
		}
	}
	f.Movements <- mov
}

func (f *Field) conflict(cc CellConflict) {
	atk := cc.Attacker
	tgt := cc.Target
	atkowner := atk.Owner
	tgtowner := tgt.Owner
	// same player, friendly units are merged into the cell
	if atkowner == tgtowner {
		tgt.Stationed += atk.Moving
	} else {
		cellVires := tgt.Stationed - atk.Moving
		// cell died, change owner
		if cellVires < 0 {
			tgtowner.Cells -= 1
			atkowner.Cells += 1
			// target has no cells left, eliminate owner from game
			if tgtowner.Cells == 0 {
				f.removePlayer(tgtowner)
			}
			tgt.Owner = atkowner
		}
	}
	f.updateVires(cc.Attacker, 0)
	f.Notifier.CellConflicts <- cc
}

func (f *Field) findCellConflict(attacker *Movement, target *Cell) {
	speed := Abs(attacker.Direction)
	dist := Abs(SubVec(target.Body.Location, attacker.Body.Location))
	delay := dist / speed
	cc := CellConflict{target, attacker}
	f.Ops.Lock()
	defer f.Ops.Unlock()
	f.Ops.Start(cc, time.Duration(delay)*time.Second, func() {
		f.conflict(cc)
	})
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
	f.findCellConflict(mov, target)
	movements := <-f.Movements
	movements = append(movements, mov)
	f.Movements <- movements
}

type CellAttack struct {
	Mov  *Movement
	Cell *Cell
}
