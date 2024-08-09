package commands

import (
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/graph"
)

func StartGraph(node *graph.Item, char *character.Character) {
	childLevels := getRequiredLevels(node.Children())
	childItems := getRequiredItems(node.Children())

	_ = childLevels
	_ = childItems

	//switch n := node.(type) {
	//case *graph.Item:
	//	// Check if item in inventory
	//	//  Do nothing
	//	// Check if item in bank
	//	//  Move to bank, withdraw
	//	// Else, execute children
	//	//  Children need the item node to be able to craft and/or get required item/quantities
	//	for _, child := range n.Children() {
	//		executeChild(child, n)
	//	}
	//}

}

func getRequiredLevels(nodes []graph.Node) map[string]int {
	levels := map[string]int{}

	for _, node := range nodes {
		if skillNode, ok := node.(*graph.Skill); ok {
			if levels[skillNode.Skill] < skillNode.Level {
				levels[skillNode.Skill] = skillNode.Level
			}
		}
	}

	return levels
}

// TODO: Should this accept a quantity?
func getRequiredItems(nodes []graph.Node) map[*game.Item]int {
	items := map[*game.Item]int{}

	for _, node := range nodes {
		if itemNode, ok := node.(*graph.Item); ok {
			items[itemNode.Item] += itemNode.Quantity
		}
	}

	return items
}
