package graph

import (
	"github.com/ahornerr/artifacts/game"
)

type Item struct {
	Item     *game.Item
	Quantity int
	//Source   Node
	children []Node
}

func NewItem(item *game.Item, quantity int) *Item {
	var sources []Node

	// Acquired via harvesting
	for _, resource := range game.Resources.ResourcesForItem(item) {
		sources = append(sources, NewResource(resource, item))
	}

	// Acquired via monsters
	monsters := game.Monsters.MonstersForItem(item)
	for _, monster := range monsters {
		sources = append(sources, NewMonster(monster))
	}

	// Acquired via crafting
	if item.Crafting != nil {
		//sources = append(sources, NewCrafting(item.Crafting))
		for craftingItem, craftingQuantity := range item.Crafting.Items {
			sources = append(sources, NewItem(craftingItem, quantity*craftingQuantity))
		}

	}

	// Acquired via tasking
	if item.SubType == "task" {
		sources = append(sources, NewTask())
	}

	//if len(sources) > 1 {
	//	sources = []Node{NewOption(sources...)}
	//}

	return &Item{
		Item:     item,
		Quantity: quantity,
		children: sources,
	}
}

func (i *Item) Type() Type {
	return TypeItem
}

func (i *Item) Children() []Node {
	return i.children
}

func (i *Item) String() string {
	//return fmt.Sprintf("%dx %s", i.Quantity, i.Item.Name)
	return i.Item.Name
}

func (i *Item) MarshalJSON() ([]byte, error) {
	return marshal(i)
}
