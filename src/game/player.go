package game

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

func (p *Player) move(direction string) map[string]interface{} {
  x, y := p.x, p.y
  rows, cols := p.game.board.Rows, p.game.board.Cols
  switch direction {
  case "up":
    x -= 1
  case "down":
    x += 1
  case "left":
    y -= 1
  case "right":
    y += 1
  }
  var f map[string]interface{}
  if validPosition(x, y, rows, cols) {
    f["x"] = x
    f["y"] = y
    return f
  }
  return nil
}
