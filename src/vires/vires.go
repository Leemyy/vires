// vires is a simple multiplayer RTS game.
package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/mhuisi/fcfg/jsoncfg"
	"github.com/mhuisi/vires/src/room"
)

type Config struct {
	UseTLS         bool
	PrivateKeyPath string
	CertPath       string
}

func (c *Config) Copy() interface{} {
	clone := *c
	return &clone
}

var (
	defaultCfg = &Config{
		false,
		"",
		"",
	}
	cfg *Config
)

const (
	roomIDPattern = "{roomid:[0-9]+}"
)

var (
	upgrader  = websocket.Upgrader{}
	roomTmpl  = template.Must(template.ParseFiles("./res/room.html"))
	rooms     = map[string]*room.Room{}
	roomQuits = make(chan *room.Room, 32)
)

// roomID gets the id of the current room from the url of an http request.
func roomID(r *http.Request) string {
	return mux.Vars(r)["roomid"]
}

func onRoom(w http.ResponseWriter, r *http.Request) {
	id := roomID(r)
	roomTmpl.Execute(w, id)
}

func connectToRoom(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade already transmits an http error on error
		return
	}
	id := roomID(r)
	ro, ok := rooms[id]
	if !ok {
		ro = room.NewRoom(id, roomQuits)
		rooms[id] = ro
	}
	ro.Connect(ws)
}

func startMatch(w http.ResponseWriter, r *http.Request) {
	id := roomID(r)
	ro, ok := rooms[id]
	if !ok {
		http.Error(w, "No room with this ID exists.", 400)
		return
	}
	started := ro.StartMatch()
	if !started {
		http.Error(w, "Match already started.", 400)
		return
	}
	fmt.Fprintf(w, "Match started.")
}

func quitRooms() {
	for r := range roomQuits {
		delete(rooms, r.ID())
	}
}

func main() {
	cfgLoader := jsoncfg.New("cfg.json", defaultCfg)
	cfgT, err := cfgLoader.Load()
	if err != nil {
		log.Fatalf("Cannot load config file: %s\n", err)
	}
	cfg = cfgT.(*Config)
	r := mux.NewRouter()
	// Rooms
	r.PathPrefix("/res").Handler(http.StripPrefix("/res", http.FileServer(http.Dir("./res/"))))
	r.HandleFunc(fmt.Sprintf("/%s", roomIDPattern), onRoom).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/c", roomIDPattern), connectToRoom)
	r.HandleFunc(fmt.Sprintf("/%s/s", roomIDPattern), startMatch)
	go quitRooms()
	http.Handle("/", r)
	fmt.Println("Webserver starting.")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatalf("Cannot start webserver: %s\n", err)
	}
}
