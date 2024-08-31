package game

import (
	"context"
	"github.com/ahornerr/artifacts/httperror"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

type monsters struct {
	client   *client.ClientWithResponses
	monsters map[string]*Monster
	drops    map[*Item][]*Monster
}

func newMonsters(c *client.ClientWithResponses) *monsters {
	return &monsters{
		client:   c,
		monsters: map[string]*Monster{},
		drops:    map[*Item][]*Monster{},
	}
}

func (m *monsters) Get(monsterCode string) *Monster {
	return m.monsters[monsterCode]
}

func (m *monsters) GetAll() map[string]*Monster {
	return m.monsters
}

func (m *monsters) MonstersForItem(item *Item) []*Monster {
	return m.drops[item]
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

		for _, monsterSchema := range resp.JSON200.Data {
			monster := &Monster{
				Code:    monsterSchema.Code,
				Name:    monsterSchema.Name,
				Stats:   StatsFromMonster(monsterSchema),
				Level:   monsterSchema.Level,
				MaxGold: monsterSchema.MaxGold,
				MinGold: monsterSchema.MinGold,
				Loot:    map[*Item]Drop{},
			}

			for _, drop := range monsterSchema.Drops {
				item := Items.Get(drop.Code)

				monster.Loot[item] = Drop{
					MaxQuantity: drop.MaxQuantity,
					MinQuantity: drop.MinQuantity,
					Rate:        drop.Rate,
				}

				if _, ok := m.drops[item]; !ok {
					m.drops[item] = []*Monster{}
				}

				m.drops[item] = append(m.drops[item], monster)
			}

			m.monsters[monsterSchema.Code] = monster
		}

		if len(resp.JSON200.Data) < size {
			break
		}

		page++
	}

	return nil
}
