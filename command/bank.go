package command

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
)

var (
	MoveToBank = NewSimple("Moving to bank", func(ctx context.Context, char *character.Character) error {
		return char.MoveClosest(ctx, game.Maps.GetBanks())
	})

	DepositAll = NewSimple("Depositing all in bank", func(ctx context.Context, char *character.Character) error {
		return char.BankAll(ctx)
	})
)

var (
	SequenceMoveToBankAndDepositAll = NewSequence(func(ctx context.Context, char *character.Character) []Command {
		return []Command{MoveToBank, DepositAll}
	})

	SequenceBankWhenFull = NewSequence(func(ctx context.Context, char *character.Character) []Command {
		if char.IsInventoryFull() {
			return []Command{SequenceMoveToBankAndDepositAll}
		}
		return nil
	})
)

func Deposit(itemCode string, quantity int) Command {
	return NewSimple(fmt.Sprintf("Depositing %d %s", quantity, itemCode), func(ctx context.Context, char *character.Character) error {
		_, err := char.DepositBank(ctx, itemCode, quantity)
		return err
	})
}

func Withdraw(itemCode string, quantity int) Command {
	return NewSimple(fmt.Sprintf("Withdrawing %d %s", quantity, itemCode), func(ctx context.Context, char *character.Character) error {
		_, err := char.WithdrawBank(ctx, itemCode, quantity)
		// TODO: Handle 404 "Item not found."
		// TODO: Handle 478 "Missing item or insufficient quantity in your inventory."
		return err
	})
}

func WithdrawItems(items map[string]int) Command {
	return NewSequence(func(ctx context.Context, char *character.Character) []Command {
		var commands []Command
		for itemCode, quantity := range items {
			commands = append(commands, Withdraw(itemCode, quantity))
		}
		return commands
	})
}
