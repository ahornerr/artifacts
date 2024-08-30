package game

import (
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"log"
	"strings"
)

type Stats struct {
	Hp          int
	Restore     int
	Haste       int
	BoostHp     int
	Attack      map[string]int
	Resistance  map[string]int
	Damage      map[string]int
	BoostDamage map[string]int
}

func (s *Stats) Add(other *Stats) {
	s.Hp += other.Hp
	s.Restore += other.Restore
	s.Haste += other.Haste
	s.BoostHp += other.BoostHp
	for element, attack := range other.Attack {
		s.Attack[element] += attack
	}
	for element, resistance := range other.Resistance {
		s.Resistance[element] += resistance
	}
	for element, damage := range other.Damage {
		s.Damage[element] += damage
	}
	for element, boost := range other.BoostDamage {
		s.BoostDamage[element] += boost
	}
}

func (s *Stats) Remove(other *Stats) {
	s.Hp -= other.Hp
	s.Restore -= other.Restore
	s.Haste -= other.Haste
	s.BoostHp -= other.BoostHp
	for element, attack := range other.Attack {
		s.Attack[element] -= attack
	}
	for element, resistance := range other.Resistance {
		s.Resistance[element] -= resistance
	}
	for element, damage := range other.Damage {
		s.Damage[element] -= damage
	}
	for element, boost := range other.BoostDamage {
		s.BoostDamage[element] -= boost
	}
}

func AccumulatedStats(items map[string]*Item) *Stats {
	accumulated := &Stats{
		Attack:      map[string]int{"fire": 0, "water": 0, "earth": 0, "air": 0},
		Resistance:  map[string]int{"fire": 0, "water": 0, "earth": 0, "air": 0},
		Damage:      map[string]int{"fire": 0, "water": 0, "earth": 0, "air": 0},
		BoostDamage: map[string]int{"fire": 0, "water": 0, "earth": 0, "air": 0},
	}
	for _, item := range items {
		accumulated.Add(item.Stats)
	}
	return accumulated
}

func AccumulatedStatsItemCodes(itemCodes map[string]string) *Stats {
	items := map[string]*Item{}
	for slot, itemCode := range itemCodes {
		if itemCode == "" {
			continue
		}
		item := Items.Get(itemCode)
		if item == nil {
			continue
		}
		items[slot] = item
	}
	return AccumulatedStats(items)
}

func (s Stats) GetDamageAgainst(other *Stats) float64 {
	totalDamage := 0.0

	for element, attack := range s.Attack {
		damage := s.Damage[element]
		resistance := other.Resistance[element]

		totalDamage += float64(attack) *
			(1 + float64(damage)/100.0) *
			(1 - float64(resistance)/100.0) *
			(1 - float64(resistance)/1000.0)
	}

	return totalDamage
}

func StatsFromMonster(monsterSchema client.MonsterSchema) *Stats {
	return &Stats{
		Hp: monsterSchema.Hp,
		Attack: map[string]int{
			"air":   monsterSchema.AttackAir,
			"earth": monsterSchema.AttackEarth,
			"fire":  monsterSchema.AttackFire,
			"water": monsterSchema.AttackWater,
		},
		Resistance: map[string]int{
			"air":   monsterSchema.ResAir,
			"earth": monsterSchema.ResEarth,
			"fire":  monsterSchema.ResFire,
			"water": monsterSchema.ResWater,
		},
	}
}

func StatsFromItem(itemSchema client.ItemSchema) *Stats {
	if itemSchema.Effects == nil || len(*itemSchema.Effects) == 0 {
		return nil
	}

	stats := &Stats{
		Attack:      map[string]int{},
		Resistance:  map[string]int{},
		Damage:      map[string]int{},
		BoostDamage: map[string]int{},
	}

	for _, effect := range *itemSchema.Effects {
		switch {
		case effect.Name == "hp":
			stats.Hp = effect.Value
		case effect.Name == "restore":
			stats.Restore = effect.Value
		case effect.Name == "boost_hp":
			stats.BoostHp = effect.Value
		case effect.Name == "haste":
			// For food
			stats.Haste = effect.Value
		case effect.Name == "woodcutting":
			// Value is an integer that reduces cooldown time by Value%
			stats.Attack[effect.Name] = -effect.Value
		case effect.Name == "fishing":
			stats.Attack[effect.Name] = -effect.Value
		case effect.Name == "mining":
			stats.Attack[effect.Name] = -effect.Value
		case strings.HasPrefix(effect.Name, "dmg_"):
			element := strings.TrimPrefix(effect.Name, "dmg_")
			stats.Damage[element] = effect.Value
		case strings.HasPrefix(effect.Name, "attack_"):
			element := strings.TrimPrefix(effect.Name, "attack_")
			stats.Attack[element] = effect.Value
		case strings.HasPrefix(effect.Name, "res_"):
			element := strings.TrimPrefix(effect.Name, "res_")
			stats.Resistance[element] = effect.Value
		case strings.HasPrefix(effect.Name, "boost_dmg_"):
			// Value is a percentage (0-100), comes from food
			element := strings.TrimPrefix(effect.Name, "boost_dmg_")
			stats.BoostDamage[element] = effect.Value
		default:
			log.Fatal("Unrecognized effect named", effect.Name, "with value", effect.Value)
		}
	}

	return stats
}

//func (s Stats) GetAttacks() map[string]float64 {
//	attacks := map[string]float64{}
//	for element, attack := range s.Attack {
//		damage := s.Damage[element]
//		attacks[element] = float64(attack) * (1.0 + (float64(damage) / 100.0))
//	}
//	return attacks
//}
//
//func (s Stats) GetAttacksWithDamageReduction(attacks map[string]float64) map[string]float64 {
//	attacksWithDamageReduction := map[string]float64{}
//	for element, attack := range attacks {
//		resistance := s.Resistance[element]
//		attacksWithDamageReduction[element] = attack * (1.0 - (float64(resistance) / 100.0))
//	}
//	return attacksWithDamageReduction
//}
//
//func (s Stats) GetBlocking() map[string]float64 {
//	blocking := map[string]float64{}
//	for element, resistance := range s.Resistance {
//		if resistance != 0 {
//			blocking[element] = 100.0 / (float64(resistance) * 0.01) * 0.01
//		}
//	}
//	return blocking
//}
