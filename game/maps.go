package game

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/httperror"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"sync"
)

type maps struct {
	client *client.ClientWithResponses
	maps   map[string]map[string][]Location
	mux    sync.Mutex
}

func newMaps(c *client.ClientWithResponses) *maps {
	return &maps{
		client: c,
		maps:   map[string]map[string][]Location{},
	}
}

func (m *maps) GetBanks() []Location {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.maps["bank"]["bank"]
}

func (m *maps) GetResources(resourceCode string) []Location {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.maps["resource"][resourceCode]
}

func (m *maps) GetWorkshops(skill string) []Location {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.maps["workshop"][skill]
}

func (m *maps) GetMonsters(monsterCode string) []Location {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.maps["monster"][monsterCode]
}

func (m *maps) GetTaskMasters(taskType string) []Location {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.maps["tasks_master"][taskType]
}

func (m *maps) load(ctx context.Context) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	page := 1
	size := 100

	m.maps = map[string]map[string][]Location{}

	for {
		resp, err := m.client.GetAllMapsMapsGetWithResponse(ctx, &client.GetAllMapsMapsGetParams{
			Page: &page,
			Size: &size,
		})
		if err != nil {
			return err
		} else if resp.JSON200 == nil {
			return httperror.NewHTTPError(resp.StatusCode(), resp.Body)
		}

		for _, tile := range resp.JSON200.Data {
			content, err := tile.Content.AsMapContentSchema()
			if err != nil {
				return err
			}

			if content.Type == "" || content.Code == "" {
				continue
			}

			if _, ok := m.maps[content.Type]; !ok {
				m.maps[content.Type] = map[string][]Location{}
			}

			if _, ok := m.maps[content.Type][content.Code]; !ok {
				m.maps[content.Type][content.Code] = []Location{}
			}

			locationName := ""
			switch content.Type {
			case "monster":
				locationName = Monsters.Get(content.Code).Name
			case "resource":
				locationName = Resources.Get(content.Code).Name
			case "workshop":
				locationName = fmt.Sprintf("%s workshop", content.Code)
			case "bank":
				locationName = "bank"
			case "grand_exchange":
				locationName = "grand exchange"
			case "tasks_master":
				locationName = fmt.Sprintf("%s task master", content.Code)
			default:
				locationName = content.Code
			}

			m.maps[content.Type][content.Code] = append(m.maps[content.Type][content.Code], Location{
				Name: locationName,
				X:    tile.X,
				Y:    tile.Y,
			})
		}

		if len(resp.JSON200.Data) < size {
			break
		}

		page++
	}

	return nil
}
