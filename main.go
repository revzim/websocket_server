package main

import (
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"

)

var addr = flag.String("addr", ":8081", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)
	hub := NewHub()
	router := mux.NewRouter()
	router.Host(":8081")
	router.HandleFunc("/ws/{room}", hub.HandleWS).Methods("GET")
	router.HandleFunc("/ws_freeroamlobby/{room}", hub.HandleMainLobbyWS).Methods("GET")
	http.Handle("/", router)


	log.Printf("http_err: %v", http.ListenAndServe(":8081", nil))
}
