package stopper

import (
	"context"
	"github.com/ahornerr/artifacts/character"
)

func StopAtLevel(skill string, level int) Stopper {
	return stopAtLevel{skill: skill, desiredLevel: level}
}

type stopAtLevel struct {
	skill        string
	desiredLevel int
}

func (s stopAtLevel) ShouldStop(ctx context.Context, char *character.Character, quantity int) (bool, error) {
	if char.GetLevel(s.skill) >= s.desiredLevel {
		return true, nil
	}
	return false, nil
}
