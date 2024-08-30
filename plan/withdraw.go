package plan

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"math"
)

type Withdraw struct {
	item     *game.Item
	quantity int
	children []Node
}

func NewWithdraw(item *game.Item, quantity int, children ...Node) *Withdraw {
	return &Withdraw{
		item:     item,
		quantity: quantity,
		children: children,
	}
}

func (w *Withdraw) Name() string {
	return fmt.Sprintf("Withdraw %d %s", w.quantity, w.item.Name)
}

func (w *Withdraw) Children() []Node {
	return w.children
}

func (w *Withdraw) Weight(char *character.Character) float64 {
	if char.Bank()[w.item.Code] < w.quantity {
		// No items in bank not a possible branch. Set max weight
		return math.MaxFloat32
	}

	// TODO: Double check this cooldown/weight
	return 3.0
}

func (w *Withdraw) Execute(ctx context.Context, char *character.Character) error {
	_, err := char.WithdrawBank(ctx, w.item.Code, w.quantity)
	return err
}
