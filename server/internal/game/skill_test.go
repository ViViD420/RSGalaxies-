package game

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestExperienceCurveBoundaries(t *testing.T) {
	tests := []struct{ level, xp int }{{1, 0}, {2, 83}, {10, 1_154}, {50, 101_333}, {99, 13_034_431}}
	for _, test := range tests {
		if got := ExperienceForLevel(test.level); got != test.xp {
			t.Errorf("ExperienceForLevel(%d) = %d, want %d", test.level, got, test.xp)
		}
		if got := LevelForExperience(test.xp); got != test.level {
			t.Errorf("LevelForExperience(%d) = %d, want %d", test.xp, got, test.level)
		}
	}
}

func TestLevelForExperienceBetweenThresholds(t *testing.T) {
	if got := LevelForExperience(82); got != 1 {
		t.Fatalf("level at 82 XP = %d, want 1", got)
	}
	if got := LevelForExperience(84); got != 2 {
		t.Fatalf("level at 84 XP = %d, want 2", got)
	}
	if got := LevelForExperience(MaxSkillExperience); got != MaxSkillLevel {
		t.Fatalf("level at XP cap = %d, want %d", got, MaxSkillLevel)
	}
}

func TestSkillDefinitionsAreDefensiveAndStable(t *testing.T) {
	definitions := SkillDefinitions()
	if len(definitions) != 20 {
		t.Fatalf("got %d skills, want exactly 20", len(definitions))
	}
	definitions[0].ID = "corrupted"
	if IsKnownSkill("corrupted") || !IsKnownSkill("woodcutting") {
		t.Fatal("skill definitions were mutable through returned slice")
	}
}

func TestSkillCategoriesAndCombatMasteryEligibility(t *testing.T) {
	wantMastery := []string{
		"blades", "combat", "defense", "magic", "melee", "pistols", "polearms",
		"ranged", "rifles", "strength", "tactics", "unarmed",
	}
	for _, id := range wantMastery {
		if !isCombatMasterySkill(id) {
			t.Errorf("%q should be a combat mastery skill", id)
		}
	}
	for _, id := range []string{"cooking", "crafting", "farming", "foraging", "mining", "scavenging", "smithing", "woodcutting", "unknown"} {
		if isCombatMasterySkill(id) {
			t.Errorf("%q must not count as combat mastery", id)
		}
	}

	seen := make(map[string]bool, len(skillDefinitions))
	for _, definition := range SkillDefinitions() {
		if definition.ID == "" || seen[definition.ID] {
			t.Errorf("empty or duplicate skill ID %q", definition.ID)
		}
		seen[definition.ID] = true
		switch definition.Category {
		case "combat", "gathering", "production":
		default:
			t.Errorf("skill %q has unsupported category %q", definition.ID, definition.Category)
		}
		if definition.combatMastery && definition.Category != "combat" {
			t.Errorf("non-combat skill %q is marked for combat mastery", definition.ID)
		}
	}
	if len(wantMastery) != 12 {
		t.Fatalf("test fixture expected 12 combat mastery skills, got %d", len(wantMastery))
	}
}

func TestCombatMasteryEligibilityIsNotSerialized(t *testing.T) {
	data, err := json.Marshal(SkillDefinitions())
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" {
		t.Fatal("serialized skill catalog is empty")
	}
	for _, forbidden := range []string{"combatMastery", "category", "jedi", "unlock"} {
		if contains := string(data); contains != "" && jsonFieldPresent(contains, forbidden) {
			t.Errorf("public skill catalog unexpectedly exposes %q", forbidden)
		}
	}
}

func jsonFieldPresent(data, field string) bool {
	var values []map[string]any
	if err := json.Unmarshal([]byte(data), &values); err != nil {
		return false
	}
	for _, value := range values {
		if _, ok := value[field]; ok {
			return true
		}
	}
	return false
}


func TestEnsureSkillProgressBackfillsWithoutResettingExistingProgress(t *testing.T) {
	existing := map[string]SkillProgress{
		"combat": {Experience: 1234, Level: 5},
		"woodcutting": {Experience: 83, Level: 2},
	}
	got := EnsureSkillProgress(existing)
	if got["combat"] != (SkillProgress{Experience: 1234, Level: 5}) {
		t.Fatalf("existing combat progress was changed: %+v", got["combat"])
	}
	if got["woodcutting"] != (SkillProgress{Experience: 83, Level: 2}) {
		t.Fatalf("existing woodcutting progress was changed: %+v", got["woodcutting"])
	}
	if got["magic"] != (SkillProgress{Level: 1}) || got["pistols"] != (SkillProgress{Level: 1}) || got["rifles"] != (SkillProgress{Level: 1}) {
		t.Fatalf("new skills were not initialized at level 1: magic=%+v pistols=%+v rifles=%+v", got["magic"], got["pistols"], got["rifles"])
	}
	if len(got) != len(SkillDefinitions()) {
		t.Fatalf("got %d initialized skills, want %d", len(got), len(SkillDefinitions()))
	}
}

func TestAwardExperience(t *testing.T) {
	player, err := NewPlayer("p-1", "Wayfarer", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	leveled, err := player.AwardExperience("woodcutting", 100)
	if err != nil {
		t.Fatal(err)
	}
	if !leveled || player.Skills["woodcutting"].Level != 2 || player.Skills["woodcutting"].Experience != 100 {
		t.Fatalf("unexpected skill state: %+v", player.Skills["woodcutting"])
	}
	if _, err := player.AwardExperience("not-a-skill", 10); !errors.Is(err, ErrUnknownSkill) {
		t.Fatalf("unknown skill error = %v", err)
	}
	if _, err := player.AwardExperience("woodcutting", 0); !errors.Is(err, ErrInvalidExperience) {
		t.Fatalf("zero award error = %v", err)
	}
	if _, err := player.AwardExperience("woodcutting", MaxSkillExperience); !errors.Is(err, ErrExperienceOverflow) {
		t.Fatalf("overflow error = %v", err)
	}
}
