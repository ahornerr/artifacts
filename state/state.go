package state

import (
	"context"
	"github.com/ahornerr/artifacts/character"
)

type State[T any] func(ctx context.Context, char *character.Character, args T) (State[T], error)

func Run[T any](ctx context.Context, char *character.Character, start State[T], args T) error {
	var err error
	current := start
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		current, err = current(ctx, char, args)
		if err != nil {
			return err
		}
		if current == nil {
			return nil
		}
	}
}

type Runner func(ctx context.Context, char *character.Character) error

//func Sequence(runners ...Runner) Runner {
//	return func(ctx context.Context, char *character.Character) error {
//		for _, runner := range runners {
//			if err := runner(ctx, char); err != nil {
//				return err
//			}
//		}
//		return nil
//	}
//}

func Loop(runners ...Runner) Runner {
	return func(ctx context.Context, char *character.Character) error {
		for {
			for _, runner := range runners {
				if err := runner(ctx, char); err != nil {
					return err
				}
			}
		}
	}
}
