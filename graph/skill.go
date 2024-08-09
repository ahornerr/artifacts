package graph

import "fmt"

type Skill struct {
	Skill string
	Level int
}

func NewSkill(skill string, level int) Node {
	return &Skill{
		Skill: skill,
		Level: level,
	}
}

func (s *Skill) Type() Type {
	return TypeSkill
}

func (s *Skill) Children() []Node {
	return nil
}

func (s *Skill) String() string {
	return fmt.Sprintf("Level %d %s", s.Level, s.Skill)
}

func (s *Skill) MarshalJSON() ([]byte, error) {
	return marshal(s)
}
