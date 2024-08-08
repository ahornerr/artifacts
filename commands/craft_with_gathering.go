package commands

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/command"
	"github.com/ahornerr/artifacts/game"
)

//func CraftWithGathering(itemCode string, stop stopper.Stopper) command.Command {
//	item := game.Items.Get(itemCode)
//
//	totalCraftable, _ := getNumCanCraft(*item.Crafting)
//}

func GetItem(itemCode string, quantity int) command.Command {
	item := game.Items.Get(itemCode)

	return command.NewFunctional(
		func() string {
			return fmt.Sprintf("Get %d %s", quantity, item.Name)
		},
		func(ctx context.Context, char *character.Character) ([]command.Command, error) {
			// Have items in bank
			if char.Bank()[item.Code] > quantity {
				return []command.Command{MoveToBank, Withdraw(item.Code, quantity)}, nil
			}

			// Item is craftable
			if item.Crafting != nil {
				var sequence []command.Command
				for craftingItem, craftingQuantity := range item.Crafting.Items {
					// TODO: This doesn't work when trying to make more than an inventory worth
					sequence = append(sequence, GetItem(craftingItem.Code, craftingQuantity*quantity))
				}
				sequence = append(sequence, MoveToClosestWorkshop(item), Craft(item, quantity))
				return sequence, nil
			}

			// Item is from a resource
			drops := game.Resources.Drops(itemCode)
			_ = drops

			return nil, nil
		},
	)
}
