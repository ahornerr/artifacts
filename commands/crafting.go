package commands

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/command"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/stopper"
	"math"
)

func NewCraftingLoop(itemCode string, stop stopper.Stopper) command.Command {
	item := game.Items.Get(itemCode)

	// Override the original stopper to stop when we're out of crafting materials
	newStop := stopper.NewStopper(func(ctx context.Context, char *character.Character, quantity int) (bool, error) {
		shouldStop, err := stop.ShouldStop(ctx, char, quantity)
		if shouldStop || err != nil {
			return shouldStop, err
		}

		totalCraftable, _ := getNumCanCraft(*item.Crafting, char.Inventory, char.Bank())
		return totalCraftable == 0, nil
	})

	// This sequence is called repeatedly for each subsequent craft action.
	// This is important because we need to recalculate the number of items being crafted and check the stopper.
	return command.NewLoop(
		newStop,
		func(iteration int) string {
			return fmt.Sprintf("Crafting %s (made %d)", item.Name, iteration)
		},
		BankForCrafting(item),
		MoveToClosestWorkshop(item),
		CraftOne(item),
	)
}

// Calculate the maximum number of items we can craft given the resources in our inventory and bank
func getNumCanCraft(crafting game.Crafting, inventory map[string]int, bank map[string]int) (totalCraftable int, inventoryCraftable int) {
	totalCraftable = math.MaxInt32
	inventoryCraftable = math.MaxInt32

	for craftingItem, quantity := range crafting.Items {
		haveInv := inventory[craftingItem.Code]
		haveBank := bank[craftingItem.Code]

		totalCraftable = min(totalCraftable, (haveInv+haveBank)/quantity)
		if totalCraftable == 0 {
			return 0, 0 // We don't have enough of this item to even make one
		}

		inventoryCraftable = min(inventoryCraftable, haveInv/quantity)
	}

	return
}

func BankForCrafting(item *game.Item) command.Command {
	return command.NewFunctional(
		func() string {
			return fmt.Sprintf("Withdrawing materials for %s", item.Name)
		},
		func(ctx context.Context, char *character.Character) ([]command.Command, error) {
			if item.Crafting == nil {
				return nil, fmt.Errorf("item cannot be crafted")
			}

			// Withdraw as many items from the bank as we can, then craft the resources one by one, checking the stopper between
			// This works around an issue of calculating number of items to craft on each iteration.
			// totalCraftable cannot be 0 because we check it in the stopper.
			totalCraftable, inventoryCraftable := getNumCanCraft(*item.Crafting, char.Inventory, char.Bank())

			switch {
			case inventoryCraftable == totalCraftable:
				// Nothing in the bank worth withdrawing
				return nil, nil
			case inventoryCraftable > 0:
				// If we have some items to craft in our inventory already, just do it
				return nil, nil
			}

			// Figure out what we need to withdraw from the bank to get a full inventory,
			// unless the bank doesn't have a full inventory's worth we just withdraw what we can.
			maxInventorySizeCanCraft := char.MaxInventoryItems() / item.Crafting.InventoryRequired()

			withdrawMultiplier := min(maxInventorySizeCanCraft, totalCraftable)

			itemsToWithdraw := map[string]int{}
			for craftingItem, quantity := range item.Crafting.Items {
				itemsToWithdraw[craftingItem.Code] = withdrawMultiplier * quantity
			}

			return []command.Command{
				MoveToBank,
				DepositAll,
				WithdrawItems(itemsToWithdraw),
			}, nil
		},
	)
}

func MoveToClosestWorkshop(item *game.Item) command.Command {
	return command.NewSimple(
		fmt.Sprintf("Moving to closest workshop for crafting %s", item.Name),
		func(ctx context.Context, char *character.Character) error {
			return char.MoveClosest(ctx, game.Maps.GetWorkshops(item.Crafting.Skill))
		},
	)
}

func CraftOne(item *game.Item) command.Command {
	return command.NewSimple(
		fmt.Sprintf("Crafting %s", item.Name),
		func(ctx context.Context, char *character.Character) error {
			_, err := char.Craft(ctx, item.Code, 1)
			return err
		},
	)
}
