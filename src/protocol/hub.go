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
  "net/url"
  "strconv"

  "github.com/coff33un/game-server-ms/src/common"
  "github.com/coff33un/game-server-ms/src/game"
  "github.com/gorilla/websocket"
  "github.com/gorilla/mux"
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

func (h *Hub) getRoom(id string) (*Room, error) {
  if room, ok := h.Byid[id]; ok {
    return room, nil
  } else {
    return nil, errors.New("Room Not Found")
  }
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
  common.DisableCors(&w)
  switch(r.Method) {
  case http.MethodOptions:
    w.Header().Set("Access-Control-Allow-Methdods","OPTIONS,GET")
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
  common.DisableCors(&w)
  switch(r.Method) {
  case http.MethodOptions:
    w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,PUT")
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

func (h *Hub) SetupRoom(w http.ResponseWriter, r *http.Request) {
  common.DisableCors(&w)
  fmt.Println("Setup Room")
  fmt.Println(r.Method)
  switch(r.Method) {
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
    fmt.Println("Setup Room", v)
    room.SetupGame(v)
    // Enqueue Players World Update in game update channel
    /*for player := v.players {
      room.game.Update <- interface(map[string]interface{
        "id": player.id,
        "type": "move",
        "x": 0,
        "y": 0,
      })
    }*/
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"id": room.ID})
  }
}

func (h *Hub) StartRoom(w http.ResponseWriter, r *http.Request) {
  common.DisableCors(&w)
  switch(r.Method) {
  case http.MethodOptions:
    w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,PUT")
  case http.MethodPut:
    room, err := h.getRoom404(w, r)
    if err != nil {
      fmt.Println(err)
      log.Println(err)
      return
    }
    if (room.Running) {
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

type Query struct {
  Data struct {
    User struct {
      Id int
      Handle string
      Email string
      Guest bool
    }
  }
}

func TokenQuery(token string) (*Query, error){
  url := "http://" + os.Getenv("GRAPHQL_URL") + "/graphql?query=" + url.QueryEscape(`{ user {id handle email guest} }`)
  bearer := "Bearer " + token
  req, err := http.NewRequest("GET", url, nil)
  fmt.Println("url:", url)
  req.Header.Add("Authorization", bearer)
  req.Header.Add("Accept", "application/json")
  client := &http.Client{}
  r, err := client.Do(req)
  if err != nil {
      log.Println("Error on response.\n[ERRO] -", err)
      return nil, err
  }
  var f Query
  err = json.NewDecoder(r.Body).Decode(&f)
  if err != nil {
    log.Println(err)
    return nil, err
  }
  return &f, nil
}

func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
  room, err := h.getRoom404(w, r)
  if err != nil {
    log.Println(err)
    return
  }
  err = common.ParseFormBadRequest(w, r)
  if err != nil {
    log.Println(err)
    return
  }
  token := r.Form.Get("token")
  var id string
  if os.Getenv("NO_AUTH") != "" {
    id = token
  } else {
    f, err := TokenQuery(token)
    if err != nil {
      log.Println(err)
      return
    }
    id = strconv.Itoa(f.Data.User.Id)
  }

  fmt.Println("The id is:", id)
  if room.OkToConnectPlayer(id) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
      fmt.Println("Error", err)
      log.Println(err)
      return
    }
    client := &Client{room: room, ID: id, conn: conn, send: make(chan interface{})}
    client.room.register <- client
    fmt.Println("ServeWS: Registered new client to room")
    go client.WritePump()
    go client.ReadPump()
  }
}
