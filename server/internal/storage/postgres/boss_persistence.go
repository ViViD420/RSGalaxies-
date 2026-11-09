package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
)

var ErrBossEncounterCycleMismatch = errors.New("boss encounter cycle does not match active enemy cycle")

// loadBossEncounterStateForUpdate must run inside the caller's combat
// transaction. Locking the parent spawn row first gives boss ticks the same
// serialization point as enemy AI, player attacks and encounter-cycle changes.
func loadBossEncounterStateForUpdate(
	ctx context.Context,
	tx *sql.Tx,
	enemyID string,
	encounterID string,
) (game.BossEncounterState, string, bool, error) {
	if enemyID == "" || encounterID == "" {
		return game.BossEncounterState{}, "", false, ErrBossEncounterCycleMismatch
	}

	var activeCycle sql.NullString
	if err := tx.QueryRowContext(ctx,
		"SELECT encounter_id FROM enemy_spawn_state WHERE enemy_id=$1 FOR UPDATE",
		enemyID,
	).Scan(&activeCycle); err != nil {
		return game.BossEncounterState{}, "", false, fmt.Errorf("lock boss enemy cycle: %w", err)
	}
	if !activeCycle.Valid || activeCycle.String != encounterID {
		return game.BossEncounterState{}, "", false, ErrBossEncounterCycleMismatch
	}

	var state game.BossEncounterState
	var definitionID string
	var activeAbility sql.NullString
	var telegraphStarted, telegraphUntil sql.NullTime
	var cooldownJSON []byte
	err := tx.QueryRowContext(ctx, `SELECT definition_id, definition_version, phase_id,
		phase_started_at, active_ability_id, telegraph_started_at, telegraph_until,
		ability_cooldowns, next_ability_index, last_updated_at, defeated
		FROM boss_encounter_states WHERE enemy_id=$1 AND encounter_id=$2 FOR UPDATE`,
		enemyID, encounterID,
	).Scan(
		&definitionID, &state.DefinitionVersion, &state.PhaseID,
		&state.PhaseStartedAt, &activeAbility, &telegraphStarted, &telegraphUntil,
		&cooldownJSON, &state.NextAbilityIndex, &state.LastUpdatedAt, &state.Defeated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return game.BossEncounterState{}, "", false, nil
	}
	if err != nil {
		return game.BossEncounterState{}, "", false, fmt.Errorf("load locked boss encounter state: %w", err)
	}

	state.EncounterID = encounterID
	state.PhaseStartedAt = state.PhaseStartedAt.UTC()
	state.LastUpdatedAt = state.LastUpdatedAt.UTC()
	if activeAbility.Valid {
		state.ActiveAbilityID = activeAbility.String
	}
	if telegraphStarted.Valid {
		state.TelegraphStartedAt = telegraphStarted.Time.UTC()
	}
	if telegraphUntil.Valid {
		state.TelegraphUntil = telegraphUntil.Time.UTC()
	}
	if err := json.Unmarshal(cooldownJSON, &state.AbilityCooldowns); err != nil {
		return game.BossEncounterState{}, "", false, fmt.Errorf("decode boss ability cooldowns: %w", err)
	}
	if state.AbilityCooldowns == nil {
		state.AbilityCooldowns = make(map[string]time.Time)
	}
	for abilityID, readyAt := range state.AbilityCooldowns {
		if abilityID == "" || readyAt.IsZero() {
			return game.BossEncounterState{}, "", false, fmt.Errorf("invalid persisted cooldown for boss ability %q", abilityID)
		}
		state.AbilityCooldowns[abilityID] = readyAt.UTC()
	}
	if state.DefinitionVersion <= 0 || state.PhaseID == "" ||
		state.PhaseStartedAt.IsZero() || state.LastUpdatedAt.IsZero() ||
		state.NextAbilityIndex < 0 ||
		(state.ActiveAbilityID == "") != (state.TelegraphStartedAt.IsZero() && state.TelegraphUntil.IsZero()) ||
		(!state.TelegraphStartedAt.IsZero() && state.TelegraphUntil.Before(state.TelegraphStartedAt)) ||
		(state.Defeated && state.ActiveAbilityID != "") {
		return game.BossEncounterState{}, "", false, errors.New("invalid persisted boss encounter state")
	}
	return state, definitionID, true, nil
}

// persistBossEncounterState writes the complete deterministic boss state in
// the same transaction as enemy/player combat. The parent-cycle lock prevents
// an old encounter worker from writing into a respawned boss cycle.
func persistBossEncounterState(
	ctx context.Context,
	tx *sql.Tx,
	enemyID string,
	definitionID string,
	state game.BossEncounterState,
) error {
	if enemyID == "" || definitionID == "" || state.EncounterID == "" ||
		state.DefinitionVersion <= 0 || state.PhaseID == "" ||
		state.PhaseStartedAt.IsZero() || state.LastUpdatedAt.IsZero() ||
		state.NextAbilityIndex < 0 ||
		(state.ActiveAbilityID == "") != (state.TelegraphStartedAt.IsZero() && state.TelegraphUntil.IsZero()) ||
		(!state.TelegraphStartedAt.IsZero() && state.TelegraphUntil.Before(state.TelegraphStartedAt)) ||
		(state.Defeated && state.ActiveAbilityID != "") {
		return errors.New("invalid boss encounter state")
	}
	for abilityID, readyAt := range state.AbilityCooldowns {
		if abilityID == "" || readyAt.IsZero() {
			return fmt.Errorf("invalid cooldown for boss ability %q", abilityID)
		}
	}
	cooldowns := state.AbilityCooldowns
	if cooldowns == nil {
		cooldowns = map[string]time.Time{}
	}
	cooldownJSON, err := json.Marshal(cooldowns)
	if err != nil {
		return fmt.Errorf("encode boss ability cooldowns: %w", err)
	}

	var activeCycle sql.NullString
	if err := tx.QueryRowContext(ctx,
		"SELECT encounter_id FROM enemy_spawn_state WHERE enemy_id=$1 FOR UPDATE",
		enemyID,
	).Scan(&activeCycle); err != nil {
		return fmt.Errorf("lock boss enemy cycle before save: %w", err)
	}
	if !activeCycle.Valid || activeCycle.String != state.EncounterID {
		return ErrBossEncounterCycleMismatch
	}

	result, err := tx.ExecContext(ctx, `INSERT INTO boss_encounter_states (
		enemy_id, encounter_id, definition_id, definition_version, phase_id,
		phase_started_at, active_ability_id, telegraph_started_at, telegraph_until,
		ability_cooldowns, next_ability_index, last_updated_at, defeated
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11,$12,$13)
	ON CONFLICT (enemy_id, encounter_id) DO UPDATE SET
		definition_id=EXCLUDED.definition_id,
		definition_version=EXCLUDED.definition_version,
		phase_id=EXCLUDED.phase_id,
		phase_started_at=EXCLUDED.phase_started_at,
		active_ability_id=EXCLUDED.active_ability_id,
		telegraph_started_at=EXCLUDED.telegraph_started_at,
		telegraph_until=EXCLUDED.telegraph_until,
		ability_cooldowns=EXCLUDED.ability_cooldowns,
		next_ability_index=EXCLUDED.next_ability_index,
		last_updated_at=EXCLUDED.last_updated_at,
		defeated=EXCLUDED.defeated
	WHERE boss_encounter_states.last_updated_at <= EXCLUDED.last_updated_at`,
		enemyID, state.EncounterID, definitionID, state.DefinitionVersion,
		state.PhaseID, state.PhaseStartedAt.UTC(), nullableString(state.ActiveAbilityID),
		nullableTime(state.TelegraphStartedAt), nullableTime(state.TelegraphUntil),
		string(cooldownJSON), state.NextAbilityIndex, state.LastUpdatedAt.UTC(), state.Defeated,
	)
	if err != nil {
		return fmt.Errorf("persist boss encounter state: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect boss encounter persistence result: %w", err)
	}
	if rowsAffected != 1 {
		return errors.New("stale boss encounter state was not persisted")
	}
	return nil
}
