package game

import (
	"math"
	"time"
)

// EnemyBehaviorState is authoritative server state for one live enemy instance.
// It is deliberately independent of a client frame rate: callers advance it
// with server timestamps and persist it alongside the enemy spawn record.
type EnemyBehaviorState string

const (
	EnemyIdle EnemyBehaviorState = "idle"
	EnemyChasing EnemyBehaviorState = "chasing"
	EnemyTelegraphing EnemyBehaviorState = "telegraphing"
	EnemyReturning EnemyBehaviorState = "returning"
)

// EnemyAIState is instance state, not shared definition data.
type EnemyAIState struct {
	Behavior EnemyBehaviorState
	Position WorldPosition
	TargetCharacterID string
	TelegraphUntil time.Time
	NextAttackAt time.Time
	LastUpdatedAt time.Time
}

// EnemyAIConfig contains server-owned tuning values.
type EnemyAIConfig struct {
	AggroRange float64
	LeashRange float64
	MoveSpeed float64 // world units per second
	AttackRange float64
	AttackCooldown time.Duration
	TelegraphDuration time.Duration
}

// EnemyAIResult reports an event for the authoritative combat/persistence layer.
// AttackRequested means the enemy completed its telegraph in range; the caller
// must apply damage transactionally and must never accept client-reported damage.
type EnemyAIResult struct {
	State EnemyAIState
	StateChanged bool
	AttackRequested bool
	ReachedSpawn bool
}

// AdvanceEnemyAI advances one enemy by one server tick. It does not mutate the
// player or award damage/XP; callers must commit the returned state and any
// requested attack in the same transaction.
func AdvanceEnemyAI(definition EnemyDefinition, config EnemyAIConfig, state EnemyAIState, player WorldPosition, now time.Time) EnemyAIResult {
	if state.Position.ZoneID == "" {
		state.Position = definition.Position
	}
	if state.Behavior == "" {
		state.Behavior = EnemyIdle
	}
	previous := state.Behavior
	if state.LastUpdatedAt.IsZero() {
		state.LastUpdatedAt = now
	} else if now.Before(state.LastUpdatedAt) {
		// Server clocks can differ slightly across replicas. Never move the
		// persisted simulation clock backwards when a stale tick arrives.
		now = state.LastUpdatedAt
	}
	delta := now.Sub(state.LastUpdatedAt).Seconds()
	if delta < 0 { delta = 0 }
	if delta > 1 { delta = 1 } // bound catch-up movement after pauses
	state.LastUpdatedAt = now

	sameZone := player.ZoneID == state.Position.ZoneID
	distance := math.Inf(1)
	if sameZone { distance = math.Hypot(player.X-state.Position.X, player.Z-state.Position.Z) }
	spawnDistance := math.Hypot(state.Position.X-definition.Position.X, state.Position.Z-definition.Position.Z)

	if !sameZone || spawnDistance > config.LeashRange || distance > config.LeashRange {
		state.Behavior = EnemyReturning
		moveToward(&state.Position, definition.Position, config.MoveSpeed*delta)
		if math.Hypot(state.Position.X-definition.Position.X, state.Position.Z-definition.Position.Z) < 0.01 {
			state.Position = definition.Position
			state.Behavior = EnemyIdle
			state.TelegraphUntil = time.Time{}
			state.NextAttackAt = time.Time{}
			return EnemyAIResult{State: state, StateChanged: previous != state.Behavior, ReachedSpawn: true}
		}
		return EnemyAIResult{State: state, StateChanged: previous != state.Behavior}
	}

	switch state.Behavior {
	case EnemyIdle, EnemyReturning:
		if distance <= config.AggroRange { state.Behavior = EnemyChasing }
	case EnemyChasing:
		if distance <= config.AttackRange {
			if !now.Before(state.NextAttackAt) {
				state.Behavior = EnemyTelegraphing
				state.TelegraphUntil = now.Add(config.TelegraphDuration)
			}
		} else {
			moveToward(&state.Position, player, config.MoveSpeed*delta)
		}
	case EnemyTelegraphing:
		// Moving out of attack range safely interrupts the telegraphed attack.
		if distance > config.AttackRange {
			state.Behavior = EnemyChasing
			state.TelegraphUntil = time.Time{}
		} else if !now.Before(state.TelegraphUntil) {
			state.Behavior = EnemyChasing
			state.TelegraphUntil = time.Time{}
			state.NextAttackAt = now.Add(config.AttackCooldown)
			return EnemyAIResult{State: state, StateChanged: previous != state.Behavior, AttackRequested: true}
		}
	}
	return EnemyAIResult{State: state, StateChanged: previous != state.Behavior}
}

func moveToward(from *WorldPosition, target WorldPosition, maxDistance float64) {
	if maxDistance <= 0 || from.ZoneID != target.ZoneID { return }
	dx, dz := target.X-from.X, target.Z-from.Z
	distance := math.Hypot(dx, dz)
	if distance == 0 { return }
	step := math.Min(distance, maxDistance)
	from.X += dx / distance * step
	from.Z += dz / distance * step
}

 
// EnemySnapshot is the read-only, zone-scoped state clients use to render
// server-authoritative enemy instances. It contains no account/player secrets.
type EnemySnapshot struct {
	EnemyID string `json:"enemyId"`
	Name string `json:"name"`
	ZoneID string `json:"zoneId"`
	X float64 `json:"x"`
	Z float64 `json:"z"`
	Health int `json:"health"`
	MaxHealth int `json:"maxHealth"`
	Behavior EnemyBehaviorState `json:"behavior"`
	TelegraphUntil time.Time `json:"telegraphUntil,omitempty"`
	RespawnAt time.Time `json:"respawnAt,omitempty"`
	LastUpdatedAt time.Time `json:"lastUpdatedAt,omitempty"`
	BossID string `json:"bossId,omitempty"`
	BossDefinitionVersion int `json:"bossDefinitionVersion,omitempty"`
	BossPhaseID string `json:"bossPhaseId,omitempty"`
	BossPhaseStartedAt *time.Time `json:"bossPhaseStartedAt,omitempty"`
	BossActiveAbilityID string `json:"bossActiveAbilityId,omitempty"`
	BossTelegraphStartedAt *time.Time `json:"bossTelegraphStartedAt,omitempty"`
	BossTelegraphUntil *time.Time `json:"bossTelegraphUntil,omitempty"`
}
