package state

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"log"
)

type CollectItemsArgs struct {
	Item                              *game.Item
	Quantity                          int
	includeBank                       bool
	includeAllInventoriesAndEquipment bool
	characters                        map[string]*character.Character
}

// CollectItems collects the materials required to craft the item in the desired quantity
func CollectItems(itemCode string, quantity int, includeBank bool, includeAllInventoriesAndEquipment bool, characters map[string]*character.Character) Runner {
	item := game.Items.Get(itemCode)
	return func(ctx context.Context, char *character.Character) error {
		err := Run(ctx, char, CollectItemsLoop, NewCollectItemsArgs(item, quantity, includeBank, includeAllInventoriesAndEquipment, characters))
		if err != nil {
			return CollectErr{
				Item: item,
				Err:  err,
			}
		}
		return nil
	}
}

func NewCollectItemsArgs(item *game.Item, quantity int, includeBank, includeAllInventoriesAndEquipment bool, characters map[string]*character.Character) *CollectItemsArgs {
	return &CollectItemsArgs{
		Item:                              item,
		Quantity:                          quantity,
		includeBank:                       includeBank,
		includeAllInventoriesAndEquipment: includeAllInventoriesAndEquipment,
		characters:                        characters,
	}
}

func CollectItemsLoop(ctx context.Context, char *character.Character, args *CollectItemsArgs) (State[*CollectItemsArgs], error) {
	item := args.Item
	quantity := args.Quantity

	have := 0
	if args.includeBank {
		have = char.Bank()[item.Code]
	}
	if args.includeAllInventoriesAndEquipment {
		for _, char := range args.characters {
			have += char.Inventory[item.Code]

			for _, equipItemCode := range char.Equipment {
				if equipItemCode == item.Code {
					have++
				}
			}
		}
	} else {
		have += char.Inventory[item.Code]
	}
	need := quantity - have
	if need <= 0 {
		return nil, nil
	}

	char.PushState("Collecting %d %s", need, item.Name)
	defer char.PopState()

	// Item comes from a resource
	resources := game.Resources.ResourcesForItem(item)
	if len(resources) > 0 {
		if len(resources) > 1 {
			// TODO: If this is possible, figure out which one is easier to get
			log.Println("Found multiple resources for item", item.Code)
		}
		resource := resources[0]
		rate := resource.Loot[item].Rate
		if rate != 1 {
			char.PopState()
			char.PushState("Harvesting %d %s from %s (1/%d)", need, item.Name, resource.Name, rate)
		}
		runner := Harvest(resource.Code, func(c *character.Character, _ *HarvestArgs) bool {
			have := char.Inventory[item.Code]
			if args.includeBank {
				have += char.Bank()[item.Code]
			}
			return have >= quantity
		})
		err := runner(ctx, char)

		if err != nil {
			return nil, err
		}
	}

	monsters := game.Monsters.MonstersForItem(item)
	if len(monsters) > 0 {
		if len(monsters) > 1 {
			// TODO: If this is possible, figure out which one is easier to get
			log.Println("Found multiple monsters for item", item.Code)
		}
		monster := monsters[0]
		rate := monster.Loot[item].Rate
		if rate != 1 {
			char.PopState()
			char.PushState("Fighting %s for %d %s (1/%d)", monster.Name, need, item.Name, rate)
		}
		fightArgs := NewFightArgs(monster.Code, func(c *character.Character, _ *FightArgs) bool {
			have := char.Inventory[item.Code]
			if args.includeBank {
				have += char.Bank()[item.Code]
			}
			return have >= quantity
		}, nil)
		err := Run(ctx, char, FightLoop, fightArgs)
		if err != nil {
			return nil, err
		}
		if fightArgs.NumFights() == 0 {
			// Didn't attempt to fight, must be unwinnable
			return nil, FightErr{Monster: monster}
		}
	}

	if item.Crafting != nil {
		for reqItem, reqQuantity := range item.Crafting.Items {
			err := CollectItems(reqItem.Code, reqQuantity*need, true, false, args.characters)(ctx, char)
			if err != nil {
				return nil, err
			}
		}

		runner := Craft(item.Code, need, false, nil)
		err := runner(ctx, char)
		if err != nil {
			return nil, err
		}
	}

	if item.SubType == "task" {
		taskItemArgs := func(c *character.Character, _ *TaskItemArgs) bool {
			have := c.Inventory[item.Code]
			if args.includeBank {
				have += c.Bank()[item.Code]
			}
			return have >= quantity
		}
		runner := TaskItem(item.Code, quantity, taskItemArgs)
		err := runner(ctx, char)
		if err != nil {
			return nil, err
		}
	}

	return CollectItemsLoop, nil
}

type CollectErr struct {
	Item *game.Item
	Err  error
}

func (e CollectErr) Error() string {
	return fmt.Sprintf("cannot collect item %s: %v", e.Item.Name, e.Err)
}

func (e CollectErr) Unwrap() error {
	return e.Err
}

type FightErr struct {
	Monster *game.Monster
}

func (e FightErr) Error() string {
	return fmt.Sprintf("cannot win fight against %s", e.Monster.Name)
}
