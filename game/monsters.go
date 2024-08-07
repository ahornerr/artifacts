package game

import (
	"context"
	"github.com/ahornerr/artifacts/httperror"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

type monsters struct {
	client   *client.ClientWithResponses
	monsters map[string]*Monster
}

func newMonsters(c *client.ClientWithResponses) *monsters {
	return &monsters{
		client:   c,
		monsters: map[string]*Monster{},
	}
}

func (m *monsters) Get(monsterCode string) *Monster {
	return m.monsters[monsterCode]
}

func (m *monsters) load(ctx context.Context) error {
	page := 1
	size := 100

	m.monsters = map[string]*Monster{}

	for {
		resp, err := m.client.GetAllMonstersMonstersGetWithResponse(ctx, &client.GetAllMonstersMonstersGetParams{
			Page: &page,
			Size: &size,
		})
		if err != nil {
			return err
		} else if resp.JSON200 == nil {
			return httperror.NewHTTPError(resp.StatusCode(), resp.Body)
		}

		for _, monster := range resp.JSON200.Data {
			m.monsters[monster.Code] = &Monster{
				MonsterSchema: monster,
				Stats:         StatsFromMonster(monster),
			}
		}

		if len(resp.JSON200.Data) < size {
			break
		}

		page++
	}

	return nil
}
