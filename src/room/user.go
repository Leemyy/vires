package room

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

const (
	txBufSize   = 1024
	rxBufSize   = 1024
	ErrorPacket = "error"
)

type RX struct {
	Type    string
	Version string
	Data    json.RawMessage
}

type TX struct {
	Type    string
	Version string
	Data    interface{}
	done    chan struct{}
}

type User struct {
	Conn    *websocket.Conn
	quit    chan struct{}
	Version string
	tx      chan TX
	rx      chan RX
}

func (u *User) transmitPackets() {
	for {
		select {
		case <-u.quit:
			return
		case packet := u.tx:
			err := u.Conn.WriteJSON(packet)
			close(packet.done)
			if err != nil {
				// does not send error to user to avoid recursive erroring
				u.Disconnect()
				return
			}
		}
	}
}

func (u *User) receivePackets() {
	for {
		select {
		case <-u.quit:
			return
		default:
			p := RX{}
			// issue: after a disconnect this waits until json is received
			err := u.Conn.ReadJSON(&p)
			if err != nil {
				u.SendFatalf("Could not read from your websocket connection: %s", err)
				return
			}
			u.rx <- p
		}
	}
}

func ConnectUser(c *websocket.Conn, protocolVersion string) *User {
	u := &User{
		Conn:    c,
		Quit:    make(chan struct{}),
		Version: protocolVersion,
		tx:      make(chan TX, txBufSize),
		rx:      make(chan RX, rxBufSize),
	}
	go u.transmitPackets(u)
	go u.receivePackets(u)
	return u
}

func (u *User) Disconnect() error {
	close(u.quit)
	close(u.rx)

	return u.Conn.Close()
}

func (u *User) Send(typ string, v interface{}) chan struct{} {
	done := make(chan struct{}, 1)
	u.tx <- TX{typ, u.Version, v, done}
	return done
}

func (u *User) Receive() <-chan RX {
	return u.rx
}

func (u *User) SendErrorf(s string, v ...interface{}) chan struct{} {
	return u.Send(ErrorPacket, fmt.Sprintf(s, v))
}

func (u *User) SendFatalf(s string, v ...interface{}) {
	// wait until message is transmitted
	<-u.SendErrorf(s, v)
	u.Disconnect()
}
