package command

import (
	"context"
	"github.com/ahornerr/artifacts/character"
)

func SequenceFunc(execute func(ctx context.Context, char *character.Character) []Command) Command {
	return NewFunctional(
		func() string {
			return ""
		},
		func(ctx context.Context, char *character.Character) ([]Command, error) {
			return execute(ctx, char), nil
		},
	)
}

func Sequence(sequence ...Command) Command {
	return SequenceFunc(func(ctx context.Context, char *character.Character) []Command {
		return sequence
	})
}
