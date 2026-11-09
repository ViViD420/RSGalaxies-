package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
)

func newEncounterID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil { return "", fmt.Errorf("generate encounter id: %w", err) }
	return hex.EncodeToString(raw[:]), nil
}

func ensureEncounterID(ctx context.Context, tx *sql.Tx, enemyID string, resetCycle bool) (string, error) {
	var current sql.NullString
	if err := tx.QueryRowContext(ctx, "SELECT encounter_id FROM enemy_spawn_state WHERE enemy_id=$1", enemyID).Scan(&current); err != nil {
		return "", fmt.Errorf("read encounter cycle id: %w", err)
	}
	if !resetCycle && current.Valid && current.String != "" { return current.String, nil }
	id, err := newEncounterID()
	if err != nil { return "", err }
	if _, err := tx.ExecContext(ctx, "UPDATE enemy_spawn_state SET encounter_id=$2 WHERE enemy_id=$1", enemyID, id); err != nil {
		return "", fmt.Errorf("persist encounter cycle id: %w", err)
	}
	return id, nil
}

func recordEncounterContribution(ctx context.Context, tx *sql.Tx, encounterID, enemyID, accountID string, characterID game.PlayerID, experienceSkillID string, damageDealt, damageReceived int, now time.Time) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO enemy_encounter_participants(
		encounter_id,enemy_id,account_id,character_id,damage_dealt,damage_received,action_count,first_contribution_at,last_contribution_at,experience_skill_id)
		VALUES($1,$2,$3,$4,$5,$6,1,$7,$7,$8)
		ON CONFLICT(encounter_id,character_id) DO UPDATE SET
			damage_dealt=enemy_encounter_participants.damage_dealt+EXCLUDED.damage_dealt,
			damage_received=enemy_encounter_participants.damage_received+EXCLUDED.damage_received,
			action_count=enemy_encounter_participants.action_count+1,
			last_contribution_at=EXCLUDED.last_contribution_at,
			experience_skill_id=CASE WHEN EXCLUDED.damage_dealt>0 THEN EXCLUDED.experience_skill_id ELSE enemy_encounter_participants.experience_skill_id END`,
		encounterID,enemyID,accountID,string(characterID),damageDealt,damageReceived,now.UTC(),experienceSkillID)
	if err != nil { return fmt.Errorf("record encounter contribution: %w", err) }
	return nil
}

func recordEncounterDamageReceived(ctx context.Context, tx *sql.Tx, encounterID, enemyID string, characterID game.PlayerID, damage int, _ time.Time) error {
	if damage <= 0 { return nil }
	_, err := tx.ExecContext(ctx, `UPDATE enemy_encounter_participants SET
		damage_received=damage_received+$4,action_count=action_count+1
		WHERE encounter_id=$1 AND enemy_id=$2 AND character_id=$3`,
		encounterID,enemyID,string(characterID),damage)
	if err != nil { return fmt.Errorf("record encounter damage received: %w", err) }
	return nil
}

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
	var zoneID, targetCharacterID sql.NullString
	var x, z sql.NullFloat64
	var telegraphUntil, nextAttackAt, lastUpdatedAt sql.NullTime
	err := reader.QueryRowContext(ctx, `SELECT health, initialized, respawn_at, behavior,
		position_zone_id, position_x, position_z, target_character_id,
		telegraph_until, next_attack_at, last_ai_updated_at
		FROM enemy_spawn_state WHERE enemy_id=$1`, enemy.ID).Scan(
		&health, &initialized, &respawnAt, &behavior, &zoneID, &x, &z,
		&targetCharacterID, &telegraphUntil, &nextAttackAt, &lastUpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, sql.NullTime{}, game.EnemyAIState{Behavior: game.EnemyIdle, Position: enemy.Position}, nil
	}
	if err != nil { return 0, false, sql.NullTime{}, game.EnemyAIState{}, fmt.Errorf("read enemy AI state: %w", err) }
	state := game.EnemyAIState{Behavior: game.EnemyBehaviorState(behavior), Position: enemy.Position}
	if zoneID.Valid && x.Valid && z.Valid { state.Position = game.WorldPosition{ZoneID: zoneID.String, X: x.Float64, Z: z.Float64} }
	if targetCharacterID.Valid { state.TargetCharacterID = targetCharacterID.String }
	if telegraphUntil.Valid { state.TelegraphUntil = telegraphUntil.Time.UTC() }
	if nextAttackAt.Valid { state.NextAttackAt = nextAttackAt.Time.UTC() }
	if lastUpdatedAt.Valid { state.LastUpdatedAt = lastUpdatedAt.Time.UTC() }
	return health, initialized, respawnAt, state, nil
}

func (r *AccountRepository) activeEnemyCandidates(ctx context.Context, enemy game.EnemyDefinition, state game.EnemyAIState, now time.Time) ([]enemyTargetCandidate, error) {
	config := enemyAIConfig(enemy)
	rows, err := r.db.QueryContext(ctx, `SELECT p.account_id, p.character_id FROM zone_presence p
		LEFT JOIN enemy_spawn_state spawn ON spawn.enemy_id=$11
		LEFT JOIN enemy_encounter_participants contribution
			ON contribution.enemy_id=spawn.enemy_id
			AND contribution.encounter_id=spawn.encounter_id
			AND contribution.character_id=p.character_id
		WHERE p.zone_id=$1 AND p.last_seen_at>$2
		AND p.position_x BETWEEN $3 AND $4 AND p.position_z BETWEEN $5 AND $6
		ORDER BY CASE WHEN p.character_id=$9 THEN 0 ELSE 1 END,
			(COALESCE(contribution.damage_dealt,0)*2
			 + COALESCE(contribution.damage_received,0)
			 + COALESCE(contribution.action_count,0)) DESC,
			POWER(p.position_x-$7,2)+POWER(p.position_z-$8,2), p.character_id
		LIMIT $10`,
		state.Position.ZoneID, now.UTC().Add(-game.PresenceLease),
		state.Position.X-config.LeashRange, state.Position.X+config.LeashRange,
		state.Position.Z-config.LeashRange, state.Position.Z+config.LeashRange,
		state.Position.X, state.Position.Z, state.TargetCharacterID, game.MaxZonePresence, enemy.ID)
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
// advanceBossEncounterInTransaction advances one boss and, if its
// telegraph resolves, applies player damage/Defense XP in the same transaction
// as the boss state. A nil player represents an empty spawn tick and can only
// interrupt an active ability, never resolve one.
func advanceBossEncounterInTransaction(
	ctx context.Context,
	tx *sql.Tx,
	enemy game.EnemyDefinition,
	definition game.BossEncounterDefinition,
	encounterID string,
	health int,
	aiState *game.EnemyAIState,
	player *game.Player,
	accountID string,
	now time.Time,
) error {
	bossState, storedDefinitionID, found, err := loadBossEncounterStateForUpdate(ctx, tx, enemy.ID, encounterID)
	if err != nil {
		return err
	}
	if found && storedDefinitionID != definition.ID {
		return fmt.Errorf("boss definition mismatch: stored %q, active %q", storedDefinitionID, definition.ID)
	}

	targetPosition := game.WorldPosition{}
	if player != nil {
		targetPosition = player.Position
	}
	result, err := game.AdvanceBossEncounter(definition, bossState, game.BossEncounterInput{
		EncounterID: encounterID,
		Health: health,
		MaxHealth: enemy.MaxHealth,
		BossPosition: aiState.Position,
		TargetPosition: targetPosition,
		Now: now.UTC(),
	})
	if err != nil {
		return fmt.Errorf("advance boss encounter %s: %w", definition.ID, err)
	}

	if result.AbilityResolved {
		if player == nil {
			return errors.New("boss ability resolved without a locked target player")
		}
		ability, ok := game.FindBossAbility(definition, result.AbilityID)
		if !ok {
			return fmt.Errorf("resolved boss ability %q is missing from definition", result.AbilityID)
		}
		attack, err := player.ReceiveBossAttack(result.Damage, aiState.Position, ability.MinRange, ability.MaxRange, now)
		if err != nil {
			return fmt.Errorf("resolve boss ability %s: %w", ability.ID, err)
		}
		if player.Health <= 0 {
			aiState.TargetCharacterID = ""
		}
		if err := recordEncounterDamageReceived(ctx, tx, encounterID, enemy.ID, player.ID, attack.DamageReceived, now); err != nil {
			return err
		}
		if err := updateOwnedPlayer(ctx, tx, accountID, *player); err != nil {
			return fmt.Errorf("persist boss ability target: %w", err)
		}
	}

	return persistBossEncounterState(ctx, tx, enemy.ID, definition.ID, result.State)
}

func (r *AccountRepository) TickEnemyAI(ctx context.Context, enemyID string, now time.Time) error {
	enemy, ok := game.FindEnemy(enemyID)
	if !ok {
		return game.ErrUnknownEnemy
	}
	return r.tickEnemyAIForDefinition(ctx, enemy, now)
}

func (r *AccountRepository) tickEnemyAIForDefinition(ctx context.Context, enemy game.EnemyDefinition, now time.Time) error {
	enemyID := enemy.ID
	now = now.UTC()
	_, _, _, observedState, err := readEnemyAIState(ctx, r.db, enemy)
	if err != nil { return err }
	candidates, err := r.activeEnemyCandidates(ctx, enemy, observedState, now)
	if err != nil { return err }
	bossDefinition, isBoss := game.FindBossEncounterForEnemy(enemy.ID)
	aiConfig := enemyAIConfig(enemy)
	if isBoss {
		// Boss abilities own their range and telegraph. Disable the ordinary
		// enemy attack channel while retaining shared movement and leash logic.
		aiConfig.AttackRange = -1
		aiConfig.TelegraphDuration = 0
	}

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

		resetEncounter := !initialized || health <= 0
		if !initialized { health, initialized = enemy.MaxHealth, true }
		if health <= 0 {
			if respawnAt.Valid && now.Before(respawnAt.Time) { _ = tx.Rollback(); return nil }
			health, respawnAt = enemy.MaxHealth, sql.NullTime{}
			state = game.EnemyAIState{Behavior: game.EnemyIdle, Position: enemy.Position}
		}
		encounterID, err := ensureEncounterID(ctx, tx, enemyID, resetEncounter)
		if err != nil { _ = tx.Rollback(); return err }
		if state.TargetCharacterID != "" && state.TargetCharacterID != string(candidate.characterID) &&
			(state.Behavior == game.EnemyChasing || state.Behavior == game.EnemyTelegraphing) {
			var stickyTargetStillActive bool
			err := tx.QueryRowContext(ctx, `SELECT EXISTS(
				SELECT 1 FROM zone_presence p
				JOIN players target_player ON target_player.id=p.character_id AND target_player.account_id=p.account_id
				WHERE p.character_id=$1 AND p.zone_id=$2 AND p.last_seen_at>$3
				AND p.position_x BETWEEN $4 AND $5 AND p.position_z BETWEEN $6 AND $7
				AND target_player.health>0)`,
				state.TargetCharacterID, state.Position.ZoneID, now.Add(-game.PresenceLease),
				state.Position.X-enemyAIConfig(enemy).LeashRange, state.Position.X+enemyAIConfig(enemy).LeashRange,
				state.Position.Z-enemyAIConfig(enemy).LeashRange, state.Position.Z+enemyAIConfig(enemy).LeashRange,
			).Scan(&stickyTargetStillActive)
			if err != nil { _ = tx.Rollback(); return fmt.Errorf("revalidate sticky enemy target: %w", err) }
			if stickyTargetStillActive { _ = tx.Rollback(); return nil }
		}
		result := game.AdvanceEnemyAI(enemy, aiConfig, state, player.Position, now)
		if (result.State.Behavior == game.EnemyChasing || result.State.Behavior == game.EnemyTelegraphing) && player.Health > 0 {
			result.State.TargetCharacterID = string(player.ID)
		} else {
			result.State.TargetCharacterID = ""
		}
		if isBoss {
			result.State.TelegraphUntil = time.Time{}
			result.State.NextAttackAt = time.Time{}
			if err := advanceBossEncounterInTransaction(ctx, tx, enemy, bossDefinition, encounterID,
				health, &result.State, &player, candidate.accountID, now); err != nil {
				_ = tx.Rollback()
				return err
			}
		} else if result.AttackRequested {
			attackResult, err := player.ReceiveEnemyAttack(enemy, result.State.Position, now)
			if err != nil {
				_ = tx.Rollback()
				if errors.Is(err, game.ErrEnemyOutOfRange) || errors.Is(err, game.ErrPlayerDefeated) { return nil }
				return fmt.Errorf("resolve telegraphed enemy attack: %w", err)
			}
			if player.Health <= 0 { result.State.TargetCharacterID = "" }
			if err := recordEncounterDamageReceived(ctx, tx, encounterID, enemyID, player.ID, attackResult.DamageReceived, now); err != nil {
				_ = tx.Rollback(); return err
			}
			if err := updateOwnedPlayer(ctx, tx, candidate.accountID, player); err != nil {
				_ = tx.Rollback(); return fmt.Errorf("persist enemy attack target: %w", err)
			}
		}
		if err := persistEnemyAIState(ctx, tx, enemyID, health, initialized, respawnAt, encounterID, result.State); err != nil {
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
	resetEncounter := !initialized || health <= 0
	if !initialized { health, initialized = enemy.MaxHealth, true }
	if health <= 0 {
		if respawnAt.Valid && now.Before(respawnAt.Time) { return nil }
		health, respawnAt = enemy.MaxHealth, sql.NullTime{}
		state = game.EnemyAIState{Behavior: game.EnemyIdle, Position: enemy.Position}
	}
	encounterID, err := ensureEncounterID(ctx, tx, enemyID, resetEncounter)
	if err != nil { return err }
	result := game.AdvanceEnemyAI(enemy, enemyAIConfig(enemy), state, game.WorldPosition{}, now)
	result.State.TargetCharacterID = ""
	if err := persistEnemyAIState(ctx, tx, enemyID, health, initialized, respawnAt, encounterID, result.State); err != nil { return err }
	if err := tx.Commit(); err != nil { return fmt.Errorf("commit idle enemy tick: %w", err) }
	return nil
}

func readEnemyAIStateForUpdate(ctx context.Context, tx *sql.Tx, enemy game.EnemyDefinition) (int, bool, sql.NullTime, game.EnemyAIState, error) {
	var health int
	var initialized bool
	var respawnAt sql.NullTime
	var behavior string
	var zoneID, targetCharacterID sql.NullString
	var x, z sql.NullFloat64
	var telegraphUntil, nextAttackAt, lastUpdatedAt sql.NullTime
	err := tx.QueryRowContext(ctx, `SELECT health, initialized, respawn_at, behavior,
		position_zone_id, position_x, position_z, target_character_id,
		telegraph_until, next_attack_at, last_ai_updated_at
		FROM enemy_spawn_state WHERE enemy_id=$1 FOR UPDATE`, enemy.ID).Scan(
		&health, &initialized, &respawnAt, &behavior, &zoneID, &x, &z,
		&targetCharacterID, &telegraphUntil, &nextAttackAt, &lastUpdatedAt,
	)
	if err != nil { return 0, false, sql.NullTime{}, game.EnemyAIState{}, fmt.Errorf("lock enemy AI state: %w", err) }
	state := game.EnemyAIState{Behavior: game.EnemyBehaviorState(behavior), Position: enemy.Position}
	if zoneID.Valid && x.Valid && z.Valid { state.Position = game.WorldPosition{ZoneID: zoneID.String, X: x.Float64, Z: z.Float64} }
	if targetCharacterID.Valid { state.TargetCharacterID = targetCharacterID.String }
	if telegraphUntil.Valid { state.TelegraphUntil = telegraphUntil.Time.UTC() }
	if nextAttackAt.Valid { state.NextAttackAt = nextAttackAt.Time.UTC() }
	if lastUpdatedAt.Valid { state.LastUpdatedAt = lastUpdatedAt.Time.UTC() }
	return health, initialized, respawnAt, state, nil
}

func nullableString(value string) any {
	if value == "" { return nil }
	return value
}

func persistEnemyAIState(ctx context.Context, tx *sql.Tx, enemyID string, health int, initialized bool, respawnAt sql.NullTime, encounterID string, state game.EnemyAIState) error {
	var respawnValue any
	if respawnAt.Valid { respawnValue = respawnAt.Time.UTC() }
	_, err := tx.ExecContext(ctx, `UPDATE enemy_spawn_state SET health=$2, initialized=$3,
		behavior=$4, position_zone_id=$5, position_x=$6, position_z=$7,
		telegraph_until=$8, next_attack_at=$9, last_ai_updated_at=$10,
		target_character_id=$11, respawn_at=$12, encounter_id=$13 WHERE enemy_id=$1`,
		enemyID, health, initialized, string(state.Behavior), state.Position.ZoneID,
		state.Position.X, state.Position.Z, nullableTime(state.TelegraphUntil),
		nullableTime(state.NextAttackAt), state.LastUpdatedAt.UTC(), nullableString(state.TargetCharacterID), respawnValue, encounterID)
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

 
// ListEnemySnapshotsForAccount returns authoritative enemy instances in the
// requesting character's current zone. Character ownership is checked in SQL.
func (r *AccountRepository) ListEnemySnapshotsForAccount(ctx context.Context, accountID string, characterID game.PlayerID, now time.Time) ([]game.EnemySnapshot, error) {
	var zoneID string
	err := r.db.QueryRowContext(ctx, `SELECT zone_id FROM players WHERE account_id=$1 AND id=$2`, accountID, string(characterID)).Scan(&zoneID)
	if errors.Is(err, sql.ErrNoRows) { return nil, game.ErrPlayerNotFound }
	if err != nil { return nil, fmt.Errorf("resolve enemy snapshot zone: %w", err) }
	snapshots := make([]game.EnemySnapshot, 0)
	for _, enemy := range game.EnemyDefinitions() {
		if enemy.Position.ZoneID != zoneID { continue }
		snapshot := game.EnemySnapshot{
			EnemyID: enemy.ID, Name: enemy.Name, ZoneID: enemy.Position.ZoneID,
			X: enemy.Position.X, Z: enemy.Position.Z, Health: enemy.MaxHealth,
			MaxHealth: enemy.MaxHealth, Behavior: game.EnemyIdle,
		}
		var health int
		var initialized bool
		var behavior string
		var storedZone sql.NullString
		var x, z sql.NullFloat64
		var telegraphUntil, respawnAt, lastUpdatedAt sql.NullTime
		var bossID, bossPhaseID, bossAbilityID sql.NullString
		var bossPhaseStartedAt, bossTelegraphStartedAt, bossTelegraphUntil sql.NullTime
		var bossDefinitionVersion sql.NullInt64
		err := r.db.QueryRowContext(ctx, `SELECT s.health, s.initialized, s.behavior, s.position_zone_id,
			s.position_x, s.position_z, s.telegraph_until, s.respawn_at, s.last_ai_updated_at,
			b.definition_id, b.definition_version, b.phase_id, b.phase_started_at,
			b.active_ability_id, b.telegraph_started_at, b.telegraph_until
			FROM enemy_spawn_state s
			LEFT JOIN boss_encounter_states b
				ON b.enemy_id=s.enemy_id AND b.encounter_id=s.encounter_id
			WHERE s.enemy_id=$1`, enemy.ID).Scan(
			&health, &initialized, &behavior, &storedZone, &x, &z,
			&telegraphUntil, &respawnAt, &lastUpdatedAt,
			&bossID, &bossDefinitionVersion, &bossPhaseID, &bossPhaseStartedAt,
			&bossAbilityID, &bossTelegraphStartedAt, &bossTelegraphUntil)
		if err != nil && !errors.Is(err, sql.ErrNoRows) { return nil, fmt.Errorf("load enemy snapshot %s: %w", enemy.ID, err) }
		if err == nil {
			if initialized { snapshot.Health = health }
			if behavior != "" { snapshot.Behavior = game.EnemyBehaviorState(behavior) }
			if storedZone.Valid && x.Valid && z.Valid { snapshot.ZoneID, snapshot.X, snapshot.Z = storedZone.String, x.Float64, z.Float64 }
			if telegraphUntil.Valid { snapshot.TelegraphUntil = telegraphUntil.Time.UTC() }
			if respawnAt.Valid { snapshot.RespawnAt = respawnAt.Time.UTC() }
			if lastUpdatedAt.Valid { snapshot.LastUpdatedAt = lastUpdatedAt.Time.UTC() }
			if bossID.Valid {
				snapshot.BossID = bossID.String
				if bossDefinitionVersion.Valid { snapshot.BossDefinitionVersion = int(bossDefinitionVersion.Int64) }
				if bossPhaseID.Valid { snapshot.BossPhaseID = bossPhaseID.String }
				if bossPhaseStartedAt.Valid { value := bossPhaseStartedAt.Time.UTC(); snapshot.BossPhaseStartedAt = &value }
				if bossAbilityID.Valid { snapshot.BossActiveAbilityID = bossAbilityID.String }
				if bossTelegraphStartedAt.Valid { value := bossTelegraphStartedAt.Time.UTC(); snapshot.BossTelegraphStartedAt = &value }
				if bossTelegraphUntil.Valid { value := bossTelegraphUntil.Time.UTC(); snapshot.BossTelegraphUntil = &value }
			}
			if snapshot.Health <= 0 && snapshot.RespawnAt.IsZero() { snapshot.RespawnAt = now.UTC() }
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, nil
}

 
// ListEncounterParticipants returns the server-recorded contribution ledger for
// one spawn life. Reward settlement must consume this persisted ledger rather
// than accepting client-supplied contribution values.
func (r *AccountRepository) ListEncounterParticipants(ctx context.Context, enemyID, encounterID string) ([]game.EncounterParticipant, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT encounter_id,enemy_id,account_id,character_id,
		damage_dealt,experience_skill_id,damage_received,action_count,first_contribution_at,last_contribution_at
		FROM enemy_encounter_participants WHERE enemy_id=$1 AND encounter_id=$2
		ORDER BY damage_dealt DESC, first_contribution_at, character_id`, enemyID, encounterID)
	if err != nil { return nil, fmt.Errorf("list encounter participants: %w", err) }
	defer rows.Close()
	participants := make([]game.EncounterParticipant, 0)
	for rows.Next() {
		var participant game.EncounterParticipant
		if err := rows.Scan(&participant.EncounterID,&participant.EnemyID,&participant.AccountID,
			&participant.CharacterID,&participant.DamageDealt,&participant.ExperienceSkillID,&participant.DamageReceived,
			&participant.ActionCount,&participant.FirstContributionAt,&participant.LastContributionAt); err != nil {
			return nil, fmt.Errorf("scan encounter participant: %w", err)
		}
		participants = append(participants, participant)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate encounter participants: %w", err) }
	return participants, nil
}
