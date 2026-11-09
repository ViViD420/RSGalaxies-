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
// CombatMastery is deliberately private: it is progression metadata for future
// server-side unlocks and must never be serialized into public catalogs.
type SkillDefinition struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"-"`
	ParentID    string `json:"-"`
	parent      bool
	coSkill     bool
	combatMastery bool
}

var skillDefinitions = []SkillDefinition{
	{ID: "blades", Name: "Blades", Description: "Master edged melee weapons and their techniques.", Category: "combat", ParentID: "melee", combatMastery: true},
	{ID: "cooking", Name: "Cooking", Description: "Prepare food and provisions from gathered ingredients.", Category: "production"},
	{ID: "crafting", Name: "Crafting", Description: "Shape materials into useful tools, equipment and goods.", Category: "production"},
	{ID: "defense", Name: "Defense", Description: "Train defensive technique and qualify to use protective equipment.", Category: "combat", coSkill: true},
	{ID: "farming", Name: "Farming", Description: "Cultivate crops and manage productive land.", Category: "gathering"},
	{ID: "foraging", Name: "Foraging", Description: "Find useful plants, fungi and natural materials.", Category: "gathering"},
	{ID: "magic", Name: "Magic", Description: "Study and use offensive, defensive and utility magic.", Category: "combat", ParentID: "ranged", combatMastery: true},
	{ID: "melee", Name: "Melee", Description: "Develop broad close-quarters combat fundamentals.", Category: "combat", parent: true},
	{ID: "mining", Name: "Mining", Description: "Extract ores, stone and mineral resources.", Category: "gathering"},
	{ID: "pistols", Name: "Pistols", Description: "Improve proficiency with one-handed firearms.", Category: "combat", ParentID: "ranged", combatMastery: true},
	{ID: "polearms", Name: "Polearms", Description: "Master staffs, spears and other long-reach melee weapons.", Category: "combat", ParentID: "melee", combatMastery: true},
	{ID: "ranged", Name: "Ranged", Description: "Develop broad ranged-combat fundamentals.", Category: "combat", parent: true},
	{ID: "rifles", Name: "Rifles", Description: "Improve proficiency with long guns and precision firearms.", Category: "combat", ParentID: "ranged", combatMastery: true},
	{ID: "scavenging", Name: "Scavenging", Description: "Recover components and useful materials from ruins.", Category: "gathering"},
	{ID: "smithing", Name: "Smithing", Description: "Refine ores and work metals into equipment.", Category: "production"},
	{ID: "strength", Name: "Strength", Description: "Develop physical power that contributes to maximum melee damage.", Category: "combat", coSkill: true},
	{ID: "unarmed", Name: "Unarmed", Description: "Master hand-to-hand combat without a weapon.", Category: "combat", ParentID: "melee", combatMastery: true},
	{ID: "woodcutting", Name: "Woodcutting", Description: "Harvest timber and other renewable plant resources.", Category: "gathering"},
}

func SkillDefinitions() []SkillDefinition {
	result := make([]SkillDefinition, len(skillDefinitions))
	copy(result, skillDefinitions)
	return result
}

// EnsureSkillProgress adds level-1 entries for newly introduced skills while preserving
// every existing progress record. It keeps older persisted characters compatible.
func EnsureSkillProgress(skills map[string]SkillProgress) map[string]SkillProgress {
	if skills == nil {
		skills = make(map[string]SkillProgress, len(skillDefinitions))
	}
	for _, definition := range skillDefinitions {
		if _, exists := skills[definition.ID]; !exists {
			skills[definition.ID] = SkillProgress{Level: 1}
		}
	}
	return skills
}

func skillDefinition(id string) (SkillDefinition, bool) {
	id = strings.TrimSpace(strings.ToLower(id))
	for _, definition := range skillDefinitions { if definition.ID == id { return definition, true } }
	return SkillDefinition{}, false
}

func skillParentID(id string) string {
	definition, ok := skillDefinition(id)
	if !ok { return "" }
	return definition.ParentID
}

func IsKnownSkill(id string) bool {
	id = strings.TrimSpace(strings.ToLower(id))
	for _, definition := range skillDefinitions {
		if definition.ID == id {
			return true
		}
	}
	return false
}

// isCombatMasterySkill is intentionally server-internal. Do not expose mastery
// eligibility in public skill catalogs or character skill payloads.
func isCombatMasterySkill(id string) bool {
	id = strings.TrimSpace(strings.ToLower(id))
	for _, definition := range skillDefinitions {
		if definition.ID == id {
			return definition.combatMastery
		}
	}
	return false
}

// ExperienceForLevel returns the cumulative XP threshold for a level (1-99).
func ExperienceForLevel(level int) int {
	if level <= 1 {
		return 0
	}
	if level > MaxSkillLevel {
		level = MaxSkillLevel
	}
	total := 0
	for current := 1; current < level; current++ {
		total += int(math.Floor(float64(current) + 300*math.Pow(2, float64(current)/7)))
	}
	return total / 4
}

func LevelForExperience(experience int) int {
	if experience < 0 {
		return 1
	}
	if experience > MaxSkillExperience {
		experience = MaxSkillExperience
	}
	level := 1
	for level < MaxSkillLevel && experience >= ExperienceForLevel(level+1) {
		level++
	}
	return level
}
