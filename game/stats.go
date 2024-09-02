package game

import (
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"log"
	"strings"
)

type Stats struct {
	Hp      uint16
	Restore int8
	Haste   int8
	BoostHp int8

	AttackFire  int8
	AttackWater int8
	AttackEarth int8
	AttackAir   int8

	AttackWoodcutting int8
	AttackMining      int8
	AttackFishing     int8

	ResistFire  int8
	ResistWater int8
	ResistEarth int8
	ResistAir   int8

	ResistWoodcutting int8
	ResistMining      int8
	ResistFishing     int8

	DamageFire  int8
	DamageWater int8
	DamageEarth int8
	DamageAir   int8

	IsTool     bool
	IsResource bool

	// TODO: BoostDamage for each element
}

func (s *Stats) Add(other *Stats) {
	s.Hp += other.Hp
	s.Restore += other.Restore
	s.Haste += other.Haste
	s.BoostHp += other.BoostHp

	s.AttackFire += other.AttackFire
	s.AttackWater += other.AttackWater
	s.AttackEarth += other.AttackEarth
	s.AttackAir += other.AttackAir

	s.AttackWoodcutting += other.AttackWoodcutting
	s.AttackFishing += other.AttackFishing
	s.AttackMining += other.AttackMining

	s.ResistFire += other.ResistFire
	s.ResistWater += other.ResistWater
	s.ResistEarth += other.ResistEarth
	s.ResistAir += other.ResistAir

	s.ResistWoodcutting += other.ResistWoodcutting
	s.ResistFishing += other.ResistFishing
	s.ResistMining += other.ResistMining

	s.DamageFire += other.DamageFire
	s.DamageWater += other.DamageWater
	s.DamageEarth += other.DamageEarth
	s.DamageAir += other.DamageAir
}

func AccumulatedStats(items map[string]*Item) *Stats {
	accumulated := &Stats{}
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

	if s.AttackAir > 0 {
		totalDamage += float64(s.AttackAir) *
			(1 + float64(s.DamageAir)/100.0) *
			(1 - float64(other.ResistAir)/100.0) *
			(1 - float64(other.ResistAir)/1000.0)
	}

	if s.AttackFire > 0 {
		totalDamage += float64(s.AttackFire) *
			(1 + float64(s.DamageFire)/100.0) *
			(1 - float64(other.ResistFire)/100.0) *
			(1 - float64(other.ResistFire)/1000.0)
	}

	if s.AttackWater > 0 {
		totalDamage += float64(s.AttackWater) *
			(1 + float64(s.DamageWater)/100.0) *
			(1 - float64(other.ResistWater)/100.0) *
			(1 - float64(other.ResistWater)/1000.0)
	}

	if s.AttackEarth > 0 {
		totalDamage += float64(s.AttackEarth) *
			(1 + float64(s.DamageEarth)/100.0) *
			(1 - float64(other.ResistEarth)/100.0) *
			(1 - float64(other.ResistEarth)/1000.0)
	}

	if other.IsResource {
		if s.AttackWoodcutting > 0 && other.ResistWoodcutting < 0 {
			totalDamage += float64(s.AttackWoodcutting) *
				(1 - float64(other.ResistWoodcutting)/100.0) *
				(1 - float64(other.ResistWoodcutting)/1000.0)
		}

		if s.AttackMining > 0 && other.ResistMining < 0 {
			totalDamage += float64(s.AttackMining) *
				(1 - float64(other.ResistMining)/100.0) *
				(1 - float64(other.ResistMining)/1000.0)
		}

		if s.AttackFishing > 0 && other.ResistFishing < 0 {
			totalDamage += float64(s.AttackFishing) *
				(1 - float64(other.ResistFishing)/100.0) *
				(1 - float64(other.ResistFishing)/1000.0)
		}
	}

	return totalDamage
}

func StatsFromMonster(monsterSchema client.MonsterSchema) *Stats {
	return &Stats{
		Hp: uint16(monsterSchema.Hp),

		AttackFire:  int8(monsterSchema.AttackFire),
		AttackWater: int8(monsterSchema.AttackWater),
		AttackEarth: int8(monsterSchema.AttackEarth),
		AttackAir:   int8(monsterSchema.AttackAir),

		ResistFire:  int8(monsterSchema.ResFire),
		ResistWater: int8(monsterSchema.ResWater),
		ResistEarth: int8(monsterSchema.ResEarth),
		ResistAir:   int8(monsterSchema.ResAir),
	}
}

func StatsFromItem(itemSchema client.ItemSchema) *Stats {
	if itemSchema.Effects == nil || len(*itemSchema.Effects) == 0 {
		return nil
	}

	stats := &Stats{}

	for _, effect := range *itemSchema.Effects {
		value := int8(effect.Value)
		switch {
		case effect.Name == "hp":
			stats.Hp = uint16(effect.Value)
		case effect.Name == "restore":
			stats.Restore = value
		case effect.Name == "boost_hp":
			stats.BoostHp = value
		case effect.Name == "haste":
			// For food
			stats.Haste = value
		case effect.Name == "woodcutting":
			// Value is an integer that reduces cooldown time by Value%
			stats.AttackWoodcutting = int8(-effect.Value)
			stats.IsTool = true
		case effect.Name == "fishing":
			stats.AttackFishing = int8(-effect.Value)
			stats.IsTool = true
		case effect.Name == "mining":
			stats.AttackMining = int8(-effect.Value)
			stats.IsTool = true
		case strings.HasPrefix(effect.Name, "dmg_"):
			element := strings.TrimPrefix(effect.Name, "dmg_")
			switch element {
			case "air":
				stats.DamageAir = value
			case "water":
				stats.DamageWater = value
			case "fire":
				stats.DamageFire = value
			case "earth":
				stats.DamageEarth = value
			default:
				panic(element)
			}
		case strings.HasPrefix(effect.Name, "attack_"):
			element := strings.TrimPrefix(effect.Name, "attack_")
			switch element {
			case "air":
				stats.AttackAir = value
			case "water":
				stats.AttackWater = value
			case "fire":
				stats.AttackFire = value
			case "earth":
				stats.AttackEarth = value
			default:
				panic(element)
			}
		case strings.HasPrefix(effect.Name, "res_"):
			element := strings.TrimPrefix(effect.Name, "res_")
			switch element {
			case "air":
				stats.ResistAir = value
			case "water":
				stats.ResistWater = value
			case "fire":
				stats.ResistFire = value
			case "earth":
				stats.ResistEarth = value
			default:
				panic(element)
			}
		case strings.HasPrefix(effect.Name, "boost_dmg_"):
			// Value is a percentage (0-100), comes from food
			//element := strings.TrimPrefix(effect.Name, "boost_dmg_")
			//stats.BoostDamage[element] = effect.Value
		default:
			log.Fatal("Unrecognized effect named ", effect.Name, " with value ", value)
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
