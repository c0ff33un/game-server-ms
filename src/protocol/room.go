package protocol

import (
  "fmt"
  //"context"
  //"time"
  "math/rand"
  "strconv"

  //"github.com/segmentio/ksuid"
  "github.com/coff33un/game-server-ms/src/game"
  //"github.com/coff33un/game-server-ms/src/common"
  //"go.mongodb.org/mongo-driver/bson"
  //"go.mongodb.org/mongo-driver/mongo/options"
  //"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Room struct {
  ID string `json: "id"`
  hub *Hub `json: "-"`
  game *game.Game `json: "-"`

  Ready bool `json: "ready"`
  Setup bool `json: "setup"`
  Running bool `json: "runnning"`

  clients map[*Client]bool `json: "-"`
  byid map[string]*Client

  register chan *Client `json: "-"`
  unregister chan *Client `json: "-"`
  broadcast chan interface{} `json: "-"`
}

func genId() string {
  return strconv.Itoa(rand.Intn(10000))
}

func NewRoom(h *Hub) *Room {
  id := genId()
  for h.Byid[id] != nil {
    id = genId()
  }
  fmt.Println("New RoomID:", id)
  room := &Room{
    ID: id,
    hub: h,
    clients: make(map[*Client]bool),
    byid: make(map[string]*Client),

    register: make(chan *Client),
    unregister: make(chan *Client),
    broadcast: make(chan interface{}),
  }
  room.hub.register <- room
  return room
}

func (r *Room) SetupGame(v game.SetupGameMessage) {
  fmt.Println("Setting up game.")

  var players []string
  for player := range r.clients {
    players = append(players, player.ID)
  }
  r.game = game.NewGame(v, r.broadcast, players)
  r.Setup = true
  r.Ready = true
  fmt.Println("Room Setup")
}

func (r *Room) StartGame() {
  if (r.Ready && r.Setup) {
    for player := range r.clients {
      x, y := r.game.Begin.X, r.game.Begin.Y
      r.broadcast <- interface{}(map[string]interface{}{
        "type": "move",
        "id": player.ID,
        "x": y,
        "y": x,
      })
    }
    r.broadcast <- interface{}(map[string]interface{}{
      "type": "setup",
      "grid": r.game.Board.Grid,
    })
    r.Running = true
    go r.game.Run()
  } else {
    fmt.Println("Room Not Setup")
  }
}

func (r *Room) StopGame() {
  if (!r.Running) {
    fmt.Println("Cannot Stop A Room that is Not Running")
  } else {
    fmt.Println("Try Stop Game")
    r.Running = false
    r.Ready = false
    r.Setup = false
    close(r.game.Update)
    fmt.Println("Gracefully Stopped Game")
  }
}

func (r *Room) writeRoomInfo() {
  /*db := r.hub.db
  ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
  err := db.Ping(ctx, readpref.Primary())
  collection := client.Database("taurus").Collection("rooms")

  ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
  res, err := collection.InsertOne(ctx, bson.D())*/
}

func (r *Room) closeRoom() {
  r.hub.unregister <- r
  close(r.register)
  close(r.unregister)
  close(r.broadcast)
  for client := range r.clients {
    close(client.send)
    delete(r.clients, client)
  }
}

func (r *Room) playerConnected(id string) bool {
  if c, ok := r.byid[id]; ok { // check registered
    _, ok := r.clients[c] // check WS Connection
    return ok
  }
  return false
}

func (r *Room) OkToConnectPlayer(id string) bool {
  if (r.playerConnected(id)) {
    fmt.Println("User already Connected to Room")
    return false // there can't be two connections of the same user
  }
  if _, ok := r.byid[id]; ok {
    fmt.Println("User is registrered and Disconnected")
    return true
  }
  fmt.Println("Room len", len(r.byid))
  return len(r.byid) < 3 && !r.Running
}

func (r *Room) Run() {
  defer func() {
    fmt.Println("Room Died holy shit")
  }()
  for {
    fmt.Println("Room Run here..")
    select {
    case client := <-r.register:
      r.clients[client] = true
      r.byid[client.ID] = client
      case client := <-r.unregister: // websocket closed
      fmt.Println("Unregistering Client:", client.ID)
      if _, ok := r.clients[client]; ok {
        delete(r.clients, client)
        close(client.send)
      }
    case message := <-r.broadcast:
      f := message.(map[string]interface{})
      fmt.Println(f["type"])
      if f["type"] == "win" {
        f["handle"] = r.byid[f["id"].(string)].Handle
        r.StopGame()
      }
      for client := range r.clients {
        select {
        case client.send <- message:
        default:
          close(client.send)
          delete(r.clients, client)
        }
      }
    }
  }
}

