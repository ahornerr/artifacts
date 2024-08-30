package plan

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
)

type Move struct {
	locations []game.Location
	children  []Node
}

func NewMove(locations []game.Location, children ...Node) *Move {
	return &Move{
		locations: locations,
		children:  children,
	}
}

func (m *Move) Name() string {
	return fmt.Sprintf("Move to %s", m.locations[0].Name)
}

func (m *Move) Weight(char *character.Character) float64 {
	_, distance := char.ClosestOf(m.locations)
	return distance * 5 // Each tile moved incurs a 5 second cooldown
}

func (m *Move) Children() []Node {
	// TODO: Is this the best way to handle this?
	return m.children
}

func (m *Move) Execute(ctx context.Context, char *character.Character) error {
	return char.MoveClosest(ctx, m.locations)
}
