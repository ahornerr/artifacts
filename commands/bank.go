package commands

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/command"
	"github.com/ahornerr/artifacts/game"
)

var (
	MoveToBank = command.NewSimple(
		"Moving to bank",
		func(ctx context.Context, char *character.Character) error {
			return char.MoveClosest(ctx, game.Maps.GetBanks())
		},
	)

	DepositAll = command.NewSimple(
		"Depositing all in bank",
		func(ctx context.Context, char *character.Character) error {
			return char.DepositAll(ctx)
		},
	)
)

var (
	SequenceMoveToBankAndDepositAll = command.Sequence(MoveToBank, DepositAll)

	SequenceBankWhenFull = command.SequenceFunc(func(ctx context.Context, char *character.Character) []command.Command {
		if char.IsInventoryFull() {
			return []command.Command{SequenceMoveToBankAndDepositAll}
		}
		return nil
	})
)

func Deposit(itemCode string, quantity int) command.Command {
	return command.NewSimple(
		fmt.Sprintf("Depositing %d %s", quantity, itemCode),
		func(ctx context.Context, char *character.Character) error {
			_, err := char.DepositBank(ctx, itemCode, quantity)
			return err
		},
	)
}

func Withdraw(itemCode string, quantity int) command.Command {
	return command.NewSimple(
		fmt.Sprintf("Withdrawing %d %s", quantity, itemCode),
		func(ctx context.Context, char *character.Character) error {
			_, err := char.WithdrawBank(ctx, itemCode, quantity)
			// TODO: Handle 404 "Item not found."
			// TODO: Handle 478 "Missing item or insufficient quantity in your inventory."
			return err
		},
	)
}

func WithdrawItems(items map[string]int) command.Command {
	var sequence []command.Command
	for itemCode, quantity := range items {
		sequence = append(sequence, Withdraw(itemCode, quantity))
	}

	return command.Sequence(sequence...)
}
