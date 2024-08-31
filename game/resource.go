package game

type Resource struct {
	Code  string
	Name  string
	Loot  map[*Item]Drop
	Skill string
	Level int
}

func (r Resource) String() string {
	return r.Name
}

type Drop struct {
	MaxQuantity int
	MinQuantity int
	Rate        int
}
