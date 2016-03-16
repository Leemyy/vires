// Package room maintains the game field of vires.
package room

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/mhuisi/logg"
	"github.com/mhuisi/vires/src/ent"
	"github.com/mhuisi/vires/src/game"
	"github.com/mhuisi/vires/src/transm"
)

type userConn struct {
	id   ent.ID
	conn *websocket.Conn
	send chan transm.TX
}

// Room represents a single entity
// that hosts a match.
type Room struct {
	id    string
	users map[*websocket.Conn]userConn
	// current userid
	uid ent.ID
	// joining users
	join chan *websocket.Conn
	// users to disconnect
	kill chan userConn
	// channel to process packets from users
	read     chan transm.RX
	gameMsgs *transm.Transmitter
	// starts a match and echos back if it was started
	startMatch chan chan<- bool
	notifyQuit chan<- *Room
	quit       chan struct{}
	field      *game.Field
}

// NewRoom creates a new room and launches
// the handler for inbound connections
// and starting the match.
func NewRoom(id string, notifyQuit chan<- *Room) *Room {
	r := &Room{
		id:         id,
		users:      map[*websocket.Conn]userConn{},
		uid:        1,
		join:       make(chan *websocket.Conn, 16),
		kill:       make(chan userConn, 16),
		read:       make(chan transm.RX, 512),
		gameMsgs:   &transm.Transmitter{},
		startMatch: make(chan chan<- bool),
		notifyQuit: notifyQuit,
		quit:       make(chan struct{}),
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
			logg.Info("Error while writing to conn of user %d in room %s: %s", c.id, r.id, err)
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
		p := transm.MakeRX(c.id)
		err := c.conn.ReadJSON(&p)
		if err != nil {
			logg.Info("Error while reading from conn of user %d in room %s: %s", c.id, r.id, err)
			r.kill <- c
			return
		}
		r.read <- p
	}
}

// killUser disconnects a user completly,
// cleaning up all of his resources.
func (r *Room) killUser(c userConn) {
	u, ok := r.users[c.conn]
	if !ok {
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
	if len(r.users) == 0 {
		close(r.quit)
	}
	logg.Info("Killing player %d in room %s.", u.id, r.id)
}

// send sends a message to the specified user.
//
// If the send channel is blocked because the user
// is too slow, the user is killed.
func (r *Room) send(c userConn, tx transm.TX) {
	select {
	case c.send <- tx:
	default:
		logg.Info("User %d in room %s too slow, killing user.", c.id, r.id)
		r.killUser(c)
	}
}

// broadcast sends a message to everyone in the room.
func (r *Room) broadcast(tx transm.TX) {
	logg.Debug("Sending packet '%s' in room %s: %+v", tx.Type, r.id, tx.Data)
	for _, userConn := range r.users {
		r.send(userConn, tx)
	}
}

// handleRX parses the actual packet, checks it for validity
// in the context and notifies the game when necessary.
func (r *Room) handleRX(p transm.RX) {
	if p.Version != transm.Version {
		return
	}
	unmarshal := func(v interface{}) error {
		err := json.Unmarshal(p.Data, v)
		logg.Debug("Received packet '%s' in room %s: %+v", p.Type, r.id, v)
		return err
	}
	switch p.Type {
	case "Movement":
		if r.field == nil {
			return
		}
		m := transm.ReceivedMovement{}
		err := unmarshal(&m)
		if err != nil {
			return
		}
		r.field.Move(p.Sender(), m.Source, m.Dest)
	}
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
	start := r.startMatch
	quit := r.quit
	for {
		select {
		case conn := <-join:
			id := r.uid
			send := make(chan transm.TX, 64)
			uconn := userConn{id, conn, send}
			go r.userReader(uconn)
			go r.userWriter(uconn)
			r.uid++
			r.users[conn] = uconn
			logg.Info("Client %s joined room %s and was assigned ID %d.", conn.RemoteAddr(), r.id, id)
			r.send(uconn, transm.MakeOwnID(id))
			r.broadcast(transm.MakeUserJoined(id))
		case u := <-kill:
			r.killUser(u)
		case m := <-read:
			r.handleRX(m)
		case p := <-gameMsgs.Packets():
			switch data := p.Data.(type) {
			case transm.Winner:
				logg.Info("Player %d won the match in room %s!", data, r.id)
				r.field.Close()
				// set nil to block future movements
				r.field = nil
				r.gameMsgs.Disable()
			}
			r.broadcast(p)
		case started := <-start:
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
		case <-quit:
			logg.Info("Closing room %s.", r.id)
			r.notifyQuit <- r
			// we assume that when a quit msg is sent, proper cleanup has already occured
			return
		}
	}
}

// ID gets the room id.
func (r *Room) ID() string {
	return r.id
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
