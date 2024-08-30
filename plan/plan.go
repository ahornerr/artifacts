package plan

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
)

type Plan struct {
	Options []Node
}

func FromItem(item *game.Item, quantity int) *Plan {
	return &Plan{
		Options: OptionsForItem(item, quantity),
	}
}

type Node interface {
	Name() string
	Children() []Node
	Weight(char *character.Character) float64
	Execute(ctx context.Context, char *character.Character) error
}

func NewTask(item *game.Item, quantity int, children ...Node) Node {
	return nil
}

func OptionsForItem(item *game.Item, quantity int, children ...Node) []Node {
	options := []Node{
		// TODO: We can probably combine these two for convenience sake
		NewMove(game.Maps.GetBanks(), NewWithdraw(item, quantity, children...)),
	}

	if item.Crafting != nil {
		options = append(options, NewCrafting(item, quantity, item.Crafting, children...))
	}

	for _, resource := range game.Resources.ResourcesForItem(item) {
		options = append(options, NewHarvest(item, quantity, resource, children...))
	}

	for _, monster := range game.Monsters.MonstersForItem(item) {
		options = append(options, NewMonsters(item, quantity, monster, children...))
	}

	if item.SubType == "task" {
		options = append(options, NewTask(item, quantity, children...))
	}

	return options
}

func avgDropPerAction(drops []game.Drop, itemCode string) float64 {
	// Weight calculation here is pretty simple.
	// We want some quantity of items from the resource, and we know its drop rate.
	// From here we can calculate the average number of drops per harvest.
	// Weight is roughly correlated with time, so lower weights are more desirable.
	// Take the quantity and divide by the average drop per harvest (lower drop rate increases weight).
	// This also means that higher quantity increases rate too.
	var avgDrop float64
	for _, drop := range drops {
		if drop.Item.Code == itemCode {
			avgDropQuantity := float64(drop.MinQuantity+drop.MaxQuantity) / 2.0
			avgDrop = avgDropQuantity * 1 / float64(drop.Rate)
			break
		}
	}

	return avgDrop

	//if avgDropPerHarvest == 0 {
	//	return 0
	//}
	//
	//// Don't divide by zero if the resource doesn't drop it (for some reason)
	//weight := math.MaxFloat64
	//if avgDropPerHarvest > 0 {
	//	//weight = float64(quantity) / avgDropPerHarvest
	//	weight = 1.0 / avgDropPerHarvest
	//}
	//
	//return weight
}
