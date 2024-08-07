package game

import (
	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

type Item struct {
	// Code Item code. This is the item's unique identifier (ID).
	Code string

	// Name Item name
	Name string

	// Type Item type. In the case of armor, it will be the name of the slot e.g. "boots"
	Type string

	// Level Item level.
	Level int

	// Crafting will be nil if the object cannot be crafted.
	Crafting *Crafting

	// DropsFrom resources that drop this item
	//DropsFrom []client.ResourceSchema

	// Effects List of object effects. For equipment, it will include item stats.
	Effects *[]client.ItemEffectSchema

	Stats Stats
}

// TODO: This might be wrong
func (i Item) GetAllSkillRequirements() map[string]int {
	skills := map[string]int{}
	if i.Crafting != nil {
		skills[i.Crafting.Skill] = i.Crafting.Level

		for craftingItem := range i.Crafting.Items {
			for skill, level := range craftingItem.GetAllSkillRequirements() {
				if level > skills[skill] {
					skills[skill] = level
				}
			}
		}
	}
	return skills
}
