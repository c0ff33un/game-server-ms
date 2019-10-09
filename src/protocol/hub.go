package protocol

import (
  "fmt"
  "log"
  "net/http"

  "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
  ReadBufferSize: 1024,
  WriteBufferSize: 1024,
  CheckOrigin: func(r *http.Request) bool {return true},
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {

	// Registered Room Games
	rooms map[*Room]bool
  byid map[string]*Room

	// Registers new rooms.
	register chan *Room
	// Unregister rooms.
	unregister chan *Room
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Room),
		unregister: make(chan *Room),
		rooms:    make(map[*Room]bool),
		byid: make(map[string]*Room),
	}
}

func (h *Hub) Run() {
	for {
	  fmt.Println("Hub run here")
		select {
		case room := <-h.register:
			h.rooms[room] = true
			h.byid[room.ID] = room
		case room := <-h.unregister:
			if _, ok := h.rooms[room]; ok {
				delete(h.rooms, room)
				delete(h.byid, room.ID)
			}
	  }
	}
}

func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
  err := r.ParseForm()
  if err != nil {
    log.Println(err)
  }
  id := r.Form.Get("id")
  // Room Exists
  fmt.Println("ServeWS Accessed")
  if room, ok := h.byid[id]; ok {
    fmt.Println("Room exists")
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
      fmt.Println("Error")
      fmt.Println(err)
      log.Println(err)
    }
    fmt.Println("Here")
    client := &Client{room: room, conn: conn, send: make(chan []byte)}
    fmt.Println("Passes")
    client.room.register <- client
    fmt.Println("Registered new client to room")
    go client.WritePump()
    go client.ReadPump()
  } else {
    w.WriteHeader(http.StatusNotFound)
    fmt.Fprint(w, "No room found")
  }
}
