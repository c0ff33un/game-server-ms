package game

import (
//"fmt"
)

type Player struct {
	X       int `json:"x"`
	Y       int `json:"y"`
	ID      string
	Stamina float32 `json:"stamina"`
	Running bool    `json:"running"`
	Dead    bool    `json:"dead"`
	game    *Game
}

func NewPlayer(x, y int, id string, game *Game) *Player {
	return &Player{
		X:    x,
		Y:    y,
		ID:   id,
		game: game,
	}
}

func (p *Player) MoveMessage() map[string]interface{} {
	return map[string]interface{}{
		"type": "move",
		"x":    p.X,
		"y":    p.Y,
		"id":   p.ID,
	}
}

func (p *Player) move(direction string) map[string]interface{} {
	x, y := p.X, p.Y
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
	if p.game.validPosition(x, y) {
		p.X = x
		p.Y = y
		return p.MoveMessage()
	}
	return nil
}
