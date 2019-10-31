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
  Begin struct {
    X int
    Y int
  }
  broadcast chan interface{}
  Update chan interface{}
  Close chan bool
}

func NewGame(v SetupGameMessage, broadcast chan interface{}, players []string) *Game {
  b := &Board{
    Rows: v.Rows,
    Cols: v.Cols,
    Grid: v.Grid,
  }
  game := &Game{
    Board : b,
    Exit: v.Exit,
    Begin: v.Begin,
    broadcast : broadcast,
    Close : make(chan bool),
    players : make (map[string]*Player),
    Update : make(chan interface{}),
  }
  x, y := v.Begin.Y, v.Begin.X
  for _, player := range players {
    fmt.Println(player)
    game.players[player] = NewPlayer(x, y, game)
    // To-do
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
      x, y := json["x"], json["y"]
      g.broadcast <- interface{}(json)
      if x == g.Exit.Y && y == g.Exit.X {
        fmt.Println("User", id, "won")
        g.broadcast <- interface{}(map[string]interface{}{
          "type": "win",
          "id": id,
        })
        g.Close <- true
      }
    }
  }
}

func (g *Game) Run() {
  for {
    fmt.Println("Game Run here..")
    select {
    case message := <-g.Update:
      if message != nil {
        g.updateWorld(message)
      }
    }
  }
}
