package game

import (
	"context"
	"github.com/ahornerr/artifacts/httperror"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

type resources struct {
	client    *client.ClientWithResponses
	resources map[string]*Resource
	drops     map[*Item][]*Resource
}

func newResources(c *client.ClientWithResponses) *resources {
	return &resources{
		client:    c,
		resources: map[string]*Resource{},
		drops:     map[*Item][]*Resource{},
	}
}

func (r *resources) Get(resourceCode string) *Resource {
	return r.resources[resourceCode]
}

func (r *resources) ResourcesForItem(item *Item) []*Resource {
	return r.drops[item]
}

func (r *resources) load(ctx context.Context) error {
	page := 1
	size := 100

	r.resources = map[string]*Resource{}
	r.drops = map[*Item][]*Resource{}

	for {
		resp, err := r.client.GetAllResourcesResourcesGetWithResponse(ctx, &client.GetAllResourcesResourcesGetParams{
			Page: &page,
			Size: &size,
		})
		if err != nil {
			return err
		} else if resp.JSON200 == nil {
			return httperror.NewHTTPError(resp.StatusCode(), resp.Body)
		}

		for _, resourceSchema := range resp.JSON200.Data {
			resource := &Resource{
				Code:  resourceSchema.Code,
				Name:  resourceSchema.Name,
				Skill: string(resourceSchema.Skill),
				Level: resourceSchema.Level,
				Loot:  map[*Item]Drop{},
			}

			for _, drop := range resourceSchema.Drops {
				item := Items.Get(drop.Code)

				resource.Loot[item] = Drop{
					MaxQuantity: drop.MaxQuantity,
					MinQuantity: drop.MinQuantity,
					Rate:        drop.Rate,
				}

				if _, ok := r.drops[item]; !ok {
					r.drops[item] = []*Resource{}
				}

				r.drops[item] = append(r.drops[item], resource)
			}

			r.resources[resource.Code] = resource
		}

		if len(resp.JSON200.Data) < size {
			break
		}

		page++
	}

	return nil
}
