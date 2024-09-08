package state

import (
	"context"
	"github.com/ahornerr/artifacts/character"
)

type EventArgs struct {
	//Rewards map[string]int
	//Task    *client.TaskSchema
	//
	//Drops   map[string]int
	//Gold    int
	//Xp      int
	//Results []client.FightSchemaResult

	stop func(*character.Character, *EventArgs) bool
}

func Event(stop func(*character.Character, *EventArgs) bool) Runner {
	return func(ctx context.Context, char *character.Character) error {
		return Run(ctx, char, EventLoop, NewEventArgs(stop))
	}
}

func NewEventArgs(stop func(*character.Character, *EventArgs) bool) *EventArgs {
	return &EventArgs{
		stop: stop,
	}
}

func EventLoop(ctx context.Context, char *character.Character, args *EventArgs) (State[*EventArgs], error) {
	if args.stop != nil && args.stop(char, args) {
		return nil, nil
	}

	char.PushState("Doing event")
	defer char.PopState()

	//for eventType, eventCodes := range game.Events.Events() {
	//	for eventCode, locations := range eventCodes {
	//
	//	}
	//}
	//
	//// Complete task
	//if char.Task != "" && char.TaskProgress == char.TaskTotal {
	//	err := MoveToClosest(ctx, char, game.Maps.GetTaskMasters("monsters"))
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	reward, err := char.CompleteTask(ctx)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	args.Rewards[reward.Code] += reward.Quantity
	//}
	//
	//// Get new task
	//if char.Task == "" {
	//	err := MoveToClosest(ctx, char, game.Maps.GetTaskMasters("monsters"))
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	args.Task, err = char.NewTask(ctx)
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	//
	//// Bank if full, get better equipment, move to monster, fight once
	//fightArgs := NewFightArgs(char.Task, func(c *character.Character, args *FightArgs) bool {
	//	return args.NumFights() > 0
	//})
	//err := Run(ctx, char, FightLoop, fightArgs)
	//if err != nil {
	//	return nil, err
	//}
	//if fightArgs.NumFights() == 0 {
	//	// Didn't attempt to fight, must be unwinnable
	//	return nil, nil
	//}
	//
	//args.Results = append(args.Results, fightArgs.Results...)
	//args.Xp += fightArgs.Xp
	//args.Gold += fightArgs.Gold
	//for itemCode, quantity := range fightArgs.Drops {
	//	args.Drops[itemCode] += quantity
	//}

	return EventLoop, nil
}
