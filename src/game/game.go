package game

import (
  "fmt"
)

type Board struct {
  Rows int `json:"rows"` // grid size
  Cols int `json:"cols"`
  grid []bool // wall or not
  x int
  y int // exit
}

type Game struct {
  board *Board
  broadcast chan interface{}
  players map[string]*Player

  Update chan interface{}
}

func NewGame(rows, cols int, broadcast chan interface{}) *Game {
  b := &Board{
    Rows: rows,
    Cols: cols,
  }
  return &Game{
    board : b,
    Update : make(chan interface{}),
    broadcast : broadcast,
  }
}

func (g *Game) updateWorld(f interface{}) {
  m := f.(map[string]interface{})
  switch m["type"].(string) {
  case "move":
    id, direction := m["id"].(string), m["direction"].(string)
    result := g.players[id].move(direction)
    if result != nil {
      result["id"] = id
      g.broadcast <- interface{}(result)
    }
    break
  }

}

func (g *Game) Run() {
  for player := range players {

  }
  for {
    fmt.Println("Game Run here..")
    select {
    case message := <-g.Update:
      g.updateWorld(message)
    }
  }
}
