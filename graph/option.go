package graph

import "fmt"

type Option struct {
	Options []Node
}

func NewOption(options ...Node) Node {
	return &Option{Options: options}
}

func (o *Option) Type() Type {
	return TypeOption
}

func (o *Option) Children() []Node {
	return o.Options
}

func (o *Option) String() string {
	return fmt.Sprintf("%d options", len(o.Options))
}

func (o *Option) MarshalJSON() ([]byte, error) {
	return marshal(o)
}
