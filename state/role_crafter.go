package state

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/gofiber/fiber/v3/log"
	"math"
	"reflect"
	"slices"
)

var (
	levelMilestones = []int{5, 10, 15, 20, 25, 30, 35}
	craftingSkills  = []string{"weaponcrafting", "gearcrafting", "jewelrycrafting"}
)

func itemsForTraining(char *character.Character, skill string) []*game.Item {
	charLevel := char.GetLevel(skill)
	if charLevel >= maxLevel {
		return nil
	}

	return game.Items.ForTrainingCraftingSkill(skill, charLevel)
}

func RoleCrafter(characters map[string]*character.Character, crafterWants chan game.ItemQuantity) Runner {
	return func(ctx context.Context, char *character.Character) error {
		for {
			err := crafter(ctx, char, characters, crafterWants)
			if err != nil {
				log.Errorf("%s %v", char.Name, err)
			}
		}
	}
}

func crafter(ctx context.Context, char *character.Character, characters map[string]*character.Character, crafterWants chan game.ItemQuantity) error {
	did, err := doMonsterEvent(ctx, char)
	if err != nil {
		return err
	} else if did {
		return nil
	}

	did, err = doTask(ctx, char)
	if err != nil {
		return err
	} else if did {
		return nil
	}

	did, err = doCraftBetterItems(ctx, char, characters, crafterWants)
	if err != nil {
		return err
	} else if did {
		return nil
	}

	did, err = doCraftingTraining(ctx, char, crafterWants)
	if err != nil {
		return err
	} else if did {
		return nil
	}

	return nil
}

func doMonsterEvent(ctx context.Context, char *character.Character) (bool, error) {
	for monsterCode := range game.Events.Events()["monster"] {
		monster := game.Monsters.Get(monsterCode)
		bestEquipment := char.GetBestOwnedEquipment(monster.Stats)
		if bestEquipment.TurnsToKillMonster > bestEquipment.TurnsToKillPlayer {
			// Can't win the fight
			continue
		}

		// TODO: Choose which monster instead of the first one we can kill
		return true, Fight(monsterCode, nil, nil)(ctx, char)
	}

	return false, nil
}

func doTask(ctx context.Context, char *character.Character) (bool, error) {
	// TODO: Support other task types
	if char.TaskType == "monsters" && char.Task != "" {
		lastEvents := game.Events.Events()

		monster := game.Monsters.Get(char.Task)
		bestEquipment := char.GetBestOwnedEquipment(monster.Stats)
		// Make sure we can win the fight
		if bestEquipment.TurnsToKillMonster < bestEquipment.TurnsToKillPlayer {
			args := NewTaskArgs(func(c *character.Character, args *TaskArgs) bool {
				return !reflect.DeepEqual(lastEvents, game.Events.Events())
			})

			return true, Run(ctx, char, TaskLoop, args)
		}
	}

	return false, nil
}

func doCraftBetterItems(ctx context.Context, char *character.Character, characters map[string]*character.Character, crafterWants chan game.ItemQuantity) (bool, error) {
	totalItemQuantity := func(itemCode string) int {
		quantity := char.Bank()[itemCode]
		for _, c := range characters {
			quantity += c.Inventory[itemCode]
			for _, equipped := range c.Equipment {
				if equipped == itemCode {
					quantity++
				}
			}
		}
		return quantity
	}

	//var itemsToMake []game.ItemQuantity
	for _, level := range levelMilestones {
		charactersNeedingThisLevelItem := 0

		for _, c := range characters {
			combatLevel := c.GetLevel("combat")

			// If the player is > 5 levels higher than this item then it's not worth making
			if level > combatLevel {
				continue
			}
			if level+5 < combatLevel {
				continue
			}
			charactersNeedingThisLevelItem++
		}

		if charactersNeedingThisLevelItem == 0 {
			continue
		}

		items := game.Items.ForLevel(level)

		// Sort items lowest cost first
		slices.SortFunc(items, func(a, b *game.Item) int {
			return game.Cost(a.Code) - game.Cost(b.Code)
		})

		for _, item := range items {
			// TODO: Ignore tools for now
			if item.SubType == "tool" {
				continue
			}
			// TODO: Select only equippable items somehow
			if item.Type == "resource" {
				continue
			}

			desiredQuantity := charactersNeedingThisLevelItem
			if item.Type == "ring" {
				desiredQuantity *= 2
			}

			remainingQuantity := desiredQuantity - totalItemQuantity(item.Code)
			if remainingQuantity <= 0 {
				continue
			}

			return true, distributeAndMake(ctx, char, crafterWants, item, remainingQuantity, false)
		}

	}

	// Make the first item in the list
	return false, nil
}

func doCraftingTraining(ctx context.Context, char *character.Character, crafterWants chan game.ItemQuantity) (bool, error) {
	for _, level := range levelMilestones {
		for _, skill := range craftingSkills {
			if char.GetLevel(skill) >= level {
				continue
			}

			// TODO: Take into account materials we have in inventory/bank
			lowestCost := math.MaxInt32
			var lowestItem *game.Item
			for _, item := range itemsForTraining(char, skill) {
				cost := game.Cost(item.Code)
				if cost < lowestCost {
					lowestCost = cost
					lowestItem = item
				}
			}

			if lowestItem == nil {
				// This should never happen
				return false, fmt.Errorf("unable to find item for training %s", skill)
			}

			quantityToMakeAtATime := 5

			startXp := char.GetXP(skill)
			err := distributeAndMake(ctx, char, crafterWants, lowestItem, quantityToMakeAtATime, true)
			if err != nil {
				return true, err
			}

			if startXp == char.GetXP(skill) {
				return true, fmt.Errorf("not getting any %s XP from making %s", skill, lowestItem.Name)
			}

			return true, err
		}
	}

	return false, nil
}

func distributeAndMake(ctx context.Context, char *character.Character, crafterWants chan game.ItemQuantity, item *game.Item, quantity int, recycle bool) error {
	numHarvesters := 4 // TODO: Make this configurable?

	var totalCost int
	for reqItem, reqQuantity := range item.Crafting.Items {
		// Account for items in the bank and inventory
		totalQuantity := quantity * reqQuantity
		remainingQuantity := totalQuantity - char.Bank()[reqItem.Code]
		remainingQuantity -= char.Inventory[reqItem.Code]
		if remainingQuantity <= 0 {
			continue
		}
		totalCost += game.Cost(reqItem.Code) * remainingQuantity
	}

	avgHarvesterCost := totalCost / numHarvesters

	for reqItem, reqQuantity := range item.Crafting.Items {
		// Account for items in the bank
		totalQuantity := reqQuantity * quantity
		remainingQuantity := totalQuantity - char.Bank()[reqItem.Code]
		remainingQuantity -= char.Inventory[reqItem.Code]
		if remainingQuantity <= 0 {
			continue
		}

		itemCost := game.Cost(reqItem.Code)

		for remainingQuantity > 0 {
			thisQuantity := max(1, min(remainingQuantity, avgHarvesterCost/itemCost))
			crafterWants <- game.ItemQuantity{
				Item:     reqItem,
				Quantity: thisQuantity,
			}
			remainingQuantity -= thisQuantity
		}
	}

	args := NewMakeXArgs(item.Code, quantity, recycle, func(c *character.Character, args *MakeXArgs) bool {
		if args.Made >= quantity {
			// We made one batch
			return true
		}

		return false
	})

	return Run(ctx, char, MakeXLoop, args)
}
