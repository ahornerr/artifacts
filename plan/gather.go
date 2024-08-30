package plan

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
)

type Gather struct {
	resource *game.Resource
	dropRate float64
	children []Node
}

func NewGather(resource *game.Resource, desiredItemCode string, child ...Node) *Gather {
	return &Gather{
		resource: resource,
		dropRate: avgDropPerAction(resource.Loot, desiredItemCode),
		children: child,
	}
}

func (a *Gather) Name() string {
	return "Gather " + a.resource.Name
}

func (a *Gather) Weight(char *character.Character) float64 {
	// Certain items have effects that reduce cooldown. Those should be taken into account here
	cooldown := 20.0

	// TODO: Do we need to check everything or just the weapon?
	weapon, ok := char.Equipment["weapon"]
	if ok {
		attack := game.Items.Get(weapon).Stats.Attack[a.resource.Skill]
		fmt.Println(attack)
		// TODO: Reduce cooldown
	}

	return cooldown * 1.0 / a.dropRate // TODO: Double check this math
}

func (a *Gather) Children() []Node {
	// TODO: Is this the best way to handle this?
	return a.children
}

func (a *Gather) Execute(ctx context.Context, char *character.Character) error {
	_, err := char.Gather(ctx)
	return err
}
