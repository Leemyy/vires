package transm

import "github.com/mhuisi/vires/src/game/ent"

type CollisionMovement struct {
	ID     ent.ID
	Moving ent.Vires
}

func makeCollMov(m *ent.Movement) CollisionMovement {
	return CollisionMovement{m.ID(), m.Moving()}
}

type Collision struct {
	A, B CollisionMovement
}

type ConflictCell struct {
	ID        ent.ID
	Stationed ent.Vires
	Owner     ent.ID
}

func makeConflCell(c *ent.Cell) ConflictCell {
	return ConflictCell{c.ID(), c.Stationed(), c.Owner().ID()}
}

type Conflict struct {
	Movement ent.ID
	Cell     ConflictCell
}

type EliminatedPlayer ent.ID

type Winner ent.ID

type Transmitter struct {
	collisions        chan *Collision
	conflicts         chan *Conflict
	eliminatedPlayers chan *EliminatedPlayer
	winner            chan *Winner
}

func (t *Transmitter) Open() {
	t.collisions = make(chan *Collision, 1024)
	t.conflicts = make(chan *Conflict, 512)
	t.eliminatedPlayers = make(chan *EliminatedPlayer, 16)
	t.winner = make(chan *Winner)

}

func (t *Transmitter) Disable() {
	t.collisions = nil
	t.conflicts = nil
	t.eliminatedPlayers = nil
	t.winner = nil
}

func (t *Transmitter) Collide(a, b *ent.Movement) {
	t.collisions <- &Collision{makeCollMov(a), makeCollMov(b)}
}

func (t *Transmitter) Conflict(m *ent.Movement, c *ent.Cell) {
	t.conflicts <- &Conflict{m.ID(), makeConflCell(c)}
}

func (t *Transmitter) Eliminate(p ent.Player) {
	e := EliminatedPlayer(p.ID())
	t.eliminatedPlayers <- &e
}

func (t *Transmitter) Win(p ent.Player) {
	w := Winner(p.ID())
	t.winner <- &w
}

func (t *Transmitter) Collisions() <-chan *Collision {
	return t.collisions
}

func (t *Transmitter) Conflicts() <-chan *Conflict {
	return t.conflicts
}

func (t *Transmitter) EliminatedPlayers() <-chan *EliminatedPlayer {
	return t.eliminatedPlayers
}

func (t *Transmitter) Winner() <-chan *Winner {
	return t.winner
}

type Movement struct {
	Source ent.ID
	Dest   ent.ID
}
