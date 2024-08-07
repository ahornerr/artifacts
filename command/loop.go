package command

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/stopper"
)

type Loop struct {
	sequence        []Command
	descriptionFunc func(iteration int) string
	stop            stopper.Stopper
	iterations      int
}

func NewLoop(stop stopper.Stopper, descriptionFunc func(iteration int) string, sequence ...Command) *Loop {
	return &Loop{
		sequence:        sequence,
		descriptionFunc: descriptionFunc,
		stop:            stop,
	}
}

func (c *Loop) Description() string {
	if c.descriptionFunc != nil {
		return c.descriptionFunc(c.iterations)
	}
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
