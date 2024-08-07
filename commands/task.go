package commands

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/command"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/stopper"
)

var (
	MoveToTaskmaster = command.NewSimple(
		"Moving to taskmaster",
		func(ctx context.Context, char *character.Character) error {
			return char.MoveClosest(ctx, game.Maps.GetTaskMasters("monsters"))
		},
	)

	GetNewTask = command.NewSimple(
		"Getting new task",
		func(ctx context.Context, char *character.Character) error {
			_, err := char.NewTask(ctx)
			return err
		},
	)

	CompleteTask = command.NewSimple(
		"Completing task",
		func(ctx context.Context, char *character.Character) error {
			_, err := char.CompleteTask(ctx)
			// TODO: Notify task runner with task reward or something similar?
			return err
		},
	)
)

var (
	MoveToClosestTaskMonster = command.SequenceFunc(func(ctx context.Context, char *character.Character) []command.Command {
		monster := game.Monsters.Get(char.Task)
		return []command.Command{MoveToClosestMonster(monster)}
	})

	SequenceGetNewTaskIfNone = command.SequenceFunc(func(ctx context.Context, char *character.Character) []command.Command {
		if char.Task == "" {
			return []command.Command{MoveToTaskmaster, GetNewTask}
		}
		return nil
	})

	FightTaskMonster = command.SequenceFunc(func(ctx context.Context, char *character.Character) []command.Command {
		monster := game.Monsters.Get(char.Task)
		return []command.Command{Fight(monster)}
	})

	SequenceCompleteTaskIfFinished = command.SequenceFunc(func(ctx context.Context, char *character.Character) []command.Command {
		if char.TaskProgress == char.TaskTotal {
			return []command.Command{MoveToTaskmaster, CompleteTask}
		}
		return nil
	})
)

type TaskLoop struct {
	stop             stopper.Stopper
	monster          *game.Monster
	numKilled        int
	lastProgress     int
	fightsWithoutWin int
}

func NewTaskLoop(stop stopper.Stopper) command.Command {
	return &TaskLoop{stop: stop}
}

func (t *TaskLoop) Description() string {
	return fmt.Sprintf("Task %s (killed %d)", t.monster.Name, t.numKilled)
}

func (t *TaskLoop) Execute(ctx context.Context, char *character.Character) ([]command.Command, error) {
	stop, err := t.stop.ShouldStop(ctx, char, t.numKilled)
	if stop || err != nil {
		return nil, err
	}

	taskChanged := char.Task != t.monster.Code
	if taskChanged {
		t.monster = game.Monsters.Get(char.Task)
		t.numKilled = 0
		t.fightsWithoutWin = 0
	} else {
		if t.lastProgress == char.TaskProgress {
			t.fightsWithoutWin++
		} else {
			t.lastProgress = char.TaskProgress
			t.numKilled++
			t.fightsWithoutWin = 0
		}
	}

	if t.fightsWithoutWin > 3 {
		return []command.Command{NewFightLoop("skeleton", t.stop)}, nil
	}

	return []command.Command{
		SequenceCompleteTaskIfFinished,
		SequenceGetNewTaskIfNone,
		SequenceBankWhenFull,
		SequenceEquipBestEquipmentForTask,
		MoveToClosestTaskMonster,
		FightTaskMonster,
		t,
	}, nil
}
