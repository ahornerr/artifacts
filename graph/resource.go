package graph

import (
	"github.com/ahornerr/artifacts/game"
)

type Resource struct {
	Resource *game.Resource
	Item     *game.Item
	//Skill    Node
}

func NewResource(resource *game.Resource, item *game.Item) Node {
	return &Resource{
		Resource: resource,
		Item:     item,
		//Skill:    NewSkill(resource.Skill, resource.Level),
	}
}

func (r *Resource) Type() Type {
	return TypeResource
}

func (r *Resource) Children() []Node {
	//return []Node{r.Skill}
	return nil
}

func (r *Resource) String() string {
	return r.Resource.Name
}

func (r *Resource) MarshalJSON() ([]byte, error) {
	return marshal(r)
}
