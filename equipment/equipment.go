package main

import (
	"encoding/json"
	"github.com/ahornerr/artifacts/game"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
)

var equipmentTypes = map[string]bool{
	"amulet":     true,
	"body_armor": true,
	"boots":      true,
	"helmet":     true,
	"leg_armor":  true,
	"ring":       true,
	"shield":     true,
	"weapon":     true,
}

type EquipmentSet struct {
	Equipment          map[string]*game.Item
	TurnsToKillMonster int
	TurnsToKillPlayer  int
}

func NewEquipmentSet(other *EquipmentSet) *EquipmentSet {
	s := &EquipmentSet{
		Equipment: map[string]*game.Item{},
	}

	if other != nil {
		// Copy best set to a new set, might not be the most efficient thing
		for slot, item := range other.Equipment {
			s.Equipment[slot] = item
		}
	}

	return s
}

func (s EquipmentSet) MarshalJSON() ([]byte, error) {
	j := struct {
		Equipment          map[string]string `json:"equipment"`
		Stats              game.Stats        `json:"stats"`
		TurnsToKillMonster int               `json:"turns_to_kill_monster"`
		TurnsToKillPlayer  int               `json:"turns_to_kill_player"`
		Winnable           bool              `json:"winnable"`
	}{
		Equipment:          map[string]string{},
		Stats:              *game.AccumulatedStats(s.Equipment),
		TurnsToKillMonster: s.TurnsToKillMonster,
		TurnsToKillPlayer:  s.TurnsToKillPlayer,
		Winnable:           s.TurnsToKillMonster < s.TurnsToKillPlayer && s.TurnsToKillMonster <= 50,
	}

	for slot, item := range s.Equipment {
		j.Equipment[slot] = item.Code
	}

	return json.Marshal(j)
}

type LevelSet map[*game.Monster]EquipmentSet

func (s LevelSet) MarshalJSON() ([]byte, error) {
	mapWithMonsterCodes := map[string]EquipmentSet{}
	for monster, levelSet := range s {
		mapWithMonsterCodes[monster.Code] = levelSet
	}

	return json.Marshal(mapWithMonsterCodes)
}

func main() {
	levelGroups := map[int]bool{}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	var err error
	//file, err := os.Create("./cpu.pprof")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//err = pprof.StartCPUProfile(file)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer pprof.StopCPUProfile()

	slotsEquipment := map[string][]*game.Item{}
	for _, item := range game.Items.GetAll() {
		if _, ok := equipmentTypes[item.Type]; !ok {
			continue
		}
		// item.Stats should never be nil
		slot := item.Type
		if _, ok := slotsEquipment[slot]; !ok {
			slotsEquipment[slot] = []*game.Item{}
		}
		slotsEquipment[slot] = append(slotsEquipment[slot], item)
		levelGroups[item.Level] = true
	}

	levelGroups = map[int]bool{1: true, 10: true}
	//levelGroups = map[int]bool{1: true}

	slotsEquipmentForLevel := map[int]map[string][]*game.Item{}
	for level := range levelGroups {
		slotsEquipmentForLevel[level] = map[string][]*game.Item{}
		for slot, slotEquipment := range slotsEquipment {
			slotsEquipmentForLevel[level][slot] = []*game.Item{}
			for _, equipment := range slotEquipment {
				if equipment.Level > level {
					continue
				}
				// TODO: Filter out irrelevant items that are too low level
				if equipment.Level < level-5 {
					continue
				}
				// TODO: Exclude tools
				if equipment.SubType == "tool" {
					continue
				}
				if equipment.Level != level {
					continue
				}
				slotsEquipmentForLevel[level][slot] = append(slotsEquipmentForLevel[level][slot], equipment)
			}
			if len(slotsEquipmentForLevel[level][slot]) == 0 {
				delete(slotsEquipmentForLevel[level], slot)
			}
		}
	}

	bestLevelSets := map[int]LevelSet{}
	for level := range levelGroups {
		bestLevelSets[level] = computeBestEquipmentSetsForLevel(level, slotsEquipmentForLevel[level])
	}

	file, err := os.Create("./best_equipment.json")
	if err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(bestLevelSets)
	if err != nil {
		log.Fatal(err)
	}
}

type MonsterEquipment struct {
	Monster   *game.Monster
	Equipment EquipmentSet
}

func computeBestEquipmentSetsForLevel(level int, slotsEquipment map[string][]*game.Item) LevelSet {
	bestSets := LevelSet{}

	monsters := game.Monsters.GetAll()
	resultsChan := make(chan MonsterEquipment, len(monsters))

	for _, monster := range monsters {
		go func(m *game.Monster) {
			equipment := computeBestEquipmentSetForLevelAndMonster(level, slotsEquipment, m)
			log.Println("Computed", m.Name, "level", level)
			resultsChan <- MonsterEquipment{
				Monster:   m,
				Equipment: equipment,
			}
		}(monster)

		bestSets[monster] = computeBestEquipmentSetForLevelAndMonster(level, slotsEquipment, monster)
	}

	for {
		me := <-resultsChan
		bestSets[me.Monster] = me.Equipment
		if len(bestSets) == len(monsters) {
			break
		}
	}

	return bestSets
}

func computeBestEquipmentSetForLevelAndMonster(level int, slotsEquipment map[string][]*game.Item, monster *game.Monster) EquipmentSet {
	set := NewEquipmentSet(nil)
	stats := game.AccumulatedStats(set.Equipment)
	turnsToKillMonster, turnsToKillPlayer := computeBestForRestOfSet(set, stats, level, slotsEquipment, monster)
	set.TurnsToKillPlayer = turnsToKillPlayer
	set.TurnsToKillMonster = turnsToKillMonster
	return *set
}

func computeBestForRestOfSet(set *EquipmentSet, stats *game.Stats, level int, slotsEquipment map[string][]*game.Item, monster *game.Monster) (int, int) {
	if len(set.Equipment) == len(slotsEquipment) {
		turnsToKillMonster, turnsToKillPlayer := calculateTurns(stats, monster, level)
		return turnsToKillMonster, turnsToKillPlayer
	}

	//bestEquipment := NewEquipmentSet(nil)
	//fewestTurnsToKillMonster := math.MaxInt32
	//mostTurnsToKillPlayer := -math.MaxInt32

	// Pick a slot. Pick an item. Put an item in the slot. Calculate all other slots

	for slot, items := range slotsEquipment {
		if set.Equipment[slot] != nil {
			continue
		}
		fewestTurnsToKillMonsterInSlot := math.MaxInt32
		mostTurnsToKillPlayerInSlot := -math.MaxInt32
		var bestItem *game.Item
		for _, item := range items {
			set.Equipment[slot] = item
			stats.Add(item.Stats)

			turnsToKillMonster, turnsToKillPlayer := computeBestForRestOfSetInner(set, stats, level, slotsEquipment, monster)

			// Minimize turnsToKillMonster, maximize turnsToKillPlayer
			if turnsToKillMonster < fewestTurnsToKillMonsterInSlot {
				mostTurnsToKillPlayerInSlot = turnsToKillPlayer
				fewestTurnsToKillMonsterInSlot = turnsToKillMonster
				//bestEquipment.Equipment[slot] = item
				bestItem = item
			}

			// Same number of turns to kill the monster but more turns to kill the player (better resistance)
			if turnsToKillMonster == fewestTurnsToKillMonsterInSlot {
				if turnsToKillPlayer > mostTurnsToKillPlayerInSlot {
					mostTurnsToKillPlayerInSlot = turnsToKillPlayer
					//bestEquipment.Equipment[slot] = item
					bestItem = item
				}
			}

			stats.Remove(item.Stats)
			delete(set.Equipment, slot)
		}
		if bestItem != nil {
			set.Equipment[slot] = bestItem
			stats.Add(bestItem.Stats)
		}
	}

	return calculateTurns(stats, monster, level)
}

func computeBestForRestOfSetInner(set *EquipmentSet, stats *game.Stats, level int, slotsEquipment map[string][]*game.Item, monster *game.Monster) (int, int) {
	if len(set.Equipment) == len(slotsEquipment) {
		turnsToKillMonster, turnsToKillPlayer := calculateTurns(stats, monster, level)
		return turnsToKillMonster, turnsToKillPlayer
	}

	//bestEquipment := NewEquipmentSet(nil)
	//fewestTurnsToKillMonster := math.MaxInt32
	//mostTurnsToKillPlayer := -math.MaxInt32

	// Pick a slot. Pick an item. Put an item in the slot. Calculate all other slots

	for slot, items := range slotsEquipment {
		if set.Equipment[slot] != nil {
			continue
		}
		fewestTurnsToKillMonsterInSlot := math.MaxInt32
		mostTurnsToKillPlayerInSlot := -math.MaxInt32
		var bestItem *game.Item
		for _, item := range items {
			set.Equipment[slot] = item
			stats.Add(item.Stats)

			turnsToKillMonster, turnsToKillPlayer := computeBestForRestOfSetInner(set, stats, level, slotsEquipment, monster)

			// Minimize turnsToKillMonster, maximize turnsToKillPlayer
			if turnsToKillMonster < fewestTurnsToKillMonsterInSlot {
				mostTurnsToKillPlayerInSlot = turnsToKillPlayer
				fewestTurnsToKillMonsterInSlot = turnsToKillMonster
				//bestEquipment.Equipment[slot] = item
				bestItem = item
			}

			// Same number of turns to kill the monster but more turns to kill the player (better resistance)
			if turnsToKillMonster == fewestTurnsToKillMonsterInSlot {
				if turnsToKillPlayer > mostTurnsToKillPlayerInSlot {
					mostTurnsToKillPlayerInSlot = turnsToKillPlayer
					//bestEquipment.Equipment[slot] = item
					bestItem = item
				}
			}

			stats.Remove(item.Stats)
			delete(set.Equipment, slot)
		}
		if bestItem != nil {
			//set.Equipment[slot] = bestItem
			stats.Add(bestItem.Stats)

			defer func(s *game.Stats) {
				stats.Remove(s)
			}(bestItem.Stats)
		}
	}

	return calculateTurns(stats, monster, level)
}

func calculateTurns(stats *game.Stats, monster *game.Monster, level int) (int, int) {
	playerAttack := int(stats.GetDamageAgainst(monster.Stats))
	monsterAttack := int(monster.Stats.GetDamageAgainst(stats))

	playerHp := stats.Hp + 120 + (5 * level)

	var turnsToKillMonster int
	if playerAttack <= 0 {
		turnsToKillMonster = math.MaxInt32
	} else {
		turnsToKillMonster = int(math.Ceil(float64(monster.Stats.Hp) / float64(playerAttack)))
	}

	var turnsToKillPlayer int
	if monsterAttack <= 0 {
		turnsToKillPlayer = math.MaxInt32
	} else {
		turnsToKillPlayer = int(math.Ceil(float64(playerHp) / float64(monsterAttack)))
	}

	return turnsToKillMonster, turnsToKillPlayer
}
