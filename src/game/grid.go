type Board struct {
  N int `json:"n"` // grid size
  M int `json:"m"`
  grid []bool // wall or not
  x, y int // exit
}


