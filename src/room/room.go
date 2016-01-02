package room

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
)

const (
	Version = "0.1"
)

type Room struct {
	users chan []*User
}

func NewRoom() *Room {
	users := make(chan []*User, 1)
	users <- []*User{}
	return &Room{users}
}

func (r *Room) handlePackets(u *User) {
	rx := u.Receive()
	for p := range rx {
		if p.Version != Version {
			u.SendFatalf("Invalid protocol version.")
			return
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
			u.SendFatalf("Unknown package type.")
		}
	}
}

func (r *Room) Connect(c *websocket.Conn) {
	u := ConnectUser(c, Version)
	us := <-r.users
	us = append(us, u)
	r.users <- us
	go r.handlePackets(u)
	u.Send("Foo", "Hello, World")
}

func (r *Room) Disconnect(c *websocket.Conn) error {
	users := <-r.users
	for i, u := range users {
		if u.Conn == c {
			u.Disconnect()
			users = append(users[:i], users[i+1:]...)
			return nil
		}
	}
	return errors.New("Connection to disconnect not found.")
}
