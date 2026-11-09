package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
)

func TestPostgresEnemyDefeatCreatesOneDurableRewardClaim(t *testing.T) {
	db := integrationDatabase(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	repo := NewAccountRepository(db)
	account, _, err := repo.Register(ctx, fmt.Sprintf("enemy-settlement-%d@example.test", now.UnixNano()), "correct-horse-battery-staple", now)
	if err != nil { t.Fatal(err) }
	playerID := game.PlayerID(fmt.Sprintf("enemy-settlement-player-%d", now.UnixNano()))
	player, err := game.NewPlayer(playerID, "SettlementHero", now)
	if err != nil { t.Fatal(err) }
	player.Position.X = 8
	if err := repo.CreateCharacter(ctx, account.ID, player); err != nil { t.Fatal(err) }

	const enemyID = "frontier-rat-001"
	_, _ = db.ExecContext(ctx, "DELETE FROM enemy_encounter_reward_claims WHERE enemy_id=$1", enemyID)
	_, _ = db.ExecContext(ctx, "DELETE FROM enemy_encounter_settlements WHERE enemy_id=$1", enemyID)
	_, _ = db.ExecContext(ctx, "DELETE FROM enemy_encounter_participants WHERE enemy_id=$1", enemyID)
	_, _ = db.ExecContext(ctx, "DELETE FROM enemy_spawn_state WHERE enemy_id=$1", enemyID)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_encounter_reward_claims WHERE enemy_id=$1", enemyID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_encounter_settlements WHERE enemy_id=$1", enemyID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_encounter_participants WHERE enemy_id=$1", enemyID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_spawn_state WHERE enemy_id=$1", enemyID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM players WHERE id=$1", string(playerID))
		_, _ = db.ExecContext(context.Background(), "DELETE FROM accounts WHERE id=$1", account.ID)
	})

	var final game.CombatResult
	for hit := 0; hit < 6; hit++ {
		_, result, err := repo.AttackEnemyForAccount(ctx, account.ID, playerID, enemyID, now.Add(time.Duration(hit+2)*time.Second))
		if err != nil { t.Fatalf("attack %d: %v", hit+1, err) }
		final = result
	}
	if !final.EnemyDefeated { t.Fatalf("sixth attack did not defeat enemy: %+v", final) }
	if final.Experience != 0 || final.LootItemID != "" {
		t.Fatalf("combat response granted direct kill rewards instead of claims: %+v", final)
	}

	var settlements, claims int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM enemy_encounter_settlements WHERE enemy_id=$1", enemyID).Scan(&settlements); err != nil { t.Fatal(err) }
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM enemy_encounter_reward_claims WHERE enemy_id=$1 AND character_id=$2", enemyID, string(playerID)).Scan(&claims); err != nil { t.Fatal(err) }
	if settlements != 1 || claims != 1 { t.Fatalf("settlements=%d claims=%d, want one each", settlements, claims) }

	var status, skillID, parentSkillID string
	var xp, parentXP, lootQuantity int
	var lootItemID sql.NullString
	err = db.QueryRowContext(ctx, `SELECT status,experience_skill_id,experience_amount,
		parent_skill_id,parent_experience_amount,loot_item_id,loot_quantity
		FROM enemy_encounter_reward_claims WHERE enemy_id=$1 AND character_id=$2`,
		enemyID, string(playerID)).Scan(&status,&skillID,&xp,&parentSkillID,&parentXP,&lootItemID,&lootQuantity)
	if err != nil { t.Fatal(err) }
	if status != "pending" || skillID != "unarmed" || xp != 25 ||
		parentSkillID != "melee" || parentXP != 6 ||
		!lootItemID.Valid || lootItemID.String != "rat_pelt" || lootQuantity != 1 {
		t.Fatalf("durable reward claim = status:%s skill:%s xp:%d parent:%s/%d loot:%v/%d",
			status,skillID,xp,parentSkillID,parentXP,lootItemID,lootQuantity)
	}
	var playerSkills string
	if err := db.QueryRowContext(ctx, "SELECT skills::text FROM players WHERE id=$1", string(playerID)).Scan(&playerSkills); err != nil { t.Fatal(err) }
	if playerSkills == "" { t.Fatal("player skill state unexpectedly empty") }
}

func TestPostgresEncounterRewardClaimIsAtomicAndIdempotent(t *testing.T) {
	db := integrationDatabase(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	repo := NewAccountRepository(db)
	account, _, err := repo.Register(ctx, fmt.Sprintf("reward-claim-flow-%d@example.test", now.UnixNano()), "correct-horse-battery-staple", now)
	if err != nil { t.Fatal(err) }
	playerID := game.PlayerID(fmt.Sprintf("reward-claim-flow-player-%d", now.UnixNano()))
	player, err := game.NewPlayer(playerID, "ClaimWayfarer", now)
	if err != nil { t.Fatal(err) }
	player.Inventory["rat_pelt"] = 100
	if err := repo.CreateCharacter(ctx, account.ID, player); err != nil { t.Fatal(err) }

	const enemyID = "reward-claim-flow-enemy"
	const encounterID = "reward-claim-flow-cycle"
	if _, err := db.ExecContext(ctx, `INSERT INTO enemy_encounter_settlements(enemy_id,encounter_id,settled_at)
		VALUES($1,$2,$3)`, enemyID, encounterID, now); err != nil { t.Fatal(err) }
	if _, err := db.ExecContext(ctx, `INSERT INTO enemy_encounter_reward_claims(
		enemy_id,encounter_id,account_id,character_id,experience_skill_id,experience_amount,
		parent_skill_id,parent_experience_amount,loot_item_id,loot_quantity,status,created_at)
		VALUES($1,$2,$3,$4,'blades',25,'melee',6,'rat_pelt',1,'pending',$5)`,
		enemyID,encounterID,account.ID,string(playerID),now); err != nil { t.Fatal(err) }
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_encounter_reward_claims WHERE enemy_id=$1", enemyID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_encounter_settlements WHERE enemy_id=$1", enemyID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM players WHERE id=$1", string(playerID))
		_, _ = db.ExecContext(context.Background(), "DELETE FROM accounts WHERE id=$1", account.ID)
	})

	if _, _, err := repo.ClaimEncounterRewardForAccount(ctx, account.ID, playerID, enemyID, encounterID, now); !errors.Is(err, game.ErrRewardInventoryFull) {
		t.Fatalf("full inventory claim error=%v, want ErrRewardInventoryFull", err)
	}
	var status string
	if err := db.QueryRowContext(ctx, "SELECT status FROM enemy_encounter_reward_claims WHERE enemy_id=$1 AND character_id=$2", enemyID,string(playerID)).Scan(&status); err != nil { t.Fatal(err) }
	if status != "pending" { t.Fatalf("full-inventory claim status=%q, want pending", status) }
	unchanged, err := repo.GetCharacter(ctx, account.ID, playerID)
	if err != nil { t.Fatal(err) }
	if unchanged.Skills["blades"].Experience != 0 || unchanged.Skills["melee"].Experience != 0 {
		t.Fatalf("full inventory consumed XP: blades=%d melee=%d", unchanged.Skills["blades"].Experience, unchanged.Skills["melee"].Experience)
	}

	tx, err := db.BeginTx(ctx,nil)
	if err != nil { t.Fatal(err) }
	locked, err := repo.lockOwnedPlayer(ctx,tx,account.ID,playerID)
	if err != nil { t.Fatal(err) }
	locked.Inventory["rat_pelt"] = 99
	if err := updateOwnedPlayer(ctx,tx,account.ID,locked); err != nil { t.Fatal(err) }
	if err := tx.Commit(); err != nil { t.Fatal(err) }

	claimedPlayer, claim, err := repo.ClaimEncounterRewardForAccount(ctx,account.ID,playerID,enemyID,encounterID,now.Add(time.Second))
	if err != nil { t.Fatalf("claim after freeing stack space: %v",err) }
	if claim.Status != "claimed" || claimedPlayer.Inventory["rat_pelt"] != 100 ||
		claimedPlayer.Skills["blades"].Experience != 25 || claimedPlayer.Skills["melee"].Experience != 6 {
		t.Fatalf("claimed reward state: claim=%+v inventory=%d blades=%d melee=%d",
			claim,claimedPlayer.Inventory["rat_pelt"],claimedPlayer.Skills["blades"].Experience,claimedPlayer.Skills["melee"].Experience)
	}
	replayedPlayer, replayedClaim, err := repo.ClaimEncounterRewardForAccount(ctx,account.ID,playerID,enemyID,encounterID,now.Add(2*time.Second))
	if err != nil { t.Fatalf("idempotent claim retry: %v",err) }
	if replayedClaim.Status != "claimed" || replayedPlayer.Inventory["rat_pelt"] != 100 ||
		replayedPlayer.Skills["blades"].Experience != 25 || replayedPlayer.Skills["melee"].Experience != 6 {
		t.Fatalf("claim retry duplicated reward: claim=%+v inventory=%d blades=%d melee=%d",
			replayedClaim,replayedPlayer.Inventory["rat_pelt"],replayedPlayer.Skills["blades"].Experience,replayedPlayer.Skills["melee"].Experience)
	}
}
