package graph

import (
	"github.com/ahornerr/artifacts/game"
)

type Monster struct {
	Monster *game.Monster
}

func NewMonster(monster *game.Monster) Node {
	return &Monster{Monster: monster}
}

func (m *Monster) Type() Type {
	return TypeMonster
}

func (m *Monster) Children() []Node {
	return nil
}

func (m *Monster) String() string {
	return m.Monster.Name
}

func (m *Monster) MarshalJSON() ([]byte, error) {
	return marshal(m)
}
