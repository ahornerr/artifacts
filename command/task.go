package command

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/stopper"
)

var (
	MoveToTaskmaster = NewSimple("Moving to taskmaster", func(ctx context.Context, char *character.Character) error {
		return char.MoveClosest(ctx, game.Maps.GetTaskMasters("monsters"))
	})

	GetNewTask = NewSimple("Getting new task", func(ctx context.Context, char *character.Character) error {
		_, err := char.NewTask(ctx)
		return err
	})

	CompleteTask = NewSimple("Completing task", func(ctx context.Context, char *character.Character) error {
		_, err := char.CompleteTask(ctx)
		// TODO: Notify task runner with task reward or something similar?
		return err
	})
)

var (
	MoveToClosestTaskMonster = NewSequence(func(ctx context.Context, char *character.Character) []Command {
		return []Command{MoveToClosestMonster(char.Task)}
	})

	SequenceGetNewTaskIfNone = NewSequence(func(ctx context.Context, char *character.Character) []Command {
		if char.Task == "" {
			return []Command{MoveToTaskmaster, GetNewTask}
		}
		return nil
	})

	FightTaskMonster = NewSequence(func(ctx context.Context, char *character.Character) []Command {
		return []Command{Fight(char.Task)}
	})

	SequenceCompleteTaskIfFinished = NewSequence(func(ctx context.Context, char *character.Character) []Command {
		if char.TaskProgress == char.TaskTotal {
			return []Command{MoveToTaskmaster, CompleteTask}
		}
		return nil
	})
)

type TaskLoop struct {
	stop             stopper.Stopper
	monsterCode      string
	numKilled        int
	lastProgress     int
	fightsWithoutWin int
}

func NewTaskLoop(stop stopper.Stopper) Command {
	return &TaskLoop{stop: stop}
}

func (t *TaskLoop) Description() string {
	return fmt.Sprintf("Fighting %s (killed %d)", t.monsterCode, t.numKilled)
}

func (t *TaskLoop) Execute(ctx context.Context, char *character.Character) ([]Command, error) {
	stop, err := t.stop.ShouldStop(ctx, char, t.numKilled)
	if stop || err != nil {
		return nil, err
	}

	taskChanged := char.Task != t.monsterCode
	if taskChanged {
		t.monsterCode = char.Task
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
		return []Command{NewFightLoop("skeleton", t.stop)}, nil
	}

	return []Command{
		SequenceCompleteTaskIfFinished,
		SequenceGetNewTaskIfNone,
		SequenceBankWhenFull,
		SequenceEquipBestEquipmentForTask,
		MoveToClosestTaskMonster,
		FightTaskMonster,
		t,
	}, nil
}
