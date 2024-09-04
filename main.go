package main

import (
	"context"
	"github.com/ahornerr/artifacts/bank"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/client"
	"github.com/ahornerr/artifacts/game"
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

	//totalItemQuantity := func(itemCode string) int {
	//	quantity := theBank.Items[itemCode]
	//
	//	for _, char := range characters {
	//		quantity += char.Inventory[itemCode]
	//
	//		for _, equipItemCode := range char.Equipment {
	//			if equipItemCode == itemCode {
	//				quantity++
	//			}
	//		}
	//	}
	//	return quantity
	//}

	characterStates := map[string]state.Runner{
		//"curlyBoy1": state.TaskCrafting(characters),
		//"curlyBoy1": state.CollectItems(game.Items.Get("copper_ring"), 20),
		//"curlyBoy2": state.Harvest("copper_rocks", func(c *character.Character, args *state.HarvestArgs) bool {
		//	return false
		//}),
		//"curlyBoy3": state.Harvest("copper_rocks", func(c *character.Character, args *state.HarvestArgs) bool {
		//	return false
		//}),
		//"curlyBoy4": state.Harvest("copper_rocks", func(c *character.Character, args *state.HarvestArgs) bool {
		//	return false
		//}),
		//"curlyBoy5": state.Harvest("copper_rocks", func(c *character.Character, args *state.HarvestArgs) bool {
		//	return false
		//}),
		//"curlyBoy2": state.Fight("chicken", func(c *character.Character, args *state.FightArgs) bool {
		//	return false
		//}),
		//"curlyBoy3": state.Fight("chicken", func(c *character.Character, args *state.FightArgs) bool {
		//	return false
		//}),
		//"curlyBoy4": state.Fight("chicken", func(c *character.Character, args *state.FightArgs) bool {
		//	return false
		//}),
		//"curlyBoy5": state.Fight("chicken", func(c *character.Character, args *state.FightArgs) bool {
		//	return false
		//}),
		"curlyBoy1": state.Loop(
			//state.CollectItems(game.Items.Get("adventurer_pants"), 5, true, characters),
			//state.CollectItems(game.Items.Get("slime_shield"), 5, true, characters),
			//state.CollectItems(game.Items.Get("fire_and_earth_amulet"), 5, true, characters),
			//state.CollectItems(game.Items.Get("air_and_water_amulet"), 5, true, characters),
			//state.CollectItems(game.Items.Get("mushmush_jacket"), 5, true, characters),
			//state.CollectItems(game.Items.Get("lucky_wizard_hat"), 5, true, characters),
			//state.CollectItems(game.Items.Get("mushmush_wizard_hat"), 5, true, characters),

			state.CollectInventory(game.Items.Get("iron_boots"), func(c *character.Character, args *state.CollectInventoryArgs) bool {
				return c.GetLevel("gearcrafting") >= 20
			}),

			state.CollectInventory(game.Items.Get("iron_ring"), func(c *character.Character, args *state.CollectInventoryArgs) bool {
				return c.GetLevel("jewelrycrafting") >= 20
			}),

			state.CollectInventory(game.Items.Get("iron_dagger"), func(c *character.Character, args *state.CollectInventoryArgs) bool {
				return c.GetLevel("weaponcrafting") >= 20
			}),

			state.Task(func(c *character.Character, args *state.TaskArgs) bool {
				return false
			}),
		),
		"curlyBoy2": state.Loop(
			state.Harvest("iron_rocks", func(c *character.Character, args *state.HarvestArgs) bool {
				return c.IsInventoryFull()
			}),
			state.Craft("iron", 0, func(c *character.Character, args *state.CraftingArgs) bool {
				return c.Inventory["iron_ore"] < 8
			}),
		),
		"curlyBoy3": state.Loop(
			state.Harvest("iron_rocks", func(c *character.Character, args *state.HarvestArgs) bool {
				return c.IsInventoryFull()
			}),
			state.Craft("iron", 0, func(c *character.Character, args *state.CraftingArgs) bool {
				return c.Inventory["iron_ore"] < 8
			}),
		),
		//"curlyBoy4": state.Loop(
		//	state.Harvest("iron_rocks", func(c *character.Character, args *state.HarvestArgs) bool {
		//		return c.IsInventoryFull()
		//	}),
		//	state.Craft("iron", 0, func(c *character.Character, args *state.CraftingArgs) bool {
		//		return c.Inventory["iron_ore"] < 8
		//	}),
		//),
		//"curlyBoy5": state.Loop(
		//	state.Harvest("iron_rocks", func(c *character.Character, args *state.HarvestArgs) bool {
		//		return c.IsInventoryFull()
		//	}),
		//	state.Craft("iron", 0, func(c *character.Character, args *state.CraftingArgs) bool {
		//		return c.Inventory["iron_ore"] < 8
		//	}),
		//),
		//"curlyBoy2": state.Fight("chicken", func(c *character.Character, args *state.FightArgs) bool {
		//	return false
		//}),
		//"curlyBoy3": state.Fight("chicken", func(c *character.Character, args *state.FightArgs) bool {
		//	return false
		//}),
		"curlyBoy4": state.Fight("chicken", func(c *character.Character, args *state.FightArgs) bool {
			return false
		}),
		"curlyBoy5": state.Fight("chicken", func(c *character.Character, args *state.FightArgs) bool {
			return false
		}),
		//"curlyBoy2": state.Loop(
		//	state.Task(func(c *character.Character, args *state.TaskArgs) bool {
		//		return false
		//	}),
		//	state.Harvest("spruce_tree", func(c *character.Character, args *state.HarvestArgs) bool {
		//		return c.IsInventoryFull()
		//	}),
		//	state.Craft("spruce_plank", 0, func(c *character.Character, args *state.CraftingArgs) bool {
		//		return c.Inventory["spruce_log"] < 8
		//	}),
		//),
		//"curlyBoy3": state.Loop(
		//	state.Task(func(c *character.Character, args *state.TaskArgs) bool {
		//		return false
		//	}),
		//	state.Harvest("spruce_tree", func(c *character.Character, args *state.HarvestArgs) bool {
		//		return c.IsInventoryFull()
		//	}),
		//	state.Craft("spruce_plank", 0, func(c *character.Character, args *state.CraftingArgs) bool {
		//		return c.Inventory["spruce_log"] < 8
		//	}),
		//),
		//"curlyBoy4": state.Loop(
		//	state.Task(func(c *character.Character, args *state.TaskArgs) bool {
		//		return false
		//	}),
		//	state.Harvest("spruce_tree", func(c *character.Character, args *state.HarvestArgs) bool {
		//		return c.IsInventoryFull()
		//	}),
		//	state.Craft("spruce_plank", 0, func(c *character.Character, args *state.CraftingArgs) bool {
		//		return c.Inventory["spruce_log"] < 8
		//	}),
		//),
		//"curlyBoy5": state.Loop(
		//	state.Task(func(c *character.Character, args *state.TaskArgs) bool {
		//		return false
		//	}),
		//	state.Harvest("spruce_tree", func(c *character.Character, args *state.HarvestArgs) bool {
		//		return c.IsInventoryFull()
		//	}),
		//	state.Craft("spruce_plank", 0, func(c *character.Character, args *state.CraftingArgs) bool {
		//		return c.Inventory["spruce_log"] < 8
		//	}),
		//),

		//"curlyBoy1": state.Loop(
		//state.Craft("iron_boots", func(c *character.Character, args *state.CraftingArgs) bool {
		//	return totalItemQuantity("iron_boots") >= 5
		//}),
		//state.Craft("iron_helm", func(c *character.Character, args *state.CraftingArgs) bool {
		//	return totalItemQuantity("iron_helm") >= 5
		//}),
		//state.Task(func(c *character.Character, args *state.TaskArgs) bool {
		//	// We've completed one task
		//	if len(args.Rewards) > 0 {
		//		return true
		//	}
		//	// We can't seem to win this fight
		//	if args.NumFights() >= 3 && args.NumWins() == 0 {
		//		return true
		//	}
		//	return false
		//}),
		//state.Harvest("iron_rocks", func(c *character.Character, args *state.HarvestArgs) bool {
		//	c.IsInventoryFull()
		//}),
		//state.Fight("cow", func(c *character.Character, args *state.FightArgs) bool {
		//	return args.Drops["feather"] >= 10
		//}),
		//),
		//"curlyBoy2": state.Fight("cow", func(c *character.Character, args *state.FightArgs) bool {
		//	return args.Drops["cowhide"] >= 30
		//}),
		//"curlyBoy3": state.Fight("cow", func(c *character.Character, args *state.FightArgs) bool {
		//	return args.Drops["cowhide"] >= 30
		//}),
		//"curlyBoy4": state.Fight("cow", func(c *character.Character, args *state.FightArgs) bool {
		//	return args.Drops["cowhide"] >= 30
		//}),
		//"curlyBoy5": state.Fight("cow", func(c *character.Character, args *state.FightArgs) bool {
		//	return args.Drops["cowhide"] >= 30
		//}),
		//"curlyBoy2": state.Task(func(c *character.Character, args *state.TaskArgs) bool {
		//	// We can't seem to win this fight
		//	if args.NumFights() >= 3 && args.NumWins() == 0 {
		//		return true
		//	}
		//	return false
		//}),
		//"curlyBoy3": state.Task(func(c *character.Character, args *state.TaskArgs) bool {
		//	// We can't seem to win this fight
		//	if args.NumFights() >= 3 && args.NumWins() == 0 {
		//		return true
		//	}
		//	return false
		//}),
		//"curlyBoy4": state.Task(func(c *character.Character, args *state.TaskArgs) bool {
		//	// We can't seem to win this fight
		//	if args.NumFights() >= 3 && args.NumWins() == 0 {
		//		return true
		//	}
		//	return false
		//}),
		//"curlyBoy5": state.Task(func(c *character.Character, args *state.TaskArgs) bool {
		//	// We can't seem to win this fight
		//	if args.NumFights() >= 3 && args.NumWins() == 0 {
		//		return true
		//	}
		//	return false
		//}),
		//state.Loop(
		//	state.Harvest("iron_rocks", func(c *character.Character, args *state.HarvestArgs) bool {
		//		return c.IsInventoryFull()
		//	}),
		//	state.Craft("iron", func(c *character.Character, args *state.CraftingArgs) bool {
		//		return c.Inventory["iron_ore"] < 8
		//	}),
		//),
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

	onNewClient := func() {
		// Iterate over the character slice since it's ordered
		for _, charName := range characterNames {
			events <- Event{Character: characters[charName]}
		}
		events <- Event{Bank: theBank.Items}
	}

	server := httpServer(events, onNewClient)
	log.Fatal(server.Listen(":8080"))
}
