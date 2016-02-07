package game

import (
	"time"

	"github.com/mhuisi/vires/src/game/ent"
	"github.com/mhuisi/vires/src/timed"
	"github.com/mhuisi/vires/src/vec"
)

const (
	ReplicationInterval = 1 * time.Second
)

type Field struct {
	Players         map[ent.ID]ent.Player
	Cells           map[ent.ID]*ent.Cell
	Movements       map[ent.ID]*ent.Movement
	MovementID      ent.ID
	Collisions      map[ent.ID]ent.Collision
	CollisionID     ent.ID
	Ops             *timed.Timed
	StopReplication func() bool
	Notifier        StateNotifier
	Size            vec.V
}

type StateNotifier struct {
	Collisions      chan<- ent.Collision
	Conflicts       chan<- *ent.Movement
	EliminatePlayer chan<- ent.Player
	Victory         chan<- ent.Player
}

func NewField(players []string, notifier StateNotifier) *Field {
	// mapgen algorithm here ...

	ps := make(map[ent.ID]ent.Player, len(players))
	for i, p := range players {
		id := ent.ID(i)
		ps[id] = ent.NewPlayer(id, p)
	}
	f := &Field{
		Players: ps,
		// change to cells from mapgen algorithm later!
		Cells:       map[ent.ID]*ent.Cell{},
		Movements:   map[ent.ID]*ent.Movement{},
		MovementID:  0,
		Collisions:  map[ent.ID]ent.Collision{},
		CollisionID: 0,
		Ops:         timed.New(),
		Notifier:    notifier,
		// change to size from mapgen algorithm later!
		Size: vec.V{},
	}
	f.startReplication()
	return f
}

func (f *Field) startReplication() {
	var replicate func(time.Time)
	start := func() { f.StopReplication = f.Ops.Start(time.Now().Add(ReplicationInterval), replicate) }
	replicate = func(time.Time) {
		for _, c := range f.Cells {
			c.Replicate()
		}
		start()
	}
	start()
}

func (f *Field) Close() {
	f.Ops.Close()
	f.StopReplication()
}

func (f *Field) checkDominationVictory() {
	if len(f.Players) > 1 {
		return
	}
	var winner ent.Player
	// get first winner
	for _, winner = range f.Players {
		break
	}
	f.Notifier.Victory <- winner
	f.Close()
}

func (f *Field) removeCollisions(m *ent.Movement) {
	for id, c := range f.Collisions {
		if c.A == m || c.B == m {
			delete(f.Collisions, id)
			c.Stop()
		}
	}
}

func (f *Field) removeMovement(m *ent.Movement) {
	id := m.ID()
	delete(f.Movements, id)
	m.Stop()
}

func (f *Field) viresChanged(m *ent.Movement) {
	f.removeCollisions(m)
	if m.IsDead() {
		f.removeMovement(m)
	} else {
		f.findCollisions(m)
	}
}

func (f *Field) collide(c ent.Collision) {
	c.A.Collide(c.B)
	f.viresChanged(c.A)
	f.viresChanged(c.B)
}

func (f *Field) findCollisions(m *ent.Movement) {
	for _, mv := range f.Movements {
		collideAt, collides := m.CollidesWith(mv)
		if !collides {
			continue
		}
		var c ent.Collision
		stopCollision := f.Ops.Start(collideAt, func(time.Time) {
			f.collide(c)
		})
		c = ent.Collision{
			ID:   f.CollisionID,
			A:    mv,
			B:    m,
			Stop: stopCollision,
		}
		f.Collisions[c.ID] = c
		f.CollisionID++
	}
}

func (f *Field) conflict(mv *ent.Movement) {
	defender := mv.Target().Owner()
	mv.Conflict()
	if defender.IsDead() {
		delete(f.Players, defender.ID())
		f.checkDominationVictory()
	}
	f.removeMovement(mv)
	f.Notifier.Conflicts <- mv
}

func (f *Field) Move(src, tgt *ent.Cell) {
	f.Ops.Start(time.Now(), func(time.Time) {
		mid := f.MovementID
		mov := src.Move(mid, tgt)
		at := ent.ConflictAt(mov)
		mov.Stop = f.Ops.Start(at, func(time.Time) {
			f.conflict(mov)
		})
		f.Movements[mid] = mov
		f.MovementID++
		f.findCollisions(mov)
	})

}
