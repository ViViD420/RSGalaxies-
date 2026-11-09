package game

import (
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
	if !result.AttackRequested || result.State.NextAttackAt != now.Add(6*time.Second) { t.Fatalf("attack event=%+v", result) }
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
	if result.State.Behavior != EnemyIdle || !result.ReachedSpawn || result.State.Position.X != 0 { t.Fatalf("home result=%+v", result) }
}

func TestEnemyAIRejectsCrossZoneChaseAndBoundsLongTick(t *testing.T) {
	enemy, config, now := testEnemyAI()
	state := EnemyAIState{Behavior:EnemyChasing, Position:WorldPosition{ZoneID:DefaultWorldZoneID, X:0, Z:0}}
	player := WorldPosition{ZoneID:"another-zone", X:1, Z:0}
	result := AdvanceEnemyAI(enemy, config, state, player, now.Add(time.Hour))
	if result.State.Behavior != EnemyIdle || !result.ReachedSpawn { t.Fatalf("cross-zone leash result=%+v", result) }
}
