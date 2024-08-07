package command

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/stopper"
)

func Fight(monsterCode string) Command {
	return NewSimple(fmt.Sprintf("Fighting %s", monsterCode),
		func(ctx context.Context, char *character.Character) error {
			_, err := char.Fight(ctx)
			return err
		},
	)
}

func MoveToClosestMonster(monsterCode string) Command {
	return NewSimple(fmt.Sprintf("Moving to closest %s monster", monsterCode),
		func(ctx context.Context, char *character.Character) error {
			return char.MoveClosest(ctx, game.Maps.GetMonsters(monsterCode))
		},
	)
}

func NewFightLoop(monsterCode string, stop stopper.Stopper) Command {
	return NewLoop(
		stop,
		SequenceBankWhenFull,
		EquipBestEquipmentForMonster(monsterCode),
		MoveToClosestMonster(monsterCode),
		Fight(monsterCode),
	)
}
