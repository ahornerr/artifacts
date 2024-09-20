package state

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"log"
)

const (
	// Keep some amount of coins around for exchanging tasks
	coinsToKeep = 5

	coinsRequiredForRewardItems = 3
	coinsRewardedFromTask       = 2
	coinsRequiredToExchangeTask = 1

	tasksCoinItemCode = "tasks_coin"
)

type TaskItemArgs struct {
	Item     *game.Item
	Quantity int

	stop func(*character.Character, *TaskItemArgs) bool
}

func TaskItem(itemCode string, quantity int, stop func(*character.Character, *TaskItemArgs) bool) Runner {
	return func(ctx context.Context, char *character.Character) error {
		return Run(ctx, char, TaskItemLoop, NewTaskItemArgs(itemCode, quantity, stop))
	}
}

func NewTaskItemArgs(itemCode string, quantity int, stop func(*character.Character, *TaskItemArgs) bool) *TaskItemArgs {
	return &TaskItemArgs{
		Item:     game.Items.Get(itemCode),
		Quantity: quantity,
		stop:     stop,
	}
}

func TaskItemLoop(ctx context.Context, char *character.Character, args *TaskItemArgs) (State[*TaskItemArgs], error) {
	// Repeat until stop condition
	if args.stop != nil && args.stop(char, args) {
		return nil, nil
	}

	// If we have some task coins, turn them in (hopefully we get some of the item we want)
	haveCoins := char.Inventory[tasksCoinItemCode] + char.Bank()[tasksCoinItemCode]
	availableCoins := haveCoins - coinsToKeep
	if availableCoins >= coinsRequiredForRewardItems {
		err := MoveToBankAndDepositAll(ctx, char)
		if err != nil {
			return nil, err
		}

		// TODO: Account for a coins quantity greater than inventory space
		err = Withdraw(ctx, char, tasksCoinItemCode, availableCoins)
		if err != nil {
			return nil, err
		}

		err = MoveToClosest(ctx, char, game.Maps.GetTaskMasters("monsters"))
		if err != nil {
			return nil, err
		}

		for {
			_, err = char.ExchangeTask(ctx)
			if err != nil {
				return nil, err
			}

			if char.Inventory[args.Item.Code] > args.Quantity {
				return nil, nil
			}
			if char.Inventory[tasksCoinItemCode] < coinsRequiredForRewardItems {
				return TaskItemLoop, nil
			}
		}
	}

	// If we don't have enough task coins, do a task
	taskArgs := NewTaskArgs(func(c *character.Character, args *TaskArgs) bool {
		return args.TasksCompleted > 0
	})
	err := Run(ctx, char, TaskLoop, taskArgs)
	if err != nil {
		return nil, err
	}

	// If we can't complete the task, get a new one (assuming we have task coins)
	if taskArgs.TasksCompleted == 0 {
		log.Println(char.Name, "Cannot complete task")
		if haveCoins < coinsRequiredToExchangeTask {
			return nil, fmt.Errorf("cannot complete task and no coins available to exchange task")
		}

		err = MoveToBankAndDepositAll(ctx, char)
		if err != nil {
			return nil, err
		}

		err = Withdraw(ctx, char, tasksCoinItemCode, coinsRequiredToExchangeTask)
		if err != nil {
			return nil, err
		}

		err = MoveToClosest(ctx, char, game.Maps.GetTaskMasters("monsters"))
		if err != nil {
			return nil, err
		}

		err = char.CancelTask(ctx)
		if err != nil {
			return nil, err
		}

		return TaskItemLoop, nil
	}

	return TaskItemLoop, nil
}
