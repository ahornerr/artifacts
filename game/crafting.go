package game

import (
	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

type Crafting struct {
	// Quantity of items crafted
	Quantity int

	// Skill required to craft the item
	Skill string

	// Level The skill level required to craft the item
	Level int

	// Items List of items required to craft the item
	Items map[*Item]int `json:"-"`
}

func (c Crafting) InventoryRequired() int {
	inventoryRequired := 0
	for _, quantity := range c.Items {
		inventoryRequired += quantity
	}
	return inventoryRequired
}

func craftingFromSchema(itemSchemaCraft *client.ItemSchema_Craft) (*Crafting, error) {
	if itemSchemaCraft == nil {
		return nil, nil
	}

	craft, err := itemSchemaCraft.AsCraftSchema()
	if err != nil {
		return nil, err
	}

	craftingItems := map[*Item]int{}
	for _, cr := range *craft.Items {
		// This is kind of hacky, but we're going to replace the keys with actual item references later
		craftingItems[&Item{Code: cr.Code}] = cr.Quantity
	}

	return &Crafting{
		Skill:    string(*craft.Skill),
		Quantity: *craft.Quantity,
		Level:    *craft.Level,
		Items:    craftingItems,
	}, nil
}
