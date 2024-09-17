package state

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/httperror"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"log"
	"reflect"
	"strings"
)

type FightArgs struct {
	Monster *game.Monster
	Drops   map[string]int
	Gold    int
	Xp      int
	Results []client.FightSchemaResult

	lastBank map[string]int
	stop     func(*character.Character, *FightArgs) bool
	bankWhen func(*character.Character, *FightArgs) bool
}

func (t *FightArgs) NumFights() int {
	return len(t.Results)
}

func (t *FightArgs) NumWins() int {
	num := 0
	for _, r := range t.Results {
		if r == client.Win {
			num++
		}
	}
	return num
}

func (t *FightArgs) NumLosses() int {
	num := 0
	for _, r := range t.Results {
		if r == client.Lose {
			num++
		}
	}
	return num
}

func Fight(monsterCode string, stop func(*character.Character, *FightArgs) bool, bankWhen func(*character.Character, *FightArgs) bool) Runner {
	return func(ctx context.Context, char *character.Character) error {
		return Run(ctx, char, FightLoop, NewFightArgs(monsterCode, stop, bankWhen))
	}
}

func NewFightArgs(monsterCode string, stop func(*character.Character, *FightArgs) bool, bankWhen func(*character.Character, *FightArgs) bool) *FightArgs {
	return &FightArgs{
		Monster:  game.Monsters.Get(monsterCode),
		Drops:    map[string]int{},
		stop:     stop,
		bankWhen: bankWhen,
	}
}

func FightLoop(ctx context.Context, char *character.Character, args *FightArgs) (State[*FightArgs], error) {
	// Repeat until stop condition
	if args.stop != nil && args.stop(char, args) {
		return nil, nil
	}

	char.PushState("Fighting %s", args.Monster.Name)
	defer char.PopState()

	if args.NumFights() > 0 {
		char.PushState(fmt.Sprintf("Won %d, lost %d", args.NumWins(), args.NumLosses()))
		defer char.PopState()
	}

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

	locations := game.Maps.GetMonsters(args.Monster.Code)
	if len(locations) == 0 {
		log.Println("No locations found for monster", args.Monster.Name)
		return nil, nil
	}

	// Bank if full
	if char.IsInventoryFull() || (args.bankWhen != nil && args.bankWhen(char, args)) {
		err := MoveToBankAndDepositAll(ctx, char)
		if err != nil {
			return nil, err
		}
	}

	// Equip best equipment
	if !reflect.DeepEqual(args.lastBank, char.Bank()) {
		err := EquipBestEquipment(ctx, char, args.Monster.Stats)
		if err != nil {
			log.Println(char.Name, err)
			return nil, nil
		}
		args.lastBank = char.Bank()
	}

	// Move to the closest monster
	err := MoveToClosest(ctx, char, locations)
	if err != nil {
		return nil, err
	}

	// Fight monster
	result, err := char.Fight(ctx)
	if err != nil {
		if httperror.ErrIsNotFoundOnMap(err) {
			return nil, nil
		}
		return nil, err
	}

	args.Results = append(args.Results, result.Result)
	args.Xp += result.Xp
	args.Gold += result.Gold
	for _, drop := range result.Drops {
		args.Drops[drop.Code] += drop.Quantity
	}

	return FightLoop, err
}
