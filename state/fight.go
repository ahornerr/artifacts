package state

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"log"
	"reflect"
)

type FightArgs struct {
	Monster *game.Monster
	Drops   map[string]int
	Gold    int
	Xp      int
	Results []client.FightSchemaResult

	lastBank map[string]int
	stop     func(*character.Character, *FightArgs) bool
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

func Fight(monsterCode string, stop func(*character.Character, *FightArgs) bool) Runner {
	return func(ctx context.Context, char *character.Character) error {
		return Run(ctx, char, FightLoop, NewFightArgs(monsterCode, stop))
	}
}

func NewFightArgs(monsterCode string, stop func(*character.Character, *FightArgs) bool) *FightArgs {
	return &FightArgs{
		Monster: game.Monsters.Get(monsterCode),
		Drops:   map[string]int{},
		stop:    stop,
	}
}

func FightLoop(ctx context.Context, char *character.Character, args *FightArgs) (State[*FightArgs], error) {
	// Repeat until stop condition
	if args.stop(char, args) {
		return nil, nil
	}

	if args.NumFights() == 0 {
		char.PushState("Fighting %s", args.Monster.Name)
	} else {
		char.PushState("Fighting %s (won %d, lost %d)", args.Monster.Name, args.NumWins(), args.NumLosses())
	}
	defer char.PopState()

	// Bank if full
	if char.IsInventoryFull() {
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
	err := MoveToClosest(ctx, char, game.Maps.GetMonsters(args.Monster.Code))
	if err != nil {
		return nil, err
	}

	// Fight monster
	result, err := char.Fight(ctx)
	if err != nil {
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
