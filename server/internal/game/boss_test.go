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
	if len(definitions) != 4 {
		t.Fatalf("boss definition count=%d, want 4", len(definitions))
	}
	for _, def := range definitions {
		if err := ValidateBossEncounterDefinition(def); err != nil {
			t.Fatalf("catalog definition %s is invalid: %v", def.EnemyID, err)
		}
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

func TestStonebackColossusIsRegisteredWithBossAndRewardDefinitions(t *testing.T) {
	enemy, ok := FindEnemy("stoneback-colossus")
	if !ok {
		t.Fatal("Stoneback Colossus is missing from the enemy catalog")
	}
	if enemy.MaxHealth != 100 || enemy.Defense != 2 || enemy.Experience != 500 ||
		enemy.Attack != 0 || enemy.LootTableID != "stoneback-colossus-harvest" ||
		enemy.Position.ZoneID != DefaultWorldZoneID || enemy.RespawnSeconds != 300 {
		t.Fatalf("unexpected Stoneback enemy definition: %+v", enemy)
	}
	boss, ok := FindBossEncounterForEnemy(enemy.ID)
	if !ok || boss.ID != "stoneback-colossus" {
		t.Fatalf("Stoneback boss definition lookup = %+v, found=%v", boss, ok)
	}
	if _, ok := FindLootTable(enemy.LootTableID); !ok {
		t.Fatalf("Stoneback loot table %q is missing", enemy.LootTableID)
	}
}

func TestKraytJadSovereignThreePhaseBossEncounterAndLootTable(t *testing.T) {
	enemy, ok := FindEnemy("krayt-jad-sovereign")
	if !ok {
		t.Fatal("Ancient Krayt-Jad Sovereign is missing from the enemy catalog")
	}
	if enemy.MaxHealth != 320 || enemy.Defense != 8 || enemy.Experience != 1200 ||
		enemy.Attack != 0 || enemy.LootTableID != "krayt-jad-sovereign-harvest" ||
		enemy.Position.ZoneID != DefaultWorldZoneID || enemy.RespawnSeconds != 300 {
		t.Fatalf("unexpected Krayt-Jad Sovereign enemy definition: %+v", enemy)
	}
	definition, ok := FindBossEncounterForEnemy(enemy.ID)
	if !ok || definition.ID != "krayt-jad-sovereign" || len(definition.Phases) != 3 {
		t.Fatalf("Krayt-Jad Sovereign boss definition = %+v, found=%v", definition, ok)
	}

	start := time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC)
	input := func(health int, now time.Time) BossEncounterInput {
		return BossEncounterInput{
			EncounterID:    "krayt-jad-cycle-1",
			Health:         health,
			MaxHealth:      enemy.MaxHealth,
			BossPosition:   enemy.Position,
			TargetPosition: WorldPosition{ZoneID: DefaultWorldZoneID, X: enemy.Position.X + 2.0, Z: enemy.Position.Z},
			Now:            now,
		}
	}

	// Phase 1 (100% HP -> armored-carapace, starts rancor-cleave)
	p1, err := AdvanceBossEncounter(definition, BossEncounterState{}, input(320, start))
	if err != nil {
		t.Fatal(err)
	}
	if p1.State.PhaseID != "armored-carapace" || !p1.AbilityStarted || p1.AbilityID != "rancor-cleave" {
		t.Fatalf("phase 1 start = %+v", p1)
	}
	p1Resolved, err := AdvanceBossEncounter(definition, p1.State, input(320, start.Add(1500*time.Millisecond)))
	if err != nil {
		t.Fatal(err)
	}
	if !p1Resolved.AbilityResolved || p1Resolved.AbilityID != "rancor-cleave" || p1Resolved.Damage != 16 {
		t.Fatalf("phase 1 rancor-cleave resolution = %+v", p1Resolved)
	}

	// Phase 2 (<= 65% HP -> 200/320 = 62.5% -> molten-fury)
	p2Transition, err := AdvanceBossEncounter(definition, p1Resolved.State, input(200, start.Add(2*time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	if !p2Transition.PhaseChanged || p2Transition.State.PhaseID != "molten-fury" {
		t.Fatalf("phase 2 transition = %+v", p2Transition)
	}

	// Phase 3 (<= 30% HP -> 90/320 = 28.125% -> sovereign-cataclysm)
	p3Transition, err := AdvanceBossEncounter(definition, p2Transition.State, input(90, start.Add(3*time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	if !p3Transition.PhaseChanged || p3Transition.State.PhaseID != "sovereign-cataclysm" {
		t.Fatalf("phase 3 transition = %+v", p3Transition)
	}

	// Verify multi-roll boss loot table produces guaranteed boss core + scales + 3 independent bonus rolls
	drops, err := RollLootTableDrops(enemy.LootTableID, []LootRollInput{
		{DropRoll: 0, QuantityRoll: 9999}, // Guaranteed 1: sovereign_jad_fire_core x2
		{DropRoll: 0, QuantityRoll: 9999}, // Guaranteed 2: ancient_krayt_dragon_scale x4
		{DropRoll: 0, QuantityRoll: 9999}, // Bonus 1: krayt_wyrm_pearl x2
		{DropRoll: 0, QuantityRoll: 0},    // Bonus 2: krayt_fang_whip_blade x1
		{DropRoll: 0, QuantityRoll: 0},    // Bonus 3: torva_gorax_warplate x1
	})
	if err != nil {
		t.Fatalf("roll krayt-jad-sovereign-harvest: %v", err)
	}
	if len(drops) != 5 ||
		drops[0] != (ItemQuantity{ItemID: "sovereign_jad_fire_core", Quantity: 2}) ||
		drops[1] != (ItemQuantity{ItemID: "ancient_krayt_dragon_scale", Quantity: 4}) ||
		drops[2] != (ItemQuantity{ItemID: "krayt_wyrm_pearl", Quantity: 2}) ||
		drops[3] != (ItemQuantity{ItemID: "krayt_fang_whip_blade", Quantity: 1}) ||
		drops[4] != (ItemQuantity{ItemID: "torva_gorax_warplate", Quantity: 1}) {
		t.Fatalf("unexpected Krayt-Jad Sovereign boss drops: %+v", drops)
	}
}

func TestTzKalMythosaurEmperorThreePhaseBossEncounterAndLootTable(t *testing.T) {
	enemy, ok := FindEnemy("tzkal-mythosaur-emperor")
	if !ok {
		t.Fatal("tzkal-mythosaur-emperor missing from enemy catalog")
	}
	if enemy.MaxHealth != 620 || enemy.Defense != 14 || enemy.Experience != 2600 ||
		enemy.Attack != 0 || enemy.LootTableID != "tzkal-mythosaur-emperor-harvest" ||
		enemy.Position.ZoneID != DefaultWorldZoneID || enemy.RespawnSeconds != 420 {
		t.Fatalf("unexpected TzKal-Mythosaur Emperor enemy definition: %+v", enemy)
	}
	definition, ok := FindBossEncounterForEnemy(enemy.ID)
	if !ok || definition.ID != "tzkal-mythosaur-emperor" || len(definition.Phases) != 3 {
		t.Fatalf("TzKal-Mythosaur Emperor boss definition = %+v, found=%v", definition, ok)
	}

	start := time.Date(2026, 11, 18, 16, 0, 0, 0, time.UTC)
	input := func(health int, now time.Time) BossEncounterInput {
		return BossEncounterInput{
			EncounterID:    "tzkal-mythosaur-cycle-1",
			Health:         health,
			MaxHealth:      enemy.MaxHealth,
			BossPosition:   enemy.Position,
			TargetPosition: WorldPosition{ZoneID: DefaultWorldZoneID, X: enemy.Position.X + 2.0, Z: enemy.Position.Z},
			Now:            now,
		}
	}

	// Phase 1 (100% HP -> beskar-aegis, starts mythosaur-tusk-rend)
	p1, err := AdvanceBossEncounter(definition, BossEncounterState{}, input(620, start))
	if err != nil {
		t.Fatal(err)
	}
	if p1.State.PhaseID != "beskar-aegis" || !p1.AbilityStarted || p1.AbilityID != "mythosaur-tusk-rend" {
		t.Fatalf("phase 1 start = %+v", p1)
	}
	p1Resolved, err := AdvanceBossEncounter(definition, p1.State, input(620, start.Add(1500*time.Millisecond)))
	if err != nil {
		t.Fatal(err)
	}
	if !p1Resolved.AbilityResolved || p1Resolved.AbilityID != "mythosaur-tusk-rend" || p1Resolved.Damage != 24 {
		t.Fatalf("phase 1 mythosaur-tusk-rend resolution = %+v", p1Resolved)
	}

	// Phase 2 (<= 65% HP -> 390/620 = 62.9% -> infernal-overdrive)
	p2Transition, err := AdvanceBossEncounter(definition, p1Resolved.State, input(390, start.Add(2*time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	if !p2Transition.PhaseChanged || p2Transition.State.PhaseID != "infernal-overdrive" {
		t.Fatalf("phase 2 transition = %+v", p2Transition)
	}

	// Phase 3 (<= 30% HP -> 180/620 = 29.0% -> omnium-apocalypse)
	p3Transition, err := AdvanceBossEncounter(definition, p2Transition.State, input(180, start.Add(3*time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	if !p3Transition.PhaseChanged || p3Transition.State.PhaseID != "omnium-apocalypse" {
		t.Fatalf("phase 3 transition = %+v", p3Transition)
	}

	// Verify 5-roll boss loot table produces guaranteed infernal heart + leviathan crests + 3 independent bonus rolls
	drops, err := RollLootTableDrops(enemy.LootTableID, []LootRollInput{
		{DropRoll: 0, QuantityRoll: 9999}, // Guaranteed 1: tzkal_infernal_heart x2
		{DropRoll: 0, QuantityRoll: 9999}, // Guaranteed 2: mythosaur_leviathan_crest x4
		{DropRoll: 0, QuantityRoll: 9999}, // Bonus 1: terentatek_blood_spine x2
		{DropRoll: 0, QuantityRoll: 0},    // Bonus 2: zaros_starforge_godsword x1
		{DropRoll: 0, QuantityRoll: 0},    // Bonus 3: justiciar_mandalore_hauberk x1
	})
	if err != nil {
		t.Fatalf("roll tzkal-mythosaur-emperor-harvest: %v", err)
	}
	if len(drops) != 5 ||
		drops[0] != (ItemQuantity{ItemID: "tzkal_infernal_heart", Quantity: 2}) ||
		drops[1] != (ItemQuantity{ItemID: "mythosaur_leviathan_crest", Quantity: 4}) ||
		drops[2] != (ItemQuantity{ItemID: "terentatek_blood_spine", Quantity: 2}) ||
		drops[3] != (ItemQuantity{ItemID: "zaros_starforge_godsword", Quantity: 1}) ||
		drops[4] != (ItemQuantity{ItemID: "justiciar_mandalore_hauberk", Quantity: 1}) {
		t.Fatalf("unexpected TzKal-Mythosaur Emperor boss drops: %+v", drops)
	}
}

func TestZamorakZilloApocalypseThreePhaseBossEncounterAndLootTable(t *testing.T) {
	enemy, ok := FindEnemy("zamorak-zillo-apocalypse")
	if !ok {
		t.Fatal("zamorak-zillo-apocalypse missing from enemy catalog")
	}
	if enemy.MaxHealth != 1100 || enemy.Defense != 22 || enemy.Experience != 4800 ||
		enemy.Attack != 0 || enemy.LootTableID != "zamorak-zillo-apocalypse-harvest" ||
		enemy.Position.ZoneID != DefaultWorldZoneID || enemy.RespawnSeconds != 540 {
		t.Fatalf("unexpected Zamorakian Zillo-Nihilus Worldbreaker enemy definition: %+v", enemy)
	}
	definition, ok := FindBossEncounterForEnemy(enemy.ID)
	if !ok || definition.ID != "zamorak-zillo-apocalypse" || len(definition.Phases) != 3 {
		t.Fatalf("Zamorakian Zillo-Nihilus Worldbreaker boss definition = %+v, found=%v", definition, ok)
	}

	start := time.Date(2026, 11, 18, 17, 0, 0, 0, time.UTC)
	input := func(health int, now time.Time) BossEncounterInput {
		return BossEncounterInput{
			EncounterID:    "zamorak-zillo-cycle-1",
			Health:         health,
			MaxHealth:      enemy.MaxHealth,
			BossPosition:   enemy.Position,
			TargetPosition: WorldPosition{ZoneID: DefaultWorldZoneID, X: enemy.Position.X + 2.0, Z: enemy.Position.Z},
			Now:            now,
		}
	}

	// Phase 1 (100% HP -> tectonic-dominion, starts zillo-tectonic-tail-crush)
	p1, err := AdvanceBossEncounter(definition, BossEncounterState{}, input(1100, start))
	if err != nil {
		t.Fatal(err)
	}
	if p1.State.PhaseID != "tectonic-dominion" || !p1.AbilityStarted || p1.AbilityID != "zillo-tectonic-tail-crush" {
		t.Fatalf("phase 1 start = %+v", p1)
	}
	p1Resolved, err := AdvanceBossEncounter(definition, p1.State, input(1100, start.Add(1500*time.Millisecond)))
	if err != nil {
		t.Fatal(err)
	}
	if !p1Resolved.AbilityResolved || p1Resolved.AbilityID != "zillo-tectonic-tail-crush" || p1Resolved.Damage != 34 {
		t.Fatalf("phase 1 zillo-tectonic-tail-crush resolution = %+v", p1Resolved)
	}

	// Phase 2 (<= 65% HP -> 700/1100 = 63.6% -> chaos-edict-ascension)
	p2Transition, err := AdvanceBossEncounter(definition, p1Resolved.State, input(700, start.Add(2*time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	if !p2Transition.PhaseChanged || p2Transition.State.PhaseID != "chaos-edict-ascension" {
		t.Fatalf("phase 2 transition = %+v", p2Transition)
	}

	// Phase 3 (<= 30% HP -> 320/1100 = 29.1% -> worldbreaker-cataclysm)
	p3Transition, err := AdvanceBossEncounter(definition, p2Transition.State, input(320, start.Add(3*time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	if !p3Transition.PhaseChanged || p3Transition.State.PhaseID != "worldbreaker-cataclysm" {
		t.Fatalf("phase 3 transition = %+v", p3Transition)
	}

	// Verify 5-roll boss loot table produces guaranteed chaos core + sovereign horns + 3 independent bonus rolls
	drops, err := RollLootTableDrops(enemy.LootTableID, []LootRollInput{
		{DropRoll: 0, QuantityRoll: 9999}, // Guaranteed 1: zamorak_nihilus_chaos_core x2
		{DropRoll: 0, QuantityRoll: 9999}, // Guaranteed 2: zillo_leviathan_sovereign_horn x4
		{DropRoll: 0, QuantityRoll: 9999}, // Bonus 1: vonduun_amphistaff_chitin x2
		{DropRoll: 0, QuantityRoll: 0},    // Bonus 2: ek_zekkil_kyber_greatsaber x1
		{DropRoll: 0, QuantityRoll: 0},    // Bonus 3: zillo_tectonic_bulwark_cuirass x1
	})
	if err != nil {
		t.Fatalf("roll zamorak-zillo-apocalypse-harvest: %v", err)
	}
	if len(drops) != 5 ||
		drops[0] != (ItemQuantity{ItemID: "zamorak_nihilus_chaos_core", Quantity: 2}) ||
		drops[1] != (ItemQuantity{ItemID: "zillo_leviathan_sovereign_horn", Quantity: 4}) ||
		drops[2] != (ItemQuantity{ItemID: "vonduun_amphistaff_chitin", Quantity: 2}) ||
		drops[3] != (ItemQuantity{ItemID: "ek_zekkil_kyber_greatsaber", Quantity: 1}) ||
		drops[4] != (ItemQuantity{ItemID: "zillo_tectonic_bulwark_cuirass", Quantity: 1}) {
		t.Fatalf("unexpected Zamorakian Zillo-Nihilus Worldbreaker boss drops: %+v", drops)
	}
}
