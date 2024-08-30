package plan

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
)

type GetItem struct {
	item     *game.Item
	quantity int
	children []Node
}

func NewGetItem(item *game.Item, quantity int, children ...Node) *GetItem {
	return &GetItem{
		item:     item,
		quantity: quantity,
		children: OptionsForItem(item, quantity, children...),
	}
}

func (g *GetItem) Name() string {
	return fmt.Sprintf("Get %d %s", g.quantity, g.item.Name)
}

func (g *GetItem) Weight(char *character.Character) float64 {
	return 0
}

func (g *GetItem) Children() []Node {
	// TODO: Is this the best way to handle this?
	return g.children
}

func (g *GetItem) Execute(ctx context.Context, char *character.Character) error {
	return nil
}
