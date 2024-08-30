package state

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
)

func Deposit(ctx context.Context, char *character.Character, items map[*game.Item]int) error {
	for item, quantity := range items {
		_, err := char.DepositBank(ctx, item.Code, quantity)
		if err != nil {
			return err
		}
	}
	return nil
}

func DepositAll(ctx context.Context, char *character.Character) error {
	return char.DepositAll(ctx)
}

func MoveToBank(ctx context.Context, char *character.Character) error {
	return char.MoveClosest(ctx, game.Maps.GetBanks())
}

func MoveToBankAndDepositAll(ctx context.Context, char *character.Character) error {
	char.PushState("Moving to bank to deposit all")
	defer char.PopState()

	err := MoveToBank(ctx, char)
	if err != nil {
		return err
	}

	return DepositAll(ctx, char)
}

func Withdraw(ctx context.Context, char *character.Character, itemCode string, quantity int) error {
	_, err := char.WithdrawBank(ctx, itemCode, quantity)
	// TODO: Handle 404 "Item not found."
	// TODO: Handle 478 "Missing item or insufficient quantity in your inventory."
	return err
}

func WithdrawItems(ctx context.Context, char *character.Character, items map[string]int) error {
	for itemCode, quantity := range items {
		err := Withdraw(ctx, char, itemCode, quantity)
		if err != nil {
			return err
		}
	}
	return nil
}
