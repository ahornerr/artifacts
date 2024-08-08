package game

import (
	"context"
	"github.com/ahornerr/artifacts/httperror"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

type resources struct {
	client    *client.ClientWithResponses
	resources map[string]client.ResourceSchema
	drops     map[string][]client.ResourceSchema
}

// TODO: It might be about time to create a Resource type...
//  Similar to how crafting works

func newResources(c *client.ClientWithResponses) *resources {
	return &resources{
		client:    c,
		resources: map[string]client.ResourceSchema{},
		drops:     map[string][]client.ResourceSchema{},
	}
}

func (r *resources) Get(resourceCode string) client.ResourceSchema {
	return r.resources[resourceCode]
}

func (r *resources) Drops(itemCode string) []client.ResourceSchema {
	return r.drops[itemCode]
}

func (r *resources) load(ctx context.Context) error {
	page := 1
	size := 100

	r.resources = map[string]client.ResourceSchema{}
	r.drops = map[string][]client.ResourceSchema{}

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

		for _, resource := range resp.JSON200.Data {
			r.resources[resource.Code] = resource

			for _, d := range resource.Drops {
				itemCode := d.Code
				if _, ok := r.drops[itemCode]; !ok {
					r.drops[itemCode] = []client.ResourceSchema{}
				}

				r.drops[itemCode] = append(r.drops[itemCode], resource)
			}
		}

		if len(resp.JSON200.Data) < size {
			break
		}

		page++
	}

	return nil
}
