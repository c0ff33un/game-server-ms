package game

import (
	"encoding/json"
	"io"
	"log"
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
}

func NewGame(body io.ReadCloser, players []string) (*Game, error) {
	var v SetupGameMessage
	err := json.NewDecoder(body).Decode(&v)
	if err != nil {
		return nil, err
	}
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
	}
	x, y := v.Begin.Y, v.Begin.X
	for _, player := range players {
		game.Players[player] = NewPlayer(x, y, player, game)
	}
	return game, nil
}

type BoardMessage struct {
	Message
	Grid []bool `json:"grid"`
}

func (g *Game) getBoardMessage() interface{} {
	return BoardMessage{
		Message{Type: "start"},
		g.Board.Grid,
	}
}

func (g *Game) GetStatus() []interface{} {
	var res []interface{}
	res = append(res, g.getBoardMessage())
	for _, player := range g.Players {
		res = append(res, player.getMoveMessage())
	}
	return res
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

type GameSimulator interface {
	simulateGame(g *Game, message interface{}) []interface{}
}

type ClassicSimulator struct{}

type Message struct {
	Type string `json:"type"`
}

type ClientMoveMessage struct {
	PlayerMessage
	Direction string `json:"direction"`
}

func (s *ClassicSimulator) simulateGame(g *Game, m interface{}) []interface{} {
	var res []interface{}
	switch v := m.(type) {
	case ClientMoveMessage:
		id, direction := v.Id, v.Direction
		player, ok := g.Players[id]
		if !ok {
			log.Println("player not found")
			return nil
		}
		move := player.move(direction)
		if move.Type != "" {
			x, y := move.X, move.Y
			res = append(res, move)
			if x == g.Exit.Y && y == g.Exit.X {
				log.Println("User", id, "won")
				res = append(res, player.getWinMessage())
			}
		}
	}
	return res
}

func ClassicGameRunner(body io.ReadCloser, broadcast chan interface{}, players []string) (*GameRunner, error) {
	game, err := NewGame(body, players)
	if err != nil {
		return nil, err
	}
	return &GameRunner{
		Game:          game,
		GameSimulator: &ClassicSimulator{},
		broadcast:     broadcast,
		Update:        make(chan interface{}),
		Quit:          make(chan bool),
	}, nil
}

type GameRunner struct {
	GameSimulator
	*Game

	broadcast chan interface{}
	Update    chan interface{}
	Quit      chan bool
}

func (g *GameRunner) Run() {
	for _, f := range g.GetStatus() {
		g.broadcast <- f
	}
	for {
		log.Println("Game Run here..")
		select {
		case <-g.Quit:
			return
		case message := <-g.Update:
			result := g.simulateGame(g.Game, message)
			for _, message := range result {
				g.broadcast <- message
			}
		}
	}
}
