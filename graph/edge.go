package graph

type Edge interface {
	String() string
	//Do(context.Context, *character.Character) error
}

type EdgeFight struct {
}

func (e EdgeFight) String() string {
	return "Fight"
}
