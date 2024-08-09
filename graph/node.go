package graph

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Type string

const (
	TypeCrafting Type = "crafting"
	TypeItem     Type = "item"
	TypeMonster  Type = "monster"
	TypeResource Type = "resource"
	TypeOption   Type = "option"
	TypeSkill    Type = "skill"
	TypeTask     Type = "task"
)

type Node interface {
	Type() Type
	Children() []Node
	String() string
	MarshalJSON() ([]byte, error)
}

type nodeStruct struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Children    []Node `json:"children,omitempty"`
}

func marshal(n Node) ([]byte, error) {
	return json.Marshal(nodeStruct{
		Type:        string(n.Type()),
		Description: n.String(),
		Children:    n.Children(),
	})
}

func Print(node Node, depth int) {
	n := fmt.Sprintf("[%s] %s", strings.Title(string(node.Type())), node.String())
	fmt.Printf("%s- %s\n", strings.Repeat("  ", depth), n)
	for _, child := range node.Children() {
		Print(child, depth+1)
	}
}
