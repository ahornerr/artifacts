package graph2

//import (
//	"github.com/ahornerr/artifacts/character"
//	"github.com/ahornerr/artifacts/game"
//	"github.com/dominikbraun/graph"
//	"github.com/dominikbraun/graph/draw"
//	"log"
//	"os"
//	"reflect"
//	"strconv"
//)
//
//var equipmentTypes = map[string]bool{
//	"amulet":     true,
//	"body_armor": true,
//	"boots":      true,
//	"helmet":     true,
//	"leg_armor":  true,
//	"ring":       true,
//	"shield":     true,
//	"weapon":     true,
//}
//
//func BestEquipmentAgainst(target game.Stats, char *character.Character) {
//	equipmentTypeItems := map[string][]*game.Item{}
//	var allValidItems []*game.Item
//	for _, item := range game.Items.GetAll() {
//		if item.Level > char.GetLevel("combat") {
//			continue
//		}
//
//		if _, ok := equipmentTypes[item.Type]; !ok {
//			continue
//		}
//
//		if reflect.DeepEqual(item.Stats, game.Stats{}) {
//			// Item has no stats
//			continue
//		}
//		if equipmentTypeItems[item.Type] == nil {
//			equipmentTypeItems[item.Type] = []*game.Item{}
//		}
//		equipmentTypeItems[item.Type] = append(equipmentTypeItems[item.Type], item)
//		allValidItems = append(allValidItems, item)
//	}
//
//	//itemHash := func(item *game.Item) string {
//	//	return item.Code
//	//}
//
//	g := graph.New(graph.StringHash, graph.Weighted(), graph.Rooted(), graph.Acyclic(), graph.Directed())
//
//	//start := &game.Item{Code: }
//	err := g.AddVertex("Start")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	//end := &game.Item{Code: "End"}
//	err = g.AddVertex("End")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	elements := []string{"fire", "water", "air", "earth", "woodcutting", "fishing", "mining"}
//
//	for _, element := range elements {
//		err = g.AddVertex(element)
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
//	//
//	//	err = g.AddEdge("Start", element)
//	//	if err != nil {
//	//		log.Fatal(err)
//	//	}
//
//	var prevItems []*game.Item
//	for _, items := range equipmentTypeItems {
//		for _, item := range items {
//
//			err = g.AddVertex(item.Code)
//			if err != nil {
//				log.Fatal(err)
//			}
//
//			if len(prevItems) == 0 {
//				err = g.AddEdge("Start", item.Code)
//				if err != nil {
//					log.Fatal(err)
//				}
//			}
//
//			for element, attack := range item.Stats.Attack {
//				err = g.AddEdge(item.Code, element, graph.EdgeWeight(attack), graph.EdgeAttribute("label", strconv.Itoa(attack)))
//				if err != nil {
//					log.Fatal(err)
//				}
//
//				//for _, prevItem := range prevItems {
//				//	err = g.AddEdge(element, prevItem.Code, graph.EdgeWeight(attack), graph.EdgeAttribute("label", strconv.Itoa(attack)))
//				//	if err != nil {
//				//		log.Fatal(err)
//				//	}
//				//}
//			}
//		}
//
//		prevItems = items
//	}
//
//	for _, prevItem := range prevItems {
//		//prevAttack := prevItem.Stats.Attack[element]
//		//if prevAttack == 0 {
//		//	continue
//		//}
//
//		err = g.AddEdge(prevItem.Code, "End")
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
//	//}
//
//	//var prevVertexes map[string]game.Stats
//	//for _, items := range equipmentTypeItems {
//	//	newVertexes := map[string]game.Stats{}
//	//
//	//	for _, item := range items {
//	//		if len(prevVertexes) == 0 {
//	//			vertexName := "Start, " + item.Code
//	//			err = g.AddVertex(vertexName)
//	//			if err != nil {
//	//				log.Fatal(err)
//	//			}
//	//
//	//			newVertexes[vertexName] = item.Stats
//	//		}
//	//
//	//		for prevVertexName, prevVertexStats := range prevVertexes {
//	//			vertexName := prevVertexName + ", " + item.Code
//	//			err = g.AddVertex(vertexName)
//	//			if err != nil {
//	//				log.Fatal(err)
//	//			}
//	//
//	//			accumulatedStats := prevVertexStats.Add(item.Stats)
//	//
//	//			damage := accumulatedStats.GetDamageAgainst(target)
//	//			//resistance := target.GetDamageAgainst(accumulatedStats)
//	//			//score := damage - resistance
//	//			score := 10000 - damage
//	//
//	//			err = g.AddEdge(prevVertexName, vertexName, graph.EdgeWeight(int(score)), graph.EdgeAttribute("label", strconv.Itoa(int(score))))
//	//			if err != nil {
//	//				log.Fatal(err)
//	//			}
//	//
//	//			newVertexes[vertexName] = accumulatedStats
//	//		}
//	//	}
//	//
//	//	prevVertexes = newVertexes
//	//}
//
//	//var prevItems []*game.Item
//	//for _, items := range equipmentTypeItems {
//	//
//	//	for _, item := range items {
//	//		err = g.AddVertex(item)
//	//		if err != nil {
//	//			log.Fatal(err)
//	//		}
//	//		if len(prevStats) == 0 {
//	//			err = g.AddEdge("Start", item.Code, graph.EdgeData(item.Stats))
//	//			if err != nil {
//	//				log.Fatal(err)
//	//			}
//	//
//	//			newStats[item.Code] = item.Stats
//	//
//	//			edge, err := g.Edge("Start", item.Code)
//	//			if err != nil {
//	//				log.Fatal(err)
//	//			}
//	//			newEdges = append(newEdges, edge)
//	//		} else {
//	//			newStats[item.Code] = item.Stats
//	//		}
//	//		for prevItemCode, prevStat := range prevStats {
//	//			accumulated := prevStat.Add(item.Stats)
//	//
//	//			damage := accumulated.GetDamageAgainst(target)
//	//			//resistance := target.GetDamageAgainst(accumulated)
//	//			//score := damage - resistance
//	//			score := 10000 - damage
//	//			//err := g.AddEdge(item.Code, prevItem.Code, )
//	//
//	//			err = g.AddEdge(prevItemCode, item.Code, graph.EdgeWeight(int(score)), graph.EdgeAttribute("label", strconv.Itoa(int(score))), graph.EdgeData(accumulated))
//	//			if err != nil {
//	//				log.Fatal(err)
//	//			}
//	//		}
//	//
//	//		//for _, prevEdge := range prevEdges {
//	//		//	prevStats := prevEdge.Properties.Data.(game.Stats)
//	//		//
//	//		//	accumulated := prevStats.Add(item.Stats)
//	//		//
//	//		//	damage := accumulated.GetDamageAgainst(target)
//	//		//	//resistance := target.GetDamageAgainst(accumulated)
//	//		//	//score := damage - resistance
//	//		//	score := 10000 - damage
//	//		//	//err := g.AddEdge(item.Code, prevItem.Code, )
//	//		//
//	//		//	err = g.AddEdge(prevEdge.Target.Code, item.Code, graph.EdgeWeight(int(score)), graph.EdgeAttribute("label", strconv.Itoa(int(score))), graph.EdgeData(accumulated))
//	//		//	if err != nil {
//	//		//		log.Fatal(err)
//	//		//	}
//	//		//
//	//		//	edge, err := g.Edge(prevEdge.Target.Code, item.Code)
//	//		//	if err != nil {
//	//		//		log.Fatal(err)
//	//		//	}
//	//		//	newEdges = append(newEdges, edge)
//	//		//}
//	//	}
//	//
//	//	prevStats = newStats
//	//}
//
//	//for _, prevEdge := range prevEdges {
//	//	err = g.AddEdge(prevItemCode, "End")
//	//	if err != nil {
//	//		log.Fatal(err)
//	//	}
//	//}
//
//	//paths, err := graph.AllPathsBetween(g, "Start", "End")
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//	//_ = paths
//	//
//	//edges, err := g.Edges()
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//	//_ = edges
//
//	//accumulated := game.AccumulatedStats(map[string]string{"1": item.Code, "2": prevItem.Code})
//	//damage := accumulated.GetDamageAgainst(target)
//	////resistance := target.GetDamageAgainst(accumulated)
//	////score := damage - resistance
//	//score := 10000 - damage
//	//err := g.AddEdge(item.Code, prevItem.Code, graph.EdgeWeight(int(score)), graph.EdgeAttribute("label", strconv.Itoa(int(score))))
//
//	//for _, item := range allValidItems {
//	//	for _, otherItem := range allValidItems {
//	//		if item == otherItem || item.Type == otherItem.Type {
//	//			continue
//	//		}
//	//		fmt.Println(item.Code, "->", otherItem.Code)
//	//		accumulated := game.AccumulatedStats(map[string]string{"1": item.Code, "2": otherItem.Code})
//	//		damage := accumulated.GetDamageAgainst(target)
//	//		resistance := target.GetDamageAgainst(accumulated)
//	//		score := damage - resistance
//	//		err := g.AddEdge(item.Code, otherItem.Code, graph.EdgeWeight(int(score)), graph.EdgeAttribute("label", strconv.Itoa(int(score))))
//	//		if err != nil {
//	//			if !errors.Is(err, graph.ErrEdgeAlreadyExists) {
//	//				log.Fatal(err)
//	//			}
//	//		}
//	//	}
//	//}
//
//	//for equipmentType1, items1 := range equipmentTypeItems {
//	//	for _, item1 := range items1 {
//	//		for equipmentType2, items2 := range equipmentTypeItems {
//	//			if equipmentType1 == equipmentType2 {
//	//				continue
//	//			}
//	//			for _, item2 := range items2 {
//	//				fmt.Println(item1.Code, "->", item2.Code)
//	//				err := g.AddEdge(item1.Code, item2.Code) // TODO weights
//	//				if err != nil {
//	//					log.Fatal(err)
//	//				}
//	//			}
//	//		}
//	//	}
//	//}
//
//	//g, err = graph.MinimumSpanningTree(g)
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//
//	//shortest, err := graph.ShortestPath(g, "Start", "End")
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//	//
//	//_ = shortest
//
//	//g, err = graph.MaximumSpanningTree(g)
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//
//	//graph.Rooted()
//
//	//g, err = graph.StronglyConnectedComponents(g)
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//
//	//g, err = graph.TransitiveReduction(g)
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//
//	file, _ := os.Create("./equipment.dot")
//	err = draw.DOT(g, file, draw.GraphAttribute("label", "ogre"))
//	if err != nil {
//		log.Fatal(err)
//	}
//}
