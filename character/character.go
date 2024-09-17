package character

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/bank"
	"github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/httperror"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"log"
	"math"
	"slices"
	"time"
)

var equipmentSlotOrder = []string{
	"weapon",
	"helmet",
	"amulet",
	"body_armor",
	"shield",
	"ring1",
	"ring2",
	"leg_armor",
	"boots",
}

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
	log.Printf("%s: %s\n", c.Name, state)
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

func (c *Character) update(ctx context.Context, char client.CharacterSchema, waitForCooldown bool) {
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
		select {
		case <-ctx.Done():
		case <-time.After(time.Until(c.CooldownExpires)):
		}
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

	c.update(ctx, resp.JSON200.Data, false)

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

	c.update(ctx, resp.JSON200.Data.Character, true)

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

	c.update(ctx, resp.JSON200.Data.Character, true)

	return result, nil
}

func (c *Character) Gather(ctx context.Context) (*client.SkillInfoSchema, error) {
	resp, err := c.client.ActionGatheringMyNameActionGatheringPostWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(ctx, resp.JSON200.Data.Character, true)

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

	c.update(ctx, resp.JSON200.Data.Character, true)

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
	c.update(ctx, resp.JSON200.Data.Character, true)

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
	c.update(ctx, resp.JSON200.Data.Character, true)

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

func (c *Character) NewTask(ctx context.Context) (*client.TaskSchema, error) {
	c.PushState("Getting new task")
	defer c.PopState()

	resp, err := c.client.ActionAcceptNewTaskMyNameActionTaskNewPostWithResponse(ctx, c.Name)
	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(ctx, resp.JSON200.Data.Character, true)

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

	c.update(ctx, resp.JSON200.Data.Character, true)

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

	c.update(ctx, resp.JSON200.Data.Character, true)

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

	c.update(ctx, resp.JSON200.Data.Character, true)

	return nil
}

func (c *Character) Recycle(ctx context.Context, itemCode string, quantity int) (*client.RecyclingItemsSchema, error) {
	c.PushState("Recycling %d %s", quantity, itemCode)
	defer c.PopState()

	resp, err := c.client.ActionRecyclingMyNameActionRecyclingPostWithResponse(ctx, c.Name, client.ActionRecyclingMyNameActionRecyclingPostJSONRequestBody{
		Code:     itemCode,
		Quantity: &quantity,
	})

	if err != nil {
		return nil, err
	} else if resp.JSON200 == nil {
		return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
	}

	c.update(ctx, resp.JSON200.Data.Character, true)

	return &resp.JSON200.Data.Details, nil
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
	Haste              int8
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

		if item.Level > c.GetLevel("combat") {
			continue
		}

		if targetStats.IsResource != (item.SubType == "tool") {
			continue
		}

		// TODO: Exclude items that are too low level/never going to be the best

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

	basePlayerHp := 115 + 5*c.GetLevel("combat")

	set := NewEquipmentSet(nil)
	for slot, itemCode := range c.Equipment {
		if itemCode != "" {
			set.Equipment[slot] = game.Items.Get(itemCode)
		}
	}

	for _, slot := range equipmentSlotOrder {
		var newItems []*game.Item
		items := slotsEquipment[slot]

		highestLevel := 0
		for _, item := range items {
			if item.Level > highestLevel {
				highestLevel = item.Level
			}
		}
		for _, item := range items {
			if item.Level >= highestLevel-5 {
				newItems = append(newItems, item)
			}
		}
		slices.SortFunc(newItems, func(a, b *game.Item) int {
			if a.Code < b.Code {
				return -1
			}
			if a.Code > b.Code {
				return 1
			}
			return 0
		})
		slotsEquipment[slot] = newItems
	}

	var slots []string
	for _, slot := range equipmentSlotOrder {
		items := slotsEquipment[slot]
		if len(items) == 0 {
			continue
		}
		if len(items) == 1 {
			if items[0] != set.Equipment[slot] {
				set.Equipment[slot] = items[0]
			}
			continue
		}

		slots = append(slots, slot)
	}

	turnsToKillMonster, turnsToKillPlayer, haste := computeBestForRestOfSet(set, targetStats, basePlayerHp, slotsEquipment, slots)
	set.TurnsToKillPlayer = turnsToKillPlayer
	set.TurnsToKillMonster = turnsToKillMonster
	set.Haste = haste
	return set
}

func computeBestForRestOfSet(set *EquipmentSet, targetStats *game.Stats, basePlayerHp int, slotsEquipment map[string][]*game.Item, slots []string) (int, int, int8) {
	fewestTurnsToKill, mostTurnsToDie, bestHaste := calculateTurns(set, targetStats, basePlayerHp)

	// Pick a slot. Pick an item. Put an item in the slot. Calculate all other slots
	for _, slot := range slots {
		items := slotsEquipment[slot]
		if len(items) == 0 {
			continue
		}

		newSlots := slots[1:]
		bestItem := set.Equipment[slot]

		for _, item := range items {
			set.Equipment[slot] = item

			otherSlotsEquipment := slotsEquipment
			if slot == "weapon" {
				// Filter out other slots that don't contribute to this weapon's attack bonuses
				otherSlotsEquipment = map[string][]*game.Item{}
				for otherSlot, otherItems := range slotsEquipment {
					if otherSlot == "weapon" {
						continue
					}
					for _, otherItem := range otherItems {
						if (item.Stats.AttackAir > 0 && (otherItem.Stats.AttackAir > 0 || otherItem.Stats.DamageAir > 0)) ||
							(item.Stats.AttackEarth > 0 && (otherItem.Stats.AttackEarth > 0 || otherItem.Stats.DamageEarth > 0)) ||
							(item.Stats.AttackFire > 0 && (otherItem.Stats.AttackFire > 0 || otherItem.Stats.DamageFire > 0)) ||
							(item.Stats.AttackWater > 0 && (otherItem.Stats.AttackWater > 0 || otherItem.Stats.DamageWater > 0)) {
							otherSlotsEquipment[otherSlot] = append(otherSlotsEquipment[otherSlot], otherItem)
						}
					}
				}
			}

			turnsToKill, turnsToDie, haste := computeBestForRestOfSet(set, targetStats, basePlayerHp, otherSlotsEquipment, newSlots)

			if bestItem == nil ||
				(turnsToKill < fewestTurnsToKill) ||
				(turnsToKill == fewestTurnsToKill && turnsToDie > mostTurnsToDie) ||
				(turnsToKill == fewestTurnsToKill && turnsToDie == mostTurnsToDie && haste > bestHaste) {

				fewestTurnsToKill = turnsToKill
				mostTurnsToDie = turnsToDie
				bestHaste = haste
				bestItem = item
			}
		}

		set.Equipment[slot] = bestItem
	}

	return fewestTurnsToKill, mostTurnsToDie, bestHaste
}

func calculateTurns(set *EquipmentSet, targetStats *game.Stats, basePlayerHp int) (int, int, int8) {
	equipmentStats := game.AccumulatedStats(set.Equipment)

	playerAttack := equipmentStats.GetDamageAgainst(targetStats)
	monsterAttack := targetStats.GetDamageAgainst(equipmentStats)

	playerHp := int(equipmentStats.Hp) + basePlayerHp

	turnsToKillMonster := math.MaxUint8
	if playerAttack > 0 {
		turnsToKillMonster = int(targetStats.Hp) / playerAttack
	}

	turnsToKillPlayer := math.MaxUint8
	if monsterAttack > 0 {
		turnsToKillPlayer = playerHp / monsterAttack
	}

	return turnsToKillMonster, turnsToKillPlayer, equipmentStats.Haste
}
