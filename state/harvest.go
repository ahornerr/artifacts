package state

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"math"
	"strings"
)

type HarvestArgs struct {
	Resource *game.Resource
	Count    int
	Drops    map[string]int
	Xp       int

	stop func(*character.Character, *HarvestArgs) bool
}

func Harvest(resourceCode string, stop func(*character.Character, *HarvestArgs) bool) Runner {
	return func(ctx context.Context, char *character.Character) error {
		return Run(ctx, char, HarvestLoop, &HarvestArgs{
			Resource: game.Resources.Get(resourceCode),
			Drops:    map[string]int{},
			stop:     stop,
		})
	}
}

func HarvestLoop(ctx context.Context, char *character.Character, args *HarvestArgs) (State[*HarvestArgs], error) {
	// Repeat until stop condition
	if args.stop(char, args) {
		return nil, nil
	}

	var drops []string
	for itemCode, count := range args.Drops {
		drops = append(drops, fmt.Sprintf("%d %s", count, game.Items.Get(itemCode).Name))
	}

	char.PushState("Harvesting %s (got %s)", args.Resource.Name, strings.Join(drops, ", "))
	defer char.PopState()

	// Bank if full
	if char.IsInventoryFull() {
		err := MoveToBankAndDepositAll(ctx, char)
		if err != nil {
			return nil, err
		}
	}

	// Equip best equipment
	skillStats := &game.Stats{
		Resistance: map[string]int{
			args.Resource.Skill: -math.MaxInt32,
		},
	}
	err := EquipBestEquipment(ctx, char, skillStats)
	if err != nil {
		return nil, err
	}

	// Move to the closest resource
	err = MoveToClosest(ctx, char, game.Maps.GetResources(args.Resource.Code))
	if err != nil {
		return nil, err
	}

	result, err := char.Gather(ctx)
	if err != nil {
		return nil, err
	}

	args.Count++
	args.Xp += result.Xp
	for _, drop := range result.Items {
		args.Drops[drop.Code] += drop.Quantity
	}

	return HarvestLoop, nil
}
