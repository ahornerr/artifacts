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
		"curlyBoy1": state.Loop(
			state.CollectItems("serpent_skin_legs_armor", 5, true, true, characters),
			state.CollectItems("magic_wizard_hat", 5, true, true, characters),

			state.MakeX("life_ring", 10, true, func(character *character.Character, args *state.MakeXArgs) bool {
				return character.GetLevel("jewelrycrafting") >= 20
			}),

			state.CollectItems("dreadful_amulet", 5, true, true, characters),
			state.CollectItems("dreadful_ring", 10, true, true, characters),
			state.CollectItems("skull_ring", 10, true, true, characters),

			state.MakeX("greater_wooden_staff", 10, true, func(character *character.Character, args *state.MakeXArgs) bool {
				return character.GetLevel("weaponcrafting") >= 20
			}),

			state.MakeX("battlestaff", 5, true, func(character *character.Character, args *state.MakeXArgs) bool {
				return character.GetLevel("weaponcrafting") >= 25
			}),

			state.CollectItems("skull_wand", 5, true, true, characters),
			state.CollectItems("dreadful_staff", 5, true, true, characters),
		),
		"curlyBoy2": state.Loop(
			state.MakeX("feather", 10, false, func(character *character.Character, args *state.MakeXArgs) bool {
				return args.Made >= 50
			}),
			state.MakeX("blue_slimeball", 10, false, func(character *character.Character, args *state.MakeXArgs) bool {
				return args.Made >= 100
			}),
			state.MakeX("wolf_bone", 10, false, func(character *character.Character, args *state.MakeXArgs) bool {
				return args.Made >= 50
			}),
		),
		"curlyBoy3": state.Loop(
			state.MakeX("hardwood_plank", 18, false, func(character *character.Character, args *state.MakeXArgs) bool {
				return character.Bank()["birch_wood"] < 5
			}),
			state.MakeX("feather", 10, false, func(character *character.Character, args *state.MakeXArgs) bool {
				return args.Made >= 50
			}),
			state.MakeX("blue_slimeball", 10, false, func(character *character.Character, args *state.MakeXArgs) bool {
				return args.Made >= 100
			}),
			state.MakeX("wolf_bone", 10, false, func(character *character.Character, args *state.MakeXArgs) bool {
				return args.Made >= 50
			}),
		),
		"curlyBoy4": state.Loop(
			state.MakeX("feather", 10, false, func(character *character.Character, args *state.MakeXArgs) bool {
				return args.Made >= 50
			}),
			state.MakeX("blue_slimeball", 10, false, func(character *character.Character, args *state.MakeXArgs) bool {
				return args.Made >= 100
			}),
			state.MakeX("wolf_bone", 10, false, func(character *character.Character, args *state.MakeXArgs) bool {
				return args.Made >= 50
			}),
		),
		"curlyBoy5": state.Loop(
			state.MakeX("hardwood_plank", 18, false, func(character *character.Character, args *state.MakeXArgs) bool {
				return character.Bank()["birch_wood"] < 5
			}),
			state.MakeX("feather", 10, false, func(character *character.Character, args *state.MakeXArgs) bool {
				return args.Made >= 50
			}),
			state.MakeX("blue_slimeball", 10, false, func(character *character.Character, args *state.MakeXArgs) bool {
				return args.Made >= 100
			}),
			state.MakeX("wolf_bone", 10, false, func(character *character.Character, args *state.MakeXArgs) bool {
				return args.Made >= 50
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
