package commands

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/command"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/stopper"
)

func NewFightLoop(monsterCode string, stop stopper.Stopper) command.Command {
	monster := game.Monsters.Get(monsterCode)

	return command.NewLoop(
		stop,
		func(iteration int) string {
			return fmt.Sprintf("Fighting %s (killed %d)", monster.Name, iteration)
		},
		SequenceBankWhenFull,
		EquipBestEquipmentForMonster(monster),
		MoveToClosestMonster(monster),
		Fight(monster),
	)
}

func Fight(monster *game.Monster) command.Command {
	return command.NewSimple(
		fmt.Sprintf("Fighting %s", monster.Name),
		func(ctx context.Context, char *character.Character) error {
			_, err := char.Fight(ctx)
			return err
		},
	)
}

func MoveToClosestMonster(monster *game.Monster) command.Command {
	return command.NewSimple(
		fmt.Sprintf("Moving to closest %s", monster.Name),
		func(ctx context.Context, char *character.Character) error {
			return char.MoveClosest(ctx, game.Maps.GetMonsters(monster.Code))
		},
	)
}
