package main

import (
	"context"
	"github.com/ahornerr/artifacts/bank"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/client"
	_ "github.com/ahornerr/artifacts/game"
	"github.com/ahornerr/artifacts/state"
	"log"
	"os"
	"time"
)

func main() {
	client, err := client.New(os.Getenv("ARTIFACTS_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

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

	//sequenceGetItem := func(itemCode string, quantity int) command.Command {
	//	return command.Sequence(
	//		commands.SequenceMoveToBankAndDepositAll,
	//		commands.GetItem(itemCode, quantity),
	//		commands.SequenceMoveToBankAndDepositAll,
	//	)
	//}

	//mainSequence := func() command.Command {
	//	return command.Sequence(
	//		//commands.NewCraftingLoop("steel", stopper.Never{}),
	//		commands.NewHarvestLoop("iron_rocks", stopper.AtQuantity(1000)),
	//		commands.NewCraftingLoop("steel", stopper.Never{}),
	//	)
	//}

	// commands.NewCraftingLoop("steel", stopper.AtQuantity(20))
	// commands.NewCraftingLoop("iron_helm", stopper.AtLevel("gearcrafting", 20)),
	// commands.NewFightLoop("cow", stopper.Never{})
	//commands.NewTaskLoop(stopper.Never{})

	//crafterSequence := command.NewLoop(
	//	stopper.Never{},
	//	func(iteration int) string {
	//		return fmt.Sprintf("Main sequence: %d", iteration)
	//	},
	//	//commands.NewCraftingLoop("copper", stopper.Never{}),
	//	//commands.NewCraftingLoop("copper_dagger", stopper.AtLevel("weaponcrafting", 5)),
	//	//commands.NewCraftingLoop("ash_plank", stopper.Never{}),
	//	//commands.NewCraftingLoop("wooden_shield", stopper.AtLevel("gearcrafting", 5)),
	//	//commands.NewCraftingLoop("copper_ring", stopper.AtLevel("jewelrycrafting", 5)),
	//	//commands.NewHarvestLoop("ash_tree", stopper.AtQuantity(100)),
	//	//commands.NewHarvestLoop("copper_rocks", stopper.AtQuantity(100)),
	//	//commands.NewFightLoop("chicken", stopper.Never{}),
	//
	//	commands.NewFightLoop("chicken", stopper.NewStopper(func(ctx context.Context, char *character.Character, quantity int) (bool, error) {
	//		numFeather := char.Inventory["feather"] + char.Bank()["feather"]
	//		return numFeather >= 4*5+2*4, nil
	//	})),
	//
	//	commands.NewFightLoop("yellow_slime", stopper.NewStopper(func(ctx context.Context, char *character.Character, quantity int) (bool, error) {
	//		numSlimes := char.Inventory["yellow_slimeball"] + char.Bank()["yellow_slimeball"]
	//		return numSlimes >= 10, nil
	//	})),
	//
	//	commands.NewFightLoop("green_slime", stopper.NewStopper(func(ctx context.Context, char *character.Character, quantity int) (bool, error) {
	//		numSlimes := char.Inventory["green_slimeball"] + char.Bank()["green_slimeball"]
	//		return numSlimes >= 10, nil
	//	})),
	//
	//	commands.NewTaskLoop(stopper.Never{}),
	//
	//	commands.NewFightLoop("blue_slime", stopper.NewStopper(func(ctx context.Context, char *character.Character, quantity int) (bool, error) {
	//		numSlimes := char.Inventory["blue_slimeball"] + char.Bank()["blue_slimeball"]
	//		return numSlimes >= 10, nil
	//	})),
	//
	//	//commands.NewFightLoop("yellow_slime", stopper.NewStopper(func(ctx context.Context, char *character.Character, quantity int) (bool, error) {
	//	//	numSlimes := char.Inventory["yellow_slimeball"] + char.Bank()["yellow_slimeball"]
	//	//	return numSlimes >= 10, nil
	//	//})),
	//)

	//subFighterSequence := func() command.Command {
	//	return command.Sequence(
	//		commands.NewTaskLoop(stopper.Never{}),
	//		commands.NewFightLoop("chicken", stopper.AtLevel("combat", 5)),
	//		commands.NewFightLoop("yellow_slime", stopper.AtLevel("combat", 7)),
	//		commands.NewFightLoop("green_slime", stopper.AtLevel("combat", 8)),
	//		commands.NewFightLoop("blue_slime", stopper.AtLevel("combat", 9)),
	//		commands.NewFightLoop("red_slime", stopper.AtLevel("combat", 10)),
	//		commands.NewHarvestLoop("copper_rocks", stopper.Never{}),
	//	)
	//}

	//woodcutThenMine := func() command.Command {
	//	return command.Sequence(
	//		commands.NewHarvestLoop("ash_tree", stopper.AtLevel("woodcutting", 10)),
	//		commands.NewHarvestLoop("copper_rocks", stopper.AtLevel("mining", 10)),
	//		commands.NewHarvestLoop("spruce_tree", stopper.AtLevel("woodcutting", 20)),
	//		commands.NewHarvestLoop("iron_rocks", stopper.AtLevel("mining", 20)),
	//	)
	//}
	//
	//mineThenWoodcut := func() command.Command {
	//	return command.Sequence(
	//		commands.NewHarvestLoop("copper_rocks", stopper.AtLevel("mining", 10)),
	//		commands.NewHarvestLoop("ash_tree", stopper.AtLevel("woodcutting", 10)),
	//		commands.NewHarvestLoop("iron_rocks", stopper.AtLevel("mining", 20)),
	//		commands.NewHarvestLoop("spruce_tree", stopper.AtLevel("woodcutting", 20)),
	//	)
	//}

	//characterCommands := map[string]command.Command{
	//	"curlyBoy1": subFighterSequence(),
	//	"curlyBoy2": subFighterSequence(),
	//	"curlyBoy3": subFighterSequence(),
	//	"curlyBoy4": subFighterSequence(),
	//	"curlyBoy5": subFighterSequence(),
	//	//"curlyBoy2": woodcutThenMine(),
	//	//"curlyBoy3": woodcutThenMine(),
	//	//"curlyBoy4": mineThenWoodcut(),
	//	//"curlyBoy5": mineThenWoodcut(),
	//}

	characterNames := []string{
		"curlyBoy1", "curlyBoy2", "curlyBoy3", "curlyBoy4", "curlyBoy5",
	}
	characters := map[string]*character.Character{}
	for _, charName := range characterNames {
		char := character.NewCharacter(client, theBank, characterUpdates, charName)
		_, err = char.Get(ctx)
		if err != nil {
			log.Fatal(err)
		}
		characters[charName] = char
	}

	totalItemQuantity := func(itemCode string) int {
		quantity := theBank.Items[itemCode]

		for _, char := range characters {
			quantity += char.Inventory[itemCode]

			for _, equipItemCode := range char.Equipment {
				if equipItemCode == itemCode {
					quantity++
				}
			}
		}
		return quantity
	}

	characterStates := map[string]state.Runner{
		"curlyBoy1": state.Loop(
			state.Harvest("ash_tree", func(c *character.Character, args *state.HarvestArgs) bool {
				return c.GetLevel("woodcutting") >= 10
			}),
			state.Craft("ash_plank", func(c *character.Character, args *state.CraftingArgs) bool {
				return c.Inventory["ash_wood"] < 8
			}),
			state.Craft("spruce_plank", func(c *character.Character, args *state.CraftingArgs) bool {
				return c.Bank()["spruce_plank"]+c.Inventory["spruce_plank"] >= 10
			}),
			state.Craft("iron_axe", func(c *character.Character, args *state.CraftingArgs) bool {
				return totalItemQuantity("iron_axe") == 5
			}),
			state.Craft("iron_pickaxe", func(c *character.Character, args *state.CraftingArgs) bool {
				return totalItemQuantity("iron_pickaxe") == 5
			}),
			state.Task(func(c *character.Character, args *state.TaskArgs) bool {
				// We've completed one task
				if len(args.Rewards) > 0 {
					return true
				}
				// We can't seem to win this fight
				if args.NumFights() >= 3 && args.NumWins() == 0 {
					return true
				}
				return false
			}),
		),
		"curlyBoy2": state.Loop(
			state.Harvest("iron_rocks", func(c *character.Character, args *state.HarvestArgs) bool {
				return c.IsInventoryFull()
			}),
			state.Craft("iron", func(c *character.Character, args *state.CraftingArgs) bool {
				return c.Inventory["iron_ore"] < 8
			}),
		),
		"curlyBoy3": state.Loop(
			state.Harvest("iron_rocks", func(c *character.Character, args *state.HarvestArgs) bool {
				return c.IsInventoryFull()
			}),
			state.Craft("iron", func(c *character.Character, args *state.CraftingArgs) bool {
				return c.Inventory["iron_ore"] < 8
			}),
		),
		"curlyBoy4": state.Loop(
			state.Harvest("iron_rocks", func(c *character.Character, args *state.HarvestArgs) bool {
				return c.IsInventoryFull()
			}),
			state.Craft("iron", func(c *character.Character, args *state.CraftingArgs) bool {
				return c.Inventory["iron_ore"] < 8
			}),
		),
		"curlyBoy5": state.Loop(
			state.Harvest("iron_rocks", func(c *character.Character, args *state.HarvestArgs) bool {
				return c.IsInventoryFull()
			}),
			state.Craft("iron", func(c *character.Character, args *state.CraftingArgs) bool {
				return c.Inventory["iron_ore"] < 8
			}),
		),
	}

	for charName, charState := range characterStates {
		go func(char *character.Character, charState state.Runner) {
			char.PushState("Waiting for cooldown")
			time.Sleep(time.Until(char.CooldownExpires))
			char.PopState()

			err := charState(ctx, char)
			if err != nil {
				char.PushState("Error: %s", err)
			} else {
				char.PushState("Done")
			}
		}(characters[charName], charState)
	}

	//for charName, cmd := range characterCommands {
	//	char := character.NewCharacter(client, theBank, characterUpdates, charName)
	//
	//	_, err = char.Get(ctx)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	characters[charName] = char
	//
	//	//graph2.BestEquipmentAgainst(game.Monsters.Get("ogre").Stats, char)
	//
	//	go func(char *character.Character, command command.Command) {
	//		err := executeCommand(ctx, char, command, nil)
	//		if err != nil {
	//			log.Println("Executor error: ", err)
	//			return
	//		}
	//
	//		char.Action([]string{"Done"})
	//	}(char, cmd)
	//}

	onNewClient := func() {
		// Iterate over the slice since it's ordered
		for _, charName := range characterNames {
			events <- Event{Character: characters[charName]}
		}
		events <- Event{Bank: theBank.Items}
	}

	// TODO: Clean shutdown
	server := httpServer(events, onNewClient)
	log.Fatal(server.Listen(":8080"))
}

//func executeCommand(ctx context.Context, char *character.Character, cmd command.Command, actions []string) error {
//	tailRecursion := false
//
//	description := cmd.Description()
//	if description != "" {
//		actions = append(actions, description)
//	}
//	for {
//		char.Action(actions)
//
//		nextCommands, err := cmd.Execute(ctx, char)
//		if err != nil {
//			actions = append(actions, fmt.Sprintf("Error: %s", err.Error()))
//			char.Action(actions)
//			return err
//		}
//
//		if len(nextCommands) > 0 {
//			tailRecursion = nextCommands[len(nextCommands)-1] == cmd
//		} else {
//			tailRecursion = false
//		}
//
//		if tailRecursion {
//			nextCommands = nextCommands[:len(nextCommands)-1]
//			actions[len(actions)-1] = cmd.Description()
//		}
//
//		for _, nextCommand := range nextCommands {
//			err := executeCommand(ctx, char, nextCommand, actions)
//			if err != nil {
//				return err
//			}
//		}
//
//		if !tailRecursion {
//			return nil
//		}
//	}
//}

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
//		for _, drop := range r.Game.ResourcesForItem[item.Code] {
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
