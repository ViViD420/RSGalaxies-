package game

import (
	"errors"
	"math"
	"strings"
)

const (
	MaxSkillLevel      = 99
	MaxSkillExperience = 200_000_000
)

var (
	ErrUnknownSkill       = errors.New("unknown skill")
	ErrInvalidExperience  = errors.New("experience award must be positive")
	ErrExperienceOverflow = errors.New("experience cap reached")
)

// SkillDefinition is immutable game data. IDs are stable API and persistence keys.
type SkillDefinition struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var skillDefinitions = []SkillDefinition{
	{ID: "combat", Name: "Combat", Description: "Defeat hostile creatures and master battlefield techniques."},
	{ID: "cooking", Name: "Cooking", Description: "Prepare food and provisions from gathered ingredients."},
	{ID: "crafting", Name: "Crafting", Description: "Shape materials into useful tools, equipment and goods."},
	{ID: "farming", Name: "Farming", Description: "Cultivate crops and manage productive land."},
	{ID: "foraging", Name: "Foraging", Description: "Find useful plants, fungi and natural materials."},
	{ID: "mining", Name: "Mining", Description: "Extract ores, stone and mineral resources."},
	{ID: "ranged", Name: "Ranged", Description: "Use bows, firearms and other ranged weapons."},
	{ID: "scavenging", Name: "Scavenging", Description: "Recover components and useful materials from ruins."},
	{ID: "smithing", Name: "Smithing", Description: "Refine ores and work metals into equipment."},
	{ID: "woodcutting", Name: "Woodcutting", Description: "Harvest timber and other renewable plant resources."},
}

func SkillDefinitions() []SkillDefinition {
	result := make([]SkillDefinition, len(skillDefinitions))
	copy(result, skillDefinitions)
	return result
}

func IsKnownSkill(id string) bool {
	id = strings.TrimSpace(strings.ToLower(id))
	for _, definition := range skillDefinitions {
		if definition.ID == id { return true }
	}
	return false
}

// ExperienceForLevel returns the cumulative XP threshold for a level (1-99).
func ExperienceForLevel(level int) int {
	if level <= 1 { return 0 }
	if level > MaxSkillLevel { level = MaxSkillLevel }
	total := 0
	for current := 1; current < level; current++ {
		total += int(math.Floor(float64(current) + 300*math.Pow(2, float64(current)/7)))
	}
	return total / 4
}

func LevelForExperience(experience int) int {
	if experience < 0 { return 1 }
	if experience > MaxSkillExperience { experience = MaxSkillExperience }
	level := 1
	for level < MaxSkillLevel && experience >= ExperienceForLevel(level+1) { level++ }
	return level
}
