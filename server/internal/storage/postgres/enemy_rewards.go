package postgres

import (
	"context"
	cryptorand "crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
)


func rollEncounterLootDrops(enemy game.EnemyDefinition) ([]game.ItemQuantity, error) {
	if enemy.LootTableID == "" {
		if enemy.LootItemID == "" || enemy.LootQuantity <= 0 {
			return nil, nil
		}
		return []game.ItemQuantity{{ItemID: enemy.LootItemID, Quantity: enemy.LootQuantity}}, nil
	}

	table, ok := game.FindLootTable(enemy.LootTableID)
	if !ok {
		return nil, fmt.Errorf("resolve encounter loot table %s: %w", enemy.LootTableID, game.ErrUnknownLootTable)
	}
	rollCount := len(table.GuaranteedDrops) + len(table.BonusRolls)
	if len(table.Drops) > 0 {
		rollCount = 1
	}
	if rollCount <= 0 {
		return nil, fmt.Errorf("resolve encounter loot table %s: %w", enemy.LootTableID, game.ErrUnknownLootTable)
	}

	limit := big.NewInt(game.LootRollRange)
	rolls := make([]game.LootRollInput, rollCount)
	for index := range rolls {
		dropRoll, err := cryptorand.Int(cryptorand.Reader, limit)
		if err != nil {
			return nil, fmt.Errorf("generate encounter loot roll %d: %w", index, err)
		}
		quantityRoll, err := cryptorand.Int(cryptorand.Reader, limit)
		if err != nil {
			return nil, fmt.Errorf("generate encounter loot quantity roll %d: %w", index, err)
		}
		rolls[index] = game.LootRollInput{
			DropRoll: int(dropRoll.Int64()),
			QuantityRoll: int(quantityRoll.Int64()),
		}
	}

	drops, err := game.RollLootTableDrops(enemy.LootTableID, rolls)
	if err != nil {
		return nil, fmt.Errorf("resolve encounter loot table %s: %w", enemy.LootTableID, err)
	}
	return drops, nil
}

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
	// Every server-rolled item line is assigned once; XP is split by damage.
	lootDrops, err := rollEncounterLootDrops(enemy)
	if err != nil {
		return err
	}
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
		_, err := tx.ExecContext(ctx, `INSERT INTO enemy_encounter_reward_claims(
			enemy_id,encounter_id,account_id,character_id,experience_skill_id,
			experience_amount,parent_skill_id,parent_experience_amount,
			loot_item_id,loot_quantity,status,created_at)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,NULL,0,'pending',$9)
			ON CONFLICT(enemy_id,encounter_id,character_id) DO NOTHING`,
			enemy.ID, encounterID, participant.AccountID, string(participant.CharacterID),
			participant.ExperienceSkillID, int(xp), nullableString(parentSkillID),
			parentXP, now.UTC())
		if err != nil {
			return fmt.Errorf("create encounter reward claim for %s: %w", participant.CharacterID, err)
		}
		if index == 0 {
			for _, drop := range lootDrops {
				_, err := tx.ExecContext(ctx, `INSERT INTO enemy_encounter_reward_claim_items(
					enemy_id,encounter_id,character_id,item_id,quantity)
					VALUES($1,$2,$3,$4,$5)
					ON CONFLICT(enemy_id,encounter_id,character_id,item_id) DO NOTHING`,
					enemy.ID, encounterID, string(participant.CharacterID), drop.ItemID, drop.Quantity)
				if err != nil {
					return fmt.Errorf("create encounter reward item %s for %s: %w", drop.ItemID, participant.CharacterID, err)
				}
			}
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
	if !owned {
		return nil, game.ErrPlayerNotFound
	}

	rows, err := r.db.QueryContext(ctx, `WITH selected_claims AS (
		SELECT enemy_id,encounter_id,character_id,experience_skill_id,experience_amount,
			parent_skill_id,parent_experience_amount,loot_item_id,loot_quantity,status,created_at,claimed_at
		FROM enemy_encounter_reward_claims
		WHERE account_id=$1 AND character_id=$2
		ORDER BY created_at DESC,enemy_id,encounter_id
		LIMIT 100
	)
	SELECT claims.enemy_id,claims.encounter_id,claims.character_id,
		claims.experience_skill_id,claims.experience_amount,claims.parent_skill_id,
		claims.parent_experience_amount,claims.status,claims.created_at,claims.claimed_at,
		COALESCE(items.item_id,claims.loot_item_id),
		CASE WHEN items.item_id IS NOT NULL THEN items.quantity ELSE claims.loot_quantity END
	FROM selected_claims AS claims
	LEFT JOIN enemy_encounter_reward_claim_items AS items
		ON items.enemy_id=claims.enemy_id
		AND items.encounter_id=claims.encounter_id
		AND items.character_id=claims.character_id
	ORDER BY claims.created_at DESC,claims.enemy_id,claims.encounter_id,items.item_id`,
		accountID, string(characterID))
	if err != nil {
		return nil, fmt.Errorf("list owned encounter reward claims: %w", err)
	}
	defer rows.Close()

	type claimKey struct {
		enemyID string
		encounterID string
	}
	claims := make([]game.EncounterRewardClaim, 0)
	indexes := make(map[claimKey]int)
	for rows.Next() {
		var claim game.EncounterRewardClaim
		var parentSkillID, itemID sql.NullString
		var claimedAt sql.NullTime
		var itemQuantity sql.NullInt64
		if err := rows.Scan(&claim.EnemyID,&claim.EncounterID,&claim.CharacterID,
			&claim.ExperienceSkillID,&claim.ExperienceAmount,&parentSkillID,
			&claim.ParentExperienceAmount,&claim.Status,&claim.CreatedAt,&claimedAt,
			&itemID,&itemQuantity); err != nil {
			return nil, fmt.Errorf("scan owned encounter reward claim: %w", err)
		}
		key := claimKey{enemyID: claim.EnemyID, encounterID: claim.EncounterID}
		index, exists := indexes[key]
		if !exists {
			if parentSkillID.Valid {
				claim.ParentSkillID = parentSkillID.String
			}
			if claimedAt.Valid {
				claim.ClaimedAt = claimedAt.Time.UTC()
			}
			claim.CreatedAt = claim.CreatedAt.UTC()
			index = len(claims)
			indexes[key] = index
			claims = append(claims, claim)
		}
		if itemID.Valid && itemQuantity.Valid && itemQuantity.Int64 > 0 {
			item := game.ItemQuantity{ItemID: itemID.String, Quantity: int(itemQuantity.Int64)}
			claims[index].LootItems = append(claims[index].LootItems, item)
			if len(claims[index].LootItems) == 1 {
				claims[index].LootItemID = item.ItemID
				claims[index].LootQuantity = item.Quantity
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate owned encounter reward claims: %w", err)
	}
	return claims, nil
}

// ClaimEncounterRewardForAccount applies XP and loot once, under the character
// and claim row locks. A full item stack aborts the transaction, preserving
// both the pending claim and its XP for a later retry.
func (r *AccountRepository) ClaimEncounterRewardForAccount(ctx context.Context, accountID string, characterID game.PlayerID, enemyID, encounterID string, now time.Time) (game.Player, game.EncounterRewardClaim, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("begin encounter reward claim: %w", err)
	}
	defer tx.Rollback()

	player, err := r.lockOwnedPlayer(ctx, tx, accountID, characterID)
	if err != nil {
		return game.Player{}, game.EncounterRewardClaim{}, err
	}

	var claim game.EncounterRewardClaim
	var parentSkillID, legacyLootItemID sql.NullString
	var legacyLootQuantity int
	var claimedAt sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT enemy_id,encounter_id,character_id,
		experience_skill_id,experience_amount,parent_skill_id,parent_experience_amount,
		loot_item_id,loot_quantity,status,created_at,claimed_at
		FROM enemy_encounter_reward_claims
		WHERE account_id=$1 AND character_id=$2 AND enemy_id=$3 AND encounter_id=$4
		FOR UPDATE`, accountID,string(characterID),enemyID,encounterID).Scan(
		&claim.EnemyID,&claim.EncounterID,&claim.CharacterID,
		&claim.ExperienceSkillID,&claim.ExperienceAmount,&parentSkillID,
		&claim.ParentExperienceAmount,&legacyLootItemID,&legacyLootQuantity,
		&claim.Status,&claim.CreatedAt,&claimedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return game.Player{}, game.EncounterRewardClaim{}, game.ErrRewardClaimNotFound
	}
	if err != nil {
		return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("lock owned encounter reward claim: %w", err)
	}
	if parentSkillID.Valid {
		claim.ParentSkillID = parentSkillID.String
	}
	if claimedAt.Valid {
		claim.ClaimedAt = claimedAt.Time.UTC()
	}
	claim.CreatedAt = claim.CreatedAt.UTC()

	itemRows, err := tx.QueryContext(ctx, `SELECT item_id,quantity
		FROM enemy_encounter_reward_claim_items
		WHERE enemy_id=$1 AND encounter_id=$2 AND character_id=$3
		ORDER BY item_id
		FOR UPDATE`, enemyID,encounterID,string(characterID))
	if err != nil {
		return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("lock encounter reward items: %w", err)
	}
	for itemRows.Next() {
		var item game.ItemQuantity
		if err := itemRows.Scan(&item.ItemID,&item.Quantity); err != nil {
			itemRows.Close()
			return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("scan encounter reward item: %w", err)
		}
		claim.LootItems = append(claim.LootItems,item)
	}
	if err := itemRows.Err(); err != nil {
		itemRows.Close()
		return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("iterate encounter reward items: %w", err)
	}
	if err := itemRows.Close(); err != nil {
		return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("close encounter reward items: %w", err)
	}
	// Keep compatibility with claims created by older application versions or
	// manually seeded legacy rows that predate the backfill migration.
	if len(claim.LootItems) == 0 && legacyLootItemID.Valid && legacyLootQuantity > 0 {
		claim.LootItems = append(claim.LootItems,game.ItemQuantity{
			ItemID: legacyLootItemID.String,
			Quantity: legacyLootQuantity,
		})
	}
	if len(claim.LootItems) > 0 {
		claim.LootItemID = claim.LootItems[0].ItemID
		claim.LootQuantity = claim.LootItems[0].Quantity
	}

	if claim.Status == "claimed" {
		if err := tx.Commit(); err != nil {
			return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("commit idempotent reward claim: %w", err)
		}
		return player, claim, nil
	}

	award := func(skillID string, amount int) {
		if amount <= 0 || !game.IsKnownSkill(skillID) {
			return
		}
		player.Skills = game.EnsureSkillProgress(player.Skills)
		progress := player.Skills[skillID]
		if progress.Experience > game.MaxSkillExperience-amount {
			return
		}
		nextXP := progress.Experience + amount
		player.Skills[skillID] = game.SkillProgress{Experience:nextXP,Level:game.LevelForExperience(nextXP)}
	}
	award(claim.ExperienceSkillID,claim.ExperienceAmount)
	award(claim.ParentSkillID,claim.ParentExperienceAmount)

	// Validate every stack before mutating any inventory entry. A single full
	// stack keeps the entire claim (including XP and all other items) pending.
	for _, loot := range claim.LootItems {
		item, ok := game.FindItem(loot.ItemID)
		if !ok || loot.Quantity <= 0 || loot.Quantity > item.MaxStack {
			return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("invalid stored encounter reward item %q", loot.ItemID)
		}
		if player.Inventory[loot.ItemID] < 0 || player.Inventory[loot.ItemID] > item.MaxStack-loot.Quantity {
			return game.Player{}, game.EncounterRewardClaim{}, game.ErrRewardInventoryFull
		}
	}
	if len(claim.LootItems) > 0 && player.Inventory == nil {
		player.Inventory = make(map[string]int)
	}
	for _, loot := range claim.LootItems {
		player.Inventory[loot.ItemID] += loot.Quantity
	}

	if err := updateOwnedPlayer(ctx, tx, accountID, player); err != nil {
		return game.Player{}, game.EncounterRewardClaim{}, err
	}
	claimedAtValue := now.UTC()
	result, err := tx.ExecContext(ctx, `UPDATE enemy_encounter_reward_claims
		SET status='claimed',claimed_at=$5
		WHERE account_id=$1 AND character_id=$2 AND enemy_id=$3 AND encounter_id=$4 AND status='pending'`,
		accountID,string(characterID),enemyID,encounterID,claimedAtValue)
	if err != nil {
		return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("mark encounter reward claimed: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("inspect claimed reward update: %w", err)
	}
	if affected != 1 {
		return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("claim state changed while locked")
	}
	claim.Status = "claimed"
	claim.ClaimedAt = claimedAtValue
	if err := tx.Commit(); err != nil {
		return game.Player{}, game.EncounterRewardClaim{}, fmt.Errorf("commit encounter reward claim: %w", err)
	}
	return player, claim, nil
}
