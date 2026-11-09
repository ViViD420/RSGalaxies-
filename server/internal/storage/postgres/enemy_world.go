package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
)

type enemyTargetCandidate struct {
	accountID string
	characterID game.PlayerID
}

type enemyStateReader interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func readEnemyAIState(ctx context.Context, reader enemyStateReader, enemy game.EnemyDefinition) (int, bool, sql.NullTime, game.EnemyAIState, error) {
	var health int
	var initialized bool
	var respawnAt sql.NullTime
	var behavior string
	var zoneID sql.NullString
	var x, z sql.NullFloat64
	var telegraphUntil, nextAttackAt, lastUpdatedAt sql.NullTime
	err := reader.QueryRowContext(ctx, `SELECT health, initialized, respawn_at, behavior,
		position_zone_id, position_x, position_z, telegraph_until, next_attack_at, last_ai_updated_at
		FROM enemy_spawn_state WHERE enemy_id=$1`, enemy.ID).Scan(
		&health, &initialized, &respawnAt, &behavior, &zoneID, &x, &z,
		&telegraphUntil, &nextAttackAt, &lastUpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, sql.NullTime{}, game.EnemyAIState{Behavior: game.EnemyIdle, Position: enemy.Position}, nil
	}
	if err != nil { return 0, false, sql.NullTime{}, game.EnemyAIState{}, fmt.Errorf("read enemy AI state: %w", err) }
	state := game.EnemyAIState{Behavior: game.EnemyBehaviorState(behavior), Position: enemy.Position}
	if zoneID.Valid && x.Valid && z.Valid { state.Position = game.WorldPosition{ZoneID: zoneID.String, X: x.Float64, Z: z.Float64} }
	if telegraphUntil.Valid { state.TelegraphUntil = telegraphUntil.Time.UTC() }
	if nextAttackAt.Valid { state.NextAttackAt = nextAttackAt.Time.UTC() }
	if lastUpdatedAt.Valid { state.LastUpdatedAt = lastUpdatedAt.Time.UTC() }
	return health, initialized, respawnAt, state, nil
}

func (r *AccountRepository) activeEnemyCandidates(ctx context.Context, enemy game.EnemyDefinition, state game.EnemyAIState, now time.Time) ([]enemyTargetCandidate, error) {
	config := enemyAIConfig(enemy)
	rows, err := r.db.QueryContext(ctx, `SELECT account_id, character_id FROM zone_presence
		WHERE zone_id=$1 AND last_seen_at>$2
		AND position_x BETWEEN $3 AND $4 AND position_z BETWEEN $5 AND $6
		ORDER BY POWER(position_x-$7,2)+POWER(position_z-$8,2), character_id
		LIMIT $9`,
		state.Position.ZoneID, now.UTC().Add(-game.PresenceLease),
		state.Position.X-config.LeashRange, state.Position.X+config.LeashRange,
		state.Position.Z-config.LeashRange, state.Position.Z+config.LeashRange,
		state.Position.X, state.Position.Z, game.MaxZonePresence)
	if err != nil { return nil, fmt.Errorf("select active enemy targets: %w", err) }
	defer rows.Close()
	candidates := make([]enemyTargetCandidate, 0, game.MaxZonePresence)
	for rows.Next() {
		var candidate enemyTargetCandidate
		if err := rows.Scan(&candidate.accountID, &candidate.characterID); err != nil { return nil, fmt.Errorf("scan active enemy target: %w", err) }
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate active enemy targets: %w", err) }
	return candidates, nil
}

// TickEnemyAI advances one shared enemy spawn. It selects a fresh character
// presence, then locks that character before locking the enemy row. This
// matches player combat's player-then-enemy ordering and commits telegraph
// completion, damage, health, Defense XP and AI state atomically.
func (r *AccountRepository) TickEnemyAI(ctx context.Context, enemyID string, now time.Time) error {
	enemy, ok := game.FindEnemy(enemyID)
	if !ok { return game.ErrUnknownEnemy }
	now = now.UTC()
	_, _, _, observedState, err := readEnemyAIState(ctx, r.db, enemy)
	if err != nil { return err }
	candidates, err := r.activeEnemyCandidates(ctx, enemy, observedState, now)
	if err != nil { return err }

	for _, candidate := range candidates {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil { return fmt.Errorf("begin enemy world tick: %w", err) }
		player, err := r.lockOwnedPlayer(ctx, tx, candidate.accountID, candidate.characterID)
		if errors.Is(err, game.ErrPlayerNotFound) { _ = tx.Rollback(); continue }
		if err != nil { _ = tx.Rollback(); return fmt.Errorf("lock enemy target character: %w", err) }
		if player.Health <= 0 { _ = tx.Rollback(); continue }

		var present bool
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM zone_presence
			WHERE character_id=$1 AND account_id=$2 AND zone_id=$3 AND last_seen_at>$4)`,
			string(player.ID), candidate.accountID, player.Position.ZoneID, now.Add(-game.PresenceLease)).Scan(&present)
		if err != nil { _ = tx.Rollback(); return fmt.Errorf("revalidate enemy target presence: %w", err) }
		if !present { _ = tx.Rollback(); continue }

		if _, err := tx.ExecContext(ctx, `INSERT INTO enemy_spawn_state(enemy_id,health,initialized)
			VALUES($1,0,FALSE) ON CONFLICT(enemy_id) DO NOTHING`, enemyID); err != nil {
			_ = tx.Rollback(); return fmt.Errorf("initialize enemy spawn for AI: %w", err)
		}
		health, initialized, respawnAt, state, err := readEnemyAIStateForUpdate(ctx, tx, enemy)
		if err != nil { _ = tx.Rollback(); return err }
		if !initialized { health, initialized = enemy.MaxHealth, true }
		if health <= 0 {
			if respawnAt.Valid && now.Before(respawnAt.Time) { _ = tx.Rollback(); return nil }
			health, respawnAt = enemy.MaxHealth, sql.NullTime{}
			state = game.EnemyAIState{Behavior: game.EnemyIdle, Position: enemy.Position}
		}
		result := game.AdvanceEnemyAI(enemy, enemyAIConfig(enemy), state, player.Position, now)
		if result.AttackRequested {
			if _, err := player.ReceiveEnemyAttack(enemy, result.State.Position, now); err != nil {
				_ = tx.Rollback()
				if errors.Is(err, game.ErrEnemyOutOfRange) || errors.Is(err, game.ErrPlayerDefeated) { return nil }
				return fmt.Errorf("resolve telegraphed enemy attack: %w", err)
			}
			if err := updateOwnedPlayer(ctx, tx, candidate.accountID, player); err != nil {
				_ = tx.Rollback(); return fmt.Errorf("persist enemy attack target: %w", err)
			}
		}
		if err := persistEnemyAIState(ctx, tx, enemyID, health, initialized, respawnAt, result.State); err != nil {
			_ = tx.Rollback(); return err
		}
		if err := tx.Commit(); err != nil { return fmt.Errorf("commit enemy world tick: %w", err) }
		return nil
	}

	// No fresh eligible character is present near this spawn. Advance with an
	// absent-zone target so a chasing enemy safely returns to its spawn point.
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return fmt.Errorf("begin idle enemy tick: %w", err) }
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO enemy_spawn_state(enemy_id,health,initialized)
		VALUES($1,0,FALSE) ON CONFLICT(enemy_id) DO NOTHING`, enemyID); err != nil {
		return fmt.Errorf("initialize idle enemy spawn: %w", err)
	}
	health, initialized, respawnAt, state, err := readEnemyAIStateForUpdate(ctx, tx, enemy)
	if err != nil { return err }
	if !initialized { health, initialized = enemy.MaxHealth, true }
	if health <= 0 {
		if respawnAt.Valid && now.Before(respawnAt.Time) { return nil }
		health, respawnAt = enemy.MaxHealth, sql.NullTime{}
		state = game.EnemyAIState{Behavior: game.EnemyIdle, Position: enemy.Position}
	}
	result := game.AdvanceEnemyAI(enemy, enemyAIConfig(enemy), state, game.WorldPosition{}, now)
	if err := persistEnemyAIState(ctx, tx, enemyID, health, initialized, respawnAt, result.State); err != nil { return err }
	if err := tx.Commit(); err != nil { return fmt.Errorf("commit idle enemy tick: %w", err) }
	return nil
}

func readEnemyAIStateForUpdate(ctx context.Context, tx *sql.Tx, enemy game.EnemyDefinition) (int, bool, sql.NullTime, game.EnemyAIState, error) {
	var health int
	var initialized bool
	var respawnAt sql.NullTime
	var behavior string
	var zoneID sql.NullString
	var x, z sql.NullFloat64
	var telegraphUntil, nextAttackAt, lastUpdatedAt sql.NullTime
	err := tx.QueryRowContext(ctx, `SELECT health, initialized, respawn_at, behavior,
		position_zone_id, position_x, position_z, telegraph_until, next_attack_at, last_ai_updated_at
		FROM enemy_spawn_state WHERE enemy_id=$1 FOR UPDATE`, enemy.ID).Scan(
		&health, &initialized, &respawnAt, &behavior, &zoneID, &x, &z,
		&telegraphUntil, &nextAttackAt, &lastUpdatedAt,
	)
	if err != nil { return 0, false, sql.NullTime{}, game.EnemyAIState{}, fmt.Errorf("lock enemy AI state: %w", err) }
	state := game.EnemyAIState{Behavior: game.EnemyBehaviorState(behavior), Position: enemy.Position}
	if zoneID.Valid && x.Valid && z.Valid { state.Position = game.WorldPosition{ZoneID: zoneID.String, X: x.Float64, Z: z.Float64} }
	if telegraphUntil.Valid { state.TelegraphUntil = telegraphUntil.Time.UTC() }
	if nextAttackAt.Valid { state.NextAttackAt = nextAttackAt.Time.UTC() }
	if lastUpdatedAt.Valid { state.LastUpdatedAt = lastUpdatedAt.Time.UTC() }
	return health, initialized, respawnAt, state, nil
}

func persistEnemyAIState(ctx context.Context, tx *sql.Tx, enemyID string, health int, initialized bool, respawnAt sql.NullTime, state game.EnemyAIState) error {
	var respawnValue any
	if respawnAt.Valid { respawnValue = respawnAt.Time.UTC() }
	_, err := tx.ExecContext(ctx, `UPDATE enemy_spawn_state SET health=$2, initialized=$3,
		behavior=$4, position_zone_id=$5, position_x=$6, position_z=$7,
		telegraph_until=$8, next_attack_at=$9, last_ai_updated_at=$10, respawn_at=$11
		WHERE enemy_id=$1`,
		enemyID, health, initialized, string(state.Behavior), state.Position.ZoneID,
		state.Position.X, state.Position.Z, nullableTime(state.TelegraphUntil),
		nullableTime(state.NextAttackAt), state.LastUpdatedAt.UTC(), respawnValue)
	if err != nil { return fmt.Errorf("persist enemy AI world state: %w", err) }
	return nil
}

// RunEnemyAITick processes all catalogued spawns. The caller owns cadence and
// lifecycle; multiple server instances remain safe through PostgreSQL row locks.
func (r *AccountRepository) RunEnemyAITick(ctx context.Context, now time.Time) error {
	for _, enemy := range game.EnemyDefinitions() {
		if err := r.TickEnemyAI(ctx, enemy.ID, now); err != nil {
			return fmt.Errorf("tick enemy %s: %w", enemy.ID, err)
		}
	}
	return nil
}
