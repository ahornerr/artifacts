package game

import (
	"slices"
)

type Monster struct {
	Code    string
	Name    string
	Stats   *Stats
	Level   int
	MaxGold int
	MinGold int
	Loot    map[*Item]Drop
}

func (m Monster) String() string {
	return m.Name
}

func (m Monster) GetWeaknesses() []string {
	var weaknesses []string
	for element := range m.Stats.Resistance {
		weaknesses = append(weaknesses, element)
	}

	slices.SortFunc(weaknesses, func(a, b string) int {
		return m.Stats.Resistance[a] - m.Stats.Resistance[b]
	})

	return weaknesses
}
