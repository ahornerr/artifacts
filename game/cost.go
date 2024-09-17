package game

import (
	"math"
)

func Cost(itemCode string) int {
	item := Items.Get(itemCode)

	if item.Crafting != nil {
		return fromCrafting(item.Crafting)
	}

	resources := Resources.ResourcesForItem(item)
	if len(resources) > 0 {
		return fromResources(item, resources)
	}

	monsters := Monsters.MonstersForItem(item)
	if len(monsters) > 0 {
		return fromMonsters(item, monsters)
	}

	if item.SubType == "task" {
		// TODO
		// Some arbitrarily high number for now
		return 30000
	}

	if item.Type == "currency" {
		return 5000
	}

	if item.Code == "wooden_stick" {
		// Weird edge case
		return 5000
	}

	panic("How did we get here?")
}

func fromCrafting(crafting *Crafting) int {
	var cost int
	for material, quantity := range crafting.Items {
		cost += Cost(material.Code) * quantity
	}
	return cost
}

func fromResources(item *Item, resources []*Resource) int {
	lowestResourceCost := math.MaxInt32
	for _, resource := range resources {
		drop := resource.Loot[item]
		avgDropQuantity := float64(drop.MinQuantity+drop.MaxQuantity) / 2.0
		cost := avgDropQuantity * float64(drop.Rate)

		// Some multiplier based on roughly how long it takes to harvest the resource
		cost *= float64(resource.Level)

		if int(cost) < lowestResourceCost {
			lowestResourceCost = int(cost)
		}
	}
	return lowestResourceCost
}

func fromMonsters(item *Item, monsters []*Monster) int {
	lowestMonsterCost := math.MaxInt32
	for _, monster := range monsters {
		drop := monster.Loot[item]
		avgDropQuantity := float64(drop.MinQuantity+drop.MaxQuantity) / 2.0

		// Expected kills, via https://www.reddit.com/r/2007scape/comments/gpz0dw/drop_rates_and_probabilities_very_long_post/
		confidence := 0.75
		pDrop := 1.0 / float64(drop.Rate)
		pNoDrop := 1.0 - pDrop
		expectedKills := math.Log(1.0-confidence) / math.Log(pNoDrop)

		cost := avgDropQuantity * expectedKills

		// Some multiplier based on roughly how long it takes to kill the monster
		cost *= float64(monster.Level)

		if int(cost) < lowestMonsterCost {
			lowestMonsterCost = int(cost)
		}
	}
	return lowestMonsterCost
}
