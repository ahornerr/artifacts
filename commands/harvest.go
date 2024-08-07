package commands

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/command"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/stopper"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

func NewHarvestLoop(resourceCode string, stop stopper.Stopper) command.Command {
	resource := game.Resources.Get(resourceCode)

	return command.NewLoop(
		stop,
		func(iteration int) string {
			return fmt.Sprintf("Harvesting %s (got %d)", resource.Name, iteration)
		},
		SequenceBankWhenFull,
		EquipBestEquipmentForResource(resource),
		MoveToClosestResource(resource),
		Harvest(resource),
	)
}

func MoveToClosestResource(resource client.ResourceSchema) command.Command {
	return command.NewSimple(
		fmt.Sprintf("Moving to closest %s", resource.Name),
		func(ctx context.Context, char *character.Character) error {
			locations := game.Maps.GetResources(resource.Code)
			if len(locations) == 0 {
				return fmt.Errorf("no locations found for resource %s", resource.Code)
			}
			return char.MoveClosest(ctx, locations)
		},
	)
}

func Harvest(resource client.ResourceSchema) command.Command {
	return command.NewSimple(
		fmt.Sprintf("Harvesting %s", resource.Name),
		func(ctx context.Context, char *character.Character) error {
			_, err := char.Gather(ctx)
			return err
		},
	)
}
