package state

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/httperror"
	"log"
	"math"
)

type CraftingArgs struct {
	Item         *game.Item
	Quantity     int
	BankWhenDone bool

	Made    int
	Xp      int
	Crafted map[string]int

	stop func(*character.Character, *CraftingArgs) bool
}

func Craft(itemCode string, quantity int, bankWhenDone bool, stop func(*character.Character, *CraftingArgs) bool) Runner {
	return func(ctx context.Context, char *character.Character) error {
		return Run(ctx, char, CraftingLoop, NewCraftArgs(itemCode, quantity, bankWhenDone, stop))
	}
}

func NewCraftArgs(itemCode string, quantity int, bankWhenDone bool, stop func(*character.Character, *CraftingArgs) bool) *CraftingArgs {
	return &CraftingArgs{
		Item:         game.Items.Get(itemCode),
		Quantity:     quantity,
		BankWhenDone: bankWhenDone,
		Crafted:      map[string]int{},
		stop:         stop,
	}
}

func CraftingLoop(ctx context.Context, char *character.Character, args *CraftingArgs) (State[*CraftingArgs], error) {
	if char.GetLevel(args.Item.Crafting.Skill) < args.Item.Crafting.Level {
		log.Println(char.Name, args.Item.Crafting.Skill, "too low level to craft")
		return nil, nil
	}

	need := args.Quantity
	if need <= 0 {
		need = math.MaxInt32
	}

	// Check stop condition
	if args.Made >= need || (args.stop != nil && args.stop(char, args)) {
		if args.BankWhenDone && len(char.Inventory) > 0 {
			err := MoveToBankAndDepositAll(ctx, char)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	}

	if args.Made == 0 {
		char.PushState("Crafting %s", args.Item.Name)
	} else {
		char.PushState("Crafting %s (made %d)", args.Item.Name, args.Made)
	}
	defer char.PopState()

	if args.Item.Crafting == nil {
		return nil, fmt.Errorf("item cannot be crafted")
	}

	// Withdraw as many items from the bank as we can, then craft the resources one by one.
	// This works around an issue of calculating number of items to craft on each iteration.
	totalCraftable, inventoryCraftable := getNumCanCraft(args.Item.Crafting, char.Inventory, char.Bank())
	toWithdraw := map[string]int{}

	numToCraft := 0
	switch {
	case totalCraftable == 0:
		// We ran out of crafting materials
		return nil, nil
	case inventoryCraftable == totalCraftable:
		// Nothing in the bank worth withdrawing
		numToCraft = inventoryCraftable
	case inventoryCraftable > 0:
		// If we have some items to craft in our inventory already, just do it
		numToCraft = inventoryCraftable
	default:
		// Figure out what we need to withdraw from the bank to get a full inventory,
		// unless the bank doesn't have a full inventory's worth we just withdraw what we can.
		maxInventorySizeCanCraft := char.MaxInventoryItems() / args.Item.Crafting.InventoryRequired()
		withdrawMultiplier := min(maxInventorySizeCanCraft, totalCraftable, need)
		for craftingItem, quantity := range args.Item.Crafting.Items {
			toWithdraw[craftingItem.Code] = withdrawMultiplier * quantity
		}
		numToCraft = withdrawMultiplier
	}

	if len(toWithdraw) > 0 {
		err := MoveToBankAndDepositAll(ctx, char)
		if err != nil {
			return nil, err
		}
		err = WithdrawItems(ctx, char, toWithdraw)
		if err != nil {
			// Since we don't lock the bank, it's possible that another character took the items we needed
			if httperror.ErrIsBankInsufficientQuantity(err) || httperror.ErrIsBankItemNotFound(err) {
				return CraftingLoop, nil
			}
			return nil, err
		}
	}

	// Move to the closest workshop
	err := MoveToClosest(ctx, char, game.Maps.GetWorkshops(args.Item.Crafting.Skill))
	if err != nil {
		return nil, err
	}

	// Craft one item
	result, err := char.Craft(ctx, args.Item.Code, numToCraft)
	if err != nil {
		return nil, err
	}

	args.Made += numToCraft
	args.Xp += result.Xp
	for _, drop := range result.Items {
		args.Crafted[drop.Code] += drop.Quantity
	}

	return CraftingLoop, nil
}

// Calculate the maximum number of items we can craft given the resources in our inventory and bank
func getNumCanCraft(crafting *game.Crafting, inventory map[string]int, bank map[string]int) (totalCraftable int, inventoryCraftable int) {
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
