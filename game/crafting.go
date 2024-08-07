package game

import (
	"fmt"
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
	Items map[*Item]int
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

	skill, err := craft.Skill.AsCraftSchemaSkill0()
	if err != nil {
		return nil, err
	}

	craftQuantity, err := craft.Quantity.AsCraftSchemaQuantity0()
	if err != nil {
		return nil, err
	}

	craftLevel, err := craft.Level.AsCraftSchemaLevel0()
	if err != nil {
		return nil, err
	}

	if craftQuantity > 1 {
		fmt.Println("This makes more than one item")
	}

	items := map[*Item]int{}
	for _, cr := range *craft.Items {
		// This is kind of hacky but we're going to replace the keys with actual item references later
		items[&Item{Code: cr.Code}] = cr.Quantity
	}

	return &Crafting{
		Skill:    string(skill),
		Quantity: craftQuantity,
		Level:    craftLevel,
		Items:    items,
	}, nil
}
