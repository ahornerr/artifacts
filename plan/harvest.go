package plan

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"math"
)

type Harvest struct {
	item     *game.Item
	quantity int
	resource *game.Resource
	children []Node
}

func NewHarvest(item *game.Item, quantity int, resource *game.Resource, children ...Node) *Harvest {
	var options []Node

	// TODO: Figure out looping

	// This will always be the very last child in each option
	moveAndGather := NewMove(game.Maps.GetResources(resource.Code), NewGather(resource, item.Code, children...))

	// Resource steps:
	// - Deposit inventory in bank if space is needed
	// - Move to resource
	// - Harvest resource
	// - Bank if full
	// - Continue if we need more quantity

	// Move to resource -> gather
	// Move to bank -> get better equipment -> move to resource -> gather

	resourceStats := game.Stats{
		Attack: nil,
		Resistance: map[string]int{
			resource.Skill: -math.MaxInt32,
		},
	}

	options = append(options,
		moveAndGather,
		NewMove(game.Maps.GetBanks(), NewEquipment(resourceStats, moveAndGather)),
	)

	return &Harvest{
		item:     item,
		quantity: quantity,
		resource: resource,
		children: options,
	}
}

func (h *Harvest) Name() string {
	return "Harvest " + h.resource.Name
}

func (h *Harvest) Weight(char *character.Character) float64 {
	dropRate := avgDropPerAction(h.resource.Loot, h.item.Code)

	// Certain items have effects that reduce cooldown. Those should be taken into account here
	cooldown := 20.0

	// TODO: Do we need to check everything or just the weapon?
	weapon, ok := char.Equipment["weapon"]
	if ok {
		attack := game.Items.Get(weapon).Stats.Attack[h.resource.Skill]
		fmt.Println(attack)
		// TODO: Reduce cooldown
	}

	// TODO: Double check this math
	return cooldown * 1.0 / dropRate * float64(h.quantity)
}

func (h *Harvest) Children() []Node {
	// TODO: Is this the best way to handle this?
	return h.children
}

func (h *Harvest) Execute(ctx context.Context, char *character.Character) error {
	//_, err := char.Gather(ctx)
	//return err

	return nil // TODO
}
