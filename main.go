package main

import (
	"context"
	"fmt"
	"github.com/ahornerr/artifacts/bank"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/client"
	"github.com/ahornerr/artifacts/command"
	"github.com/ahornerr/artifacts/commands"
	"github.com/ahornerr/artifacts/stopper"
	"log"
	"os"
)

func main() {
	client, err := client.New(os.Getenv("ARTIFACTS_TOKEN"))

	events := make(chan Event)

	characterUpdates := make(chan *character.Character)
	bankUpdates := make(chan map[string]int)

	go func() {
		nonBlockingWriteEvent := func(event Event) {
			select {
			case events <- event:
			default:
			}
		}

		for {
			select {
			case char := <-characterUpdates:
				nonBlockingWriteEvent(Event{Character: char})
			case bankItems := <-bankUpdates:
				nonBlockingWriteEvent(Event{Bank: bankItems})
			}
		}
	}()

	ctx := context.Background()

	theBank := bank.NewBank(client, bankUpdates)

	if _, err := theBank.Load(ctx); err != nil {
		log.Fatalf("loading bank items: %s", err)
	}

	characterCommands := map[string]command.Command{
		"curlyBoy": commands.NewCraftingLoop("iron", stopper.StopNever{}),
		//"curlyBoy2": command.NewCraftingLoop("iron", stopper.StopNever{}),
		//"curlyBoy3": command.NewCraftingLoop("iron", stopper.StopNever{}),
		//"curlyBoy4": command.NewCraftingLoop("iron", stopper.StopNever{}),
		//"curlyBoy5": command.NewCraftingLoop("iron", stopper.StopNever{}),
		//"curlyBoy":  command.NewHarvestLoop("iron_rocks", stopper.StopNever{}),
		"curlyBoy2": commands.NewHarvestLoop("iron_rocks", stopper.StopNever{}),
		"curlyBoy3": commands.NewHarvestLoop("iron_rocks", stopper.StopNever{}),
		"curlyBoy4": commands.NewHarvestLoop("iron_rocks", stopper.StopNever{}),
		"curlyBoy5": commands.NewHarvestLoop("iron_rocks", stopper.StopNever{}),
	}

	characters := map[string]*character.Character{}

	for charName, cmd := range characterCommands {
		char := character.NewCharacter(client, theBank, characterUpdates, charName)

		_, err = char.Get(ctx)
		if err != nil {
			log.Fatal(err)
		}

		characters[charName] = char

		go func(char *character.Character, command command.Command) {
			err := executeCommand(ctx, char, command, nil)
			if err != nil {
				log.Println("Executor error: ", err)
				return
			}

			char.Action([]string{"Done"})
		}(char, cmd)

	}

	onNewClient := func() {
		for charName := range characterCommands {
			events <- Event{Character: characters[charName]}
		}
		events <- Event{Bank: theBank.Items}
	}

	// TODO: Clean shutdown
	server := httpServer(events, onNewClient)
	log.Fatal(server.Listen(":8080"))
}

func executeCommand(ctx context.Context, char *character.Character, cmd command.Command, actions []string) error {
	tailRecursion := false

	description := cmd.Description()
	if description != "" {
		actions = append(actions, description)
	}
	for {
		char.Action(actions)

		nextCommands, err := cmd.Execute(ctx, char)
		if err != nil {
			actions = append(actions, fmt.Sprintf("Error: %s", err.Error()))
			char.Action(actions)
			return err
		}

		char.WaitForCooldown()

		if len(nextCommands) > 0 {
			tailRecursion = nextCommands[len(nextCommands)-1] == cmd
		} else {
			tailRecursion = false
		}

		if tailRecursion {
			nextCommands = nextCommands[:len(nextCommands)-1]
			actions[len(actions)-1] = cmd.Description()
		}

		for _, nextCommand := range nextCommands {
			err := executeCommand(ctx, char, nextCommand, actions)
			if err != nil {
				return err
			}
		}

		if !tailRecursion {
			return nil
		}
	}
}

//func (r *Executor) acquireItem(ctx context.Context, itemCode string, wantQuantity int) error {
//	r.reportAction("Acquiring %d %s", wantQuantity, itemCode)
//
//	err := r.acquireFromInventoryAndBank(ctx, itemCode, wantQuantity)
//	if err != nil {
//		return err
//	}
//
//	// Check inventory to see if we have enough
//	inventoryQuantity := r.Char.InventoryQuantity(itemCode)
//	if inventoryQuantity >= wantQuantity {
//		return nil
//	}
//
//	item := r.Game.items[itemCode]
//	requiredSkills := item.GetAllSkillRequirements()
//
//	for skill, requiredLevel := range requiredSkills {
//		err := r.TrainSkill(ctx, skill, requiredLevel)
//		if err != nil {
//			return fmt.Errorf("training: %w", err)
//		}
//	}
//
//	// Reduce the desired quantity by the quantity we already have
//	//wantQuantity -= inventoryQuantity
//
//	// TODO: It doesn't really make sense to withdraw from the bank right before we start training.
//	//  Only withdraw from the bank after training, or if we don't need training
//
//	canBeCrafted := item.Crafting != nil
//	if canBeCrafted {
//		return r.craft(ctx, itemCode, wantQuantity)
//	} else {
//		var selectedResource *client.ResourceSchema
//		for _, drop := range r.Game.Drops[item.Code] {
//			if drop.Level > r.Char.GetLevel(string(drop.Skill)) {
//				continue
//			}
//			selectedResource = &drop
//			break
//		}
//
//		return r.harvest(ctx, selectedResource.Code, wantQuantity)
//	}
//}

//func (r *Executor) taskWhenNotCrafting(ctx context.Context, itemCode string) error {
//	item := r.Game.items[itemCode]
//
//	if item.Crafting == nil {
//		return fmt.Errorf("item cannot be crafted")
//	}
//
//	recipeRequiredInvSpace := item.Crafting.TotalQuantity()
//
//	canMakeInventoryOfItem := func() bool {
//		numCanCraft := r.NumberCanCraftInBankAndInventory(*item.Crafting)
//
//		// Only stop when we can make a full inventory worth
//		fullInvQuantity := r.Char.character.InventoryMaxItems / recipeRequiredInvSpace
//		return numCanCraft >= fullInvQuantity
//	}
//
//	for {
//		err := r.taskLoop(ctx, canMakeInventoryOfItem)
//		if err != nil {
//			return fmt.Errorf("task loop: %w", err)
//		}
//
//		err = r.MoveToBankAndDepositAll(ctx)
//		if err != nil {
//			return err
//		}
//
//		numCanCraft := r.NumberCanCraftInBankAndInventory(*item.Crafting)
//
//		// Only make whole inventories
//		fullInvQuantity := r.Char.character.InventoryMaxItems / recipeRequiredInvSpace
//		numInventories := numCanCraft / fullInvQuantity
//		numToCraft := numInventories * fullInvQuantity
//
//		err = r.craft(ctx, itemCode, numToCraft)
//		if err != nil {
//			return fmt.Errorf("crafting: %w", err)
//		}
//	}
//}
