package character

import (
	"context"
	"github.com/ahornerr/artifacts/bank"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/httperror"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"math"
	"time"
)

type Character struct {
	client           *client.ClientWithResponses
	bank             *bank.Bank
	Name             string
	character        client.CharacterSchema
	CooldownExpires  time.Time
	CooldownDuration int

	TotalXp   int
	Levels    map[string]int
	Xp        map[string]int
	MaxXp     map[string]int
	Inventory map[string]int
	Equipment map[string]string

	InventoryMaxItems int
	Skin              string

	Task         string
	TaskType     string
	TaskProgress int
	TaskTotal    int

	updates chan<- *Character

	Actions []string
}

func NewCharacter(c *client.ClientWithResponses, bank *bank.Bank, updates chan<- *Character, name string) *Character {
	return &Character{
		client:    c,
		bank:      bank,
		Name:      name,
		Levels:    map[string]int{},
		Xp:        map[string]int{},
		MaxXp:     map[string]int{},
		Inventory: map[string]int{},
		Equipment: map[string]string{},
		updates:   updates,
		Actions:   []string{},
	}
}

func (c *Character) Action(actions []string) {
	c.Actions = actions
}

func (c *Character) update(new client.CharacterSchema) {
	cooldown, err := new.CooldownExpiration.AsCharacterSchemaCooldownExpiration0()
	if err != nil {
		// TODO: Log this?
		c.CooldownExpires = time.Now().Add(time.Duration(new.Cooldown) * time.Second)
	} else {
		c.CooldownExpires = cooldown
	}

	c.CooldownDuration = new.Cooldown
	c.character = new

	c.TotalXp = c.character.TotalXp
	c.InventoryMaxItems = c.character.InventoryMaxItems
	c.Skin = string(c.character.Skin)

	c.Levels["combat"] = c.character.Level
	c.Xp["combat"] = c.character.Xp
	c.MaxXp["combat"] = c.character.MaxXp

	c.Levels["cooking"] = c.character.CookingLevel
	c.Xp["cooking"] = c.character.CookingXp
	c.MaxXp["cooking"] = c.character.CookingMaxXp

	c.Levels["fishing"] = c.character.FishingLevel
	c.Xp["fishing"] = c.character.FishingXp
	c.MaxXp["fishing"] = c.character.FishingMaxXp

	c.Levels["gearcrafting"] = c.character.GearcraftingLevel
	c.Xp["gearcrafting"] = c.character.GearcraftingXp
	c.MaxXp["gearcrafting"] = c.character.GearcraftingMaxXp

	c.Levels["jewelrycrafting"] = c.character.JewelrycraftingLevel
	c.Xp["jewelrycrafting"] = c.character.JewelrycraftingXp
	c.MaxXp["jewelrycrafting"] = c.character.JewelrycraftingMaxXp

	c.Levels["mining"] = c.character.MiningLevel
	c.Xp["mining"] = c.character.MiningXp
	c.MaxXp["mining"] = c.character.MiningMaxXp

	c.Levels["weaponcrafting"] = c.character.WeaponcraftingLevel
	c.Xp["weaponcrafting"] = c.character.WeaponcraftingXp
	c.MaxXp["weaponcrafting"] = c.character.WeaponcraftingMaxXp

	c.Levels["woodcutting"] = c.character.WoodcuttingLevel
	c.Xp["woodcutting"] = c.character.WoodcuttingXp
	c.MaxXp["woodcutting"] = c.character.WoodcuttingMaxXp

	c.Inventory = map[string]int{}
	for _, item := range *c.character.Inventory {
		if item.Code != "" {
			c.Inventory[item.Code] += item.Quantity
		}
	}

	c.Equipment = c.GetEquippedItems()

	c.Task = c.character.Task
	c.TaskType = c.character.TaskType
	c.TaskProgress = c.character.TaskProgress
	c.TaskTotal = c.character.TaskTotal

	c.updates <- c
}

func (c *Character) GetPosition() (int, int) {
	return c.character.X, c.character.Y
}

func (c *Character) WaitForCooldown() {
	until := time.Until(c.CooldownExpires)
	if until > 0 {
		//c.Logger.Printf("Waiting %0.1f seconds for cooldown\n", until.Seconds())
	}
	time.Sleep(until)
}

func (c *Character) InventoryQuantity(code string) int {
	for _, inv := range *c.character.Inventory {
		if inv.Code == code {
			return inv.Quantity
		}
	}
	return 0
}

func (c *Character) MaxInventoryItems() int {
	return c.character.InventoryMaxItems
}

func (c *Character) InventoryAsMap() map[string]int {
	itemsMap := map[string]int{}
	for _, item := range *c.character.Inventory {
		itemsMap[item.Code] += item.Quantity
	}
	return itemsMap
}

func (c *Character) Bank() map[string]int {
	return c.bank.Items
}

func (c *Character) InventoryCount() int {
	count := 0
	for _, inv := range *c.character.Inventory {
		count += inv.Quantity
	}
	return count
}

func (c *Character) IsInventoryFull() bool {
	return c.InventoryCount() == c.character.InventoryMaxItems
}

func (c *Character) IsInventoryEmpty() bool {
	return c.InventoryCount() == 0
}

func (c *Character) IsAtOneOf(locations []game.Location) bool {
	for _, loc := range locations {
		if c.character.X == loc.X && c.character.Y == loc.Y {
			return true
		}
	}
	return false
}

func (c *Character) GetLevel(skill string) int {
	switch skill {
	case "combat":
		return c.character.Level
	case "cooking":
		return c.character.CookingLevel
	case "fishing":
		return c.character.FishingLevel
	case "gearcrafting":
		return c.character.GearcraftingLevel
	case "jewelrycrafting":
		return c.character.JewelrycraftingLevel
	case "mining":
		return c.character.MiningLevel
	case "weaponcrafting":
		return c.character.WeaponcraftingLevel
	case "woodcutting":
		return c.character.WoodcuttingLevel
	default:
		return 0
	}
}

func (c *Character) GetXP(skill string) int {
	switch skill {
	case "combat":
		return c.character.Xp
	case "cooking":
		return c.character.CookingXp
	case "fishing":
		return c.character.FishingXp
	case "gearcrafting":
		return c.character.GearcraftingXp
	case "jewelrycrafting":
		return c.character.JewelrycraftingXp
	case "mining":
		return c.character.MiningXp
	case "weaponcrafting":
		return c.character.WeaponcraftingXp
	case "woodcutting":
		return c.character.WoodcuttingXp
	default:
		return 0
	}
}

func (c *Character) GetMaxXP(skill string) int {
	switch skill {
	case "combat":
		return c.character.MaxXp
	case "cooking":
		return c.character.CookingMaxXp
	case "fishing":
		return c.character.FishingMaxXp
	case "gearcrafting":
		return c.character.GearcraftingMaxXp
	case "jewelrycrafting":
		return c.character.JewelrycraftingMaxXp
	case "mining":
		return c.character.MiningMaxXp
	case "weaponcrafting":
		return c.character.WeaponcraftingMaxXp
	case "woodcutting":
		return c.character.WoodcuttingMaxXp
	default:
		return 0
	}
}

func (c *Character) GetPercentToLevel(skill string) float64 {
	return float64(c.GetXP(skill)) / float64(c.GetMaxXP(skill))
}

func (c *Character) Get(ctx context.Context) (*client.CharacterResponseSchema, error) {
	resp, err := c.client.GetCharacterCharactersNameGetWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data)

	return resp.JSON200, nil
}

func (c *Character) Move(ctx context.Context, x, y int) error {
	if c.character.X == x && c.character.Y == y {
		return nil
	}

	c.WaitForCooldown()

	resp, err := c.client.ActionMoveMyNameActionMovePostWithResponse(
		ctx,
		c.Name,
		client.ActionMoveMyNameActionMovePostJSONRequestBody{X: x, Y: y},
	)

	if err != nil {
		return err
	} else if resp.JSON200 == nil {
		httpError := httperror.NewHTTPError(resp.StatusCode(), resp.Body)

		// TODO: handle other cases
		switch httpError.Code {
		case 486: // Character is locked. Action is already in progress.
			c.WaitForCooldown()
			time.Sleep(5 * time.Second)
			return nil
		case 490: // Character already at destination
			time.Sleep(5 * time.Second) // To avoid 486 character is locked. May not be working properly
			return nil
		case 499: // Character in cooldown
			c.WaitForCooldown()
			return nil
		}

		return httpError
	}

	// TODO: Can we use reflection or add a method to use with an interface to the oapi models?
	c.update(resp.JSON200.Data.Character)

	return nil
}

func (c *Character) Fight(ctx context.Context) (*client.FightSchema, error) {
	c.WaitForCooldown()

	resp, err := c.client.ActionFightMyNameActionFightPostWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character)

	return &resp.JSON200.Data.Fight, nil
}

func (c *Character) Gather(ctx context.Context) (*client.SkillInfoSchema, error) {
	c.WaitForCooldown()

	resp, err := c.client.ActionGatheringMyNameActionGatheringPostWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character)

	return &resp.JSON200.Data.Details, nil
}

func (c *Character) Craft(ctx context.Context, code string, quantity int) (*client.SkillInfoSchema, error) {
	c.WaitForCooldown()

	resp, err := c.client.ActionCraftingMyNameActionCraftingPostWithResponse(ctx, c.Name, client.ActionCraftingMyNameActionCraftingPostJSONRequestBody{
		Code:     code,
		Quantity: &quantity,
	})
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character)

	return &resp.JSON200.Data.Details, nil
}

func (c *Character) DepositBank(ctx context.Context, code string, quantity int) ([]client.SimpleItemSchema, error) {
	c.WaitForCooldown()

	resp, err := c.client.ActionDepositBankMyNameActionBankDepositPostWithResponse(ctx, c.Name, client.ActionDepositBankMyNameActionBankDepositPostJSONRequestBody{
		Code:     code,
		Quantity: quantity,
	})
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.bank.Update(resp.JSON200.Data.Bank)
	c.update(resp.JSON200.Data.Character)

	return resp.JSON200.Data.Bank, nil
}

func (c *Character) WithdrawBank(ctx context.Context, code string, quantity int) ([]client.SimpleItemSchema, error) {
	c.WaitForCooldown()

	resp, err := c.client.ActionWithdrawBankMyNameActionBankWithdrawPostWithResponse(ctx, c.Name, client.ActionWithdrawBankMyNameActionBankWithdrawPostJSONRequestBody{
		Code:     code,
		Quantity: quantity,
	})
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.bank.Update(resp.JSON200.Data.Bank)
	c.update(resp.JSON200.Data.Character)

	return resp.JSON200.Data.Bank, nil
}

func (c *Character) MoveClosest(ctx context.Context, locations []game.Location) error {
	if c.IsAtOneOf(locations) {
		return nil
	}

	location := c.ClosestOf(locations)
	return c.Move(ctx, location.X, location.Y)
}

func (c *Character) ClosestOf(locations []game.Location) game.Location {
	if len(locations) == 1 {
		return locations[0]
	}

	currentX := c.character.X
	currentY := c.character.Y

	closestLocation := locations[0]
	var closestDistance float64
	for _, location := range locations {
		distance := math.Sqrt(math.Pow(float64(location.X-currentX), 2) + math.Pow(float64(location.Y-currentY), 2))
		if distance < closestDistance {
			closestLocation = location
			closestDistance = distance
		}
	}

	return closestLocation
}

func (c *Character) BankAll(ctx context.Context) error {
	for _, inv := range *c.character.Inventory {
		if inv.Quantity <= 0 {
			continue
		}
		_, err := c.DepositBank(ctx, inv.Code, inv.Quantity)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Character) NewTask(ctx context.Context) (*client.TaskSchema, error) {
	c.WaitForCooldown()

	resp, err := c.client.ActionAcceptNewTaskMyNameActionTaskNewPostWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character)

	return &resp.JSON200.Data.Task, nil
}

func (c *Character) CompleteTask(ctx context.Context) (*client.TaskRewardSchema, error) {
	c.WaitForCooldown()

	resp, err := c.client.ActionCompleteTaskMyNameActionTaskCompletePostWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character)

	return &resp.JSON200.Data.Reward, nil
}

func (c *Character) Unequip(ctx context.Context, slot client.UnequipSchemaSlot) error {
	c.WaitForCooldown()

	resp, err := c.client.ActionUnequipItemMyNameActionUnequipPostWithResponse(ctx, c.Name, client.ActionUnequipItemMyNameActionUnequipPostJSONRequestBody{
		Slot: slot,
	})
	if err != nil {
		return err
	} else if resp.JSON200 == nil {
		return httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character)

	return nil
}

func (c *Character) Equip(ctx context.Context, slot client.EquipSchemaSlot, itemCode string) error {
	c.WaitForCooldown()

	resp, err := c.client.ActionEquipItemMyNameActionEquipPostWithResponse(ctx, c.Name, client.ActionEquipItemMyNameActionEquipPostJSONRequestBody{
		Slot: slot,
		Code: itemCode,
	})
	if err != nil {
		return err
	} else if resp.JSON200 == nil {
		return httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character)

	return nil
}

func (c *Character) GetEquippedItems() map[string]string {
	return map[string]string{
		// TODO: Add artifact1-3, consumable1-2
		"amulet":     c.character.AmuletSlot,
		"body_armor": c.character.BodyArmorSlot,
		"boots":      c.character.BootsSlot,
		"helmet":     c.character.HelmetSlot,
		"leg_armor":  c.character.LegArmorSlot,
		"ring1":      c.character.Ring1Slot,
		"ring2":      c.character.Ring2Slot,
		"shield":     c.character.ShieldSlot,
		"weapon":     c.character.WeaponSlot,
	}
}

func (c *Character) GetEquipmentStats() map[string]*game.Stats {
	stats := map[string]*game.Stats{}
	for slot, itemCode := range c.GetEquippedItems() {
		if itemCode != "" {
			stats[slot] = &game.Items.Get(itemCode).Stats
		}
	}
	return stats
}

func (c *Character) GetEquipmentUpgradesInBank(targetStats game.Stats) map[string]string {
	upgrades := map[string]string{}
	equipment := c.GetEquipmentStats()

	for slot, equippedItem := range c.GetEquippedItems() {
		equipmentStats := equipment[slot]

		bestStrength := -math.MaxFloat64
		bestItem := ""
		if equipmentStats != nil {
			bestStrength = getStrengthAgainst(targetStats, *equipmentStats)
			bestItem = equippedItem
		}

		// TODO: Check quantity because of ring1 and ring2
		for itemCode := range c.bank.Items {
			item := game.Items.Get(itemCode)
			switch {
			case item.Type == slot:
			case item.Type == "ring" && slot == "ring1":
			case item.Type == "ring" && slot == "ring2":
			default:
				continue
			}
			if item.Level > c.character.Level {
				continue
			}
			strength := getStrengthAgainst(targetStats, item.Stats)
			if strength > bestStrength {
				bestStrength = strength
				bestItem = itemCode
			}
		}

		if bestItem == "" || bestItem == equippedItem {
			continue
		}

		upgrades[slot] = bestItem
	}

	return upgrades
}

// https://docs.artifactsmmo.com/concepts/stats_and_fights
func getStrengthAgainst(target game.Stats, equipment game.Stats) float64 {
	damageFromEquipment := equipment.GetDamageAgainst(target)
	damageFromTarget := target.GetDamageAgainst(equipment)

	return damageFromEquipment - damageFromTarget
}
