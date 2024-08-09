package graph

type Task struct{}

func NewTask() Node {
	return &Task{}
}

func (t *Task) Type() Type {
	return TypeTask
}

func (t *Task) Children() []Node {
	return nil
}

func (t *Task) String() string {
	return ""
}

func (t *Task) MarshalJSON() ([]byte, error) {
	return marshal(t)
}
