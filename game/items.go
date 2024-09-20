package game

import (
	"context"
	"github.com/ahornerr/artifacts/httperror"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"slices"
)

type items struct {
	client *client.ClientWithResponses
	items  map[string]*Item
}

func newItems(c *client.ClientWithResponses) *items {
	return &items{
		client: c,
		items:  map[string]*Item{},
	}
}

func (i *items) GetAll() map[string]*Item {
	return i.items
}

func (i *items) Get(itemCode string) *Item {
	return i.items[itemCode]
}

func (i *items) ForTrainingCraftingSkill(skill string, charLevel int) []*Item {
	var items []*Item
	for _, item := range i.items {
		if item.Crafting == nil {
			continue
		}
		if item.Crafting.Skill != skill {
			continue
		}
		// More than 10 levels above we stop receiving XP
		if item.Crafting.Level > charLevel || item.Crafting.Level+10 < charLevel {
			continue
		}
		items = append(items, item)
	}
	slices.SortFunc(items, func(a, b *Item) int {
		return a.Crafting.Level - b.Crafting.Level
	})
	return items
}

// ForLevel returns items that are craftable and usable at the supplied combat level
func (i *items) ForLevel(level int) []*Item {
	var items []*Item
	for _, item := range i.items {
		if item.Crafting == nil {
			continue
		}
		if item.Level != level {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (i *items) load(ctx context.Context) error {
	page := 1
	size := 100

	i.items = map[string]*Item{}

	for {
		resp, err := i.client.GetAllItemsItemsGetWithResponse(ctx, &client.GetAllItemsItemsGetParams{
			Page: &page,
			Size: &size,
		})
		if err != nil {
			return err
		} else if resp.JSON200 == nil {
			return httperror.NewHTTPError(resp.StatusCode(), resp.Body)
		}

		for _, itemSchema := range resp.JSON200.Data {
			crafting, err := craftingFromSchema(itemSchema.Craft)
			if err != nil {
				return err
			}

			i.items[itemSchema.Code] = &Item{
				Code:     itemSchema.Code,
				Name:     itemSchema.Name,
				Type:     itemSchema.Type,
				SubType:  itemSchema.Subtype,
				Level:    itemSchema.Level,
				Effects:  itemSchema.Effects,
				Stats:    StatsFromItem(itemSchema),
				Crafting: crafting,
			}
		}

		if len(resp.JSON200.Data) < size {
			break
		}

		page++
	}

	// 2 pass approach to populate crafting items
	for _, item := range i.items {
		if item.Crafting == nil {
			continue
		}

		craftingItems := map[*Item]int{}

		for craftingItem, quantity := range item.Crafting.Items {
			resolvedItem := i.items[craftingItem.Code]
			craftingItems[resolvedItem] = quantity
		}

		item.Crafting.Items = craftingItems
	}

	return nil
}
