// vires is a simple multiplayer RTS game.
package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/mhuisi/logg"
	"github.com/mhuisi/vires/src/cfg"
	"github.com/mhuisi/vires/src/room"
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

func httpsRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://"+cfg.General.IP+r.RequestURI, http.StatusMovedPermanently)
}

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
	logg.Info("Started match in room %s.", id)
	fmt.Fprintf(w, "Match started.")
}

func quitRooms() {
	for r := range roomQuits {
		delete(rooms, r.ID())
	}
}

func startServer() {
	if cfg.General.UseTLS {
		logg.Info("Starting HTTPS webservers.")
		go func() {
			if err := http.ListenAndServeTLS(":443", cfg.General.CertPath, cfg.General.PrivateKeyPath, nil); err != nil {
				logg.Fatal("Cannot start HTTPS webserver: %s", err)
			}
		}()
		if err := http.ListenAndServe(":80", http.HandlerFunc(httpsRedirect)); err != nil {
			logg.Fatal("Cannot start HTTP webserver (https redirect): %s", err)
		}
	} else {
		logg.Info("Starting HTTP webserver.")
		if err := http.ListenAndServe(":80", nil); err != nil {
			logg.Fatal("Cannot start HTTP webserver: %s", err)
		}
	}
}

func main() {
	logg.UseDebug = cfg.General.DebugLogging
	r := mux.NewRouter()
	r.StrictSlash(true)
	// Rooms
	r.PathPrefix("/res").Handler(http.StripPrefix("/res", http.FileServer(http.Dir("res")))).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s", roomIDPattern), onRoom).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/c", roomIDPattern), connectToRoom).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/s", roomIDPattern), startMatch).Methods("GET")
	go quitRooms()
	http.Handle("/", r)
	startServer()
}
