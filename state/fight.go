package state

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

type FightArgs struct {
	Monster *game.Monster
	Result  *client.FightSchema
}

// TODO: Add a fight loop

func NewFightArgs(monsterCode string) (State[*FightArgs], *FightArgs) {
	return Fight, &FightArgs{
		Monster: game.Monsters.Get(monsterCode),
	}
}

func Fight(ctx context.Context, char *character.Character, args *FightArgs) (State[*FightArgs], error) {
	char.PushState("Fighting %s", args.Monster.Name)
	defer char.PopState()

	// Bank if full
	if char.IsInventoryFull() {
		err := MoveToBankAndDepositAll(ctx, char)
		if err != nil {
			return nil, err
		}
	}

	// Equip best equipment
	err := EquipBestEquipment(ctx, char, args.Monster.Stats)
	if err != nil {
		return nil, err
	}

	// Move to the closest monster
	err = MoveToClosest(ctx, char, game.Maps.GetMonsters(args.Monster.Code))
	if err != nil {
		return nil, err
	}

	// Fight monster
	args.Result, err = char.Fight(ctx)
	return nil, err
}
