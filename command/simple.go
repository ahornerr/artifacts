package command

import (
	"context"
	"github.com/ahornerr/artifacts/character"
)

func NewSimple(description string, execute func(ctx context.Context, char *character.Character) error) Command {
	return NewFunctional(
		func() string {
			return description
		},
		func(ctx context.Context, char *character.Character) ([]Command, error) {
			return nil, execute(ctx, char)
		},
	)
}
