package game

import (
	"context"
	"github.com/ahornerr/artifacts/httperror"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

type events struct {
	client *client.ClientWithResponses
	events map[string]map[string][]Location
}

func newEvents(c *client.ClientWithResponses) *events {
	return &events{
		client: c,
		events: map[string]map[string][]Location{},
	}
}

func (e *events) Events() map[string]map[string][]Location {
	return e.events
}

func (e *events) load(ctx context.Context) error {
	page := 1
	size := 100

	e.events = map[string]map[string][]Location{}

	for {
		resp, err := e.client.GetAllEventsEventsGetWithResponse(ctx, &client.GetAllEventsEventsGetParams{
			Page: &page,
			Size: &size,
		})
		if err != nil {
			return err
		} else if resp.JSON200 == nil {
			return httperror.NewHTTPError(resp.StatusCode(), resp.Body)
		}

		for _, event := range resp.JSON200.Data {
			content, err := event.Map.Content.AsMapContentSchema()
			if err != nil {
				return err
			}

			if content.Type == "" || content.Code == "" {
				continue
			}

			if _, ok := e.events[content.Type]; !ok {
				e.events[content.Type] = map[string][]Location{}
			}

			if _, ok := e.events[content.Type][content.Code]; !ok {
				e.events[content.Type][content.Code] = []Location{}
			}

			e.events[content.Type][content.Code] = append(e.events[content.Type][content.Code], Location{
				Name: content.Code,
				X:    event.Map.X,
				Y:    event.Map.Y,
			})
		}

		if len(resp.JSON200.Data) < size {
			break
		}

		page++
	}

	return nil
}
