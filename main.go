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
				char2 := *char
				nonBlockingWriteEvent(Event{Character: &char2})
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

	// Priority (all):
	//   1. Events - because they're sporadic, although you can't really do any until late in the game
	//   2. Tasks - because they take a while and give good rewards
	//     a. If we can't complete a task, we need better gear. Better gear may require a higher combat level.
	//        Only one character is designated to crafting gear. Wait for them to make better gear.
	//        If we're too low level for the gear, we need to train.
	//        Ideally via task (for task rewards) so consider exchanging task if possible.
	// Priority (harvester):
	//   3. Mining
	//     a. Preferred over woodcutting because we get gems
	//     b. Refining - it gives XP and the crafter character will need the resources.
	//        Don't refine everything though, iron ore for example is needed for later recipes
	//     c. Move on to mining the next resource when we are no longer receiving XP anymore, as opposed to
	//        when we have the level for the next resource. Better to stockpile some lower stuff for crafting later
	//   4. Woodcutting
	//     a. Plank make - it gives XP and the crafter character will need the resources
	//        Don't make everything though, ash wood for example is needed for later recipes
	//     c. Move on to mining the next resource when we are no longer receiving XP anymore, as opposed to
	//        when we have the level for the next resource. Better to stockpile some lower stuff for crafting later
	// Priority (crafter):
	//   3. Make 5x (10 for rings) of the highest level weapons/gear/jewelry that you can craft
	//     a. Consider buying higher level items on the GE if we don't have the level to craft them yet
	//   4. Train crafting skills
	//     a. Level crafting skills evenly, 5 levels at a time. Seems that weaponcrafting is most important at the start
	//        but later on (level 20-25) jewelrycrafting may become a bit more important.
	//     a. Figure out when we stop getting XP for crafting (levels above item). Lower level items still give XP
	//        up to a certain point and typically simpler ingredients to craft. Only move on to the next item when
	//        we're not getting XP from the lower level items
	//     b. Write a cost function that evaluates difficulty of obtaining crafting materials. Select the "cheapest"
	//        item for crafting in a given tier. Typically, this would exclude materials like jasper crystals.
	//        Take into account drop rates of items, and estimate how long it will take to kill a given monster.
	//        We can probably just look at monster HP as an analog for difficulty.
	//        Likewise with crafted items we should look at the materials required to craft them.
	//        TODO: How does this work when we have some banked materials already?

	crafterWants := make(chan game.ItemQuantity, 100)

	characterStates := map[string]state.Runner{
		"curlyBoy1": state.RoleCrafter(characters, crafterWants),
		"curlyBoy2": state.RoleHarvester(crafterWants),
		"curlyBoy3": state.RoleHarvester(crafterWants),
		"curlyBoy4": state.RoleHarvester(crafterWants),
		"curlyBoy5": state.RoleHarvester(crafterWants),
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
			char := *characters[charName]
			events <- Event{Character: &char}
		}
		events <- Event{Bank: theBank.Items}
	}

	server := httpServer(events, onNewClient)
	log.Fatal(server.Listen(":8080"))
}
