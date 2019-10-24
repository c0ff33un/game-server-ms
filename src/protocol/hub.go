package protocol



import (
  "fmt"
  "log"
  "os"
  "time"
  "context"
  "errors"
  "net"
  "net/http"
  "encoding/json"

  "github.com/gorilla/websocket"
  "github.com/coff33un/game-server-ms/src/common"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
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
	db *mongo.Client
	rooms map[*Room]bool
  Byid map[string]*Room

	// Registers new rooms.
	register chan *Room
	// Unregister rooms.
	unregister chan *Room
}

func NewHub() *Hub {
  ctx, _ := context.WithTimeout(context.Background(), 10 * time.Second)
  mongo_url := os.Getenv("MONGO_URL")
  if mongo_url == "" {
    mongo_url = "localhost:27017"
  }
  fmt.Println(mongo_url)
  addr, err := net.LookupHost(mongo_url)
  fmt.Println(addr, err)
  client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://" + mongo_url))
  if err != nil {
    log.Println(err)
  }
  err = client.Connect(ctx)
  if err != nil {
    log.Println(err)
  }
	return &Hub{
	  db: client,
		register:   make(chan *Room),
		unregister: make(chan *Room),
		rooms:    make(map[*Room]bool),
		Byid: make(map[string]*Room),
	}
}

func (h *Hub) Run() {
	for {
	  fmt.Println("Hub run here")
		select {
		case room := <-h.register:
			h.rooms[room] = true
			h.Byid[room.ID] = room
		case room := <-h.unregister:
			if _, ok := h.rooms[room]; ok {
				delete(h.rooms, room)
				delete(h.Byid, room.ID)
			}
	  }
	}
}

func (h *Hub) getRoom404(w http.ResponseWriter, r *http.Request) (*Room, error) {
  err := r.ParseForm()
  if err != nil {
    http.Error(w, "Unable to Parse Request", http.StatusBadRequest)
    return nil, err
  }
  id := r.Form.Get("id")
  if room, ok := h.Byid[id]; ok {
    return room, nil
  } else {
    http.NotFound(w, r)
    return nil, errors.New("Room Not Found")
  }
}

func (h *Hub) RoomReady(w http.ResponseWriter, r *http.Request) {
  switch(r.Method) {
  case http.MethodGet:
    room, err := h.getRoom404(w, r)
    if err != nil {
      log.Println(err)
      return
    }
    if room.Ready {
      w.WriteHeader(http.StatusOK)
      fmt.Fprintln(w, "Room Ready")
    } else {
      w.WriteHeader(http.StatusAccepted) // Room exists but pending ready
      fmt.Fprintln(w, "Room Not Ready")
    }
  }
}

func (h *Hub) SetupReady(w http.ResponseWriter, r *http.Request) {
  switch(r.Method) {
  case http.MethodGet:
    room, err := h.getRoom404(w, r)
    if err != nil {
      log.Println(err)
      return
    }
    if room.Ready && room.Setup {
      w.WriteHeader(http.StatusOK)
      fmt.Fprintln(w, "Room Setup and Ready")
    } else {
      w.WriteHeader(http.StatusAccepted) // Room exists but pending ready
      fmt.Fprintln(w, "Room Not Ready")
    }
  }
}

type SetupRoomMessage struct {
  Rows int
  Cols int
}

func (h *Hub) SetupRoom(w http.ResponseWriter, r *http.Request) {
  switch(r.Method) {
  case http.MethodOptions:
    common.DisableCors(&w)
    w.Header().Set("Access-Control-Allow-Methods", "PUT")
    w.WriteHeader(http.StatusOK)
  case http.MethodPut:
    room, err := h.getRoom404(w, r)

    if err != nil {
      log.Println(err)
      return
    }
    var v SetupRoomMessage
    err = json.NewDecoder(r.Body).Decode(&v)
    if err != nil {
      fmt.Println("Error Decoding", err)
      log.Println(err)
      return
    }
    fmt.Println("Setup Room", v)

    rows, cols := v.Rows, v.Cols
    room.SetupGame(rows, cols)
    // Enqueue Players World Update in game update channel
    /* for player := v.players {
      room.game.Update <- player
    } */
    common.DisableCors(&w)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"id": room.ID})
  }
}

func (h *Hub) StartRoom(w http.ResponseWriter, r *http.Request) {
  switch(r.Method) {
  case http.MethodOptions:
    common.DisableCors(&w)
    w.Header().Set("Access-Control-Allow-Methods", "PUT")
    w.WriteHeader(http.StatusOK)
  case http.MethodPut:
    room, err := h.getRoom404(w, r)
    if err != nil {
      fmt.Println(err)
      log.Println(err)
    }
    fmt.Println("Start Room")
    room.StartGame()
    common.DisableCors(&w)
    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"id": room.ID})
  }
}


func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
  room, err := h.getRoom404(w, r)

  if err != nil {
    log.Println(err)
    return
  }
  // Room Exists
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    fmt.Println("Error", err)
    log.Println(err)
  }
  client := &Client{room: room, ID: "", conn: conn, send: make(chan interface{})}
  client.room.register <- client
  fmt.Println("ServeWS: Registered new client to room")
  go client.WritePump()
  go client.ReadPump()
}
