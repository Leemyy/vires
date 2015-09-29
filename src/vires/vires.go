package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mhuisi/flog/weblog"
)

const staticDir = "./static/"

func onMainPage(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Connected to main page")
	// Possibly cache files?
	http.ServeFile(res, req, staticDir+"main.html")
}

func onRoom(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Connected to room id:", mux.Vars(req)["roomid"])
	http.ServeFile(res, req, staticDir+"room.html")
}

func connectToRoom(res http.ResponseWriter, req *http.Request) {

}

func main() {
	weblog.Open(".", 1024)
	r := mux.NewRouter()
	// Main page
	r.HandleFunc("/", onMainPage).Methods("GET")
	// Rooms
	roomIDPattern := "{roomid:[0-9]+}"
	r.HandleFunc(fmt.Sprintf("/%s", roomIDPattern), onRoom).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/c", roomIDPattern), connectToRoom)
	http.Handle("/", r)
	err := http.ListenAndServe(":80", nil)

	if err != nil {
		weblog.Backend().Fatalf("Cannot start webserver: %s\n", err)
	}
}
