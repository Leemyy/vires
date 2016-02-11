// Package room maintains the game field of vires.
package room

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/mhuisi/vires/src/ent"
	"github.com/mhuisi/vires/src/game"
	"github.com/mhuisi/vires/src/transm"
)

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

// TX is a packet sent to the client.
type TX struct {
	Sender  ent.ID
	Type    string
	Version string
	Data    interface{}
}

func newTX(sender ent.ID, typ string, data interface{}) TX {
	return TX{sender, typ, transm.Version, data}
}

type userConn struct {
	id   ent.ID
	conn *websocket.Conn
	send chan TX
}

// Room represents a single entity
// that hosts a match.
type Room struct {
	users map[*websocket.Conn]userConn
	// current userid
	uid ent.ID
	// joining users
	join chan *websocket.Conn
	// users to disconnect
	kill chan userConn
	// channel to process packets from users
	read     chan RX
	gameMsgs *transm.Transmitter
	// starts a match and echos back if it was started
	startMatch chan chan<- bool
	field      *game.Field
}

// NewRoom creates a new room and launches
// the handler for inbound connections
// and starting the match.
func NewRoom() *Room {
	r := &Room{
		users:      map[*websocket.Conn]userConn{},
		uid:        1,
		join:       make(chan *websocket.Conn, 16),
		kill:       make(chan userConn, 16),
		read:       make(chan RX, 512),
		gameMsgs:   &transm.Transmitter{},
		startMatch: make(chan chan<- bool),
	}
	go r.handler()
	return r
}

// userWriter writes all packets from
// send to conn as JSON.
//
// when an error occurs, the user with the specified
// userid is killed.
func (r *Room) userWriter(c userConn) {
	for tx := range c.send {
		err := c.conn.WriteJSON(tx)
		if err != nil {
			r.kill <- c
			return
		}
	}
}

// userReader reads all packets from
// userConn.conn, pre-parses them
// as RX and sends the packet
// over Room.read.
//
// when an error occurs, the user with the specified
// userid is killed.
func (r *Room) userReader(c userConn) {
	for {
		p := RX{sender: c.id}
		err := c.conn.ReadJSON(&p)
		if err != nil {
			r.kill <- c
			return
		}
		r.read <- p
	}
}

// killUser disconnects a user completly,
// cleaning up all of his resources.
func (r *Room) killUser(c userConn) {
	if _, ok := r.users[c.conn]; !ok {
		return
	}
	delete(r.users, c.conn)
	// causes error in userReader and userWriter, terminating both
	// (a duplicate disconnect msg is sent but this is handled
	// at the start of this function and not very expensive)
	c.conn.Close()
	if r.field != nil {
		r.field.DisconnectPlayer(c.id)
	}
}

// broadcast sends a message with the specified type and
// the specified payload as data by the specified sender
// to everyone in the room.
func (r *Room) broadcast(sender ent.ID, typ string, data interface{}) {
	for _, userConn := range r.users {
		select {
		case userConn.send <- newTX(sender, typ, data):
		default:
			// send channel is blocked, user is too slow: kill user
			r.killUser(userConn)
		}
	}
}

// handleRX parses the actual packet, checks it for validity
// in the context, notifies the game when necessary and returns
// the payload of the packet and if the packet was accepted.
func (r *Room) handleRX(p RX) (payload interface{}, ok bool) {
	if p.Version != transm.Version {
		return nil, false
	}
	unmarshal := func(v interface{}) error {
		return json.Unmarshal(p.Data, v)
	}
	switch p.Type {
	case "Move":
		if r.field == nil {
			return nil, false
		}
		m := transm.ReceivedMovement{}
		err := unmarshal(&m)
		if err != nil {
			return nil, false
		}
		mvid, validMovement := r.field.Move(p.sender, m.Source, m.Dest)
		if !validMovement {
			return nil, false
		}
		bm := &transm.BroadcastedMovement{mvid, p.sender, m}
		return bm, true
	}
	return nil, false
}

// handler is a monitor goroutine
// that manages anything that happens inside
// of the room, like managing users,
// remoting packets from the game to the clients,
// receiving packets and starting/stopping the game.
func (r *Room) handler() {
	join := r.join
	kill := r.kill
	read := r.read
	gameMsgs := r.gameMsgs
	// gameMsgs members are not cached
	// because they can become nil
	// to block the respective cases
	for {
		select {
		case conn := <-join:
			id := r.uid
			send := make(chan TX, 64)
			uconn := userConn{id, conn, send}
			go r.userReader(uconn)
			go r.userWriter(uconn)
			r.uid++
			r.users[conn] = uconn
			r.broadcast(0, "Join", &transm.UserJoined{id})
		case u := <-kill:
			r.killUser(u)
		case m := <-read:
			payload, validPacket := r.handleRX(m)
			if validPacket {
				r.broadcast(m.sender, m.Type, payload)
			}
		case c := <-gameMsgs.Collisions():
			r.broadcast(0, "Collision", c)
		case c := <-gameMsgs.Conflicts():
			r.broadcast(0, "Conflict", c)
		case e := <-gameMsgs.EliminatedPlayers():
			r.broadcast(0, "EliminatedPlayer", e)
		case w := <-gameMsgs.Winner():
			r.field.Close()
			// set nil to block future movements
			r.field = nil
			r.gameMsgs.Disable()
			r.broadcast(0, "Winner", w)
		case rep := <-gameMsgs.Replications():
			r.broadcast(0, "Replication", rep)
		case f := <-gameMsgs.GeneratedField():
			r.broadcast(0, "Field", f)
		case started := <-r.startMatch:
			if r.field != nil {
				started <- false
				break
			}
			started <- true
			uids := make([]ent.ID, len(r.users))
			i := 0
			for _, userConn := range r.users {
				uids[i] = userConn.id
				i++
			}
			r.gameMsgs.Open()
			r.field = game.NewField(uids, r.gameMsgs)
		}
	}
}

// StartMatch starts the match and returns whether
// the match was started or if it was already running.
func (r *Room) StartMatch() bool {
	started := make(chan bool)
	r.startMatch <- started
	return <-started
}

// Connect connects a user with the specified connection to the room.
func (r *Room) Connect(c *websocket.Conn) {
	r.join <- c
}
