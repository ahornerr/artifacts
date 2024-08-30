package main

import (
	"context"
	"github.com/ahornerr/artifacts/bank"
	"github.com/ahornerr/artifacts/character"
	"github.com/ahornerr/artifacts/client"
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
