package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"time"
	"net/http"
	"sync"
)

/* Controls a bunch of rooms */
type Hub struct {
	hub      map[string]*Room
	upgrader websocket.Upgrader
}

/* If room doesn't exist creates it then returns it */
func (h *Hub) GetRoom(name string) *Room {
	if _, ok := h.hub[name]; !ok {
		h.hub[name] = NewRoom(name)
	}
	return h.hub[name]
}

/* Get ws conn. and hands it over to correct room */
func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	room_name := params["room"]
	var wg sync.WaitGroup
	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("%s upgrade err: %s", r.RemoteAddr, err)
		// log.Printf("%s error header: %s", r.RemoteAddr, r.Header)
		return
	}
	defer c.Close()
	log.Printf("%s header: %s", r.RemoteAddr, r.Header)
	room := h.GetRoom(room_name)

	id := room.Join(c)
	/* Reads from the client's out bound channel and broadcasts it */
	
	wg.Add(1)
	go room.HandleMsg(id, &wg)
	/* Reads from client and if this loop breaks then client disconnected. */
	room.clients[id].ReadLoop(&wg)
	log.Printf("Client %d left room: %s", id, room_name)
	room.Leave(id)

	// log.Printf("Client %d left room: %s", id, room_name)

}

/* Get ws conn. and hands it over to correct room */
func (h *Hub) HandleMainLobbyWS(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	room_name := params["room"]
	var wg sync.WaitGroup
	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("%s upgrade err: %s", r.RemoteAddr, err)
		// log.Printf("%s error header: %s", r.RemoteAddr, r.Header)
		return
	}
	defer c.Close()
	log.Printf("%s header: %s", r.RemoteAddr, r.Header)
	room := h.GetRoom(room_name)

	id := room.Join(c)
	/* Reads from the client's out bound channel and broadcasts it */
	
	wg.Add(1)
	go room.HandleMainLobbyMsg(id, &wg)
	/* Reads from client and if this loop breaks then client disconnected. */
	room.clients[id].ReadLoop(&wg)
	log.Printf("Client %d left room: %s", id, room_name)
	room.Leave(id)
}

/* Constructor */
func NewHub() *Hub {
	hub := new(Hub)
	hub.hub = make(map[string]*Room)
	hub.upgrader = websocket.Upgrader{
		HandshakeTimeout: 30 * time.Second,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// log.Printf("%s header: %s", r.RemoteAddr, r.Header)
			// if r.Header["Origin"][0] == ""{
			// 	log.Printf("%s upgraded successfully.", r.RemoteAddr)
			// 	return true
			// }
			// return false
			return true
		},
	}
	return hub
}
