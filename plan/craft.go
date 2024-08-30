package plan

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
)

type Craft struct {
	item     *game.Item
	quantity int
	children []Node
}

func NewCraft(item *game.Item, quantity int, children ...Node) *Craft {
	return &Craft{
		item:     item,
		quantity: quantity,
		children: children,
	}
}

func (c *Craft) Name() string {
	return "Craft " + c.item.Name
}

func (c *Craft) Weight(char *character.Character) float64 {
	return 3 // TODO
}

func (c *Craft) Children() []Node {
	// TODO: Is this the best way to handle this?
	return c.children
}

func (c *Craft) Execute(ctx context.Context, char *character.Character) error {
	_, err := char.Craft(ctx, c.item.Code, c.quantity)
	return err
}
