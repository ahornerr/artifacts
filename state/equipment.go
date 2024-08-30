package state

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

func EquipBestEquipment(ctx context.Context, char *character.Character, stats *game.Stats) error {
	upgrades := char.GetEquipmentUpgradesInBank(stats)
	if len(upgrades) > 0 {
		char.PushState("Upgrading equipment")
		defer char.PopState()

		err := MoveToBankAndDepositAll(ctx, char)
		if err != nil {
			return err
		}

		for slot, upgradeItemCode := range upgrades {
			previousItem := char.Equipment[slot]
			if previousItem != "" {
				err = char.Unequip(ctx, client.UnequipSchemaSlot(slot))
				if err != nil {
					return err
				}

				_, err = char.DepositBank(ctx, previousItem, 1)
				if err != nil {
					return err
				}
			}

			_, err = char.WithdrawBank(ctx, upgradeItemCode, 1)
			if err != nil {
				return err
			}

			err = char.Equip(ctx, client.EquipSchemaSlot(slot), upgradeItemCode)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
