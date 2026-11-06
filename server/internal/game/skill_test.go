package game

import (
	"errors"
	"testing"
	"time"
)

func TestExperienceCurveBoundaries(t *testing.T) {
	tests := []struct { level, xp int }{{1, 0}, {2, 83}, {10, 1_154}, {50, 101_333}, {99, 13_034_431}}
	for _, test := range tests {
		if got := ExperienceForLevel(test.level); got != test.xp { t.Errorf("ExperienceForLevel(%d) = %d, want %d", test.level, got, test.xp) }
		if got := LevelForExperience(test.xp); got != test.level { t.Errorf("LevelForExperience(%d) = %d, want %d", test.xp, got, test.level) }
	}
}

func TestLevelForExperienceBetweenThresholds(t *testing.T) {
	if got := LevelForExperience(82); got != 1 { t.Fatalf("level at 82 XP = %d, want 1", got) }
	if got := LevelForExperience(84); got != 2 { t.Fatalf("level at 84 XP = %d, want 2", got) }
	if got := LevelForExperience(MaxSkillExperience); got != MaxSkillLevel { t.Fatalf("level at XP cap = %d, want %d", got, MaxSkillLevel) }
}

func TestSkillDefinitionsAreDefensiveAndStable(t *testing.T) {
	definitions := SkillDefinitions()
	if len(definitions) < 8 { t.Fatalf("got %d skills, want at least 8", len(definitions)) }
	definitions[0].ID = "corrupted"
	if IsKnownSkill("corrupted") || !IsKnownSkill("woodcutting") { t.Fatal("skill definitions were mutable through returned slice") }
}

func TestAwardExperience(t *testing.T) {
	player, err := NewPlayer("p-1", "Wayfarer", time.Now())
	if err != nil { t.Fatal(err) }
	leveled, err := player.AwardExperience("woodcutting", 100)
	if err != nil { t.Fatal(err) }
	if !leveled || player.Skills["woodcutting"].Level != 2 || player.Skills["woodcutting"].Experience != 100 {
		t.Fatalf("unexpected skill state: %+v", player.Skills["woodcutting"])
	}
	if _, err := player.AwardExperience("not-a-skill", 10); !errors.Is(err, ErrUnknownSkill) { t.Fatalf("unknown skill error = %v", err) }
	if _, err := player.AwardExperience("woodcutting", 0); !errors.Is(err, ErrInvalidExperience) { t.Fatalf("zero award error = %v", err) }
	if _, err := player.AwardExperience("woodcutting", MaxSkillExperience); !errors.Is(err, ErrExperienceOverflow) { t.Fatalf("overflow error = %v", err) }
}
