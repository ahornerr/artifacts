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

	//Drops   map[string]int
	//Gold    int
	//Xp      int
	//Results []client.FightSchemaResult

	stop func(*character.Character, *TaskArgs) bool
}

func Task(stop func(*character.Character, *TaskArgs) bool) Runner {
	return func(ctx context.Context, char *character.Character) error {
		return Run(ctx, char, TaskLoop, NewTaskArgs(stop))
	}
}

func NewTaskArgs(stop func(*character.Character, *TaskArgs) bool) *TaskArgs {
	return &TaskArgs{
		Rewards: make(map[string]int),
		//Drops:   make(map[string]int),
		stop: stop,
	}
}

func TaskLoop(ctx context.Context, char *character.Character, args *TaskArgs) (State[*TaskArgs], error) {
	// Complete task if possible
	// TODO: Support other task types
	if char.Task != "" && char.TaskProgress == char.TaskTotal {
		err := MoveToClosest(ctx, char, game.Maps.GetTaskMasters("monsters"))
		if err != nil {
			return nil, err
		}

		reward, err := char.CompleteTask(ctx)
		if err != nil {
			return nil, err
		}

		args.Rewards[reward.Code] += reward.Quantity
	}

	// Repeat until stop condition
	if args.stop != nil && args.stop(char, args) {
		return nil, nil
	}

	char.PushState("Doing task")
	defer char.PopState()

	// Get new task
	if char.Task == "" {
		err := MoveToClosest(ctx, char, game.Maps.GetTaskMasters("monsters"))
		if err != nil {
			return nil, err
		}

		args.Task, err = char.NewTask(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Bank if full, get better equipment, move to monster, fight once
	fightArgs := NewFightArgs(char.Task, func(c *character.Character, _ *FightArgs) bool {
		return args.stop != nil && args.stop(char, args) || c.TaskProgress >= c.TaskTotal
	}, nil)
	err := Run(ctx, char, FightLoop, fightArgs)
	if err != nil {
		return nil, err
	}

	if fightArgs.NumFights() == 0 {
		// Didn't attempt to fight, must be unwinnable
		return nil, nil
	}

	//args.Results = append(args.Results, fightArgs.Results...)
	//args.Xp += fightArgs.Xp
	//args.Gold += fightArgs.Gold
	//for itemCode, quantity := range fightArgs.Drops {
	//	args.Drops[itemCode] += quantity
	//}

	return TaskLoop, nil
}
