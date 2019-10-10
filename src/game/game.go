package game

import (
  "github.com/coff33un/game-server-ms/src/game/entities"
)

type Board struct {
  N int `json:"n"` // grid size
  M int `json:"m"`
  grid []bool // wall or not
  x int
  y int // exit
}

type Game struct {
  board *Board
  players []*entities.Prey
  mosnter *entities.Monster
}


func (g *Game) Run() {
  //...
}
