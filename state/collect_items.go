package state

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
)

type CollectItemsArgs struct {
	Item                              *game.Item
	Quantity                          int
	includeAllInventoriesAndEquipment bool
	characters                        map[string]*character.Character
}

// CollectItems collects the materials required to craft the item in the desired quantity
func CollectItems(item *game.Item, quantity int, includeAllInventoriesAndEquipment bool, characters map[string]*character.Character) Runner {
	return func(ctx context.Context, char *character.Character) error {
		err := Run(ctx, char, CollectItemsLoop, NewCollectItemsArgs(item, quantity, includeAllInventoriesAndEquipment, characters))
		if err != nil {
			return CollectErr{
				Item: item,
				Err:  err,
			}
		}
		return nil
	}
}

func NewCollectItemsArgs(item *game.Item, quantity int, includeAllInventoriesAndEquipment bool, characters map[string]*character.Character) *CollectItemsArgs {
	return &CollectItemsArgs{
		Item:                              item,
		Quantity:                          quantity,
		includeAllInventoriesAndEquipment: includeAllInventoriesAndEquipment,
		characters:                        characters,
	}
}

func CollectItemsLoop(ctx context.Context, char *character.Character, args *CollectItemsArgs) (State[*CollectItemsArgs], error) {
	item := args.Item
	quantity := args.Quantity

	have := char.Bank()[item.Code]
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
			fmt.Println("Found multiple resources for item", item.Code)
		}
		resource := resources[0]
		rate := resource.Loot[item].Rate
		if rate != 1 {
			char.PushState("Drop rate 1/%d", rate)
		}
		runner := Harvest(resource.Code, func(c *character.Character, args *HarvestArgs) bool {
			return char.Bank()[item.Code]+char.Inventory[item.Code] >= quantity
		})
		err := runner(ctx, char)
		if rate != 1 {
			char.PopState()
		}
		if err != nil {
			return nil, err
		}
	}

	monsters := game.Monsters.MonstersForItem(item)
	if len(monsters) > 0 {
		if len(monsters) > 1 {
			// TODO: If this is possible, figure out which one is easier to get
			fmt.Println("Found multiple monsters for item", item.Code)
		}
		monster := monsters[0]
		rate := monster.Loot[item].Rate
		if rate != 1 {
			char.PushState("Drop rate 1/%d", rate)
		}
		fightArgs := NewFightArgs(monster.Code, func(c *character.Character, args *FightArgs) bool {
			// Bail out if we lose the first 3 fights
			if args.NumFights() == 3 && args.NumLosses() == 3 {
				return true
			}
			return char.Bank()[item.Code]+char.Inventory[item.Code] >= quantity
		})
		err := Run(ctx, char, FightLoop, fightArgs)
		if rate != 1 {
			char.PopState()
		}
		if err != nil {
			return nil, err
		}
		if fightArgs.NumFights() == 3 && fightArgs.NumLosses() == 3 {
			return nil, FightErr{Monster: monster}
		}
	}

	if item.Crafting != nil {
		for reqItem, reqQuantity := range item.Crafting.Items {
			err := CollectItems(reqItem, reqQuantity*need, false, args.characters)(ctx, char)
			if err != nil {
				return nil, err
			}
		}

		runner := Craft(item.Code, need, nil)
		err := runner(ctx, char)
		if err != nil {
			return nil, err
		}
	}

	// TODO
	//if item.SubType == "task" {
	//
	//}

	return nil, nil
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
