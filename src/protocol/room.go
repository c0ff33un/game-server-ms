package protocol

import (
  "fmt"
  //"context"
  //"time"

  "github.com/segmentio/ksuid"
  "github.com/coff33un/game-server-ms/src/game"
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

  clients map[*Client]bool `json: "-"`

  register chan *Client `json: "-"`
  unregister chan *Client `json: "-"`
  broadcast chan interface{} `json: "-"`
}

func (r *Room) SetupGame(rows, cols int) {
  r.game = game.NewGame(rows, cols, r.broadcast)
  r.Setup = true
  r.Ready = true
}

func NewRoom(h *Hub) *Room {
  id := ksuid.New().String()
  room := &Room{
    ID: id,
    hub: h,
    clients: make(map[*Client]bool),
    register: make(chan *Client),
    unregister: make(chan *Client),
    broadcast: make(chan interface{}),
  }
  room.hub.register <- room
  return room
}

func (r *Room) StartGame() {
  if (r.Ready && r.Setup) {
    go r.game.Run()
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

func (r *Room) Run() {
  for {
    fmt.Println("Room Run here..")
    select {
    case client := <-r.register:
      r.clients[client] = true
    case client := <-r.unregister:
      if _, ok := r.clients[client]; ok {
        delete(r.clients, client)
        close(client.send)
      }
    case message := <-r.broadcast:
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

