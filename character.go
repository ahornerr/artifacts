package main

import (
	"context"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"log"
	"math"
	"os"
	"time"
)

type Character struct {
	Client           *client.ClientWithResponses
	bank             *Bank
	Name             string
	Character        client.CharacterSchema
	Logger           *log.Logger
	CooldownExpires  time.Time
	CooldownDuration int
}

func NewCharacter(c *client.ClientWithResponses, bank *Bank, name string) *Character {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	logger.SetPrefix("[" + name + "] ")

	return &Character{
		Client: c,
		bank:   bank,
		Name:   name,
		Logger: logger,
	}
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
	c.Character = new
}

func (c *Character) GetPosition() (int, int) {
	return c.Character.X, c.Character.Y
}

func (c *Character) WaitForCooldown() {
	until := time.Until(c.CooldownExpires)
	if until > 0 {
		//c.Logger.Printf("Waiting %0.1f seconds for cooldown\n", until.Seconds())
	}
	time.Sleep(until)
}

func (c *Character) InventoryQuantity(code string) int {
	for _, inv := range *c.Character.Inventory {
		if inv.Code == code {
			return inv.Quantity
		}
	}
	return 0
}

func (c *Character) InventoryCount() int {
	count := 0
	for _, inv := range *c.Character.Inventory {
		count += inv.Quantity
	}
	return count
}

func (c *Character) IsInventoryFull() bool {
	return c.InventoryCount() == c.Character.InventoryMaxItems
}

func (c *Character) IsAtOneOf(locations []Location) bool {
	for _, loc := range locations {
		if c.Character.X == loc.X && c.Character.Y == loc.Y {
			return true
		}
	}
	return false
}

func (c *Character) GetLevel(skill string) int {
	switch skill {
	case "combat":
		return c.Character.Level
	case "cooking":
		return c.Character.CookingLevel
	case "fishing":
		return c.Character.FishingLevel
	case "gearcrafting":
		return c.Character.GearcraftingLevel
	case "jewelrycrafting":
		return c.Character.JewelrycraftingLevel
	case "mining":
		return c.Character.MiningLevel
	case "weaponcrafting":
		return c.Character.WeaponcraftingLevel
	case "woodcutting":
		return c.Character.WoodcuttingLevel
	default:
		return 0
	}
}

func (c *Character) GetXP(skill string) int {
	switch skill {
	case "combat":
		return c.Character.Xp
	case "cooking":
		return c.Character.CookingXp
	case "fishing":
		return c.Character.FishingXp
	case "gearcrafting":
		return c.Character.GearcraftingXp
	case "jewelrycrafting":
		return c.Character.JewelrycraftingXp
	case "mining":
		return c.Character.MiningXp
	case "weaponcrafting":
		return c.Character.WeaponcraftingXp
	case "woodcutting":
		return c.Character.WoodcuttingXp
	default:
		return 0
	}
}

func (c *Character) GetMaxXP(skill string) int {
	switch skill {
	case "combat":
		return c.Character.MaxXp
	case "cooking":
		return c.Character.CookingMaxXp
	case "fishing":
		return c.Character.FishingMaxXp
	case "gearcrafting":
		return c.Character.GearcraftingMaxXp
	case "jewelrycrafting":
		return c.Character.JewelrycraftingMaxXp
	case "mining":
		return c.Character.MiningMaxXp
	case "weaponcrafting":
		return c.Character.WeaponcraftingMaxXp
	case "woodcutting":
		return c.Character.WoodcuttingMaxXp
	default:
		return 0
	}
}

func (c *Character) GetPercentToLevel(skill string) float64 {
	return float64(c.GetXP(skill)) / float64(c.GetMaxXP(skill))
}

func (c *Character) Get(ctx context.Context) (*client.CharacterResponseSchema, error) {
	resp, err := c.Client.GetCharacterCharactersNameGetWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data)

	return resp.JSON200, nil
}

func (c *Character) Move(ctx context.Context, x, y int) error {
	if c.Character.X == x && c.Character.Y == y {
		return nil
	}

	c.WaitForCooldown()

	resp, err := c.Client.ActionMoveMyNameActionMovePostWithResponse(
		ctx,
		c.Name,
		client.ActionMoveMyNameActionMovePostJSONRequestBody{X: x, Y: y},
	)

	if err != nil {
		return err
	} else if resp.JSON200 == nil {
		httpError := NewHTTPError(resp.StatusCode(), resp.Body)

		// TODO: handle other cases
		switch httpError.Code {
		case 486: // Character is locked. Action is already in progress.
			c.Logger.Println("[!] Got 'Action is already in progress' error. Cooldown:", c.CooldownExpires)
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

	resp, err := c.Client.ActionFightMyNameActionFightPostWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character)

	return &resp.JSON200.Data.Fight, nil
}

func (c *Character) Gather(ctx context.Context) (*client.SkillInfoSchema, error) {
	c.WaitForCooldown()

	resp, err := c.Client.ActionGatheringMyNameActionGatheringPostWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character)

	return &resp.JSON200.Data.Details, nil
}

func (c *Character) Craft(ctx context.Context, code string, quantity int) (*client.SkillInfoSchema, error) {
	c.WaitForCooldown()

	resp, err := c.Client.ActionCraftingMyNameActionCraftingPostWithResponse(ctx, c.Name, client.ActionCraftingMyNameActionCraftingPostJSONRequestBody{
		Code:     code,
		Quantity: &quantity,
	})
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character)

	return &resp.JSON200.Data.Details, nil
}

func (c *Character) DepositBank(ctx context.Context, code string, quantity int) ([]client.SimpleItemSchema, error) {
	c.WaitForCooldown()

	resp, err := c.Client.ActionDepositBankMyNameActionBankDepositPostWithResponse(ctx, c.Name, client.ActionDepositBankMyNameActionBankDepositPostJSONRequestBody{
		Code:     code,
		Quantity: quantity,
	})
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.bank.Update(resp.JSON200.Data.Bank)
	c.update(resp.JSON200.Data.Character)

	return resp.JSON200.Data.Bank, nil
}

func (c *Character) WithdrawBank(ctx context.Context, code string, quantity int) ([]client.SimpleItemSchema, error) {
	c.WaitForCooldown()

	resp, err := c.Client.ActionWithdrawBankMyNameActionBankWithdrawPostWithResponse(ctx, c.Name, client.ActionWithdrawBankMyNameActionBankWithdrawPostJSONRequestBody{
		Code:     code,
		Quantity: quantity,
	})
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.bank.Update(resp.JSON200.Data.Bank)
	c.update(resp.JSON200.Data.Character)

	return resp.JSON200.Data.Bank, nil
}

func (c *Character) MoveClosest(ctx context.Context, locations []Location) error {
	if c.IsAtOneOf(locations) {
		return nil
	}

	location := c.ClosestOf(locations)
	return c.Move(ctx, location.X, location.Y)
}

func (c *Character) ClosestOf(locations []Location) Location {
	if len(locations) == 1 {
		return locations[0]
	}

	currentX := c.Character.X
	currentY := c.Character.Y

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
	for _, inv := range *c.Character.Inventory {
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

	resp, err := c.Client.ActionAcceptNewTaskMyNameActionTaskNewPostWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character)

	return &resp.JSON200.Data.Task, nil
}

func (c *Character) CompleteTask(ctx context.Context) (*client.TaskRewardSchema, error) {
	c.WaitForCooldown()

	resp, err := c.Client.ActionCompleteTaskMyNameActionTaskCompletePostWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character)

	return &resp.JSON200.Data.Reward, nil
}

func (c *Character) NumberCanCraftInInventory(crafting ItemCrafting) int {
	haveAllItems := true
	canCraft := 0
	for _, craftingItem := range crafting.Items {
		craftable := c.InventoryQuantity(craftingItem.Code) / craftingItem.Quantity

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

func (c *Character) Unequip(ctx context.Context, slot client.UnequipSchemaSlot) error {
	c.WaitForCooldown()

	resp, err := c.Client.ActionUnequipItemMyNameActionUnequipPostWithResponse(ctx, c.Name, client.ActionUnequipItemMyNameActionUnequipPostJSONRequestBody{
		Slot: slot,
	})
	if err != nil {
		return err
	} else if resp.JSON200 == nil {
		return NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character)

	return nil
}

func (c *Character) Equip(ctx context.Context, slot client.EquipSchemaSlot, itemCode string) error {
	c.WaitForCooldown()

	resp, err := c.Client.ActionEquipItemMyNameActionEquipPostWithResponse(ctx, c.Name, client.ActionEquipItemMyNameActionEquipPostJSONRequestBody{
		Slot: slot,
		Code: itemCode,
	})
	if err != nil {
		return err
	} else if resp.JSON200 == nil {
		return NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character)

	return nil
}
