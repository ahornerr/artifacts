package commands

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/command"
	"github.com/ahornerr/artifacts/game"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"math"
)

var (
	SequenceEquipBestEquipmentForTask = command.SequenceFunc(func(ctx context.Context, char *character.Character) []command.Command {
		monster := game.Monsters.Get(char.Task)
		return []command.Command{EquipBestEquipmentForMonster(monster)}
	})
)

func EquipBestEquipmentForMonster(monster *game.Monster) command.Command {
	return command.Sequence(GetAndEquipUpgrades(monster.Stats))
}

func EquipBestEquipmentForResource(resource *game.Resource) command.Command {
	skillStats := game.Stats{
		Attack: nil,
		Resistance: map[string]int{
			string(resource.Skill): -math.MaxInt32,
		},
	}

	return GetAndEquipUpgrades(skillStats)
}

func Unequip(slot string) command.Command {
	return command.NewSimple(
		fmt.Sprintf("Unequipping %s", slot),
		func(ctx context.Context, char *character.Character) error {
			return char.Unequip(ctx, client.UnequipSchemaSlot(slot))
		},
	)
}

func Equip(slot, itemCode string) command.Command {
	return command.NewSimple(
		fmt.Sprintf("Equipping %s", slot),
		func(ctx context.Context, char *character.Character) error {
			return char.Equip(ctx, client.EquipSchemaSlot(slot), itemCode)
		},
	)
}

func GetAndEquipUpgrades(stats game.Stats) command.Command {
	return command.SequenceFunc(func(ctx context.Context, char *character.Character) []command.Command {
		upgrades := char.GetEquipmentUpgradesInBank(stats)
		equipment := char.GetEquippedItems()

		var sequence []command.Command
		if len(upgrades) > 0 {
			sequence = append(sequence, MoveToBank)

			for slot, upgradeItemCode := range upgrades {
				var slotSequence []command.Command

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

			sequence = append(sequence, DepositAll)
		}
		return sequence
	})
}
