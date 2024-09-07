package state

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"log"
)

type MakeXArgs struct {
	Item     *game.Item
	Quantity int
	Recycle  bool
	stop     func(character *character.Character, args *MakeXArgs) bool

	Made int
}

func NewMakeXArgs(itemCode string, quantity int, recycle bool, stop func(character *character.Character, args *MakeXArgs) bool) *MakeXArgs {
	return &MakeXArgs{
		Item:     game.Items.Get(itemCode),
		Quantity: quantity,
		Recycle:  recycle,
		stop:     stop,
	}
}

func MakeX(itemCode string, quantity int, recycle bool, stop func(character *character.Character, args *MakeXArgs) bool) Runner {
	return func(ctx context.Context, char *character.Character) error {
		return Run(ctx, char, MakeXLoop, NewMakeXArgs(itemCode, quantity, recycle, stop))
	}
}

func MakeXLoop(ctx context.Context, char *character.Character, args *MakeXArgs) (State[*MakeXArgs], error) {
	if args.stop != nil && args.stop(char, args) {
		return nil, nil
	}

	char.PushState("Making %d %s (made %d)", args.Quantity, args.Item.Name, args.Made)
	defer char.PopState()

	collectItemArgs := NewCollectItemsArgs(args.Item, args.Quantity, false, false, nil)
	err := Run(ctx, char, CollectItemsLoop, collectItemArgs)
	if err != nil {
		return nil, err
	}

	args.Made += args.Quantity

	quantityInInventory := char.Inventory[args.Item.Code]
	if args.Recycle && quantityInInventory > 0 {
		_, err = char.Recycle(ctx, args.Item.Code, quantityInInventory)
		if err != nil {
			log.Println("Error recycling:", err)
		}
	}

	err = MoveToBankAndDepositAll(ctx, char)
	if err != nil {
		return nil, err
	}

	return MakeXLoop, nil
}
