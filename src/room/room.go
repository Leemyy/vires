package room

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/mhuisi/flog/weblog"
)

const (
	version   = "0.1"
	txBufSize = 1024
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
}

type Room struct {
	users chan []*User
}

func NewRoom() *Room {
	users := make(chan []*User, 1)
	users <- []*User{}
	return &Room{users}
}

type User struct {
	conn *websocket.Conn
	quit chan struct{}
	tx   chan TX
}

func NewUser(c *websocket.Conn) *User {
	return &User{
		c,
		make(chan struct{}),
		make(chan TX, txBufSize),
	}
}

func transmitPackets(u *User) {
	go func() {
		for packet := range u.tx {
			select {
			case <-u.quit:
				return
			default:
				err := u.conn.WriteJSON(packet)
				if err != nil {
					// if we cannot send then something is
					// wrong with the server
					weblog.Backend().Fatalf("Cannot send over socket: %s\n", err)
				}
			}
		}
	}()
}

func receivePackets(u *User, r *Room) {
	go func() {
		for {
			select {
			case <-u.quit:
				return
			default:
				p := RX{}
				err := u.conn.ReadJSON(&p)
				if err != nil {
					// handle read err
				}
				if p.Version != version {
					// handle version err
				}
				unmarshal := func(v interface{}) error {
					return json.Unmarshal(p.Data, v)
				}
				switch p.Type {
				case "Foo":
					f := &struct {
						Msg string
					}{}
					err := unmarshal(f)
					if err != nil {
						// handle invalid json err
					}
					fmt.Println(f)
				default:
					// handle unknown packet type err
				}
			}
		}
	}()
}

func (r *Room) Connect(c *websocket.Conn) {
	u := NewUser(c)
	transmitPackets(u)
	us := <-r.users
	us = append(us, u)
	r.users <- us
	receivePackets(u, r)
	u.tx <- TX{"Foo", version, "Hello, World"}
}
