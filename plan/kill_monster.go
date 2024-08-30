package plan

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
)

type Monsters struct {
	item     *game.Item
	quantity int
	monster  *game.Monster
	children []Node
}

func NewMonsters(item *game.Item, quantity int, monster *game.Monster, children ...Node) *Monsters {
	var options []Node

	// TODO: Figure out looping

	// This will always be the very last child in each option
	moveAndFight := NewMove(game.Maps.GetMonsters(monster.Code), NewFight(monster, item.Code, children...))

	options = append(options,
		moveAndFight,
		NewMove(game.Maps.GetBanks(), NewEquipment(monster.Stats, moveAndFight)),
	)

	return &Monsters{
		item:     item,
		quantity: quantity,
		monster:  monster,
		children: options,
	}
}

func (h *Monsters) Name() string {
	return "Monsters " + h.monster.Code
}

func (h *Monsters) Weight(char *character.Character) float64 {
	return 0
}

func (h *Monsters) Children() []Node {
	// TODO: Is this the best way to handle this?
	return h.children
}

func (h *Monsters) Execute(ctx context.Context, char *character.Character) error {
	//_, err := char.Gather(ctx)
	//return err

	return nil // TODO
}
