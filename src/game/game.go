package game

import (
	"time"

	"github.com/mhuisi/vires/src/game/ent"
	"github.com/mhuisi/vires/src/timed"
	"github.com/mhuisi/vires/src/transm"
	"github.com/mhuisi/vires/src/vec"
)

const (
	replicationInterval = 1 * time.Second
)

type Field struct {
	players         map[ent.ID]ent.Player
	cells           map[ent.ID]*ent.Cell
	movements       map[ent.ID]*ent.Movement
	movementID      ent.ID
	ops             *timed.Timed
	stopReplication func() bool
	transmitter     transm.Transmitter
	size            vec.V
}

func NewField(players []string, t transm.Transmitter) *Field {
	// mapgen algorithm here ...

	ps := make(map[ent.ID]ent.Player, len(players))
	for i, p := range players {
		id := ent.ID(i)
		ps[id] = ent.NewPlayer(id, p)
	}
	f := &Field{
		players: ps,
		// change to cells from mapgen algorithm later!
		cells:       map[ent.ID]*ent.Cell{},
		movements:   map[ent.ID]*ent.Movement{},
		movementID:  0,
		ops:         timed.New(),
		transmitter: t,
		// change to size from mapgen algorithm later!
		size: vec.V{},
	}
	f.startReplication()
	return f
}

func (f *Field) startReplication() {
	var replicate func(time.Time)
	start := func() { f.stopReplication = f.ops.Start(time.Now().Add(replicationInterval), replicate) }
	replicate = func(time.Time) {
		for _, c := range f.cells {
			c.Replicate()
		}
		start()
	}
	start()
}

func (f *Field) Close() {
	f.ops.Close()
	f.stopReplication()
}

func (f *Field) checkDominationVictory() {
	if len(f.players) > 1 {
		return
	}
	var winner ent.Player
	// get first winner
	for _, winner = range f.players {
		break
	}
	f.transmitter.Win(winner)
}

func (f *Field) removeMovement(m *ent.Movement) {
	m.Stop()
	id := m.ID()
	delete(f.movements, id)
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
	f.transmitter.Collide(m, m2)
	f.viresChanged(m)
	f.viresChanged(m2)
}

func (f *Field) findCollisions(m *ent.Movement) {
	for _, m2 := range f.movements {
		collideAt, collides := m.CollidesWith(m2)
		if !collides {
			continue
		}
		stopCollision := f.ops.Start(collideAt, func(time.Time) {
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
	f.transmitter.Conflict(mv, target)
	if defender.IsDead() {
		defid := defender.ID()
		delete(f.players, defid)
		f.transmitter.Eliminate(defender)
		f.checkDominationVictory()
	}
}

func (f *Field) Move(src, tgt *ent.Cell) {
	f.ops.Start(time.Now(), func(time.Time) {
		mid := f.movementID
		mov := src.Move(mid, tgt)
		at := ent.ConflictAt(mov)
		mov.Stop = f.ops.Start(at, func(time.Time) {
			f.conflict(mov)
		})
		f.movements[mid] = mov
		f.movementID++
		f.findCollisions(mov)
	})
}
