package command

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/stopper"
	"math"
)

type Crafting struct {
	item       *game.Item
	stop       stopper.Stopper
	numCrafted int
}

func NewCraftingLoop(itemCode string, stop stopper.Stopper) Command {
	return &Crafting{
		item: game.Items.Get(itemCode),
		stop: stop,
	}
}

func (c *Crafting) Description() string {
	return fmt.Sprintf("Crafting %s (made %d)", c.item.Name, c.numCrafted)
}

// Execute is called recursively for each subsequent craft action.
// This is important because we need to recalculate the number of items being crafted and check the stopper.
func (c *Crafting) Execute(ctx context.Context, char *character.Character) ([]Command, error) {
	stop, err := c.stop.ShouldStop(ctx, char, c.numCrafted)
	if stop || err != nil {
		return nil, err
	}

	if c.item.Crafting == nil {
		return nil, fmt.Errorf("item cannot be crafted")
	}

	// Withdraw as many items from the bank as we can, then craft the resources one by one, checking the stopper between
	// This works around an issue of calculating number of items to craft on each iteration.
	totalCraftable, inventoryCraftable := getNumCanCraft(*c.item.Crafting, char.InventoryAsMap(), char.Bank())
	if totalCraftable == 0 {
		// No materials left to craft
		return nil, nil
	}

	// Figure out what we need to withdraw from the bank to get a full inventory.
	// Only move to the bank if we require materials or if it's the first item being crafted,
	// And we have more stuff in the bank that we could use for crafting.
	maxInventorySizeCanCraft := char.MaxInventoryItems() / c.item.Crafting.InventoryRequired()
	moreInBank := totalCraftable > inventoryCraftable
	canFitMoreInInv := inventoryCraftable < maxInventorySizeCanCraft
	mustWithdraw := inventoryCraftable == 0
	firstExecution := c.numCrafted < 1

	var sequence []Command

	if moreInBank && canFitMoreInInv && (mustWithdraw || firstExecution) {
		withdrawMultiplier := min(maxInventorySizeCanCraft, totalCraftable)

		itemsToWithdraw := map[string]int{}
		for craftingItem, quantity := range c.item.Crafting.Items {
			itemsToWithdraw[craftingItem.Code] = withdrawMultiplier * quantity
		}

		// TODO: This should probably be its own command/sequence that short-circuits if it doesn't need to perform the actions
		//  This allows us to flatten the graph a good bit
		sequence = append(sequence,
			MoveToBank,
			DepositAll,
			WithdrawItems(itemsToWithdraw))
	}

	sequence = append(sequence,
		MoveToClosestWorkshop(c.item),
		CraftOne(c.item),
		c,
	)

	defer func() {
		c.numCrafted++
	}()

	return sequence, nil
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

func MoveToClosestWorkshop(item *game.Item) Command {
	return NewSimple(fmt.Sprintf("Moving to closest workshop for crafting %s", item.Name),
		func(ctx context.Context, char *character.Character) error {
			return char.MoveClosest(ctx, game.Maps.GetWorkshops(item.Crafting.Skill))
		},
	)
}

func CraftOne(item *game.Item) Command {
	return NewSimple(fmt.Sprintf("Crafting %s", item.Name),
		func(ctx context.Context, char *character.Character) error {
			_, err := char.Craft(ctx, item.Code, 1)
			return err
		},
	)
}
