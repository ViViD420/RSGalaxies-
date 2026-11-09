package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
)

// settleEncounterRewards creates one durable, idempotent settlement and
// character-owned claims inside the caller's enemy-defeat transaction. The
// enemy row is already locked by the combat path, serializing settlement with
// every other attack against this spawn.
func settleEncounterRewards(ctx context.Context, tx *sql.Tx, enemy game.EnemyDefinition, encounterID string, now time.Time) error {
	if encounterID == "" {
		return fmt.Errorf("settle encounter rewards: missing encounter id")
	}
	insert, err := tx.ExecContext(ctx, `INSERT INTO enemy_encounter_settlements(
		enemy_id,encounter_id,settled_at,reward_policy_version)
		VALUES($1,$2,$3,1) ON CONFLICT(enemy_id,encounter_id) DO NOTHING`,
		enemy.ID, encounterID, now.UTC())
	if err != nil {
		return fmt.Errorf("create encounter settlement: %w", err)
	}
	inserted, err := insert.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect encounter settlement insert: %w", err)
	}
	if inserted == 0 {
		return nil
	}

	rows, err := tx.QueryContext(ctx, `SELECT encounter_id,enemy_id,account_id,character_id,
		damage_dealt,experience_skill_id,damage_received,action_count,
		first_contribution_at,last_contribution_at
		FROM enemy_encounter_participants
		WHERE enemy_id=$1 AND encounter_id=$2
		ORDER BY damage_dealt DESC, first_contribution_at, character_id
		FOR UPDATE`, enemy.ID, encounterID)
	if err != nil {
		return fmt.Errorf("load encounter reward participants: %w", err)
	}
	defer rows.Close()

	eligible := make([]game.EncounterParticipant, 0)
	var totalDamage int64
	for rows.Next() {
		var participant game.EncounterParticipant
		if err := rows.Scan(&participant.EncounterID, &participant.EnemyID,
			&participant.AccountID, &participant.CharacterID, &participant.DamageDealt,
			&participant.ExperienceSkillID, &participant.DamageReceived,
			&participant.ActionCount, &participant.FirstContributionAt,
			&participant.LastContributionAt); err != nil {
			return fmt.Errorf("scan encounter reward participant: %w", err)
		}
		if !game.IsEligibleForEncounterReward(participant, enemy, now) {
			continue
		}
		if !game.IsKnownSkill(participant.ExperienceSkillID) {
			participant.ExperienceSkillID = "unarmed"
		}
		eligible = append(eligible, participant)
		totalDamage += participant.DamageDealt
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate encounter reward participants: %w", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close encounter reward participants: %w", err)
	}
	if len(eligible) == 0 || totalDamage <= 0 {
		return nil
	}

	// The first participant is the deterministic highest-damage contributor.
	// One enemy drop is assigned once; XP is split proportionally by damage.
	shares := make([]int64, len(eligible))
		var distributed int64
		for index, participant := range eligible {
			shares[index] = int64(enemy.Experience) * participant.DamageDealt / totalDamage
			distributed += shares[index]
		}
		shares[0] += int64(enemy.Experience) - distributed
	for index, participant := range eligible {
		xp := shares[index]
		parentSkillID := ""
		for _, skill := range game.SkillDefinitions() {
			if skill.ID == participant.ExperienceSkillID {
				parentSkillID = skill.ParentID
				break
			}
		}
		parentXP := 0
		if parentSkillID != "" && xp > 0 {
			parentXP = int(xp / 4)
			if parentXP < 1 { parentXP = 1 }
		}
		lootItemID, lootQuantity := "", 0
		if index == 0 && enemy.LootItemID != "" && enemy.LootQuantity > 0 {
			lootItemID, lootQuantity = enemy.LootItemID, enemy.LootQuantity
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO enemy_encounter_reward_claims(
			enemy_id,encounter_id,account_id,character_id,experience_skill_id,
			experience_amount,parent_skill_id,parent_experience_amount,
			loot_item_id,loot_quantity,status,created_at)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'pending',$11)
			ON CONFLICT(enemy_id,encounter_id,character_id) DO NOTHING`,
			enemy.ID, encounterID, participant.AccountID, string(participant.CharacterID),
			participant.ExperienceSkillID, int(xp), nullableString(parentSkillID),
			parentXP, nullableString(lootItemID), lootQuantity, now.UTC())
		if err != nil {
			return fmt.Errorf("create encounter reward claim for %s: %w", participant.CharacterID, err)
		}
	}
	return nil
}


// ListEncounterRewardClaimsForAccount returns only claims owned by the
// authenticated account and selected character.
func (r *AccountRepository) ListEncounterRewardClaimsForAccount(ctx context.Context, accountID string, characterID game.PlayerID) ([]game.EncounterRewardClaim, error) {
	var owned bool
	if err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM players WHERE account_id=$1 AND id=$2)", accountID, string(characterID)).Scan(&owned); err != nil {
		return nil, fmt.Errorf("verify reward claim character ownership: %w", err)
	}
	if !owned { return nil, game.ErrPlayerNotFound }
	rows, err := r.db.QueryContext(ctx, `SELECT enemy_id,encounter_id,character_id,
		experience_skill_id,experience_amount,parent_skill_id,parent_experience_amount,
		loot_item_id,loot_quantity,status,created_at,claimed_at
		FROM enemy_encounter_reward_claims
		WHERE account_id=$1 AND character_id=$2
		ORDER BY created_at DESC,enemy_id,encounter_id
		LIMIT 100`, accountID, string(characterID))
	if err != nil { return nil, fmt.Errorf("list owned encounter reward claims: %w", err) }
	defer rows.Close()
	claims := make([]game.EncounterRewardClaim, 0)
	for rows.Next() {
		var claim game.EncounterRewardClaim
		var parentSkillID, lootItemID sql.NullString
		var claimedAt sql.NullTime
		if err := rows.Scan(&claim.EnemyID,&claim.EncounterID,&claim.CharacterID,
			&claim.ExperienceSkillID,&claim.ExperienceAmount,&parentSkillID,
			&claim.ParentExperienceAmount,&lootItemID,&claim.LootQuantity,
			&claim.Status,&claim.CreatedAt,&claimedAt); err != nil {
			return nil, fmt.Errorf("scan owned encounter reward claim: %w", err)
		}
		if parentSkillID.Valid { claim.ParentSkillID = parentSkillID.String }
		if lootItemID.Valid { claim.LootItemID = lootItemID.String }
		if claimedAt.Valid { claim.ClaimedAt = claimedAt.Time.UTC() }
		claim.CreatedAt = claim.CreatedAt.UTC()
		claims = append(claims, claim)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate owned encounter reward claims: %w", err) }
	return claims, nil
}

// ClaimEncounterRewardForAccount applies XP and loot once, under the character
// and claim row locks. A full item stack aborts the transaction, preserving
// both the pending claim and its XP for a later retry.
func (r *AccountRepository) ClaimEncounterRewardForAccount(ctx context.Context, accountID string, characterID game.PlayerID, enemyID, encounterID string, now time.Time) (game.Player, game.EncounterRewardClaim, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("begin encounter reward claim: %w", err) }
	defer tx.Rollback()
	player, err := r.lockOwnedPlayer(ctx, tx, accountID, characterID)
	if err != nil { return game.Player{}, game.EncounterRewardClaim{}, err }

	var claim game.EncounterRewardClaim
	var parentSkillID, lootItemID sql.NullString
	var claimedAt sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT enemy_id,encounter_id,character_id,
		experience_skill_id,experience_amount,parent_skill_id,parent_experience_amount,
		loot_item_id,loot_quantity,status,created_at,claimed_at
		FROM enemy_encounter_reward_claims
		WHERE account_id=$1 AND character_id=$2 AND enemy_id=$3 AND encounter_id=$4
		FOR UPDATE`, accountID,string(characterID),enemyID,encounterID).Scan(
		&claim.EnemyID,&claim.EncounterID,&claim.CharacterID,
		&claim.ExperienceSkillID,&claim.ExperienceAmount,&parentSkillID,
		&claim.ParentExperienceAmount,&lootItemID,&claim.LootQuantity,
		&claim.Status,&claim.CreatedAt,&claimedAt)
	if errors.Is(err, sql.ErrNoRows) { return game.Player{}, game.EncounterRewardClaim{}, game.ErrRewardClaimNotFound }
	if err != nil { return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("lock owned encounter reward claim: %w", err) }
	if parentSkillID.Valid { claim.ParentSkillID = parentSkillID.String }
	if lootItemID.Valid { claim.LootItemID = lootItemID.String }
	if claimedAt.Valid { claim.ClaimedAt = claimedAt.Time.UTC() }
	if claim.Status == "claimed" {
		if err := tx.Commit(); err != nil { return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("commit idempotent reward claim: %w", err) }
		return player, claim, nil
	}

	award := func(skillID string, amount int) {
		if amount <= 0 || !game.IsKnownSkill(skillID) { return }
		player.Skills = game.EnsureSkillProgress(player.Skills)
		progress := player.Skills[skillID]
		if progress.Experience > game.MaxSkillExperience-amount { return }
		nextXP := progress.Experience + amount
		player.Skills[skillID] = game.SkillProgress{Experience:nextXP,Level:game.LevelForExperience(nextXP)}
	}
	award(claim.ExperienceSkillID, claim.ExperienceAmount)
	award(claim.ParentSkillID, claim.ParentExperienceAmount)

	if claim.LootItemID != "" && claim.LootQuantity > 0 {
		item, ok := game.FindItem(claim.LootItemID)
		if !ok || claim.LootQuantity > item.MaxStack || player.Inventory[claim.LootItemID] > item.MaxStack-claim.LootQuantity {
			return game.Player{}, game.EncounterRewardClaim{}, game.ErrRewardInventoryFull
		}
		if player.Inventory == nil { player.Inventory = make(map[string]int) }
		player.Inventory[claim.LootItemID] += claim.LootQuantity
	}
	if err := updateOwnedPlayer(ctx, tx, accountID, player); err != nil {
		return game.Player{}, game.EncounterRewardClaim{}, err
	}
	claimedAtValue := now.UTC()
	result, err := tx.ExecContext(ctx, `UPDATE enemy_encounter_reward_claims
		SET status='claimed',claimed_at=$5
		WHERE account_id=$1 AND character_id=$2 AND enemy_id=$3 AND encounter_id=$4 AND status='pending'`,
		accountID,string(characterID),enemyID,encounterID,claimedAtValue)
	if err != nil { return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("mark encounter reward claimed: %w", err) }
	affected, err := result.RowsAffected()
	if err != nil { return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("inspect claimed reward update: %w", err) }
	if affected != 1 { return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("claim state changed while locked") }
	claim.Status = "claimed"
	claim.ClaimedAt = claimedAtValue
	if err := tx.Commit(); err != nil { return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("commit encounter reward claim: %w", err) }
	return player, claim, nil
}
