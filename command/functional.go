package command

import (
	"context"
	"github.com/ahornerr/artifacts/character"
)

type Functional struct {
	descriptionFunc func() string
	execute         func(ctx context.Context, char *character.Character) ([]Command, error)
}

func NewFunctional(
	descriptionFunc func() string,
	execute func(ctx context.Context, char *character.Character) ([]Command, error),
) Command {
	return &Functional{
		descriptionFunc: descriptionFunc,
		execute:         execute,
	}
}

func (d *Functional) Description() string {
	return d.descriptionFunc()
}

func (d *Functional) Execute(ctx context.Context, char *character.Character) ([]Command, error) {
	return d.execute(ctx, char)
}
