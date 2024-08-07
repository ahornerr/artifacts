package stopper

import (
	"context"
	"github.com/ahornerr/artifacts/character"
)

type Stopper interface {
	ShouldStop(ctx context.Context, char *character.Character, quantity int) (bool, error)
}

type StopNever struct{}

func (s StopNever) ShouldStop(ctx context.Context, char *character.Character, quantity int) (bool, error) {
	return false, nil
}
