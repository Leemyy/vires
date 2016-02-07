package game

import (
	"time"

	"github.com/mhuisi/vires/src/game/ent"
	"github.com/mhuisi/vires/src/timed"
	"github.com/mhuisi/vires/src/transm"
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
	Ops             *timed.Timed
	StopReplication func() bool
	Transmitter     transm.Transmitter
	Size            vec.V
}

func NewField(players []string, t transm.Transmitter) *Field {
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
		Ops:         timed.New(),
		Transmitter: t,
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
	f.Transmitter.Win(winner)
}

func (f *Field) removeMovement(m *ent.Movement) {
	m.Stop()
	id := m.ID()
	delete(f.Movements, id)
}

func (f *Field) viresChanged(m *ent.Movement) {
	m.ClearCollisions()
	if m.IsDead() {
		f.removeMovement(m)
	} else {
		f.findCollisions(m)
	}
}

func (f *Field) collide(m, m2 *ent.Movement) {
	m.Collide(m2)
	f.Transmitter.Collide(m, m2)
	f.viresChanged(m)
	f.viresChanged(m2)
}

func (f *Field) findCollisions(m *ent.Movement) {
	for _, m2 := range f.Movements {
		collideAt, collides := m.CollidesWith(m2)
		if !collides {
			continue
		}
		stopCollision := f.Ops.Start(collideAt, func(time.Time) {
			f.collide(m, m2)
		})
		m.AddCollision(m2, stopCollision)
		m2.AddCollision(m, stopCollision)
	}
}

func (f *Field) conflict(mv *ent.Movement) {
	target := mv.Target()
	defender := target.Owner()
	mv.Conflict()
	f.removeMovement(mv)
	f.Transmitter.Conflict(mv, target)
	if defender.IsDead() {
		defid := defender.ID()
		delete(f.Players, defid)
		f.Transmitter.Eliminate(defender)
		f.checkDominationVictory()
	}
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
