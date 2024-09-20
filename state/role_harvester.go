package state

import (
	"context"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/game"
	"github.com/gofiber/fiber/v3/log"
	"time"
)

const maxLevel = 35

// 28 jewelrycrafting made iron ring (level 10) got 0 xp
// 28 jewelrycrafting made life ring (level 15) got 0 xp
// 28 jewelrycrafting made steel ring (level 20) got 279 xp
// 13 levels higher, no xp

// 25 weaponcrafting made mushstaff (level 15) got 221 xp
// 10 levels higher, yes xp

// and as soon as you have a 10-level difference with a monster, resource or craft, it won't give you any more xp.

func RoleHarvester(crafterWants chan game.ItemQuantity) Runner {
	return func(ctx context.Context, char *character.Character) error {
		var forCrafter *game.ItemQuantity
		for {
			ctx, cancel := context.WithCancel(ctx)

			if forCrafter != nil {
				time.Sleep(time.Until(char.CooldownExpires))
				err := harvestForCrafter(ctx, char, *forCrafter)
				if err != nil {
					log.Errorf("%s %v", char.Name, err)
					// TODO: Should we put the ItemQuantity back on the channel?
				}
				forCrafter = nil
			} else {
				go func() {
					for {
						select {
						case <-ctx.Done():
							return
						case want := <-crafterWants:
							if !canCollect(char, want.Item) {
								crafterWants <- want
								continue
							}
							forCrafter = &want
							cancel()
							return
						}
					}
				}()

				err := harvester(ctx, char)
				if err != nil {
					log.Errorf("%s %v", char.Name, err)
				}
			}

			cancel()
		}
	}
}

func canCollect(char *character.Character, item *game.Item) bool {
	if item.Crafting != nil {
		return char.GetLevel(item.Crafting.Skill) >= item.Crafting.Level
	}

	for _, resource := range game.Resources.ResourcesForItem(item) {
		if char.GetLevel(resource.Skill) >= resource.Level {
			return true
		}
	}

	for _, monster := range game.Monsters.MonstersForItem(item) {
		// TODO: Determine if we can win against a monster
		_ = monster
		return true
	}

	// Just jasper crystal for now
	if item.SubType == "task" {
		// TODO: Let's just assume we can always collect this for now
		return true
	}

	return false
}

func harvestForCrafter(ctx context.Context, char *character.Character, crafterWants game.ItemQuantity) error {
	char.PushState("Getting %d %s for crafter", crafterWants.Quantity, crafterWants.Item)
	defer char.PopState()

	err := CollectItems(
		crafterWants.Item.Code,
		crafterWants.Quantity,
		false,
		false,
		nil,
	)(ctx, char)
	if err != nil {
		return err
	}

	return MoveToBankAndDepositAll(ctx, char)
}

func stopForEvent[T any](char *character.Character, args T) bool {
	if char.GetLevel("mining") >= 35 && len(game.Maps.GetResources("strange_rocks")) > 0 {
		return true
	}
	if char.GetLevel("woodcutting") >= 35 && len(game.Maps.GetResources("magic_tree")) > 0 {
		return true
	}
	return false
}

func harvester(ctx context.Context, char *character.Character) error {

	if char.GetLevel("mining") >= 35 && len(game.Maps.GetResources("strange_rocks")) > 0 {
		return Harvest("strange_rocks", nil)(ctx, char)
	}

	if char.GetLevel("woodcutting") >= 35 && len(game.Maps.GetResources("magic_tree")) > 0 {
		return Harvest("magic_tree", nil)(ctx, char)
	}

	// Train skills 5 levels at a time.
	// TODO: Stop at level
	if char.GetLevel("woodcutting")%5 == 0 && char.GetLevel("woodcutting") >= char.GetLevel("mining") {
		miningResources := resourcesForTraining(char, "mining")
		if len(miningResources) > 0 {
			char.PushState("Training mining")
			defer char.PopState()
			return Harvest(miningResources[0].Code, stopForEvent)(ctx, char)
		}
	}

	woodcuttingResources := resourcesForTraining(char, "woodcutting")
	if len(woodcuttingResources) > 0 {
		char.PushState("Training woodcutting")
		defer char.PopState()
		return Harvest(woodcuttingResources[0].Code, stopForEvent)(ctx, char)
	}

	fishingResources := resourcesForTraining(char, "fishing")
	if len(fishingResources) > 0 {
		char.PushState("Training fishing")
		defer char.PopState()
		return Harvest(fishingResources[0].Code, stopForEvent)(ctx, char)
	}

	//miningItems := itemsForTraining(char, "mining")
	//woodcuttingItems := itemsForTraining(char, "woodcutting")
	//
	//_ = miningItems
	//_ = woodcuttingItems

	return nil
}

func resourcesForTraining(char *character.Character, skill string) []*game.Resource {
	charLevel := char.GetLevel(skill)
	if charLevel >= maxLevel {
		return nil
	}

	return game.Resources.ResourcesForSkill(skill, charLevel)
}
