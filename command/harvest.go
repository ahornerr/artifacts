package command

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/stopper"
)

func NewHarvestLoop(resourceCode string, stop stopper.Stopper) Command {
	return NewLoop(
		stop,
		SequenceBankWhenFull,
		EquipBestEquipmentForResource(resourceCode),
		MoveToClosestResource(resourceCode),
		Gather(resourceCode),
	)
}

func MoveToClosestResource(resourceCode string) Command {
	return NewSimple(fmt.Sprintf("Moving to closest %s resource", resourceCode),
		func(ctx context.Context, char *character.Character) error {
			locations := game.Maps.GetResources(resourceCode)
			if len(locations) == 0 {
				return fmt.Errorf("no locations found for resource code %s", resourceCode)
			}
			return char.MoveClosest(ctx, locations)
		},
	)
}

func Gather(resourceCode string) Command {
	return NewSimple(fmt.Sprintf("Gathering %s", resourceCode),
		func(ctx context.Context, char *character.Character) error {
			_, err := char.Gather(ctx)
			return err
		},
	)
}
