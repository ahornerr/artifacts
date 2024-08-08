package stopper

import (
	"context"
	"github.com/ahornerr/artifacts/character"
)

func AtQuantity(quantity int) Stopper {
	return stopAtQuantity{desiredQuantity: quantity}
}

type stopAtQuantity struct {
	desiredQuantity int
}

func (s stopAtQuantity) ShouldStop(ctx context.Context, char *character.Character, quantity int) (bool, error) {
	// TODO: Check >= vs >
	if quantity >= s.desiredQuantity {
		return true, nil
	}
	return false, nil
}
