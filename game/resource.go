package game

type Resource struct {
	Code string
	Name string
	Loot []Drop

	// TODO: Are Skill and Level always populated?
	Skill string
	Level int
}

func (r Resource) String() string {
	return r.Name
}

type Drop struct {
	Item        *Item
	MaxQuantity int
	MinQuantity int
	Rate        int
}
