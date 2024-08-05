package main

import (
	"context"
	"fmt"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"slices"
)

func (r *Runner) TrainSkill(ctx context.Context, skill string, requiredLevel int) error {
	currentLevel := r.Char.GetLevel(skill)
	if currentLevel >= requiredLevel {
		return nil // Reached level target
	}

	trainingMethods := r.Game.GetTrainingMethods(skill)

	if len(trainingMethods) == 0 {
		return fmt.Errorf("could not find a training method")
	}

	var availableTrainingMethods []TrainingMethod
	for _, method := range trainingMethods {
		if method.Level > currentLevel {
			continue // We can't do this yet
		}

		// TODO: Do we also need to check that we have the levels to make all the items for crafting too?

		availableTrainingMethods = append(availableTrainingMethods, method)
	}

	bankItems := r.Game.Bank.AsMap()

	slices.SortFunc(availableTrainingMethods, func(a, b TrainingMethod) int {
		return b.Weight(bankItems) - a.Weight(bankItems)
	})

	// TODO: Should we weight the training methods so that gathering methods of the same level are ranked higher?

	// Highest level method that we can use right now.
	// TODO: Might not always be the most optimal, especially for crafting where we can have some items already
	// TODO: Higher level items are preferred but may require more harvesting.
	//  Items that we already have resources for are preferable but may give less XP.
	//   For now, just take select the highest level item (there may be multiple of the same level)
	method := availableTrainingMethods[0]

	// If crafting training, empty our inventory first
	if method.Item != nil {
		r.reportAction("Banking before training")

		err := r.Char.MoveClosest(ctx, r.Game.GetBankLocations())
		if err != nil {
			return fmt.Errorf("moving to bank: %w", err)
		}

		err = r.Char.BankAll(ctx)
		if err != nil {
			return fmt.Errorf("banking all: %w", err)
		}
	}

	// Train in a loop until we level up and call TrainSkill() again.
	// This will let us find new possible training methods if they are unlocked.
	// It's also the mechanism to exit once we've hit the required level.
	for {
		if method.Resource != nil {
			err := r.harvestTraining(ctx, method.Resource)
			if err != nil {
				return fmt.Errorf("harvesting for training: %w", err)
			}
		} else {
			err := r.craftingTraining(ctx, method.Item)
			if err != nil {
				return fmt.Errorf("crafting for training: %w", err)
			}
		}
		if r.Char.GetLevel(skill) > currentLevel {
			break
		}
	}

	return r.TrainSkill(ctx, skill, requiredLevel)
}

func (r *Runner) craftingTraining(ctx context.Context, item *Item) error {
	// First, figure out what all we have in the bank
	bankItems := r.Game.Bank.AsMap()

	// Do we have the direct child crafting requirements for this item in the bank?
	for _, craftingItem := range item.Crafting.Items {
		missingQuantity := craftingItem.Quantity - bankItems[craftingItem.Code]
		if missingQuantity > 0 {
			// We don't have enough of these to make the item. Can we craft that item?
			err := r.acquireItem(ctx, craftingItem.Code, missingQuantity)
			if err != nil {
				return fmt.Errorf("acquiring item: %w", err)
			}
		}
	}

	// Now we should have everything we need for crafting. Move to the correct workshop and make it.
	err := r.Char.MoveClosest(ctx, r.Game.GetWorkshopLocations(item.Crafting.Skill))
	if err != nil {
		return fmt.Errorf("moving to workshop")
	}

	_, err = r.Char.Craft(ctx, item.Code, 1)
	if err != nil {
		return fmt.Errorf("crafting: %w", err)
	}

	return nil
}

func (r *Runner) harvestTraining(ctx context.Context, resource *client.ResourceSchema) error {
	return r.harvest(ctx, resource.Code, 1)
}
