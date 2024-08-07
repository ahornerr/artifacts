package command

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/stopper"
)

type Loop struct {
	sequence   []Command
	stop       stopper.Stopper
	iterations int
}

func NewLoop(stop stopper.Stopper, sequence ...Command) *Loop {
	return &Loop{sequence: sequence, stop: stop}
}

func (c *Loop) Description() string {
	return fmt.Sprintf("Loop #%d", c.iterations)
}

func (c *Loop) Execute(ctx context.Context, char *character.Character) ([]Command, error) {
	stop, err := c.stop.ShouldStop(ctx, char, c.iterations)
	if stop || err != nil {
		return nil, err
	}

	defer func() {
		c.iterations++
	}()

	return append(c.sequence, c), nil
}
