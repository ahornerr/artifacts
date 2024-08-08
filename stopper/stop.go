package stopper

import (
	"context"
	"github.com/ahornerr/artifacts/character"
)

type stopFunc func(ctx context.Context, char *character.Character, quantity int) (bool, error)

type Stopper interface {
	ShouldStop(ctx context.Context, char *character.Character, quantity int) (bool, error)
}

type Never struct{}

func (s Never) ShouldStop(ctx context.Context, char *character.Character, quantity int) (bool, error) {
	return false, nil
}

type Functional struct {
	stop stopFunc
}

func NewStopper(stop stopFunc) Stopper {
	return &Functional{stop: stop}
}

func (f *Functional) ShouldStop(ctx context.Context, char *character.Character, quantity int) (bool, error) {
	return f.stop(ctx, char, quantity)
}
