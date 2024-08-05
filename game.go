package main

import (
	"context"
	"fmt"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"slices"
	"strings"
)

type Location struct {
	X int
	Y int
}

type Item struct {
	// Code Item code. This is the item's unique identifier (ID).
	Code string

	// Type Item type. In the case of armor, it will be the name of the slot e.g. "boots"
	Type string

	// Level Item level.
	Level int

	// Crafting will be nil if the object cannot be crafted.
	Crafting *ItemCrafting

	// DropsFrom resources that drop this item
	DropsFrom []client.ResourceSchema

	// Effects List of object effects. For equipment, it will include item stats.
	Effects *[]client.ItemEffectSchema

	Attack     map[string]int
	Resistance map[string]int
}

type ItemCrafting struct {
	// Skill required to craft the item
	Skill string

	// Quantity of items crafted
	Quantity int

	// Level The skill level required to craft the item
	Level int

	// Items List of items required to craft the item
	Items []CraftingItem
}

func (c ItemCrafting) TotalQuantity() int {
	totalQuantity := 0
	for _, craftingItem := range c.Items {
		totalQuantity += craftingItem.Quantity
	}
	return totalQuantity
}

type CraftingItem struct {
	Code     string
	Quantity int
	Item     *Item
}

func (i Item) GetAllSkillRequirements() map[string]int {
	skills := map[string]int{}
	if i.Crafting != nil {
		skills[i.Crafting.Skill] = i.Crafting.Level

		for _, craftingItem := range i.Crafting.Items {
			for skill, level := range craftingItem.Item.GetAllSkillRequirements() {
				if level > skills[skill] {
					skills[skill] = level
				}
			}
		}
	}
	return skills
}

type Monster struct {
	client.MonsterSchema
	Attack     map[string]int
	Resistance map[string]int
}

func (m Monster) GetWeaknesses() []string {
	var weaknesses []string
	for element := range m.Resistance {
		weaknesses = append(weaknesses, element)
	}

	slices.SortFunc(weaknesses, func(a, b string) int {
		return m.Resistance[a] - m.Resistance[b]
	})

	return weaknesses
}

type Game struct {
	Client    *client.ClientWithResponses
	Bank      *Bank
	Maps      map[string]map[string][]Location
	Resources []client.ResourceSchema
	Drops     map[string][]client.ResourceSchema
	Items     map[string]*Item
	Monsters  map[string]Monster
}

func NewGame(ctx context.Context, c *client.ClientWithResponses, bank *Bank) (*Game, error) {
	g := &Game{
		Client:    c,
		Bank:      bank,
		Maps:      map[string]map[string][]Location{},
		Resources: []client.ResourceSchema{},
		Drops:     map[string][]client.ResourceSchema{},
		Items:     map[string]*Item{},
		Monsters:  map[string]Monster{},
	}

	if err := g.loadMaps(ctx); err != nil {
		return nil, fmt.Errorf("loading maps: %w", err)
	}

	if err := g.loadResources(ctx); err != nil {
		return nil, fmt.Errorf("loading resources: %w", err)
	}

	if err := g.loadItems(ctx); err != nil {
		return nil, fmt.Errorf("loading items: %w", err)
	}

	if _, err := g.GetBankItems(ctx); err != nil {
		return nil, fmt.Errorf("loading bank items: %w", err)
	}

	if _, err := g.loadMonsters(ctx); err != nil {
		return nil, fmt.Errorf("loading monsters: %w", err)
	}

	return g, nil
}

func (g *Game) GetBankLocations() []Location {
	return g.Maps["bank"]["bank"]
}

func (g *Game) GetResourceLocations(code string) []Location {
	return g.Maps["resource"][code]
}

func (g *Game) GetWorkshopLocations(skill string) []Location {
	return g.Maps["workshop"][skill]
}

func (g *Game) GetMonsterLocations(monster string) []Location {
	return g.Maps["monster"][monster]
}

func (g *Game) GetTaskMasterLocations(taskType string) []Location {
	return g.Maps["tasks_master"][taskType]
}

func (g *Game) GetMonster(monsterCode string) Monster {
	return g.Monsters[monsterCode]
}

func (g *Game) loadMaps(ctx context.Context) error {
	page := 1
	size := 100

	for {
		resp, err := g.Client.GetAllMapsMapsGetWithResponse(ctx, &client.GetAllMapsMapsGetParams{
			Page: &page,
			Size: &size,
		})
		if err != nil {
			return err
		} else if resp.JSON200 == nil {
			return NewHTTPError(resp.StatusCode(), resp.Body)
		}

		for _, tile := range resp.JSON200.Data {
			content, err := tile.Content.AsMapContentSchema()
			if err != nil {
				return err
			}

			if _, ok := g.Maps[content.Type]; !ok {
				g.Maps[content.Type] = map[string][]Location{}
			}

			if _, ok := g.Maps[content.Type][content.Code]; !ok {
				g.Maps[content.Type][content.Code] = []Location{}
			}

			g.Maps[content.Type][content.Code] = append(g.Maps[content.Type][content.Code], Location{X: tile.X, Y: tile.Y})
		}

		if len(resp.JSON200.Data) < size {
			break
		}

		page++
	}

	return nil
}

func (g *Game) loadResources(ctx context.Context) error {
	page := 1
	size := 100

	for {
		resp, err := g.Client.GetAllResourcesResourcesGetWithResponse(ctx, &client.GetAllResourcesResourcesGetParams{
			Page: &page,
			Size: &size,
		})
		if err != nil {
			return err
		} else if resp.JSON200 == nil {
			return NewHTTPError(resp.StatusCode(), resp.Body)
		}

		for _, resource := range resp.JSON200.Data {
			g.Resources = append(g.Resources, resource)

			for _, d := range resource.Drops {
				if _, ok := g.Drops[d.Code]; !ok {
					g.Drops[d.Code] = []client.ResourceSchema{}
				}

				g.Drops[d.Code] = append(g.Drops[d.Code], resource)
			}
		}

		if len(resp.JSON200.Data) < size {
			break
		}

		page++
	}

	return nil
}

func (g *Game) loadItems(ctx context.Context) error {
	page := 1
	size := 100

	for {
		resp, err := g.Client.GetAllItemsItemsGetWithResponse(ctx, &client.GetAllItemsItemsGetParams{
			Page: &page,
			Size: &size,
		})
		if err != nil {
			return err
		} else if resp.JSON200 == nil {
			return NewHTTPError(resp.StatusCode(), resp.Body)
		}

		for _, itemSchema := range resp.JSON200.Data {
			item, err := g.buildItem(itemSchema)
			if err != nil {
				return err
			}
			g.Items[itemSchema.Code] = item
		}

		if len(resp.JSON200.Data) < size {
			break
		}

		page++
	}

	// 2 pass approach
	populateCraftingItems(g.Items)

	return nil
}

func (g *Game) loadMonsters(ctx context.Context) (map[string]Monster, error) {
	page := 1
	size := 100

	for {
		resp, err := g.Client.GetAllMonstersMonstersGetWithResponse(ctx, &client.GetAllMonstersMonstersGetParams{
			Page: &page,
			Size: &size,
		})
		if err != nil {
			return nil, err
		} else if resp.JSON200 == nil {
			return nil, NewHTTPError(resp.StatusCode(), resp.Body)
		}

		for _, monster := range resp.JSON200.Data {
			g.Monsters[monster.Code] = Monster{
				MonsterSchema: monster,
				Attack: map[string]int{
					"air":   monster.AttackAir,
					"earth": monster.AttackEarth,
					"fire":  monster.AttackFire,
					"water": monster.AttackWater,
				},
				Resistance: map[string]int{
					"air":   monster.ResAir,
					"earth": monster.ResEarth,
					"fire":  monster.ResFire,
					"water": monster.ResWater,
				},
			}
		}

		if len(resp.JSON200.Data) < size {
			break
		}

		page++
	}

	return g.Monsters, nil
}

func populateCraftingItems(items map[string]*Item) {
	for _, item := range items {
		if item.Crafting == nil {
			continue
		}

		for i, craftingItem := range item.Crafting.Items {
			item.Crafting.Items[i].Item = items[craftingItem.Code]
		}
	}
}

func (g *Game) buildItem(itemSchema client.ItemSchema) (*Item, error) {
	item := Item{
		Code:       itemSchema.Code,
		Type:       itemSchema.Type,
		Level:      itemSchema.Level,
		DropsFrom:  g.Drops[itemSchema.Code],
		Effects:    itemSchema.Effects,
		Attack:     map[string]int{},
		Resistance: map[string]int{},
	}

	for _, effect := range *itemSchema.Effects {
		switch {
		case effect.Name == "hp":
		case effect.Name == "restore":
		case effect.Name == "haste":
		case effect.Name == "boost_hp":
			// For food
		case effect.Name == "woodcutting":
			// Value is an integer that reduces cooldown time by Value%
			item.Attack[effect.Name] = -effect.Value
		case effect.Name == "fishing":
			// Value is an integer that reduces cooldown time by Value%
			item.Attack[effect.Name] = -effect.Value
		case effect.Name == "mining":
			// Value is an integer that reduces cooldown time by Value%
			item.Attack[effect.Name] = -effect.Value
		case strings.HasPrefix(effect.Name, "dmg_"):
			element := strings.TrimPrefix(effect.Name, "dmg_")
			_ = element
		case strings.HasPrefix(effect.Name, "attack_"):
			element := strings.TrimPrefix(effect.Name, "attack_")
			item.Attack[element] = effect.Value
		case strings.HasPrefix(effect.Name, "res_"):
			element := strings.TrimPrefix(effect.Name, "res_")
			item.Resistance[element] = effect.Value
		case strings.HasPrefix(effect.Name, "boost_dmg_"):
			// Value is a percentage (0-100), comes from food
			element := strings.TrimPrefix(effect.Name, "boost_dmg_")
			_ = element
		default:
			fmt.Println()
		}
	}
	if itemSchema.Craft != nil {
		craft, err := itemSchema.Craft.AsCraftSchema()
		if err != nil {
			return nil, err
		}

		skill, err := craft.Skill.AsCraftSchemaSkill0()
		if err != nil {
			return nil, err
		}

		craftQuantity, err := craft.Quantity.AsCraftSchemaQuantity0()
		if err != nil {
			return nil, err
		}

		craftLevel, err := craft.Level.AsCraftSchemaLevel0()
		if err != nil {
			return nil, err
		}

		item.Crafting = &ItemCrafting{
			Skill:    string(skill),
			Quantity: craftQuantity,
			Level:    craftLevel,
			Items:    nil,
		}

		for _, cr := range *craft.Items {
			item.Crafting.Items = append(item.Crafting.Items, CraftingItem{Code: cr.Code, Quantity: cr.Quantity})
		}
	}

	return &item, nil
}

func (g *Game) GetBankItems(ctx context.Context) ([]client.SimpleItemSchema, error) {
	page := 1
	size := 100

	var bankItems []client.SimpleItemSchema

	for {
		resp, err := g.Client.GetBankItemsMyBankItemsGetWithResponse(ctx, &client.GetBankItemsMyBankItemsGetParams{
			Page: nil,
			Size: nil,
		})
		if err != nil {
			return nil, err
		} else if resp.JSON200 == nil {
			return nil, NewHTTPError(resp.StatusCode(), resp.Body)
		}

		bankItems = append(bankItems, resp.JSON200.Data...)

		if len(resp.JSON200.Data) < size {
			break
		}

		page++
	}

	g.Bank.Update(bankItems)

	return bankItems, nil
}

func (g *Game) GetItemSource(itemCode string) {
	item := g.Items[itemCode]
	drops := g.Drops[itemCode]

	fmt.Println(drops, item)
}

type TrainingMethod struct {
	Item     *Item
	Resource *client.ResourceSchema
	Level    int
}

func (m TrainingMethod) String() string {
	if m.Item == nil {
		return fmt.Sprintf("Crafting %s level %d", m.Item.Code, m.Level)
	} else {
		return fmt.Sprintf("Harvesting %s level %d", m.Resource.Code, m.Level)
	}
}

func (g *Game) NumberCanCraftInBank(crafting ItemCrafting) int {
	bankItems := g.Bank.AsMap()

	haveAllItems := true
	canCraft := 0
	for _, craftingItem := range crafting.Items {
		craftable := bankItems[craftingItem.Code] / craftingItem.Quantity

		if craftable == 0 {
			haveAllItems = false
		} else if canCraft == 0 {
			canCraft = craftable
		} else if craftable < canCraft {
			canCraft = craftable
		}
	}

	if !haveAllItems {
		return 0
	} else {
		return canCraft
	}
}

func craftableQuantity(item *Item, bankItems map[string]int) int {
	if item.Crafting == nil {
		// Harvest only
		return bankItems[item.Code]
	}

	haveItems := true
	canCraft := 0
	for _, craftingItem := range item.Crafting.Items {
		craftable := bankItems[craftingItem.Code] / craftingItem.Quantity

		craftable += craftableQuantity(craftingItem.Item, bankItems)

		if craftable == 0 {
			haveItems = false
		} else if canCraft == 0 {
			canCraft = craftable
		} else if craftable < canCraft {
			canCraft = craftable
		}
	}

	if !haveItems {
		return 0
	} else {
		return canCraft
	}
}

func (m TrainingMethod) Weight(bankItems map[string]int) int {
	weight := m.Level

	if m.Resource != nil {
		// Harvesting gets a static increase
		weight += 5
	} else {
		// Crafting
		// TODO: May want some sort of multiplier here
		weight += craftableQuantity(m.Item, bankItems)
	}

	return weight
}

func (g *Game) GetTrainingMethods(skill string) []TrainingMethod {
	var trainingMethods []TrainingMethod

	// Harvesting training methods
	for _, resource := range g.Resources {
		if string(resource.Skill) != skill {
			continue
		}
		trainingMethods = append(trainingMethods, TrainingMethod{
			Resource: &resource,
			Level:    resource.Level,
		})
	}

	// Crafting training methods
	for _, item := range g.Items {
		if item.Crafting == nil {
			continue
		}
		if item.Crafting.Skill != skill {
			continue
		}
		trainingMethods = append(trainingMethods, TrainingMethod{
			Item:  item,
			Level: item.Level,
		})
	}

	slices.SortFunc(trainingMethods, func(a, b TrainingMethod) int {
		return b.Level - a.Level
	})

	return trainingMethods
}

func (g *Game) BestUsableBankWeapon(charLevel int, resistances map[string]int) (string, int) {
	bankItems := g.Bank.AsMap()

	bestStrength := 0
	bestItem := ""
	// TODO: Would be nice if we could reuse the BestPossibleUsableWeapon logic.
	//  bankItems would need to be a map[string]*Item which probably necessitates a new Game method
	for itemCode := range bankItems {
		item := g.Items[itemCode]
		if item.Level > charLevel {
			continue
		}
		for element, resistance := range resistances {
			strength := item.Attack[element] - resistance
			if strength > bestStrength {
				bestStrength = strength
				bestItem = itemCode
			}
		}
	}

	return bestItem, bestStrength
}

func (g *Game) BestUsableBankArmor(charLevel int, slot string, attacks map[string]int) (string, int) {
	bankItems := g.Bank.AsMap()

	bestStrength := 0
	bestItem := ""
	// TODO: Would be nice if we could reuse the BestPossibleUsableWeapon logic.
	//  bankItems would need to be a map[string]*Item which probably necessitates a new Game method
	for itemCode := range bankItems {
		item := g.Items[itemCode]
		if item.Level > charLevel {
			continue
		}
		if item.Type != slot {
			continue
		}
		for element, attack := range attacks {
			strength := item.Resistance[element] - attack
			if strength > bestStrength {
				bestStrength = strength
				bestItem = itemCode
			}
		}
	}

	return bestItem, bestStrength
}

//func (g *Game) BestPossibleUsableWeapon(charLevel int, weakness string) (string, int) {
//	bestAttack := 0
//	bestItem := ""
//	for _, item := range g.Items {
//		if item.Level > charLevel {
//			continue
//		}
//		value := item.Attack[weakness]
//		if value > bestAttack {
//			bestAttack = value
//			bestItem = item.Code
//		}
//	}
//
//	return bestItem, bestAttack
//}
