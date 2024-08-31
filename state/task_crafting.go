package state

import (
	"context"
	"errors"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"slices"
)

type TaskCraftingArgs struct {
	characters map[string]*character.Character
}

func TaskCrafting(characters map[string]*character.Character) Runner {
	return func(ctx context.Context, char *character.Character) error {
		return Run(ctx, char, TaskCraftingLoop, NewTaskCraftingArgs(characters))
	}
}

func NewTaskCraftingArgs(characters map[string]*character.Character) *TaskCraftingArgs {
	return &TaskCraftingArgs{
		characters: characters,
	}
}

func TaskCraftingLoop(ctx context.Context, char *character.Character, args *TaskCraftingArgs) (State[*TaskCraftingArgs], error) {
	char.PushState("Crafting gear for tasks")
	defer char.PopState()

	toCraft := map[*game.Item]int{}

	for _, otherChar := range args.characters {
		if otherChar.Task == "" {
			continue
		}
		var haveLevelToCraft []*game.Item

		availableUpgrades, _ := otherChar.GetEquipmentUpgrades()
		for _, availableUpgrade := range availableUpgrades {
			if char.GetLevel(availableUpgrade.Crafting.Skill) >= availableUpgrade.Crafting.Level {
				haveLevelToCraft = append(haveLevelToCraft, availableUpgrade)
			}
		}

		attackUpgrades := slices.Clone(haveLevelToCraft)
		resistUpgrades := slices.Clone(haveLevelToCraft)
		monster := game.Monsters.Get(otherChar.Task)

		slices.SortFunc(attackUpgrades, func(a, b *game.Item) int {
			dmg := int(b.Stats.GetDamageAgainst(monster.Stats) - a.Stats.GetDamageAgainst(monster.Stats))
			if dmg != 0 {
				return dmg
			}

			hp := b.Stats.Hp - a.Stats.Hp
			if hp != 0 {
				return hp
			}

			return b.Stats.Haste - b.Stats.Haste
		})

		slices.SortFunc(resistUpgrades, func(a, b *game.Item) int {
			dmg := int(monster.Stats.GetDamageAgainst(a.Stats) - monster.Stats.GetDamageAgainst(b.Stats))
			if dmg != 0 {
				return dmg
			}

			hp := b.Stats.Hp - a.Stats.Hp
			if hp != 0 {
				return hp
			}

			return b.Stats.Haste - b.Stats.Haste
		})

		//possibleUpgradesFor := map[*game.Item][]*game.Item{}
		for slot, itemCode := range otherChar.Equipment {
			equipped := game.Items.Get(itemCode)
			itemType := slot
			if itemType == "ring1" || itemType == "ring2" {
				itemType = "ring"
			}

			var attackUpgradesForSlot []*game.Item
			var resistUpgradesForSlot []*game.Item

			for _, item := range attackUpgrades {
				if item.Type == itemType {
					attackUpgradesForSlot = append(attackUpgradesForSlot, item)
				}
			}

			for _, item := range resistUpgrades {
				if item.Type == itemType {
					resistUpgradesForSlot = append(resistUpgradesForSlot, item)
				}
			}

			attackIdx := slices.Index(attackUpgradesForSlot, equipped)
			resistIdx := slices.Index(resistUpgradesForSlot, equipped)

			// TODO: What if the currentIdx is -1?
			upgrades := attackUpgradesForSlot
			currentIdx := attackIdx
			if resistIdx > -1 && resistIdx < currentIdx {
				upgrades = resistUpgradesForSlot
				currentIdx = resistIdx
			}

			if currentIdx == 0 {
				continue
			}

			//possibleUpgradesFor[equipped] = upgrades[:currentIdx]
			toCraft[upgrades[0]]++
		}

		//_ = possibleUpgradesFor
	}

	for item, quantity := range toCraft {
		have := char.Bank()[item.Code] + char.Inventory[item.Code]
		if have >= quantity {
			delete(toCraft, item)
			continue
		}

		toCraft[item] = quantity - have
	}

	// Craft items we already have materials to craft
	for item, quantity := range toCraft {
		totalCraftable, _ := getNumCanCraft(item.Crafting, char.Inventory, char.Bank())
		if totalCraftable == 0 {
			continue
		}

		craftingArgs := NewCraftArgs(item.Code, func(c *character.Character, args *CraftingArgs) bool {
			return args.Made >= quantity
		})
		err := Run(ctx, char, CraftingLoop, craftingArgs)
		if err != nil {
			return nil, err
		}

		toCraft[item] = quantity - craftingArgs.Made
		if toCraft[item] <= 0 {
			delete(toCraft, item)
		}
	}

	craftingMaterials := map[*game.Item]int{}
	for item, quantity := range toCraft {
		for reqItem, reqQuantity := range item.Crafting.Items {
			craftingMaterials[reqItem] += quantity * reqQuantity
		}
	}

	for item, quantity := range craftingMaterials {
		err := CollectItems(item, quantity)(ctx, char)
		if err != nil {
			var fightErr FightErr
			if errors.As(err, &fightErr) {
				continue
			}
			return nil, err
		}
	}

	for item, quantity := range toCraft {
		runner := Craft(item.Code, func(c *character.Character, args *CraftingArgs) bool {
			return args.Made >= quantity
		})
		err := runner(ctx, char)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}
