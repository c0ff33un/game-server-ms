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
	players map[string]*Player
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
		players: make(map[string]*Player),

		broadcast: broadcast,
		Update:    make(chan map[string]interface{}),
		Quit:      make(chan bool),
	}
	x, y := v.Begin.Y, v.Begin.X
	for _, player := range players {
		game.players[player] = NewPlayer(x, y, game)
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
			result["id"] = id
			result["type"] = "move"
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
