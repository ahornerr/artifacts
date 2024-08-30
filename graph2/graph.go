package graph2

import (
	"fmt"
	"github.com/ahornerr/artifacts/game"
	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
	"log"
	"os"
	"strconv"
)

func FromItem(item *game.Item, quantity int) {
	stringerHash := func(stringer fmt.Stringer) string {
		return stringer.String()
	}

	//g := graph.New(stringerHash, graph.Weighted(), graph.Directed(), graph.Acyclic())
	g := graph.New(stringerHash, graph.Weighted(), graph.Acyclic())

	graphItem(g, item, quantity)

	var err error

	//g, err = graph.TransitiveReduction(g)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//g, err = graph.MinimumSpanningTree(g)
	//if err != nil {
	//	log.Fatal(err)
	//}

	g, err = graph.MaximumSpanningTree(g)
	if err != nil {
		log.Fatal(err)
	}

	file, _ := os.Create("./mygraph.dot")
	err = draw.DOT(g, file)
	if err != nil {
		log.Fatal(err)
	}
}

type Action string

func (a Action) String() string {
	return string(a)
}

func graphItem(g graph.Graph[string, fmt.Stringer], item *game.Item, quantity int) {
	err := g.AddVertex(item)
	if err != nil {
		log.Fatal(err)
	}

	withdraw := Action("Withdraw " + item.Name)
	err = g.AddVertex(withdraw)
	if err != nil {
		log.Fatal(err)
	}
	err = g.AddEdge(item.Name, withdraw.String(), graph.EdgeWeight(20), graph.EdgeAttribute("label", strconv.Itoa(quantity)))
	if err != nil {
		log.Fatal(err)
	}

	if item.Crafting != nil {
		craft := Action("Craft " + item.Name)
		err = g.AddVertex(craft)
		if err != nil {
			log.Fatal(err)
		}
		err = g.AddEdge(item.Name, craft.String())
		if err != nil {
			log.Fatal(err)
		}

		for craftingItem, craftingQuantity := range item.Crafting.Items {
			weight := craftingQuantity * quantity
			graphItem(g, craftingItem, weight)
			err = g.AddEdge(craft.String(), craftingItem.Name, graph.EdgeWeight(weight), graph.EdgeAttribute("label", strconv.Itoa(weight)))
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	for _, resource := range game.Resources.ResourcesForItem(item) {
		err = g.AddVertex(resource)
		if err != nil {
			log.Fatal(err)
		}
		cooldown := 30.0
		weight := int(cooldown*1.0/avgDropPerAction(resource.Loot, item.Code)) * quantity
		err = g.AddEdge(item.Name, resource.Name, graph.EdgeWeight(weight), graph.EdgeAttribute("label", strconv.Itoa(weight)))
		//err = g.AddEdge(item.Name, resource.Name)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, monster := range game.Monsters.MonstersForItem(item) {
		err = g.AddVertex(monster)
		if err != nil {
			log.Fatal(err)
		}
		weight := monster.Level
		err = g.AddEdge(item.Name, monster.Name, graph.EdgeWeight(weight), graph.EdgeAttribute("label", strconv.Itoa(weight)))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func avgDropPerAction(drops []game.Drop, itemCode string) float64 {
	// Weight calculation here is pretty simple.
	// We want some quantity of items from the resource, and we know its drop rate.
	// From here we can calculate the average number of drops per harvest.
	// Weight is roughly correlated with time, so lower weights are more desirable.
	// Take the quantity and divide by the average drop per harvest (lower drop rate increases weight).
	// This also means that higher quantity increases rate too.
	var avgDrop float64
	for _, drop := range drops {
		if drop.Item.Code == itemCode {
			avgDropQuantity := float64(drop.MinQuantity+drop.MaxQuantity) / 2.0
			avgDrop = avgDropQuantity * 1 / float64(drop.Rate)
			break
		}
	}

	return avgDrop

	//if avgDropPerHarvest == 0 {
	//	return 0
	//}
	//
	//// Don't divide by zero if the resource doesn't drop it (for some reason)
	//weight := math.MaxFloat64
	//if avgDropPerHarvest > 0 {
	//	//weight = float64(quantity) / avgDropPerHarvest
	//	weight = 1.0 / avgDropPerHarvest
	//}
	//
	//return weight
}
