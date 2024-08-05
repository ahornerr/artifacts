package main

import (
	"fmt"
)

type NodeType int

const (
	NodeTypeSkill NodeType = iota
	NodeTypeItem
)

func (n NodeType) String() string {
	return [...]string{"Skill", "Item"}[n]
}

type Node struct {
	Type     NodeType
	Code     string
	Quantity int
	Children []Node
}

func (n Node) String() string {
	return fmt.Sprintf("[%s] %s (%d)", n.Type, n.Code, n.Quantity)
}

func (n Node) AccumulateSkills(skills map[string]int) {
	if n.Type == NodeTypeSkill {
		if skills[n.Code] < n.Quantity {
			skills[n.Code] = n.Quantity
		}
	}
	for _, child := range n.Children {
		child.AccumulateSkills(skills)
	}
}

func (n Node) AccumulateItems(items map[string]int, quantity int) {
	if n.Type == NodeTypeItem {
		items[n.Code] = items[n.Code] + (n.Quantity * quantity)
	}
	for _, child := range n.Children {
		child.AccumulateItems(items, quantity)
	}
}

func (g *Game) ComputeItemNodes(itemCode string, quantity int) Node {
	var children []Node

	g.GetItemSource(itemCode)

	//item := g.Items[itemCode]

	//if len(item.Crafting) > 0 {
	//	// Each item that requires crafting has a skill requirement
	//	levelReq := g.ComputeSkillNodes(item.Skill, item.Level)
	//	children = append(children, levelReq)
	//
	//	for _, req := range item.Crafting {
	//		children = append(children, g.ComputeItemNodes(req.Code, quantity*req.Quantity))
	//	}
	//}

	return Node{
		Type:     NodeTypeItem,
		Code:     itemCode,
		Quantity: quantity,
		Children: children,
	}
}

func (g *Game) ComputeSkillNodes(skill string, level int) Node {
	return Node{
		Type:     NodeTypeSkill,
		Code:     skill,
		Quantity: level,
	}
}
