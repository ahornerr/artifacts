package command

import (
	"context"
	"github.com/ahornerr/artifacts/character"
)

var Idle = NewSequence(func(ctx context.Context, char *character.Character) []Command {
	return nil
})

type Command interface {
	Description() string
	Execute(ctx context.Context, char *character.Character) ([]Command, error)
}

func NewSimple(description string, execute func(ctx context.Context, char *character.Character) error) Command {
	return &Simple{
		description: description,
		execute:     execute,
	}
}

type Simple struct {
	description string
	execute     func(ctx context.Context, char *character.Character) error
}

func (c Simple) Description() string {
	return c.description
}

func (c Simple) Execute(ctx context.Context, char *character.Character) ([]Command, error) {
	return nil, c.execute(ctx, char)
}

func NewSequence(execute func(ctx context.Context, char *character.Character) []Command) Command {
	return &Sequence{execute: execute}
}

type Sequence struct {
	execute func(ctx context.Context, char *character.Character) []Command
}

func (c Sequence) Description() string {
	return ""
}

func (c Sequence) Execute(ctx context.Context, char *character.Character) ([]Command, error) {
	return c.execute(ctx, char), nil
}
