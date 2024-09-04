package state

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
)

type CollectInventoryArgs struct {
	Item *game.Item
	Made int

	stop func(*character.Character, *CollectInventoryArgs) bool
}

func CollectInventory(item *game.Item, stop func(*character.Character, *CollectInventoryArgs) bool) Runner {
	return func(ctx context.Context, char *character.Character) error {
		err := Run(ctx, char, CollectInventoryLoop, NewCollectInventoryArgs(item, stop))
		if err != nil {
			return CollectErr{
				Item: item,
				Err:  err,
			}
		}
		return nil
	}
}

func NewCollectInventoryArgs(item *game.Item, stop func(*character.Character, *CollectInventoryArgs) bool) *CollectInventoryArgs {
	return &CollectInventoryArgs{
		Item: item,
		stop: stop,
	}
}

func CollectInventoryLoop(ctx context.Context, char *character.Character, args *CollectInventoryArgs) (State[*CollectInventoryArgs], error) {
	if args.stop(char, args) {
		return nil, nil
	}

	freeInvSpaces := char.MaxInventoryItems() - char.InventoryCount()

	collectItemArgs := NewCollectItemsArgs(args.Item, freeInvSpaces, false, nil)
	err := Run(ctx, char, CollectItemsLoop, collectItemArgs)
	if err != nil {
		return nil, err
	}

	args.Made += freeInvSpaces

	return nil, nil
}
