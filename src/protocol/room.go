package protocol

import (
  "fmt"

  "github.com/segmentio/ksuid"
  "github.com/coff33un/game-server-ms/src/game"
)

type Room struct {
  ID string `json: "id"`
  hub *Hub
  game *game.Game
  clients map[*Client]bool
  register chan *Client
  unregister chan *Client
  broadcast chan []byte

}

func NewRoom(h *Hub) *Room {
  id := ksuid.New().String()
  room := &Room{
    ID: id,
    hub: h,
    clients: make(map[*Client]bool),
    register: make(chan *Client),
    unregister: make(chan *Client),
    broadcast: make(chan []byte),
  }
  room.hub.register <- room
  return room
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
