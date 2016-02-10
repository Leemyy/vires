package transm

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/mhuisi/vires/src/ent"
	"github.com/mhuisi/vires/src/vec"
)

type CollisionMovement struct {
	ID        ent.ID
	Moving    ent.Vires
	Body      ent.Circle
	Direction vec.V
}

func makeCollMov(m *ent.Movement) CollisionMovement {
	return CollisionMovement{m.ID(), m.Moving(), m.Body(), m.Direction()}
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

type CellVires struct {
	ID        ent.ID
	Stationed ent.Vires
}

func makeCellVires(c *ent.Cell) CellVires {
	return CellVires{c.ID(), c.Stationed()}
}

type Replication []CellVires

type GeneratedCell struct {
	ID   ent.ID
	Body ent.Circle
}

func makeGenCell(c *ent.Cell) GeneratedCell {
	return GeneratedCell{c.ID(), c.Body()}
}

type StartCell struct {
	Owner ent.ID
	Cell  ent.ID
}

func makeStartCell(c *ent.Cell) StartCell {
	return StartCell{c.ID(), c.Owner().ID()}
}

type Field struct {
	Cells      []GeneratedCell
	StartCells []StartCell
}

type Transmitter struct {
	collisions        chan *Collision
	conflicts         chan *Conflict
	eliminatedPlayers chan *EliminatedPlayer
	replications      chan Replication
	winner            chan *Winner
	field             chan *Field
}

func (t *Transmitter) Open() {
	t.collisions = make(chan *Collision, 1024)
	t.conflicts = make(chan *Conflict, 512)
	t.eliminatedPlayers = make(chan *EliminatedPlayer, 16)
	t.replications = make(chan Replication, 512)
	t.winner = make(chan *Winner)
	t.field = make(chan *Field)

}

func (t *Transmitter) Disable() {
	t.collisions = nil
	t.conflicts = nil
	t.eliminatedPlayers = nil
	t.replications = nil
	t.winner = nil
	t.field = nil
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

func (t *Transmitter) Replicate(field map[ent.ID]*ent.Cell) {
	cvs := make([]CellVires, len(field))
	i := 0
	for _, c := range field {
		cvs[i] = makeCellVires(c)
		i++
	}
	t.replications <- cvs
}

func (t *Transmitter) GenerateField(field map[ent.ID]*ent.Cell) {
	cells := make([]GeneratedCell, len(field))
	startCells := []StartCell{}
	i := 0
	for _, c := range field {
		cells[i] = makeGenCell(c)
		if c.Owner() != nil {
			startCells = append(startCells, makeStartCell(c))
		}
		i++
	}
	t.field <- &Field{cells, startCells}
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

func (t *Transmitter) Replications() <-chan Replication {
	return t.replications
}

func (t *Transmitter) GeneratedField() <-chan *Field {
	return t.field
}

type UserJoined struct {
	ID ent.ID
}

type ReceivedMovement struct {
	Source ent.ID
	Dest   ent.ID
}

type BroadcastedMovement struct {
	ID       ent.ID
	Owner    ent.ID
	Received ReceivedMovement
}

func protocolExample() {
	v := vec.V{2.0, 3.0}
	c := ent.Circle{v, 5.0}
	cm := CollisionMovement{1, 10, c, v}
	cell := GeneratedCell{1, c}
	startCell := StartCell{1, 2}
	cv := CellVires{1, 20}
	rm := ReceivedMovement{1, 2}
	ex := []interface{}{
		"Collision (sent by the server when a collision occurs):",
		&Collision{cm, cm},
		"Conflict (sent by the server when a conflict occurs):",
		&Conflict{
			5,
			ConflictCell{
				1,
				2,
				10,
			},
		},
		"EliminatedPlayer (sent by the server when a player dies):",
		1,
		"Winner (sent by the server when a player wins the game):",
		1,
		"Replication (sent by the server when all cells replicate)",
		Replication([]CellVires{cv, cv, cv}),
		"Field (sent by the server when the field is generated)",
		&Field{
			[]GeneratedCell{cell, cell, cell, cell, cell},
			[]StartCell{startCell, startCell},
		},
		"Joined (sent by the server when a user joins the room)",
		&UserJoined{1},
		"Move (sent by the client when moving vires):",
		&rm,
		"Move (sent by the server when a player moved vires):",
		&BroadcastedMovement{
			1,
			1,
			rm,
		},
	}
	var b bytes.Buffer
	for _, v := range ex {
		m, _ := json.MarshalIndent(v, "", "\t")
		b.Write(m)
		b.WriteByte('\n')
	}
	io.Copy(os.Stdout, &b)
}

func main() {
	protocolExample()
}
