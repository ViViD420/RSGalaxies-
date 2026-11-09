package game

import (
	"errors"
	"testing"
	"time"
)

func stonebackDefinition(t *testing.T) BossEncounterDefinition {
	t.Helper()
	definition, ok := FindBossEncounter("stoneback-colossus")
	if !ok {
		t.Fatal("stoneback-colossus boss definition not found")
	}
	if err := ValidateBossEncounterDefinition(definition); err != nil {
		t.Fatalf("validate boss definition: %v", err)
	}
	return definition
}

func bossInput(encounterID string, health int, now time.Time, targetX float64) BossEncounterInput {
	return BossEncounterInput{
		EncounterID: encounterID,
		Health: health,
		MaxHealth: 100,
		BossPosition: WorldPosition{ZoneID: DefaultWorldZoneID, X: 0, Z: 0},
		TargetPosition: WorldPosition{ZoneID: DefaultWorldZoneID, X: targetX, Z: 0},
		Now: now,
	}
}

func TestBossEncounterCatalogIsValidAndDefensivelyCopied(t *testing.T) {
	definitions := BossEncounterDefinitions()
	if len(definitions) != 1 {
		t.Fatalf("boss definition count=%d, want 1", len(definitions))
	}
	if err := ValidateBossEncounterDefinition(definitions[0]); err != nil {
		t.Fatalf("catalog definition is invalid: %v", err)
	}
	definitions[0].Phases[0].AbilityIDs[0] = "corrupted"
	definitions[0].Abilities[0].Damage = 999

	again, ok := FindBossEncounter("stoneback-colossus")
	if !ok {
		t.Fatal("boss definition disappeared")
	}
	if again.Phases[0].AbilityIDs[0] != "stone-fist" || again.Abilities[0].Damage != 8 {
		t.Fatalf("catalog returned mutable nested data: %+v", again)
	}
}

func TestValidateBossEncounterDefinitionRejectsInvalidPhasesAndAbilities(t *testing.T) {
	definition := stonebackDefinition(t)
	definition.Phases[1].AbilityIDs = []string{"missing-ability"}
	if err := ValidateBossEncounterDefinition(definition); !errors.Is(err, ErrInvalidBossEncounterDefinition) {
		t.Fatalf("unknown phase ability error=%v, want ErrInvalidBossEncounterDefinition", err)
	}

	definition = stonebackDefinition(t)
	definition.Phases[1].EnterAtOrBelowPercent = 100
	if err := ValidateBossEncounterDefinition(definition); !errors.Is(err, ErrInvalidBossEncounterDefinition) {
		t.Fatalf("duplicate phase threshold error=%v, want ErrInvalidBossEncounterDefinition", err)
	}

	definition = stonebackDefinition(t)
	definition.Abilities[0].MaxRange = -1
	if err := ValidateBossEncounterDefinition(definition); !errors.Is(err, ErrInvalidBossEncounterDefinition) {
		t.Fatalf("invalid ability range error=%v, want ErrInvalidBossEncounterDefinition", err)
	}
}

func TestAdvanceBossEncounterTelegraphsResolvesAndHonorsCooldown(t *testing.T) {
	definition := stonebackDefinition(t)
	start := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	first, err := AdvanceBossEncounter(definition, BossEncounterState{}, bossInput("cycle-a", 100, start, 2))
	if err != nil {
		t.Fatal(err)
	}
	if first.State.PhaseID != "guarded" || !first.AbilityStarted || first.AbilityID != "stone-fist" {
		t.Fatalf("initial boss tick = %+v", first)
	}
	if !first.State.TelegraphUntil.Equal(start.Add(1500 * time.Millisecond)) {
		t.Fatalf("telegraph deadline=%s", first.State.TelegraphUntil)
	}

	waiting, err := AdvanceBossEncounter(definition, first.State, bossInput("cycle-a", 100, start.Add(time.Second), 2))
	if err != nil {
		t.Fatal(err)
	}
	if waiting.AbilityResolved {
		t.Fatalf("ability resolved before telegraph elapsed: %+v", waiting)
	}

	resolved, err := AdvanceBossEncounter(definition, waiting.State, bossInput("cycle-a", 100, start.Add(1500*time.Millisecond), 2))
	if err != nil {
		t.Fatal(err)
	}
	if !resolved.AbilityResolved || resolved.AbilityID != "stone-fist" || resolved.Damage != 8 {
		t.Fatalf("resolved ability = %+v", resolved)
	}

	cooling, err := AdvanceBossEncounter(definition, resolved.State, bossInput("cycle-a", 100, start.Add(2*time.Second), 2))
	if err != nil {
		t.Fatal(err)
	}
	if cooling.AbilityStarted {
		t.Fatalf("ability ignored its cooldown: %+v", cooling)
	}
}

func TestAdvanceBossEncounterInterruptsTelegraphAndRetainsCooldown(t *testing.T) {
	definition := stonebackDefinition(t)
	start := time.Date(2026, 11, 9, 13, 0, 0, 0, time.UTC)
	first, err := AdvanceBossEncounter(definition, BossEncounterState{}, bossInput("cycle-a", 100, start, 2))
	if err != nil {
		t.Fatal(err)
	}
	interrupted, err := AdvanceBossEncounter(definition, first.State, bossInput("cycle-a", 100, start.Add(time.Second), 4))
	if err != nil {
		t.Fatal(err)
	}
	if !interrupted.AbilityInterrupted || interrupted.AbilityID != "stone-fist" ||
		interrupted.State.ActiveAbilityID != "" || !interrupted.State.TelegraphUntil.IsZero() {
		t.Fatalf("out-of-range telegraph was not safely interrupted: %+v", interrupted)
	}
	cooling, err := AdvanceBossEncounter(definition, interrupted.State, bossInput("cycle-a", 100, start.Add(2*time.Second), 2))
	if err != nil {
		t.Fatal(err)
	}
	if cooling.AbilityStarted {
		t.Fatalf("interrupted ability ignored its cooldown: %+v", cooling)
	}
	ready, err := AdvanceBossEncounter(definition, cooling.State, bossInput("cycle-a", 100, start.Add(5*time.Second), 2))
	if err != nil {
		t.Fatal(err)
	}
	if !ready.AbilityStarted || ready.AbilityID != "stone-fist" {
		t.Fatalf("ability did not become available after cooldown: %+v", ready)
	}
}

func TestAdvanceBossEncounterTransitionsPhasesAndUnlocksNewAbility(t *testing.T) {
	definition := stonebackDefinition(t)
	start := time.Date(2026, 11, 9, 14, 0, 0, 0, time.UTC)
	first, err := AdvanceBossEncounter(definition, BossEncounterState{}, bossInput("cycle-a", 100, start, 2))
	if err != nil {
		t.Fatal(err)
	}
	transition, err := AdvanceBossEncounter(definition, first.State, bossInput("cycle-a", 50, start.Add(time.Second), 2))
	if err != nil {
		t.Fatal(err)
	}
	if !transition.PhaseChanged || transition.State.PhaseID != "rampage" || transition.State.ActiveAbilityID != "" {
		t.Fatalf("phase transition did not reset the active telegraph: %+v", transition)
	}
	next, err := AdvanceBossEncounter(definition, transition.State, bossInput("cycle-a", 50, start.Add(2*time.Second), 2))
	if err != nil {
		t.Fatal(err)
	}
	if !next.AbilityStarted || next.AbilityID != "ground-shatter" {
		t.Fatalf("new phase did not select its ready ability: %+v", next)
	}
	resolved, err := AdvanceBossEncounter(definition, next.State, bossInput("cycle-a", 50, start.Add(4*time.Second), 2))
	if err != nil {
		t.Fatal(err)
	}
	if !resolved.AbilityResolved || resolved.Damage != 12 {
		t.Fatalf("ground-shatter result = %+v", resolved)
	}
}

func TestAdvanceBossEncounterDoesNotRewindAndResetsOnNewCycle(t *testing.T) {
	definition := stonebackDefinition(t)
	start := time.Date(2026, 11, 9, 15, 0, 0, 0, time.UTC)
	first, err := AdvanceBossEncounter(definition, BossEncounterState{}, bossInput("cycle-a", 100, start, 2))
	if err != nil {
		t.Fatal(err)
	}
	stale, err := AdvanceBossEncounter(definition, first.State, bossInput("cycle-a", 100, start.Add(-time.Second), 2))
	if err != nil {
		t.Fatal(err)
	}
	if stale.State.LastUpdatedAt.Before(first.State.LastUpdatedAt) {
		t.Fatalf("same-cycle state clock moved backwards: before=%s after=%s", first.State.LastUpdatedAt, stale.State.LastUpdatedAt)
	}

	nextCycle, err := AdvanceBossEncounter(definition, first.State, bossInput("cycle-b", 100, start.Add(-time.Minute), 2))
	if err != nil {
		t.Fatal(err)
	}
	if nextCycle.State.EncounterID != "cycle-b" || !nextCycle.State.PhaseStartedAt.Equal(start.Add(-time.Minute)) ||
		nextCycle.State.ActiveAbilityID != "stone-fist" || len(nextCycle.State.AbilityCooldowns) != 1 {
		t.Fatalf("new cycle inherited old timing/cooldowns: %+v", nextCycle.State)
	}
}

func TestAdvanceBossEncounterDefeatIsTerminalForEncounterCycle(t *testing.T) {
	definition := stonebackDefinition(t)
	start := time.Date(2026, 11, 9, 16, 0, 0, 0, time.UTC)
	first, err := AdvanceBossEncounter(definition, BossEncounterState{}, bossInput("cycle-a", 100, start, 2))
	if err != nil {
		t.Fatal(err)
	}
	defeated, err := AdvanceBossEncounter(definition, first.State, bossInput("cycle-a", 0, start.Add(time.Second), 2))
	if err != nil {
		t.Fatal(err)
	}
	if !defeated.State.Defeated || defeated.State.ActiveAbilityID != "" {
		t.Fatalf("boss defeat did not clear active telegraph: %+v", defeated)
	}
	replayed, err := AdvanceBossEncounter(definition, defeated.State, bossInput("cycle-a", 100, start.Add(2*time.Second), 2))
	if err != nil {
		t.Fatal(err)
	}
	if !replayed.State.Defeated || replayed.AbilityStarted {
		t.Fatalf("defeated boss restarted in same cycle: %+v", replayed)
	}
}

func TestAdvanceBossEncounterRejectsCorruptTelegraphState(t *testing.T) {
	definition := stonebackDefinition(t)
	now := time.Date(2026, 11, 9, 17, 0, 0, 0, time.UTC)
	state := BossEncounterState{
		EncounterID: "cycle-a",
		DefinitionVersion: definition.Version,
		PhaseID: "guarded",
		PhaseStartedAt: now,
		ActiveAbilityID: "stone-fist",
	}
	_, err := AdvanceBossEncounter(definition, state, bossInput("cycle-a", 100, now, 2))
	if !errors.Is(err, ErrInvalidBossEncounterDefinition) {
		t.Fatalf("corrupt telegraph error=%v, want ErrInvalidBossEncounterDefinition", err)
	}
}

func TestAdvanceBossEncounterNormalizesNegativeAbilityCursor(t *testing.T) {
	definition := stonebackDefinition(t)
	now := time.Date(2026, 11, 9, 18, 0, 0, 0, time.UTC)
	state := BossEncounterState{
		EncounterID: "cycle-a",
		DefinitionVersion: definition.Version,
		PhaseID: "rampage",
		PhaseStartedAt: now,
		NextAbilityIndex: -1,
	}
	result, err := AdvanceBossEncounter(definition, state, bossInput("cycle-a", 50, now, 2))
	if err != nil {
		t.Fatal(err)
	}
	if !result.AbilityStarted || result.AbilityID != "stone-fist" {
		t.Fatalf("negative ability cursor was not normalized: %+v", result)
	}
}

func TestAdvanceBossEncounterCopiesCooldownMapBeforeMutation(t *testing.T) {
	definition := stonebackDefinition(t)
	now := time.Date(2026, 11, 9, 19, 0, 0, 0, time.UTC)
	state := BossEncounterState{
		EncounterID: "cycle-a",
		DefinitionVersion: definition.Version,
		PhaseID: "guarded",
		PhaseStartedAt: now,
		AbilityCooldowns: map[string]time.Time{
			"ground-shatter": now.Add(time.Minute),
		},
	}
	result, err := AdvanceBossEncounter(definition, state, bossInput("cycle-a", 100, now, 2))
	if err != nil {
		t.Fatal(err)
	}
	if _, mutated := state.AbilityCooldowns["stone-fist"]; mutated {
		t.Fatal("AdvanceBossEncounter mutated the caller's cooldown map")
	}
	if result.State.AbilityCooldowns["stone-fist"].IsZero() {
		t.Fatalf("returned state did not record the new cooldown: %+v", result.State.AbilityCooldowns)
	}
}

func TestFindBossEncounterForEnemyUsesSpawnIdentity(t *testing.T) {
	definition := stonebackDefinition(t)
	found, ok := FindBossEncounterForEnemy(definition.EnemyID)
	if !ok || found.ID != definition.ID || found.EnemyID != definition.EnemyID {
		t.Fatalf("boss lookup by enemy identity = %+v, found=%v", found, ok)
	}
	found.Phases[0].AbilityIDs[0] = "mutated"
	again, ok := FindBossEncounterForEnemy(definition.EnemyID)
	if !ok || again.Phases[0].AbilityIDs[0] != "stone-fist" {
		t.Fatalf("enemy identity lookup exposed mutable catalog state: %+v", again)
	}
	if _, ok := FindBossEncounterForEnemy("not-a-boss"); ok {
		t.Fatal("unknown enemy unexpectedly resolved to a boss definition")
	}
}

