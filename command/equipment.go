package command

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"math"
)

var (
	SequenceEquipBestEquipmentForTask = NewSequence(func(ctx context.Context, char *character.Character) []Command {
		return []Command{EquipBestEquipmentForMonster(char.Task)}
	})
)

func EquipBestEquipmentForMonster(monsterCode string) Command {
	return NewSequence(func(ctx context.Context, char *character.Character) []Command {
		monster := game.Monsters.Get(monsterCode)
		return []Command{GetAndEquipUpgrades(monster.Stats)}
	})
}

func EquipBestEquipmentForResource(resourceCode string) Command {
	return NewSequence(func(ctx context.Context, char *character.Character) []Command {
		skill := string(game.Resources.Get(resourceCode).Skill)

		skillStats := game.Stats{
			Attack: nil,
			Resistance: map[string]int{
				skill: -math.MaxInt32,
			},
		}

		return []Command{GetAndEquipUpgrades(skillStats)}
	})
}

func Unequip(slot string) Command {
	return NewSimple(fmt.Sprintf("Unequipping %s", slot), func(ctx context.Context, char *character.Character) error {
		return char.Unequip(ctx, client.UnequipSchemaSlot(slot))
	})
}

func Equip(slot, itemCode string) Command {
	return NewSimple(fmt.Sprintf("Equipping %s", slot), func(ctx context.Context, char *character.Character) error {
		return char.Equip(ctx, client.EquipSchemaSlot(slot), itemCode)
	})
}

func GetAndEquipUpgrades(stats game.Stats) Command {
	return NewSequence(func(ctx context.Context, char *character.Character) []Command {
		upgrades := char.GetEquipmentUpgradesInBank(stats)
		equipment := char.GetEquippedItems()

		var sequence []Command
		if len(upgrades) > 0 {
			sequence = append(sequence, MoveToBank)

			for slot, upgradeItemCode := range upgrades {
				var slotSequence []Command

				currentItem := equipment[slot]

				if currentItem != "" {
					slotSequence = append(
						slotSequence,
						Unequip(slot),
						Deposit(currentItem, 1),
					)
				}

				slotSequence = append(
					slotSequence,
					Withdraw(upgradeItemCode, 1),
					Equip(slot, upgradeItemCode),
				)

				sequence = append(sequence, slotSequence...)
			}
		}
		return sequence
	})
}
