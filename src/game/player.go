package game

import(
  //"fmt"
)

type Player struct {
  x int `json:"x"`
  y int `json:"y"`
  stamina float32 `json:"stamina"`
  running bool `json:"running"`
  dead bool `json:"dead"`
  Game *Game
}

func between(a, b, x int) bool {
  return a <= x && x <= b
}

func validPosition(x, y int, game *Game) bool {
  cols, rows := game.Board.Cols, game.Board.Rows
  if between(0, cols - 1, x) && between(0, rows - 1 , y) {
    wall := game.Board.Grid[cols * y + x]
    return !wall
  }
  return false
}

func NewPlayer(x, y int, game *Game) *Player {
  return &Player{
    x : x,
    y : y,
    Game : game,
  }
}

func (p *Player) move(direction string) interface{} {
  x, y := p.x, p.y
  switch direction {
  case "up":
    y -= 1
  case "down":
    y += 1
  case "left":
    x -= 1
  case "right":
    x += 1
  default:
  }
  f := make(map[string]interface{})
  if validPosition(x, y, p.Game) {
    f["x"] = x
    f["y"] = y
    p.x = x
    p.y = y
    return interface{}(f)
  }
  return nil
}
