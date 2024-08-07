package game

import (
	"context"
	"github.com/ahornerr/artifacts/client"
	"log"
)

var gameClient, _ = client.New("")

var Items = newItems(gameClient)
var Maps = newMaps(gameClient) // TODO: May need to reload these occasionally
var Monsters = newMonsters(gameClient)
var Resources = newResources(gameClient)

func init() {
	ctx := context.Background()

	if err := Items.load(ctx); err != nil {
		log.Fatal("Loading items failed: ", err)
	}
	if err := Maps.load(ctx); err != nil {
		log.Fatal("Loading maps failed: ", err)
	}
	if err := Monsters.load(ctx); err != nil {
		log.Fatal("Loading monsters failed: ", err)
	}
	if err := Resources.load(ctx); err != nil {
		log.Fatal("Loading resources failed: ", err)
	}
}
