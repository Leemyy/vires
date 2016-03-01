// Package transm provides types and functions
// for the data transmission between client
// and server as well as for the communication
// between the game and the server.
package transm

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/mhuisi/vires/src/ent"
	"github.com/mhuisi/vires/src/vec"
)

const Version = "0.1"

// RX is a packet received
// by the client and used for
// pre-parsing the type and the
// version of the packet.
type RX struct {
	sender  ent.ID
	Type    string
	Version string
	Data    json.RawMessage
}

// MakeRX creates a new, empty packet
// for parsing stream data by the specified
// sender.
func MakeRX(sender ent.ID) RX {
	return RX{sender: sender}
}

// Sender gets the sender of this packet.
func (rx RX) Sender() ent.ID {
	return rx.sender
}

// TX is a packet sent to the client.
type TX struct {
	Type    string
	Version string
	Data    interface{}
}

func newTX(typ string, data interface{}) TX {
	return TX{typ, Version, data}
}

// Movement represents a movement
// transmitted to the client.
type Movement struct {
	ID        ent.ID
	Owner     ent.ID
	Moving    ent.Vires
	Body      ent.Circle
	Direction vec.V
}

func newMov(m *ent.Movement) *Movement {
	return &Movement{m.ID(), m.Owner().ID(), m.Moving(), m.Body(), m.Direction()}
}

// Collision is transmitted by the server
// when a collision occurs.
type Collision struct {
	A, B *Movement
}

// ConflictCell represents a cell
// part of a conflict.
type ConflictCell struct {
	ID        ent.ID
	Stationed ent.Vires
	Owner     ent.ID
}

func makeConflCell(c *ent.Cell) ConflictCell {
	return ConflictCell{c.ID(), c.Stationed(), c.OwnerID()}
}

// Conflict is transmitted by the server
// when a conflict occurs.
type Conflict struct {
	Movement ent.ID
	Cell     ConflictCell
}

// EliminatedPlayer is transmitted by the server
// when a player dies.
type EliminatedPlayer ent.ID

// Winner is transmitted by the server
// when a player wins the match.
// If Winner is 0, nobody won the match.
type Winner ent.ID

// CellVires represents the amount of vires
// of a cell.
type CellVires struct {
	ID        ent.ID
	Stationed ent.Vires
}

func makeCellVires(c *ent.Cell) CellVires {
	return CellVires{c.ID(), c.Stationed()}
}

// Replication is transmitted by the server
// when cells replicate (i.e. gain vires through
// a growth cycle)
type Replication []CellVires

// GeneratedCell represents a cell that was
// generated as part of map generation.
type GeneratedCell struct {
	ID       ent.ID
	Body     ent.Circle
	Capacity ent.Vires
}

func makeGenCell(c *ent.Cell) GeneratedCell {
	return GeneratedCell{c.ID(), c.Body(), c.Capacity()}
}

// StartCell represents the cell a player
// starts out with when the game is started.
type StartCell struct {
	Owner ent.ID
	Cell  ent.ID
}

func makeStartCell(c *ent.Cell) StartCell {
	return StartCell{c.OwnerID(), c.ID()}
}

// Field is transmitted by the server
// when a map was generated.
type Field struct {
	Cells      []GeneratedCell
	StartCells []StartCell
	Size       vec.V
}

// Transmitter is a binding between
// the server logic and the server.
//
// When the server sided game logic
// calculates something that the user
// has to know about, it pipes that
// data into the Transmitter.
//
// The server listens to the
// Transmitter and relays the data
// to clients.
type Transmitter struct {
	packets chan TX
}

// Open opens the Transmitter,
// enabling communication.
func (t *Transmitter) Open() {
	t.packets = make(chan TX, 1024)
}

// Disable nils the Transmitter,
// disabling communication and disabling
// the transmitter as case in selects.
//
// This also makes sure that when a
// game ends, the transmitter is niled,
// so no left over data is transmitted
// after the game ended.
func (t *Transmitter) Disable() {
	t.packets = nil
}

// Packets gets the packet channel.
func (t *Transmitter) Packets() <-chan TX {
	return t.packets
}

func (t *Transmitter) sendTX(typ string, data interface{}) {
	t.packets <- newTX(typ, data)
}

// Move transmits a movement packet.
func (t *Transmitter) Move(m *ent.Movement) {
	t.sendTX("Movement", newMov(m))
}

// Collide transmits a collision packet.
func (t *Transmitter) Collide(a, b *ent.Movement) {
	t.sendTX("Collision", &Collision{newMov(a), newMov(b)})
}

// Conflict transmits a conflict packet.
func (t *Transmitter) Conflict(m *ent.Movement, c *ent.Cell) {
	t.sendTX("Conflict", &Conflict{m.ID(), makeConflCell(c)})
}

// Eliminate transmits a packet meaning that a player was eliminated.
func (t *Transmitter) Eliminate(p *ent.Player) {
	e := EliminatedPlayer(p.ID())
	t.sendTX("EliminatedPlayer", &e)
}

// Win transmits a packet meaning that a player won the game.
func (t *Transmitter) Win(p *ent.Player) {
	w := Winner(p.ID())
	t.sendTX("Winner", &w)
}

// Replicate transmits a packet containing updated
// stationed vires in cells after a replication cycle.
func (t *Transmitter) Replicate(field map[ent.ID]*ent.Cell) {
	cvs := make([]CellVires, len(field))
	i := 0
	for _, c := range field {
		cvs[i] = makeCellVires(c)
		i++
	}
	t.sendTX("Replication", cvs)
}

// GenerateField transmits a packet containing the entire
// field that was generated.
func (t *Transmitter) GenerateField(fieldCells map[ent.ID]*ent.Cell, size vec.V) {
	cells := make([]GeneratedCell, len(fieldCells))
	startCells := []StartCell{}
	i := 0
	for _, c := range fieldCells {
		cells[i] = makeGenCell(c)
		if c.Owner() != nil {
			startCells = append(startCells, makeStartCell(c))
		}
		i++
	}
	t.sendTX("Field", &Field{cells, startCells, size})
}

// UserJoined is transmitted by the server
// when at any time a user joins the room.
type UserJoined ent.ID

// MakeUserJoined creates a new UserJoined packet.
func MakeUserJoined(userID ent.ID) TX {
	return newTX("UserJoined", userID)
}

// OwnID is transmitted by the server
// to connecting clients to notify
// them about their own ID.
type OwnID ent.ID

// MakeOwnID creates a new OwnID packet.
func MakeOwnID(userID ent.ID) TX {
	return newTX("OwnID", userID)
}

// ReceivedMovement is transmitted by the client
// when he decides to send a movement.
type ReceivedMovement struct {
	Source ent.ID
	Dest   ent.ID
}

// protocolExample prints an example
// for all the top level packet types
// as defined by this package to stdout
// as a json stream.
func protocolExample() {
	v := vec.V{2.0, 3.0}
	c := ent.Circle{v, 5.0}
	mv := &Movement{1, 1, 10, c, v}
	cell := GeneratedCell{1, c, 10}
	startCell := StartCell{1, 2}
	cv := CellVires{1, 20}
	ex := []interface{}{
		"RX (packets sent to the server):",
		RX{
			Type:    "Movement",
			Version: Version,
			Data:    json.RawMessage("some payload, may be any json type, see below"),
		},
		"TX (packets sent by the server):",
		newTX("Collision", "some payload, may be any json type, see below"),
		"Payloads:",
		"Movement (sent by the server when a movement was started):",
		mv,
		"Collision (sent by the server when a collision occurs):",
		Collision{mv, mv},
		"Conflict (sent by the server when a conflict occurs):",
		Conflict{
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
		"Replication (sent by the server when all cells replicate):",
		Replication([]CellVires{cv, cv, cv}),
		"Field (sent by the server when the field is generated):",
		Field{
			[]GeneratedCell{cell, cell, cell, cell, cell},
			[]StartCell{startCell, startCell},
			v,
		},
		"UserJoined (sent by the server when a user joins the room):",
		1,
		"OwnID (sent by the server to users to tell them their ID when joining):",
		1,
		"Movement (sent by the client when moving vires):",
		ReceivedMovement{1, 2},
	}
	var b bytes.Buffer
	for _, v := range ex {
		m, _ := json.MarshalIndent(v, "", "\t")
		b.Write(m)
		b.WriteByte('\n')
	}
	io.Copy(os.Stdout, &b)
}

// main to print protocolExample when protocol changes
func main() {
	protocolExample()
}
