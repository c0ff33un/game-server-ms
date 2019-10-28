package game

import (
  "fmt"
)

type Board struct {
  Rows int `json:"rows"` // grid size
  Cols int `json:"cols"`
  Grid []bool // wall or not
  x int
  y int // exit
}

type Game struct {
  Board *Board
  players map[string]*Player

  broadcast chan interface{}
  Update chan interface{}
}

func NewGame(rows, cols int, broadcast chan interface{}, players []string) *Game {
  var grid []bool
  for i := 0; i < rows * cols; i++ {
    grid = append(grid, false)
  }
  b := &Board{
    Rows: rows,
    Cols: cols,
    Grid: grid,
  }
  game := &Game{
    Board : b,
    Update : make(chan interface{}),
    broadcast : broadcast,
    players : make (map[string]*Player),
  }

  for _, player := range players {
    fmt.Println("create player: ", player)
    game.players[player] = NewPlayer(rows / 2, cols / 2, game)
  }
  return game
}

func (g *Game) updateWorld(f interface{}) {
  m := f.(map[string]interface{})
  switch m["type"].(string) {
  case "move":
    id, direction := m["id"].(string), m["direction"].(string)
    result := g.players[id].move(direction)
    if result != nil {
      json := result.(map[string]interface{})
      json["id"] = id
      json["type"] = "move"
      fmt.Println("Valid Move");
      g.broadcast <- interface{}(json)
    }
  }
}

func (g *Game) Run() {
  for {
    fmt.Println("Game Run here..")
    select {
    case message := <-g.Update:
      g.updateWorld(message)
    }
  }
}
