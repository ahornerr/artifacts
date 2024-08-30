package plan

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
)

type Fight struct {
	monster  *game.Monster
	dropRate float64
	children []Node
}

func NewFight(monster *game.Monster, desiredItemCode string, child ...Node) *Fight {
	return &Fight{
		monster:  monster,
		dropRate: avgDropPerAction(monster.Loot, desiredItemCode),
		children: child,
	}
}

func (f *Fight) Name() string {
	return "Fight " + f.monster.Name
}

func (f *Fight) Weight(char *character.Character) float64 {
	// Certain items have effects that reduce cooldown. Those should be taken into account here
	cooldown := 20.0

	// TODO: Calculate cooldown based on fight scenario

	//weapon, ok := char.GetEquippedItems()["weapon"]
	//if ok {
	//	attack := game.Items.Get(weapon).Stats.Attack[f.resource.Skill]
	//	fmt.Println(attack)
	//	// TODO: Reduce cooldown
	//}

	return cooldown * 1.0 / f.dropRate // TODO: Double check this math
}

func (f *Fight) Children() []Node {
	// TODO: Is this the best way to handle this?
	return f.children
}

func (f *Fight) Execute(ctx context.Context, char *character.Character) error {
	_, err := char.Fight(ctx)
	return err
}
