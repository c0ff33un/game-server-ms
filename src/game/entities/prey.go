package entities

type Prey struct {
  X float32 `json:"x"`
  Y float32 `json:"y"`
  Stamina float32 `json:"stamina"`
  Running bool `json:"running"`
  dead bool `json:"dead"`
}
