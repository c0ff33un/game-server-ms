package protocol

import (
  "fmt"
  //"context"
  //"time"

  "github.com/segmentio/ksuid"
  "github.com/coff33un/game-server-ms/src/game"
  "github.com/coff33un/game-server-ms/src/common"
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

func NewRoom(h *Hub) *Room {
  id := ksuid.New().String()
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
  fmt.Println(players)
  r.game = game.NewGame(v, r.broadcast, players)
  r.Setup = true
  r.Ready = true
}

func (r *Room) StartGame() {
  if (r.Ready && r.Setup) {
    r.broadcast <- interface{}(map[string]interface{}{
      "type": "grid",
      "grid": r.game.Board.Grid,
    })
    r.Running = true
    go r.game.Run()
  } else {
    fmt.Println("Room Not Setup")
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
  if c, ok := r.byid[id]; ok {
    _, ok := r.clients[c]
    return ok
  }
  return false
}

func (r *Room) OkToConnectPlayer(id string) bool {
  if (r.playerConnected(id)) {
    return false // there can't be two connections of the same user
  }
  if _, ok := r.byid[id]; ok {
    return true
  }
  return len(r.byid) < 3
}

func (r *Room) Run() {
  defer func() {
    fmt.Println("Room Died holy shit")
  }()
  for {
    fmt.Println("Room Run here..")
    select {
    case client := <-r.register:
      fmt.Println("Register Client:", client.ID)
      r.clients[client] = true
      for c := range r.clients {
        fmt.Println("Registered Client:", c.ID)
      }
      r.byid[client.ID] = client
      /* Breaks Channel
      client.room.broadcast <- interface{}(map[string]interface{}{
        "type": "connect",
        "id" : client.ID,
      })*/
    case client := <-r.unregister: // websocket closed
      fmt.Println("Unregistering Client:", client.ID)
      if _, ok := r.clients[client]; ok {
        delete(r.clients, client)
        close(client.send)
      }
    case message := <-r.broadcast:
      fmt.Println("Broadcast message")
      common.PrintJSON(message.(map[string]interface{}))
      fmt.Println("Registered Clients:", len(r.clients))
      for c := range r.clients {
        fmt.Println("Registered Client:", c.ID)
      }
      n := len(r.broadcast)
      fmt.Println("Messages to broadcast:", n)
      for client := range r.clients {
        fmt.Println("Client ID to send message:", client.ID)
        select {
        case client.send <- message:
        default:
          fmt.Println("Closing Client:", client.ID)
          close(client.send)
          delete(r.clients, client)
        }
      }
    }
  }
}

