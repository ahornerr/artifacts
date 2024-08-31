package state

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/httperror"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

func EquipBestEquipment(ctx context.Context, char *character.Character, targetStats *game.Stats) error {
	bestEquipment := char.GetBestOwnedEquipment(targetStats)

	upgradesInInventory := map[string]*game.Item{}
	upgradesInBank := map[string]*game.Item{}
	var slotsToUnequip []string

	for slot, item := range bestEquipment.Equipment {
		if char.Equipment[slot] == item.Code {
			continue
		}
		if char.Equipment[slot] != "" {
			slotsToUnequip = append(slotsToUnequip, slot)
		}
		if char.Inventory[item.Code] > 0 {
			upgradesInInventory[slot] = item
		} else {
			upgradesInBank[slot] = item
		}
	}

	if len(upgradesInInventory) == 0 && len(upgradesInBank) == 0 {
		return nil
	}

	char.PushState("Upgrading equipment")
	defer char.PopState()

	bankBecauseInventoryFull := len(slotsToUnequip) > char.MaxInventoryItems()-char.InventoryCount()
	needToBank := len(upgradesInBank) > 0 || bankBecauseInventoryFull
	if needToBank {
		err := MoveToBankAndDepositAll(ctx, char)
		if err != nil {
			return err
		}
	}

	for _, slot := range slotsToUnequip {
		err := char.Unequip(ctx, client.UnequipSchemaSlot(slot))
		if err != nil {
			return err
		}
	}

	for slot, item := range upgradesInBank {
		err := Withdraw(ctx, char, item.Code, 1)
		if err != nil {
			if httperror.ErrIsBankItemNotFound(err) {
				return EquipBestEquipment(ctx, char, targetStats)
			}
			return err
		}

		err = char.Equip(ctx, client.EquipSchemaSlot(slot), item.Code)
		if err != nil {
			return err
		}
	}

	for slot, item := range upgradesInInventory {
		if needToBank {
			err := Withdraw(ctx, char, item.Code, 1)
			if err != nil {
				if httperror.ErrIsBankItemNotFound(err) {
					return EquipBestEquipment(ctx, char, targetStats)
				}
				return err
			}
		}

		err := char.Equip(ctx, client.EquipSchemaSlot(slot), item.Code)
		if err != nil {
			return err
		}
	}

	if needToBank {
		err := DepositAll(ctx, char)
		if err != nil {
			return err
		}
	}

	return nil
}
