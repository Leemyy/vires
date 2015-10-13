package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/mhuisi/flog/weblog"
)

const (
	roomIDPattern = "{roomid:[0-9]+}"
	staticDir     = "./static/"
	tmplDir       = "./tmpl/"
)

var (
	upgrader = websocket.Upgrader{}
	roomTmpl = template.Must(template.ParseFiles(tmplDir + "room.html"))
)

type Foo struct {
	Msg string
}

func onMainPage(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Connected to main page")
	// Possibly cache files?
	http.ServeFile(w, req, staticDir+"main.html")
}

func onRoom(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["roomid"]
	fmt.Println("Connected to room id:", id)
	roomTmpl.Execute(w, id)
}

func connectToRoom(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		weblog.Backend().Printf("Cannot open websocket connection: %s\n", err)
		return
	}
	f := &Foo{}
	err = ws.ReadJSON(f)
	if err != nil {
		weblog.Backend().Printf("Cannot read websocket message: %s", err)
		return
	}
	fmt.Println(f.Msg)
	f2 := &Foo{Msg: "Hello from the server!"}
	err = ws.WriteJSON(f2)
	if err != nil {
		weblog.Backend().Printf("Cannot write websocket message: %s", err)
	}
	// read json

}

func main() {
	weblog.Open(".", 1024)
	r := mux.NewRouter()
	// Main page
	r.HandleFunc("/", onMainPage).Methods("GET")
	// Rooms
	r.HandleFunc(fmt.Sprintf("/%s", roomIDPattern), onRoom).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/c", roomIDPattern), connectToRoom)
	http.Handle("/", r)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		weblog.Backend().Fatalf("Cannot start webserver: %s\n", err)
	}
}
