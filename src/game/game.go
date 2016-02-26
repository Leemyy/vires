package game

import (
	"time"

	"github.com/mhuisi/vires/src/ent"
	"github.com/mhuisi/vires/src/timed"
	"github.com/mhuisi/vires/src/transm"
	"github.com/mhuisi/vires/src/vec"
)

const (
	replicationInterval = 1 * time.Second
)

// Field represents a game instance of a field.
type Field struct {
	players     map[ent.ID]ent.Player
	cells       map[ent.ID]*ent.Cell
	movements   map[ent.ID]*ent.Movement
	movementID  ent.ID
	ops         *timed.Timed
	transmitter *transm.Transmitter
	size        vec.V
}

// NewField generates a new Field for the specified players,
// using the specified transmitter to notify the caller
// about things that happen in the game.
func NewField(players []ent.ID, t *transm.Transmitter) *Field {
	// mapgen algorithm here ...

	ps := make(map[ent.ID]ent.Player, len(players))
	for _, id := range players {
		ps[id] = ent.NewPlayer(id)
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
	// handle this here instead of in the caller to avoid the caller trying to read the cells
	// while we're running our game loop
	t.GenerateField(f.cells, f.size)
	f.startReplication()
	return f
}

func (f *Field) startReplication() {
	var replicate func(time.Time)
	start := func() { f.ops.Start(time.Now().Add(replicationInterval), replicate) }
	replicate = func(time.Time) {
		for _, c := range f.cells {
			c.Replicate()
		}
		f.transmitter.Replicate(f.cells)
		start()
	}
	start()
}

// Close stops all operations on the field and blocks
// until all operations have been stopped.
func (f *Field) Close() {
	f.ops.Close()
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

func (f *Field) removePlayer(p ent.ID) {
	delete(f.players, p)
	for _, m := range f.movements {
		if m.Owner().ID() == p {
			m.ClearCollisions()
			f.removeMovement(m)
		}
	}
	for _, c := range f.cells {
		if c.OwnerID() == p {
			c.Neutralize()
		}
	}
	f.checkDominationVictory()
}

func (f *Field) removeMovement(m *ent.Movement) {
	m.Stop()
	delete(f.movements, m.ID())
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
		f.transmitter.Eliminate(*defender)
		f.removePlayer(defid)
	}
}

func (f *Field) isValidMovement(attacker, src, dst ent.ID) bool {
	srcCell, ok := f.cells[src]
	if !ok {
		return false
	}
	_, ok = f.cells[dst]
	if !ok {
		return false
	}
	if srcCell.IsNeutral() {
		return false
	}
	return srcCell.OwnerID() == attacker
}

// Move moves a movement by the specified attacker
// from the specified source cell to the
// specified target cell.
func (f *Field) Move(attacker, srcid, tgtid ent.ID) {
	f.ops.Start(time.Now(), func(time.Time) {
		if !f.isValidMovement(attacker, srcid, tgtid) {
			return
		}
		mid := f.movementID
		src := f.cells[srcid]
		tgt := f.cells[tgtid]
		mov := src.Move(mid, tgt)
		f.transmitter.Move(mov)
		at := mov.ConflictAt()
		mov.Stop = f.ops.Start(at, func(time.Time) {
			f.conflict(mov)
		})
		f.movements[mid] = mov
		f.movementID++
		f.findCollisions(mov)
	})
}

// DisconnectPlayer removes the player from
// the field, stops all his actions and
// neutralizes all his cells.
func (f *Field) DisconnectPlayer(id ent.ID) {
	f.ops.Start(time.Now(), func(time.Time) {
		f.removePlayer(id)
	})
}
