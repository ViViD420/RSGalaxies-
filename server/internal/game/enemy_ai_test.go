package game

import (
	"encoding/json"
	"testing"
	"time"
)

func testEnemyAI() (EnemyDefinition, EnemyAIConfig, time.Time) {
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	enemy := EnemyDefinition{ID:"ai-test", Position:WorldPosition{ZoneID:DefaultWorldZoneID, X:0, Z:0}}
	config := EnemyAIConfig{AggroRange:8, LeashRange:20, MoveSpeed:4, AttackRange:2, AttackCooldown:3*time.Second, TelegraphDuration:time.Second}
	return enemy, config, now
}

func TestEnemyAIAggroChasesAndTelegraphsBeforeAttack(t *testing.T) {
	enemy, config, now := testEnemyAI()
	player := WorldPosition{ZoneID:DefaultWorldZoneID, X:6, Z:0}
	state := EnemyAIState{}
	result := AdvanceEnemyAI(enemy, config, state, player, now)
	if result.State.Behavior != EnemyChasing || result.AttackRequested { t.Fatalf("aggro result=%+v", result) }

	// One second of movement closes four world units.
	result = AdvanceEnemyAI(enemy, config, result.State, player, now.Add(time.Second))
	if result.State.Position.X != 4 || result.State.Behavior != EnemyChasing { t.Fatalf("chase result=%+v", result) }
	result = AdvanceEnemyAI(enemy, config, result.State, player, now.Add(2*time.Second))
	if result.State.Behavior != EnemyTelegraphing || result.AttackRequested { t.Fatalf("telegraph start=%+v", result) }
	result = AdvanceEnemyAI(enemy, config, result.State, player, now.Add(3*time.Second))
	if !result.AttackRequested || result.State.NextAttackAt != now.Add(6*time.Second) || !result.State.TelegraphUntil.IsZero() { t.Fatalf("attack event=%+v", result) }
}

func TestEnemyAIMovementInterruptsTelegraphAndLeashesHome(t *testing.T) {
	enemy, config, now := testEnemyAI()
	player := WorldPosition{ZoneID:DefaultWorldZoneID, X:1, Z:0}
	state := EnemyAIState{Behavior:EnemyChasing, Position:enemy.Position}
	result := AdvanceEnemyAI(enemy, config, state, player, now)
	if result.State.Behavior != EnemyTelegraphing { t.Fatalf("telegraph=%+v", result.State) }
	player.X = 5
	result = AdvanceEnemyAI(enemy, config, result.State, player, now.Add(time.Second))
	if result.AttackRequested || result.State.Behavior != EnemyChasing { t.Fatalf("telegraph was not interrupted: %+v", result) }

	// A player beyond the leash causes the enemy to return, not chase forever.
	result.State.Position.X = 10
	player.X = 40
	result = AdvanceEnemyAI(enemy, config, result.State, player, now.Add(2*time.Second))
	if result.State.Behavior != EnemyReturning || result.State.Position.X >= 10 { t.Fatalf("leash result=%+v", result) }
	result = AdvanceEnemyAI(enemy, config, result.State, player, now.Add(10*time.Second))
	result = AdvanceEnemyAI(enemy, config, result.State, player, now.Add(11*time.Second))
	if !result.ReachedSpawn || result.State.Position.X != 0 { t.Fatalf("spawn arrival result=%+v", result) }
	result = AdvanceEnemyAI(enemy, config, result.State, player, now.Add(12*time.Second))
	if result.State.Behavior != EnemyIdle { t.Fatalf("idle-at-spawn result=%+v", result) }
}

func TestEnemyAIRejectsCrossZoneChaseAndBoundsLongTick(t *testing.T) {
	enemy, config, now := testEnemyAI()
	state := EnemyAIState{Behavior:EnemyChasing, Position:WorldPosition{ZoneID:DefaultWorldZoneID, X:0, Z:0}}
	player := WorldPosition{ZoneID:"another-zone", X:1, Z:0}
	result := AdvanceEnemyAI(enemy, config, state, player, now.Add(time.Hour))
	if result.State.Behavior != EnemyIdle || !result.ReachedSpawn { t.Fatalf("cross-zone leash result=%+v", result) }
}

 
func TestEnemyAITimestampNeverMovesBackward(t *testing.T) {
	enemy, config, now := testEnemyAI()
	player := WorldPosition{ZoneID: DefaultWorldZoneID, X: 10, Z: 0}
	state := EnemyAIState{
		Behavior: EnemyChasing,
		Position: enemy.Position,
		LastUpdatedAt: now.Add(2 * time.Second),
	}
	result := AdvanceEnemyAI(enemy, config, state, player, now.Add(time.Second))
	if !result.State.LastUpdatedAt.Equal(now.Add(2*time.Second)) {
		t.Fatalf("simulation clock moved backwards: %s", result.State.LastUpdatedAt)
	}
	if result.State.Position.X != 0 {
		t.Fatalf("stale tick moved enemy: x=%v", result.State.Position.X)
	}
	next := AdvanceEnemyAI(enemy, config, result.State, player, now.Add(3*time.Second))
	if next.State.Position.X != 4 || !next.State.LastUpdatedAt.Equal(now.Add(3*time.Second)) {
		t.Fatalf("monotonic next tick = %+v", next.State)
	}
}

func TestEnemySnapshotBossFieldsAreOptionalAndAuthoritative(t *testing.T) {
	ordinary := EnemySnapshot{
		EnemyID: "frontier-rat-001",
		Name: "Frontier Rat",
		ZoneID: DefaultWorldZoneID,
		Health: 12,
		MaxHealth: 12,
		Behavior: EnemyIdle,
	}
	ordinaryJSON, err := json.Marshal(ordinary)
	if err != nil {
		t.Fatal(err)
	}
	var ordinaryPayload map[string]any
	if err := json.Unmarshal(ordinaryJSON, &ordinaryPayload); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		"bossId", "bossDefinitionVersion", "bossPhaseId",
		"bossPhaseStartedAt", "bossActiveAbilityId",
		"bossTelegraphStartedAt", "bossTelegraphUntil",
	} {
		if _, exists := ordinaryPayload[key]; exists {
			t.Fatalf("ordinary enemy snapshot unexpectedly includes %s: %s", key, ordinaryJSON)
		}
	}

	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	deadline := now.Add(2 * time.Second)
	boss := ordinary
	boss.BossID = "stoneback-colossus"
	boss.BossDefinitionVersion = 1
	boss.BossPhaseID = "rampage"
	boss.BossPhaseStartedAt = &now
	boss.BossActiveAbilityID = "ground-shatter"
	boss.BossTelegraphStartedAt = &now
	boss.BossTelegraphUntil = &deadline
	bossJSON, err := json.Marshal(boss)
	if err != nil {
		t.Fatal(err)
	}
	var bossPayload map[string]any
	if err := json.Unmarshal(bossJSON, &bossPayload); err != nil {
		t.Fatal(err)
	}
	if bossPayload["bossId"] != "stoneback-colossus" ||
		bossPayload["bossDefinitionVersion"] != float64(1) ||
		bossPayload["bossPhaseId"] != "rampage" ||
		bossPayload["bossActiveAbilityId"] != "ground-shatter" {
		t.Fatalf("boss snapshot fields missing or incorrect: %s", bossJSON)
	}
	if bossPayload["bossTelegraphStartedAt"] != now.Format(time.RFC3339Nano) ||
		bossPayload["bossTelegraphUntil"] != deadline.Format(time.RFC3339Nano) {
		t.Fatalf("boss telegraph timestamps incorrect: %s", bossJSON)
	}
}

