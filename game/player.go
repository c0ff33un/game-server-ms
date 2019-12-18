package game

import (
//"fmt"
	"time"
)

type Player struct {
	X       int `json:"x"`
	Y       int `json:"y"`
	Id      string
	Stamina float32 `json:"stamina"`
	Running bool    `json:"running"`
	Dead    bool    `json:"dead"`
	game    *Game
}

func NewPlayer(x, y int, id string, game *Game) *Player {
	return &Player{
		X:    x,
		Y:    y,
		Id:   id,
		game: game,
	}
}

type PlayerMessage struct {
	Message
	Id string `json:"id"`
}

func (p *Player) getPlayerMessage(Type string) PlayerMessage {
	return PlayerMessage{Message{Type}, p.Id}
}

type MoveMessage struct {
	PlayerMessage
	X int `json:"x"`
	Y int `json:"y"`
}

func (p *Player) getMoveMessage() MoveMessage {
	return MoveMessage{
		p.getPlayerMessage("move"),
		p.X,
		p.Y,
	}
}

type WinMessage struct {
	PlayerMessage
	Handle string `json:"handle"`
	ResolveTime int64 
}

func (p *Player) getWinMessage(startTime time.Time) WinMessage {
	return WinMessage{
		PlayerMessage: p.getPlayerMessage("win"), 
		ResolveTime: time.Since(startTime).Milliseconds(),
	}
}

func (p *Player) move(direction string) MoveMessage {
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
		return p.getMoveMessage()
	}
	return MoveMessage{}
}
