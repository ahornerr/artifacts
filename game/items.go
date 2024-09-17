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
	//bestAttackItems map[string][]*Item
	//bestResistItems map[string][]*Item
	//bestDamageItems map[string][]*Item
}

func newItems(c *client.ClientWithResponses) *items {
	return &items{
		client: c,
		items:  map[string]*Item{},
		//bestAttackItems: map[string][]*Item{},
		//bestResistItems: map[string][]*Item{},
		//bestDamageItems: map[string][]*Item{},
	}
}

func (i *items) GetAll() map[string]*Item {
	return i.items
}

func (i *items) Get(itemCode string) *Item {
	return i.items[itemCode]
}

func (i *items) ForSkill(skill string, charLevel int) []*Item {
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

//func (i *items) BestAttack(element string) []*Item {
//	return i.bestAttackItems[element]
//}
//
//func (i *items) BestResist(element string) []*Item {
//	return i.bestResistItems[element]
//}
//
//func (i *items) BestDamage(element string) []*Item {
//	return i.bestDamageItems[element]
//}

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
				Code:    itemSchema.Code,
				Name:    itemSchema.Name,
				Type:    itemSchema.Type,
				SubType: itemSchema.Subtype,
				Level:   itemSchema.Level,
				//DropsFrom: g.ResourcesForItem[itemSchema.Code], // TODO: Best way to populate this?
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

	//for _, item := range i.items {
	//	for element, attack := range item.Stats.Attack {
	//		if attack == 0 {
	//			continue
	//		}
	//		if i.bestAttackItems[element] == nil {
	//			i.bestAttackItems[element] = []*Item{}
	//		}
	//		i.bestAttackItems[element] = append(i.bestAttackItems[element], item)
	//	}
	//	for element, resistance := range item.Stats.Resistance {
	//		if resistance == 0 {
	//			continue
	//		}
	//		if i.bestResistItems[element] == nil {
	//			i.bestResistItems[element] = []*Item{}
	//		}
	//		i.bestResistItems[element] = append(i.bestResistItems[element], item)
	//	}
	//	for element, damage := range item.Stats.Damage {
	//		if damage == 0 {
	//			continue
	//		}
	//		if i.bestDamageItems[element] == nil {
	//			i.bestDamageItems[element] = []*Item{}
	//		}
	//		i.bestDamageItems[element] = append(i.bestDamageItems[element], item)
	//	}
	//}
	//
	//for element := range i.bestAttackItems {
	//	slices.SortFunc(i.bestAttackItems[element], func(a, b *Item) int {
	//		return b.Stats.Attack[element] - a.Stats.Attack[element]
	//	})
	//}
	//
	//for element := range i.bestResistItems {
	//	slices.SortFunc(i.bestResistItems[element], func(a, b *Item) int {
	//		return b.Stats.Resistance[element] - a.Stats.Resistance[element]
	//	})
	//}
	//
	//for element := range i.bestDamageItems {
	//	slices.SortFunc(i.bestDamageItems[element], func(a, b *Item) int {
	//		return b.Stats.Damage[element] - a.Stats.Damage[element]
	//	})
	//}

	return nil
}
