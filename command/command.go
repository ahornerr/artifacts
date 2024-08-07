package command

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
)

var Idle = SequenceFunc(func(ctx context.Context, char *character.Character) []Command {
	return nil
})

type Command interface {
	Description() string
	Execute(ctx context.Context, char *character.Character) ([]Command, error)
}

type StopCommand struct {
	reason string
}

func NewStopCommand(reason string) Command {
	return &StopCommand{reason}
}

func (s *StopCommand) Description() string {
	return fmt.Sprintf("Stopping: %s", s.reason)
}

func (s *StopCommand) Execute(ctx context.Context, char *character.Character) ([]Command, error) {
	return nil, nil
}
