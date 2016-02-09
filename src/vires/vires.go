package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/mhuisi/flog/weblog"
	"github.com/mhuisi/vires/src/room"
)

const (
	roomIDPattern = "{roomid:[0-9]+}"
	staticDir     = "./static/"
	tmplDir       = "./tmpl/"
)

var (
	upgrader = websocket.Upgrader{}
	roomTmpl = template.Must(template.ParseFiles(tmplDir + "room.html"))
	rooms    = map[string]*room.Room{}
)

func onMainPage(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Connected to main page")
	// Possibly cache files?
	http.ServeFile(w, req, staticDir+"main.html")
}

func roomID(r *http.Request) string {
	return mux.Vars(r)["roomid"]
}

func onRoom(w http.ResponseWriter, r *http.Request) {
	id := roomID(r)
	fmt.Println("Connected to room id:", id)
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
		ro = room.NewRoom()
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

func main() {
	weblog.Open(".", 1024)
	r := mux.NewRouter()
	// Main page
	r.HandleFunc("/", onMainPage).Methods("GET")
	// Rooms
	r.HandleFunc(fmt.Sprintf("/%s", roomIDPattern), onRoom).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/c", roomIDPattern), connectToRoom)
	r.HandleFunc(fmt.Sprintf("/%s/s", roomIDPattern), startMatch)
	http.Handle("/", r)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		weblog.Backend().Fatalf("Cannot start webserver: %s\n", err)
	}
}
