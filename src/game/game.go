package game

import (
  "fmt"
)

type Board struct {
  Rows int `json:"rows"` // grid size
  Cols int `json:"cols"`
  Grid []bool // wall or not
}

type SetupGameMessage struct {
  Rows int
  Cols int
  Grid []bool
  Exit struct {
    X int
    Y int
  }
  Begin struct {
    X int
    Y int
  }
  Players []struct {
    X int
    Y int
    Id string
  }
}

type Game struct {
  Board *Board
  players map[string]*Player
  Exit struct {
    X int
    Y int
  }
  broadcast chan interface{}
  Update chan interface{}
}

func NewGame(v SetupGameMessage, broadcast chan interface{}, players []string) *Game {
  b := &Board{
    Rows: v.Rows,
    Cols: v.Cols,
    Grid: v.Grid,
  }
  game := &Game{
    Board : b,
    Update : make(chan interface{}),
    Exit: v.Exit,
    broadcast : broadcast,
    players : make (map[string]*Player),
  }
  x, y := v.Begin.Y, v.Begin.X
  for _, player := range players {
    fmt.Println("create player: ", player)
    game.players[player] = NewPlayer(x, y, game)
    /* To-do
    game.Update <- interface{}(map[string]interface{}{
      "type": "move",
      "x": x,
      "y": y,
    })*/
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
