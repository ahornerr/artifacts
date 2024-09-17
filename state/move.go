package state

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
)

func Move(ctx context.Context, char *character.Character, location game.Location) error {
	if char.Location.X == location.X && char.Location.Y == location.Y {
		return nil
	}

	char.PushState("Moving to %s", location.Name)
	defer char.PopState()
	return char.Move(ctx, location.X, location.Y)
}

func MoveToClosest(ctx context.Context, char *character.Character, locations []game.Location) error {
	closest, _ := char.ClosestOf(locations)
	return Move(ctx, char, closest)
}
