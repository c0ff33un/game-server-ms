package game

import(
  "fmt"
)

type Player struct {
  x int `json:"x"`
  y int `json:"y"`
  stamina float32 `json:"stamina"`
  running bool `json:"running"`
  dead bool `json:"dead"`
  game *Game
}

func between(a, b, x int) bool {
  return a <= x && x <= b
}

func validPosition(x, y, rows, cols int) bool {
  return between(0, cols - 1, x) && between(0, rows - 1, y)
}

func NewPlayer(x, y int, game *Game) *Player {
  return &Player{
    x : x,
    y : y,
    game : game,
  }
}

func (p *Player) move(direction string) interface{} {
  x, y := p.x, p.y
  rows, cols := p.game.board.Rows, p.game.board.Cols
  fmt.Println("direction is", direction)
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
  if validPosition(x, y, rows, cols) {
    f["x"] = x
    f["y"] = y
    p.x = x
    p.y = y
    return interface{}(f)
  }
  return nil
}
