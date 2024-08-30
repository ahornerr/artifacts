package state

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

type TaskArgs struct {
	// TODO: Equipment override?
	// TODO: Bank first?
	Rewards map[string]int
	Task    *client.TaskSchema

	Drops   map[string]int
	Gold    int
	Xp      int
	Results []client.FightSchemaResult

	stop func(*character.Character, *TaskArgs) bool
}

func (t *TaskArgs) NumFights() int {
	return len(t.Results)
}

func (t *TaskArgs) NumWins() int {
	num := 0
	for _, r := range t.Results {
		if r == client.Win {
			num++
		}
	}
	return num
}

func (t *TaskArgs) NumLosses() int {
	num := 0
	for _, r := range t.Results {
		if r == client.Lose {
			num++
		}
	}
	return num
}

func Task(stop func(*character.Character, *TaskArgs) bool) Runner {
	return func(ctx context.Context, char *character.Character) error {
		return Run(ctx, char, TaskLoop, &TaskArgs{
			Rewards: make(map[string]int),
			Drops:   make(map[string]int),
			stop:    stop,
		})
	}
}

func TaskLoop(ctx context.Context, char *character.Character, args *TaskArgs) (State[*TaskArgs], error) {
	// Repeat until stop condition
	if args.stop(char, args) {
		return nil, nil
	}

	char.PushState("Doing task")
	defer char.PopState()

	// Complete task
	if char.Task != "" && char.TaskProgress == char.TaskTotal {
		err := MoveToClosest(ctx, char, game.Maps.GetTaskMasters("monster"))
		if err != nil {
			return nil, err
		}

		reward, err := char.CompleteTask(ctx)
		if err != nil {
			return nil, err
		}

		args.Rewards[reward.Code] += reward.Quantity
	}

	// Get new task
	if char.Task == "" {
		err := MoveToClosest(ctx, char, game.Maps.GetTaskMasters("monster"))
		if err != nil {
			return nil, err
		}

		args.Task, err = char.NewTask(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Bank if full, get better equipment, move to monster, fight once
	fightArgs := NewFightArgs(char.Task, func(c *character.Character, args *FightArgs) bool {
		return args.NumFights() > 0
	})
	err := Run(ctx, char, FightLoop, fightArgs)
	if err != nil {
		return nil, err
	}

	args.Results = append(args.Results, fightArgs.Results...)
	args.Xp += fightArgs.Xp
	args.Gold += fightArgs.Gold
	for itemCode, quantity := range fightArgs.Drops {
		args.Drops[itemCode] += quantity
	}

	return TaskLoop, nil
}
