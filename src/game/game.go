package game

import (
	"fmt"
)

type Board struct {
	Rows int    `json:"rows"` // grid size
	Cols int    `json:"cols"`
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
		X  int
		Y  int
		Id string
	}
}

type Game struct {
	Board   *Board
	Players map[string]*Player
	Exit    struct {
		X int
		Y int
	}
	Begin struct {
		X int
		Y int
	}
	broadcast chan map[string]interface{}
	Update    chan map[string]interface{}
	Quit      chan bool
}

func NewGame(v SetupGameMessage, broadcast chan map[string]interface{}, players []string) *Game {
	b := &Board{
		Rows: v.Rows,
		Cols: v.Cols,
		Grid: v.Grid,
	}
	game := &Game{
		Board:   b,
		Exit:    v.Exit,
		Begin:   v.Begin,
		Players: make(map[string]*Player),

		broadcast: broadcast,
		Update:    make(chan map[string]interface{}),
		Quit:      make(chan bool),
	}
	x, y := v.Begin.Y, v.Begin.X
	for _, player := range players {
		game.Players[player] = NewPlayer(x, y, player, game)
	}
	return game
}

func (g *Game) BoardMessage() map[string]interface{} {
	return map[string]interface{}{
		"type": "start",
		"grid": g.Board.Grid,
	}
}

func between(a, b, x int) bool {
	return a <= x && x <= b
}

func (game *Game) validPosition(x, y int) bool {
	cols, rows := game.Board.Cols, game.Board.Rows
	if between(0, cols-1, x) && between(0, rows-1, y) {
		wall := game.Board.Grid[cols*y+x]
		return !wall
	}
	return false
}

func (g *Game) updateWorld(m map[string]interface{}) {
	switch m["type"].(string) {
	case "move":
		id, direction := m["id"].(string), m["direction"].(string)
		result := g.Players[id].move(direction)
		if result != nil {
			x, y := result["x"], result["y"]
			g.broadcast <- result
			if x == g.Exit.Y && y == g.Exit.X {
				fmt.Println("User", id, "won")
				g.broadcast <- map[string]interface{}{
					"type": "win",
					"id":   id,
				}
			}
		}
	}
}

func (g *Game) Run() {
	for _, player := range g.Players {
		g.broadcast <- player.MoveMessage()
	}
	g.broadcast <- g.BoardMessage()
	for {
		fmt.Println("Game Run here..")
		select {
		case <-g.Quit:
			return
		case message := <-g.Update:
			g.updateWorld(message)
		}
	}
}
