package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/coff33un/game-server-ms/src/game"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {

	// Registered Room Games
	Rooms map[*Room]bool
	Byid  map[string]*Room

	// Registers new rooms.
	register chan *Room
	// Unregister rooms.
	unregister chan *Room
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Room),
		unregister: make(chan *Room),
		Rooms:      make(map[*Room]bool),
		Byid:       make(map[string]*Room),
	}
}

func ParseFormBadRequest(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to Parse Request", http.StatusBadRequest)
		return err
	}
	return nil
}

func disableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func (h *Hub) getRoom(id string) (*Room, error) {
	room, ok := h.Byid[id]
	if !ok {
		return nil, errors.New("Room Not Found")
	}
	return room, nil
}

func (h *Hub) getRoom404(w http.ResponseWriter, r *http.Request) (*Room, error) {
	vars := mux.Vars(r)
	roomid := vars["roomid"]
	room, err := h.getRoom(roomid)
	if err != nil {
		http.NotFound(w, r)
		return nil, err
	}
	return room, nil
}

func (h *Hub) RoomReady(w http.ResponseWriter, r *http.Request) {
	disableCors(&w)
	switch r.Method {
	case http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Methdods", "OPTIONS,GET")
	case http.MethodGet:
		room, err := h.getRoom404(w, r)
		if err != nil {
			log.Println(err)
			return
		}
		if !room.Ready {
			w.WriteHeader(http.StatusAccepted) // Room exists but pending ready
			fmt.Fprintln(w, "Room Not Ready")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Room Ready")
	}
}

func (h *Hub) SetupReady(w http.ResponseWriter, r *http.Request) {
	disableCors(&w)
	switch r.Method {
	case http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,PUT")
	case http.MethodGet:
		room, err := h.getRoom404(w, r)
		if err != nil {
			log.Println(err)
			return
		}
		if !room.Ready || !room.Setup {
			w.WriteHeader(http.StatusAccepted) // Room exists but pending ready
			fmt.Fprintln(w, "Room Not Ready")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Room Setup and Ready")
	}
}

func (h *Hub) SetupRoom(w http.ResponseWriter, r *http.Request) {
	disableCors(&w)
	fmt.Println("Setup Room")
	fmt.Println(r.Method)
	switch r.Method {
	case http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,PUT")
	case http.MethodPut:
		room, err := h.getRoom404(w, r)
		if err != nil {
			log.Println(err)
			return
		}
		if room.Running {
			fmt.Println("Cannot Setup Running Room")
			return
		}
		var v game.SetupGameMessage
		err = json.NewDecoder(r.Body).Decode(&v)
		if err != nil {
			fmt.Println("Error Decoding", err)
			log.Println(err)
			return
		}
		fmt.Println("Room Setup")
		room.SetupGame(v)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": room.ID})
	}
}

func (h *Hub) StartRoom(w http.ResponseWriter, r *http.Request) {
	disableCors(&w)
	switch r.Method {
	case http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,PUT")
	case http.MethodPut:
		room, err := h.getRoom404(w, r)
		if err != nil {
			fmt.Println(err)
			log.Println(err)
			return
		}
		if room.Running {
			fmt.Println("Room already Running")
			return
		}
		fmt.Println("Start Room")
		room.StartGame()
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": room.ID})
	}
}

func (hub *Hub) GetRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Println("here")
	disableCors(&w)
	switch r.Method {
	case http.MethodGet:
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			return
		}
		id := vars["roomid"]
		w.Header().Set("Content-Type", "application/json")
		if room, ok := hub.Byid[id]; ok {
			json.NewEncoder(w).Encode(room)
		} else {
			http.NotFound(w, r)
		}
	}
}

func (hub *Hub) CreateRoom(w http.ResponseWriter, r *http.Request) {
	disableCors(&w)
	switch r.Method {
	case http.MethodPost:
		room := NewRoom(hub)
		js, err := json.Marshal(map[string]string{"id": room.ID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println("Created room")
		go room.Run()
		fmt.Println("room Run running")
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	room, err := h.getRoom404(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	err = ParseFormBadRequest(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	token := r.Form.Get("token")
	var id, handle string

	f, err := TokenQuery(token)
	if err != nil {
		log.Println(err)
		return
	}
	id = strconv.Itoa(f.Data.User.Id)
	handle = f.Data.User.Handle

	fmt.Printf("The id is: %v, the handle is: %v\n", id, handle)

	if !room.OkToConnectPlayer(id) {
		fmt.Println("Not ok to connect to room")
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error", err)
		log.Println(err)
		return
	}
	client := &Client{room: room, ID: id, Handle: handle, conn: conn, send: make(chan map[string]interface{}, 256)}
	client.room.register <- client

	fmt.Println("ServeWS: Registered new client to room")
}

func (h *Hub) Run() {
	for {
		fmt.Println("Hub run here")
		select {
		case room := <-h.register:
			h.Rooms[room] = true
			h.Byid[room.ID] = room
		case room := <-h.unregister:
			if _, ok := h.Rooms[room]; ok {
				delete(h.Rooms, room)
				delete(h.Byid, room.ID)
			}
		}
	}
}
