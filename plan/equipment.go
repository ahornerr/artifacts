package plan

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
)

type Equipment struct {
	target   game.Stats
	children []Node
}

func NewEquipment(target game.Stats, children ...Node) *Equipment {
	return &Equipment{
		target:   target,
		children: children,
	}
}

func (h *Equipment) Name() string {
	return "Get better equipment"
}

func (h *Equipment) Weight(char *character.Character) float64 {
	return 0 // TODO
}

func (h *Equipment) Children() []Node {
	// TODO: Is this the best way to handle this?
	return h.children
}

func (h *Equipment) Execute(ctx context.Context, char *character.Character) error {
	// TODO: Unequip, deposit, withdraw, equip as many items as we need
	//  May also want to deposit full inventory
	return nil
}
