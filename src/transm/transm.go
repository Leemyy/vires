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

	"github.com/mhuisi/vires/src/cfg"
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
	ID     ent.ID
	Source ent.ID
	Target ent.ID
	Owner  ent.ID
	Moving ent.Vires
	Body   ent.Circle
	Speed  float64
}

func newMov(m *ent.Movement) *Movement {
	return &Movement{m.ID(), m.Source().ID(), m.Target().ID(), m.Owner().ID(), m.Moving(), m.Body(), vec.Abs(m.Direction())}
}

type CollisionMovement struct {
	ID     ent.ID
	Moving ent.Vires
	Body   ent.Circle
	Speed  float64
}

func makeCollMov(m *ent.Movement) *CollisionMovement {
	return &CollisionMovement{m.ID(), m.Moving(), m.Body(), vec.Abs(m.Direction())}
}

// Collision is transmitted by the server
// when a collision occurs.
type Collision struct {
	A, B *CollisionMovement
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

// GeneratedCell represents a cell that was
// generated as part of map generation.
type GeneratedCell struct {
	ID          ent.ID
	Body        ent.Circle
	Stationed   ent.Vires
	Capacity    ent.Vires
	Replication ent.Vires
}

func makeGenCell(c *ent.Cell) GeneratedCell {
	return GeneratedCell{c.ID(), c.Body(), c.Stationed(), c.Capacity(), c.Replication()}
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
	Cells               []GeneratedCell
	StartCells          []StartCell
	Size                vec.V
	NeutralReplication  float64
	ReplicationInterval float64
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
	t.sendTX("Collision", Collision{makeCollMov(a), makeCollMov(b)})
}

// Conflict transmits a conflict packet.
func (t *Transmitter) Conflict(m *ent.Movement, c *ent.Cell) {
	t.sendTX("Conflict", Conflict{m.ID(), makeConflCell(c)})
}

// Eliminate transmits a packet meaning that a player was eliminated.
func (t *Transmitter) Eliminate(p ent.ID) {
	t.sendTX("EliminatedPlayer", EliminatedPlayer(p))
}

// Win transmits a packet meaning that a player won the game.
func (t *Transmitter) Win(p ent.ID) {
	t.sendTX("Winner", Winner(p))
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
	t.sendTX("Field", Field{cells, startCells, size, cfg.Gameplay.NeutralReplication, cfg.Gameplay.ReplicationInterval})
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
	collMov := &CollisionMovement{1, 100, c, 20}
	cell := GeneratedCell{1, c, 4, 10, 20}
	startCell := StartCell{1, 2}
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
		&Movement{1, 1, 1, 1, 10, c, 15},
		"Collision (sent by the server when a collision occurs):",
		Collision{collMov, collMov},
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
		"Field (sent by the server when the field is generated):",
		Field{
			[]GeneratedCell{cell, cell, cell, cell, cell},
			[]StartCell{startCell, startCell},
			v,
			0.5,
			1.5,
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
