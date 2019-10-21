package game

import (

  "github.com/coff33un/game-server-ms/src/game/entities"
)

type Board struct {
  rows int `json:"rows"` // grid size
  cols int `json:"cols"`
  grid []bool // wall or not
  x int
  y int // exit
}

type Result struct {

}

type Game struct {
  board *Board
  players []*entities.Prey
  monster *entities.Monster
}

func NewGame(rows, cols int) *Game {
  b := &Board{
    rows: rows,
    cols: cols,
  }
  return &Game{
    board : b,
  }
}

func (g *Game) Run() {
  //...
}
