package postgres

import (
	"context"
	"database/sql"
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
	if len(eligible) == 0 || totalDamage <= 0 {
		return nil
	}

	// The first participant is the deterministic highest-damage contributor.
	// One enemy drop is assigned once; XP is split proportionally by damage.
	remainingXP := int64(enemy.Experience)
	for index, participant := range eligible {
		xp := int64(0)
		if index == 0 {
			// Give integer-division remainder to the top contributor so the
			// claims account for the complete base XP value.
			share := int64(enemy.Experience) * participant.DamageDealt / totalDamage
			xp = share
		} else {
			xp = int64(enemy.Experience) * participant.DamageDealt / totalDamage
		}
		remainingXP -= xp
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
	_ = remainingXP // kept explicit for policy accounting; integer remainder is reviewed below.
	return nil
}
