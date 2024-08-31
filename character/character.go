package character

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/bank"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/httperror"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"math"
	"slices"
	"time"
)

type Character struct {
	client           *client.ClientWithResponses
	bank             *bank.Bank
	Name             string
	CooldownExpires  time.Time
	CooldownDuration int

	Location game.Location

	Levels    map[string]int
	Xp        map[string]int
	MaxXp     map[string]int
	Inventory map[string]int
	Equipment map[string]string

	InventoryMaxItems int
	Skin              string
	Gold              int

	Task         string
	TaskType     string
	TaskProgress int
	TaskTotal    int

	updates chan<- *Character

	State []string
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
		State:     []string{},
	}
}

func (c *Character) PushState(stateFmt string, args ...interface{}) {
	state := fmt.Sprintf(stateFmt, args...)
	//log.Printf("Pushing %q\n", state)
	c.State = append(c.State, state)
	c.updates <- c
}

func (c *Character) PopState() {
	if len(c.State) > 0 {
		//log.Printf("Popping %q\n", c.State[len(c.State)-1])
		c.State = c.State[:len(c.State)-1]
		c.updates <- c
	}
}

func (c *Character) update(char client.CharacterSchema, waitForCooldown bool) {
	cooldown := char.CooldownExpiration
	if cooldown == nil {
		c.CooldownExpires = time.Time{}
	} else {
		c.CooldownExpires = *cooldown
	}

	c.CooldownDuration = char.Cooldown

	c.Location = game.Location{
		X: char.X,
		Y: char.Y,
	}

	c.InventoryMaxItems = char.InventoryMaxItems
	c.Skin = string(char.Skin)
	c.Gold = char.Gold

	c.Levels["combat"] = char.Level
	c.Xp["combat"] = char.Xp
	c.MaxXp["combat"] = char.MaxXp

	c.Levels["cooking"] = char.CookingLevel
	c.Xp["cooking"] = char.CookingXp
	c.MaxXp["cooking"] = char.CookingMaxXp

	c.Levels["fishing"] = char.FishingLevel
	c.Xp["fishing"] = char.FishingXp
	c.MaxXp["fishing"] = char.FishingMaxXp

	c.Levels["gearcrafting"] = char.GearcraftingLevel
	c.Xp["gearcrafting"] = char.GearcraftingXp
	c.MaxXp["gearcrafting"] = char.GearcraftingMaxXp

	c.Levels["jewelrycrafting"] = char.JewelrycraftingLevel
	c.Xp["jewelrycrafting"] = char.JewelrycraftingXp
	c.MaxXp["jewelrycrafting"] = char.JewelrycraftingMaxXp

	c.Levels["mining"] = char.MiningLevel
	c.Xp["mining"] = char.MiningXp
	c.MaxXp["mining"] = char.MiningMaxXp

	c.Levels["weaponcrafting"] = char.WeaponcraftingLevel
	c.Xp["weaponcrafting"] = char.WeaponcraftingXp
	c.MaxXp["weaponcrafting"] = char.WeaponcraftingMaxXp

	c.Levels["woodcutting"] = char.WoodcuttingLevel
	c.Xp["woodcutting"] = char.WoodcuttingXp
	c.MaxXp["woodcutting"] = char.WoodcuttingMaxXp

	c.Inventory = map[string]int{}
	for _, item := range *char.Inventory {
		if item.Code != "" {
			c.Inventory[item.Code] += item.Quantity
		}
	}

	c.Equipment = map[string]string{
		// TODO: Add artifact1-3, consumable1-2
		"amulet":     char.AmuletSlot,
		"body_armor": char.BodyArmorSlot,
		"boots":      char.BootsSlot,
		"helmet":     char.HelmetSlot,
		"leg_armor":  char.LegArmorSlot,
		"ring1":      char.Ring1Slot,
		"ring2":      char.Ring2Slot,
		"shield":     char.ShieldSlot,
		"weapon":     char.WeaponSlot,
	}

	c.Task = char.Task
	c.TaskType = char.TaskType
	c.TaskProgress = char.TaskProgress
	c.TaskTotal = char.TaskTotal

	c.updates <- c

	// Wait for cooldown
	if waitForCooldown {
		time.Sleep(time.Until(c.CooldownExpires))
	}
}

func (c *Character) MaxInventoryItems() int {
	return c.InventoryMaxItems
}

func (c *Character) Bank() map[string]int {
	return c.bank.Items
}

func (c *Character) InventoryCount() int {
	count := 0
	for _, quantity := range c.Inventory {
		count += quantity
	}
	return count
}

func (c *Character) IsInventoryFull() bool {
	return c.InventoryCount() == c.InventoryMaxItems
}

func (c *Character) IsInventoryEmpty() bool {
	return c.InventoryCount() == 0
}

func (c *Character) IsAtOneOf(locations []game.Location) bool {
	for _, loc := range locations {
		if c.Location.X == loc.X && c.Location.Y == loc.Y {
			return true
		}
	}
	return false
}

func (c *Character) GetLevel(skill string) int {
	return c.Levels[skill]
}

func (c *Character) GetXP(skill string) int {
	return c.Xp[skill]
}

func (c *Character) GetMaxXP(skill string) int {
	return c.MaxXp[skill]
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

	c.update(resp.JSON200.Data, false)

	return resp.JSON200, nil
}

func (c *Character) Move(ctx context.Context, x, y int) error {
	if c.Location.X == x && c.Location.Y == y {
		return nil
	}

	resp, err := c.client.ActionMoveMyNameActionMovePostWithResponse(
		ctx,
		c.Name,
		client.ActionMoveMyNameActionMovePostJSONRequestBody{X: x, Y: y},
	)

	if err != nil {
		return err
	} else if resp.JSON200 == nil {
		return httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character, true)

	return nil
}

func (c *Character) Fight(ctx context.Context) (*client.FightSchema, error) {
	resp, err := c.client.ActionFightMyNameActionFightPostWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	result := &resp.JSON200.Data.Fight

	c.PushState(result.Logs[len(result.Logs)-1])
	defer c.PopState()

	//if result.Result == client.Win {
	//	c.PushState("Fight won (enemy HP %d)", result.Result)
	//	defer c.PopState()
	//} else if result.Result == client.Lose {
	//	c.PushState("Fight lost")
	//	defer c.PopState()
	//}

	c.update(resp.JSON200.Data.Character, true)

	return result, nil
}

func (c *Character) Gather(ctx context.Context) (*client.SkillInfoSchema, error) {
	c.PushState("Gathering")
	defer c.PopState()

	resp, err := c.client.ActionGatheringMyNameActionGatheringPostWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character, true)

	return &resp.JSON200.Data.Details, nil
}

func (c *Character) Craft(ctx context.Context, code string, quantity int) (*client.SkillInfoSchema, error) {
	c.PushState("Crafting %d %s", quantity, code)
	defer c.PopState()

	resp, err := c.client.ActionCraftingMyNameActionCraftingPostWithResponse(ctx, c.Name, client.ActionCraftingMyNameActionCraftingPostJSONRequestBody{
		Code:     code,
		Quantity: &quantity,
	})
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character, true)

	return &resp.JSON200.Data.Details, nil
}

func (c *Character) DepositBank(ctx context.Context, code string, quantity int) ([]client.SimpleItemSchema, error) {
	c.PushState("Depositing %d %s", quantity, code)
	defer c.PopState()

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
	c.update(resp.JSON200.Data.Character, true)

	return resp.JSON200.Data.Bank, nil
}

func (c *Character) WithdrawBank(ctx context.Context, code string, quantity int) ([]client.SimpleItemSchema, error) {
	c.PushState("Withdrawing %d %s", quantity, code)
	defer c.PopState()

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
	c.update(resp.JSON200.Data.Character, true)

	return resp.JSON200.Data.Bank, nil
}

func (c *Character) MoveClosest(ctx context.Context, locations []game.Location) error {
	if c.IsAtOneOf(locations) {
		return nil
	}

	location, _ := c.ClosestOf(locations)
	return c.Move(ctx, location.X, location.Y)
}

func (c *Character) ClosestOf(locations []game.Location) (game.Location, float64) {
	closestLocation := locations[0]
	closestDistance := c.Location.DistanceTo(closestLocation)

	if len(locations) == 1 {
		return closestLocation, closestDistance
	}

	for _, location := range locations {
		distance := c.Location.DistanceTo(location)
		if distance < closestDistance {
			closestLocation = location
			closestDistance = distance
		}
	}

	return closestLocation, closestDistance
}

func (c *Character) DepositAll(ctx context.Context) error {
	for itemCode, quantity := range c.Inventory {
		_, err := c.DepositBank(ctx, itemCode, quantity)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Character) NewTask(ctx context.Context) (*client.TaskSchema, error) {
	c.PushState("Getting new task")
	defer c.PopState()

	resp, err := c.client.ActionAcceptNewTaskMyNameActionTaskNewPostWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character, true)

	return &resp.JSON200.Data.Task, nil
}

func (c *Character) CompleteTask(ctx context.Context) (*client.TaskRewardSchema, error) {
	c.PushState("Completing task")
	defer c.PopState()

	resp, err := c.client.ActionCompleteTaskMyNameActionTaskCompletePostWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character, true)

	return &resp.JSON200.Data.Reward, nil
}

func (c *Character) Unequip(ctx context.Context, slot client.UnequipSchemaSlot) error {
	c.PushState("Unequipping %s", string(slot))
	defer c.PopState()

	resp, err := c.client.ActionUnequipItemMyNameActionUnequipPostWithResponse(ctx, c.Name, client.ActionUnequipItemMyNameActionUnequipPostJSONRequestBody{
		Slot: slot,
	})
	if err != nil {
		return err
	} else if resp.JSON200 == nil {
		return httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character, true)

	return nil
}

func (c *Character) Equip(ctx context.Context, slot client.EquipSchemaSlot, itemCode string) error {
	c.PushState("Equipping %s in %s", itemCode, string(slot))
	defer c.PopState()

	resp, err := c.client.ActionEquipItemMyNameActionEquipPostWithResponse(ctx, c.Name, client.ActionEquipItemMyNameActionEquipPostJSONRequestBody{
		Slot: slot,
		Code: itemCode,
	})
	if err != nil {
		return err
	} else if resp.JSON200 == nil {
		return httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(resp.JSON200.Data.Character, true)

	return nil
}

var equipmentTypes = map[string]bool{
	"amulet":     true,
	"body_armor": true,
	"boots":      true,
	"helmet":     true,
	"leg_armor":  true,
	"ring":       true,
	"shield":     true,
	"weapon":     true,
}

type EquipmentSet struct {
	Equipment          map[string]*game.Item
	TurnsToKillMonster int
	TurnsToKillPlayer  int
	Haste              int
}

func NewEquipmentSet(other *EquipmentSet) *EquipmentSet {
	s := &EquipmentSet{
		Equipment: map[string]*game.Item{},
	}

	if other != nil {
		// Copy best set to a new set, might not be the most efficient thing
		for slot, item := range other.Equipment {
			s.Equipment[slot] = item
		}
	}

	return s
}

func (c *Character) GetEquipmentUpgrades() ([]*game.Item, []*game.Item) {
	var withinLevel []*game.Item
	var aboveLevel []*game.Item
	for _, item := range game.Items.GetAll() {
		if _, ok := equipmentTypes[item.Type]; !ok {
			continue
		}
		if item.Crafting == nil {
			continue
		}
		if item.Level > c.GetLevel("combat") {
			aboveLevel = append(aboveLevel, item)
		} else {
			withinLevel = append(withinLevel, item)
		}
	}

	return withinLevel, aboveLevel
}

func (c *Character) GetBestOwnedEquipment(targetStats *game.Stats) *EquipmentSet {
	invBankAndEquipment := map[*game.Item]bool{}
	for itemCode := range c.bank.Items {
		invBankAndEquipment[game.Items.Get(itemCode)] = true
	}
	for itemCode := range c.Inventory {
		invBankAndEquipment[game.Items.Get(itemCode)] = true
	}
	for _, itemCode := range c.Equipment {
		if itemCode != "" {
			invBankAndEquipment[game.Items.Get(itemCode)] = true
		}
	}

	slotsEquipment := map[string][]*game.Item{}
	for item := range invBankAndEquipment {
		itemType := item.Type
		if _, ok := equipmentTypes[itemType]; !ok {
			continue
		}
		// TODO: Exclude items that are too low level/never going to be the best
		if item.Level > c.GetLevel("combat") {
			continue
		}

		if itemType == "ring" {
			if _, ok := slotsEquipment["ring1"]; !ok {
				slotsEquipment["ring1"] = []*game.Item{}
			}
			slotsEquipment["ring1"] = append(slotsEquipment["ring1"], item)

			if _, ok := slotsEquipment["ring2"]; !ok {
				slotsEquipment["ring2"] = []*game.Item{}
			}
			slotsEquipment["ring2"] = append(slotsEquipment["ring2"], item)
		} else {
			if _, ok := slotsEquipment[itemType]; !ok {
				slotsEquipment[itemType] = []*game.Item{}
			}
			slotsEquipment[itemType] = append(slotsEquipment[itemType], item)
		}
	}

	basePlayerHp := 120 + 5*c.GetLevel("combat")

	set := NewEquipmentSet(nil)
	for slot, itemCode := range c.Equipment {
		if itemCode != "" {
			set.Equipment[slot] = game.Items.Get(itemCode)
		}
	}

	equipmentStats := game.AccumulatedStats(set.Equipment)
	turnsToKillMonster, turnsToKillPlayer, haste := computeBestForRestOfSet(set, equipmentStats, targetStats, basePlayerHp, slotsEquipment, nil)
	set.TurnsToKillPlayer = turnsToKillPlayer
	set.TurnsToKillMonster = turnsToKillMonster
	set.Haste = haste
	return set
}

func computeBestForRestOfSet(set *EquipmentSet, equipmentStats *game.Stats, targetStats *game.Stats, basePlayerHp int, slotsEquipment map[string][]*game.Item, ignoreSlots []string) (int, int, int) {
	fewestTurnsToKill, mostTurnsToDie, bestHaste := calculateTurns(equipmentStats, targetStats, basePlayerHp)

	// Pick a slot. Pick an item. Put an item in the slot. Calculate all other slots
	for slot, items := range slotsEquipment {
		if slices.Contains(ignoreSlots, slot) {
			continue
		}
		if len(items) == 0 {
			continue
		}

		currentItem := set.Equipment[slot]

		// No item equipped. Only one possible item for this slot, it will be the best
		if currentItem == nil && len(items) == 1 {
			set.Equipment[slot] = items[0]
			equipmentStats.Add(items[0].Stats)
			fewestTurnsToKill, mostTurnsToDie, bestHaste = calculateTurns(equipmentStats, targetStats, basePlayerHp)
			continue
		}

		// Item equipped and it's the only possible item. Do nothing
		if currentItem == items[0] && len(items) == 1 {
			continue
		}

		// This shouldn't be possible as the current item should be included in the possible items
		if currentItem != items[0] && len(items) == 1 {
			panic("This shouldn't happen")
		}

		bestItem := currentItem

		if currentItem != nil {
			equipmentStats.Remove(currentItem.Stats)
			delete(set.Equipment, slot)
		}

		for _, item := range items {
			if item == currentItem {
				// We've already done the math for this item, either at the beginning or in a previous loop
				continue
			}

			set.Equipment[slot] = item
			equipmentStats.Add(item.Stats)

			turnsToKill, turnsToDie, haste := computeBestForRestOfSet(set, equipmentStats, targetStats, basePlayerHp, slotsEquipment, append(ignoreSlots, slot))

			delete(set.Equipment, slot)
			equipmentStats.Remove(item.Stats)

			// An item is better than another if it
			//  - Kills faster than the other item while still keeping us alive
			//  - Kills the same speed, still keeps us alive, but has a better haste

			//if turnsToKill < fewestTurnsToKill ||
			//	(turnsToKill == fewestTurnsToKill && turnsToDie > mostTurnsToDie) ||
			//	(turnsToKill == fewestTurnsToKill && turnsToDie == mostTurnsToDie && haste > bestHaste) {

			if (turnsToKill < fewestTurnsToKill && turnsToKill < turnsToDie) ||
				(turnsToKill == fewestTurnsToKill && turnsToKill < turnsToDie && haste > bestHaste) {
				fewestTurnsToKill = turnsToKill
				mostTurnsToDie = turnsToDie
				bestHaste = haste
				bestItem = item
			}
		}

		// bestItem should never be nil theoretically
		set.Equipment[slot] = bestItem
		equipmentStats.Add(bestItem.Stats)
	}

	return fewestTurnsToKill, mostTurnsToDie, bestHaste
}

func calculateTurns(equipmentStats *game.Stats, targetStats *game.Stats, basePlayerHp int) (int, int, int) {
	playerAttack := int(equipmentStats.GetDamageAgainst(targetStats))
	monsterAttack := int(targetStats.GetDamageAgainst(equipmentStats))

	playerHp := equipmentStats.Hp + basePlayerHp

	var turnsToKillMonster int
	if playerAttack <= 0 {
		turnsToKillMonster = math.MaxInt32
	} else {
		turnsToKillMonster = int(math.Ceil(float64(targetStats.Hp) / float64(playerAttack)))
	}

	var turnsToKillPlayer int
	if monsterAttack <= 0 {
		turnsToKillPlayer = math.MaxInt32
	} else {
		turnsToKillPlayer = int(math.Ceil(float64(playerHp) / float64(monsterAttack)))
	}

	return turnsToKillMonster, turnsToKillPlayer, equipmentStats.Haste
}
