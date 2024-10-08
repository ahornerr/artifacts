package state

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/httperror"
	"log"
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
		return Run(ctx, char, HarvestLoop, NewHarvestArgs(resourceCode, stop))
	}
}

func NewHarvestArgs(resourceCode string, stop func(*character.Character, *HarvestArgs) bool) *HarvestArgs {
	return &HarvestArgs{
		Resource: game.Resources.Get(resourceCode),
		Drops:    map[string]int{},
		stop:     stop,
	}
}

func HarvestLoop(ctx context.Context, char *character.Character, args *HarvestArgs) (State[*HarvestArgs], error) {
	// Repeat until stop condition
	if args.stop != nil && args.stop(char, args) {
		return nil, nil
	}

	if char.GetLevel(args.Resource.Skill) < args.Resource.Level {
		log.Println(char.Name, args.Resource.Skill, "too low level to harvest")
		return nil, nil
	}

	locations := game.Maps.GetResources(args.Resource.Code)
	if len(locations) == 0 {
		log.Println("No locations found for resource", args.Resource.Name)
		return nil, nil
	}

	char.PushState("Harvesting %s", args.Resource.Name)
	defer char.PopState()

	if len(args.Drops) > 0 || args.Xp > 0 {
		drops := []string{
			fmt.Sprintf("%d XP", args.Xp),
		}
		for itemCode, count := range args.Drops {
			drops = append(drops, fmt.Sprintf("%d %s", count, game.Items.Get(itemCode).Name))
		}
		char.PushState("Got %s", strings.Join(drops, ", "))
		defer char.PopState()
	}

	// Bank if full
	if char.IsInventoryFull() {
		err := MoveToBankAndDepositAll(ctx, char)
		if err != nil {
			return nil, err
		}
	}

	// Equip best equipment
	skillStats := &game.Stats{
		IsResource: true,
		Hp:         20,
	}
	switch args.Resource.Skill {
	case "woodcutting":
		skillStats.ResistWoodcutting = math.MinInt8
	case "mining":
		skillStats.ResistMining = math.MinInt8
	case "fishing":
		skillStats.ResistFishing = math.MinInt8
	default:
		panic(args.Resource.Skill)
	}
	err := EquipBestEquipment(ctx, char, skillStats)
	if err != nil {
		return nil, err
	}

	// Move to the closest resource
	err = MoveToClosest(ctx, char, locations)
	if err != nil {
		return nil, err
	}

	result, err := char.Gather(ctx)
	if err != nil {
		if httperror.ErrIsNotFoundOnMap(err) {
			return nil, nil
		}
		return nil, err
	}

	args.Count++
	args.Xp += result.Xp
	for _, drop := range result.Items {
		args.Drops[drop.Code] += drop.Quantity
	}

	return HarvestLoop, nil
}
