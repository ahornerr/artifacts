package graph

import (
	"github.com/ahornerr/artifacts/game"
)

type Resource struct {
	Resource *game.Resource
	Skill    Node
}

func NewResource(resource *game.Resource) Node {
	return &Resource{
		Resource: resource,
		Skill:    NewSkill(resource.Skill, resource.Level),
	}
}

func (r *Resource) Type() Type {
	return TypeResource
}

func (r *Resource) Children() []Node {
	return []Node{r.Skill}
}

func (r *Resource) String() string {
	return r.Resource.Name
}

func (r *Resource) MarshalJSON() ([]byte, error) {
	return marshal(r)
}
