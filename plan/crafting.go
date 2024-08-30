package plan

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
)

type Crafting struct {
	item     *game.Item
	quantity int
	crafting *game.Crafting
	children []Node
}

func NewCrafting(item *game.Item, quantity int, crafting *game.Crafting, children ...Node) *Crafting {
	var options []Node

	//options = append(options, NewCrafting(item, quantity, item.Crafting))
	//for craftingItem, craftingQuantity := range item.Crafting.Items {
	//	children = append(OptionsForItem(craftingItem, quantity*craftingQuantity))
	//}

	moveAndCraft := NewMove(game.Maps.GetWorkshops(crafting.Skill), NewCraft(item, quantity, children...))

	// TODO: Figure out looping

	//options = append(options,
	//	moveAndCraft,
	//)

	var getItem Node
	for craftingItem, craftingQuantity := range crafting.Items {
		if getItem == nil {
			getItem = NewGetItem(craftingItem, craftingQuantity*quantity, moveAndCraft)
		} else {
			getItem = NewGetItem(craftingItem, craftingQuantity*quantity, getItem)
		}
	}

	options = append(options, getItem)

	return &Crafting{
		item:     item,
		quantity: quantity,
		crafting: crafting,
		children: options,
	}
}

func (c *Crafting) Name() string {
	return "Crafting " + c.item.Name
}

func (c *Crafting) Weight(char *character.Character) float64 {
	// Crafting doesn't cost anything on its own
	return 0
}

func (c *Crafting) Children() []Node {
	// TODO: Is this the best way to handle this?
	return c.children
}

func (c *Crafting) Execute(ctx context.Context, char *character.Character) error {
	//_, err := char.Gather(ctx)
	//return err

	return nil // TODO
}
