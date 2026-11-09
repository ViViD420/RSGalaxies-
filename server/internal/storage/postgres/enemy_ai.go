package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
)

// EnemyAITickConfig is deliberately private to the server. Clients cannot
// choose aggro, movement, attack range, cooldown, or telegraph timing.
func enemyAIConfig(enemy game.EnemyDefinition) game.EnemyAIConfig {
	return game.EnemyAIConfig{
		AggroRange: 12,
		LeashRange: 24,
		MoveSpeed: 4,
		AttackRange: enemy.InteractionRange,
		AttackCooldown: 3 * time.Second,
		TelegraphDuration: time.Second,
	}
}

// AdvanceEnemyAIForPlayer advances and persists one enemy instance while
// holding its spawn row lock. The caller supplies a position obtained from
// authoritative player storage (never from a client payload).
//
// AttackRequested is an intent for the world-combat coordinator. This method
// does not apply player damage; the coordinator must perform that damage and
// this state transition in one transaction before enabling live enemy attacks.
func (r *PlayerRepository) AdvanceEnemyAIForPlayer(
	ctx context.Context,
	enemyID string,
	playerPosition game.WorldPosition,
	now time.Time,
) (game.EnemyAIResult, error) {
	definition, ok := game.FindEnemy(enemyID)
	if !ok {
		return game.EnemyAIResult{}, game.ErrUnknownEnemy
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return game.EnemyAIResult{}, fmt.Errorf("begin enemy AI tick: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `INSERT INTO enemy_spawn_state(enemy_id, health, initialized)
		VALUES($1, 0, FALSE) ON CONFLICT(enemy_id) DO NOTHING`, enemyID); err != nil {
		return game.EnemyAIResult{}, fmt.Errorf("initialize enemy AI state: %w", err)
	}

	var health int
	var initialized bool
	var behavior string
	var zoneID sql.NullString
	var x, z sql.NullFloat64
	var telegraphUntil, nextAttackAt, lastUpdatedAt sql.NullTime
	var respawnAt sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT health, initialized, respawn_at, behavior,
		position_zone_id, position_x, position_z, telegraph_until, next_attack_at, last_ai_updated_at
		FROM enemy_spawn_state WHERE enemy_id=$1 FOR UPDATE`, enemyID).Scan(
		&health, &initialized, &respawnAt, &behavior, &zoneID, &x, &z,
		&telegraphUntil, &nextAttackAt, &lastUpdatedAt,
	)
	if err != nil {
		return game.EnemyAIResult{}, fmt.Errorf("lock enemy AI state: %w", err)
	}
	if !initialized {
		health = definition.MaxHealth
		initialized = true
	}
	if health <= 0 {
		if respawnAt.Valid && now.Before(respawnAt.Time) {
			return game.EnemyAIResult{}, game.ErrEnemyRespawning
		}
		health = definition.MaxHealth
		respawnAt = sql.NullTime{}
		behavior = string(game.EnemyIdle)
		zoneID = sql.NullString{}
		x, z = sql.NullFloat64{}, sql.NullFloat64{}
		telegraphUntil, nextAttackAt, lastUpdatedAt = sql.NullTime{}, sql.NullTime{}, sql.NullTime{}
	}

	state := game.EnemyAIState{Behavior: game.EnemyBehaviorState(behavior)}
	if zoneID.Valid && x.Valid && z.Valid {
		state.Position = game.WorldPosition{ZoneID: zoneID.String, X: x.Float64, Z: z.Float64}
	}
	if telegraphUntil.Valid { state.TelegraphUntil = telegraphUntil.Time.UTC() }
	if nextAttackAt.Valid { state.NextAttackAt = nextAttackAt.Time.UTC() }
	if lastUpdatedAt.Valid { state.LastUpdatedAt = lastUpdatedAt.Time.UTC() }

	result := game.AdvanceEnemyAI(definition, enemyAIConfig(definition), state, playerPosition, now.UTC())
	_, err = tx.ExecContext(ctx, `UPDATE enemy_spawn_state SET health=$2, initialized=TRUE,
		behavior=$3, position_zone_id=$4, position_x=$5, position_z=$6,
		telegraph_until=$7, next_attack_at=$8, last_ai_updated_at=$9
		WHERE enemy_id=$1`,
		enemyID, health, string(result.State.Behavior), result.State.Position.ZoneID,
		result.State.Position.X, result.State.Position.Z,
		nullableTime(result.State.TelegraphUntil), nullableTime(result.State.NextAttackAt),
		result.State.LastUpdatedAt.UTC(),
	)
	if err != nil {
		return game.EnemyAIResult{}, fmt.Errorf("persist enemy AI tick: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return game.EnemyAIResult{}, fmt.Errorf("commit enemy AI tick: %w", err)
	}
	return result, nil
}

func nullableTime(value time.Time) any {
	if value.IsZero() { return nil }
	return value.UTC()
}

// IsEnemyRespawning reports whether a tick was rejected because the spawn is
// still in its server-owned respawn window.
func IsEnemyRespawning(err error) bool {
	return errors.Is(err, game.ErrEnemyRespawning)
}
