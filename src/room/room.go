package room

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/mhuisi/vires/src/game"
	"github.com/mhuisi/vires/src/game/ent"
	"github.com/mhuisi/vires/src/transm"
)

const (
	Version     = "0.1"
	joinBufSize = 16
	killBufSize = 16
	rxBufSize   = 512
	txBufSize   = 32
)

type RX struct {
	sender  ent.ID
	Type    string
	Version string
	Data    json.RawMessage
}

type TX struct {
	Type    string
	Version string
	Data    interface{}
}

type user struct {
	id   ent.ID
	conn *websocket.Conn
	send chan TX
	read chan RX
}

func (u *user) writer() {
	conn := u.conn
	for tx := range u.send {
		err := conn.WriteJSON(tx)
		if err != nil {
			break
		}
	}
	u.conn.Close()
}

func (u *user) reader() {
	uid := u.id
	conn := u.conn
	for {
		p := RX{sender: uid}
		err := conn.ReadJSON(&p)
		if err != nil {
			u.conn.Close()
			return
		}
		u.read <- p
	}
}

type Room struct {
	users      map[*user]struct{}
	uid        ent.ID
	join       chan *user
	kill       chan *user
	read       chan RX
	gameMsgs   *transm.Transmitter
	startMatch chan chan<- bool
	field      *game.Field
}

func NewRoom() *Room {
	r := &Room{
		users:      map[*user]struct{}{},
		uid:        1,
		join:       make(chan *user, joinBufSize),
		kill:       make(chan *user, killBufSize),
		read:       make(chan RX, rxBufSize),
		gameMsgs:   &transm.Transmitter{},
		startMatch: make(chan chan<- bool),
	}
	go r.handler()
	return r
}

func (r *Room) broadcast(typ string, data interface{}) {
	for u := range r.users {
		select {
		case u.send <- TX{typ, Version, data}:
		default:
			// send channel is blocked, user is too slow: kill user
			r.kill <- u
		}
	}
}

func (r *Room) handleRX(p RX) (payload interface{}, ok bool) {
	if p.Version != Version {
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
		m := &transm.Movement{}
		err := unmarshal(m)
		if err != nil {
			return nil, false
		}
		validMovement := r.field.Move(p.sender, m.Source, m.Dest)
		if !validMovement {
			return nil, false
		}
		return m, true
	}
	return nil, false
}

func (r *Room) handler() {
	join := r.join
	kill := r.kill
	read := r.read
	gameMsgs := r.gameMsgs
	collisions := gameMsgs.Collisions()
	conflicts := gameMsgs.Conflicts()
	eliminated := gameMsgs.EliminatedPlayers()
	winner := gameMsgs.Winner()
	for {
		select {
		case u := <-join:
			u.id = r.uid
			r.uid++
			r.users[u] = struct{}{}
		case u := <-kill:
			delete(r.users, u)
			// close send channel -> closes conn in writer -> results in error in reader
			// -> both writer and reader are terminated
			close(u.send)
		case m := <-read:
			payload, validPacket := r.handleRX(m)
			if validPacket {
				r.broadcast(m.Type, payload)
			}
		case c := <-collisions:
			r.broadcast("Collision", c)
		case c := <-conflicts:
			r.broadcast("Conflict", c)
		case e := <-eliminated:
			r.broadcast("EliminatedPlayer", e)
		case w := <-winner:
			r.field.Close()
			// set nil to block future movements
			r.field = nil
			r.gameMsgs.Disable()
			r.broadcast("Winner", w)
		case started := <-r.startMatch:
			if r.field != nil {
				started <- false
				break
			}
			started <- true
			uids := make([]ent.ID, len(r.users))
			i := 0
			for u := range r.users {
				uids[i] = u.id
				i++
			}
			r.gameMsgs.Open()
			r.field = game.NewField(uids, r.gameMsgs)
		}
	}
}

func (r *Room) StartMatch() bool {
	started := make(chan bool)
	r.startMatch <- started
	return <-started
}

func (r *Room) Connect(c *websocket.Conn) {
	u := &user{
		conn: c,
		send: make(chan TX, txBufSize),
		read: r.read,
	}
	go u.reader()
	go u.writer()
	r.join <- u
}
