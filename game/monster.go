package game

import (
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"slices"
)

type Monster struct {
	client.MonsterSchema
	Stats Stats
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
