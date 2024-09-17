package state

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/gofiber/fiber/v3/log"
	"math"
	"reflect"
)

func itemsForTraining(char *character.Character, skill string) []*game.Item {
	charLevel := char.GetLevel(skill)
	if charLevel >= maxLevel {
		return nil
	}

	return game.Items.ForSkill(skill, charLevel)
}

func RoleCrafter(crafterWants chan game.ItemQuantity) Runner {
	return func(ctx context.Context, char *character.Character) error {
		for {
			err := crafter(ctx, char, crafterWants)
			if err != nil {
				log.Errorf("%s %v", char.Name, err)
			}
		}
	}
}

func crafter(ctx context.Context, char *character.Character, crafterWants chan game.ItemQuantity) error {
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

	// TODO: Craft better items if possible

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

func doCraftingTraining(ctx context.Context, char *character.Character, crafterWants chan game.ItemQuantity) (bool, error) {
	levels := []int{5, 10, 15, 20, 25, 30, 35}
	skills := []string{"weaponcrafting", "gearcrafting", "jewelrycrafting"}

	for _, level := range levels {
		for _, skill := range skills {
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
			numHarvesters := 4 // TODO: Make this configurable?

			var totalCost int
			for item, quantity := range lowestItem.Crafting.Items {
				// Account for items in the bank
				totalQuantity := quantity * quantityToMakeAtATime
				remainingQuantity := totalQuantity - char.Bank()[item.Code]
				if remainingQuantity <= 0 {
					continue
				}
				totalCost += game.Cost(item.Code) * remainingQuantity
			}

			avgHarvesterCost := totalCost / numHarvesters

			for item, quantity := range lowestItem.Crafting.Items {
				// Account for items in the bank
				totalQuantity := quantity * quantityToMakeAtATime
				remainingQuantity := totalQuantity - char.Bank()[item.Code]
				if remainingQuantity <= 0 {
					continue
				}

				itemCost := game.Cost(item.Code)

				for remainingQuantity > 0 {
					thisQuantity := min(remainingQuantity, avgHarvesterCost/itemCost)
					crafterWants <- game.ItemQuantity{
						Item:     item,
						Quantity: thisQuantity,
					}
					remainingQuantity -= thisQuantity
				}
			}

			startLevel := char.GetLevel(skill)
			startXp := char.GetXP(skill)

			args := NewMakeXArgs(lowestItem.Code, quantityToMakeAtATime, true, func(c *character.Character, args *MakeXArgs) bool {
				if char.GetLevel(skill) > startLevel {
					// We leveled up
					return true
				}

				if args.Made >= quantityToMakeAtATime {
					// We made one batch
					return true
				}

				return false
			})

			err := Run(ctx, char, MakeXLoop, args)
			if err != nil {
				return true, err
			}

			if startXp == char.GetXP(skill) {
				return true, fmt.Errorf("not getting any %s XP from making %s", skill, lowestItem.Name)
			}

			return true, nil
		}
	}

	return false, nil
}
