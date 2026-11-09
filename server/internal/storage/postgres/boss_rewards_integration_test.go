package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
)

func TestPostgresStonebackDefeatSettlesAndClaimsMultiItemRewards(t *testing.T) {
	db := integrationDatabase(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	repo := NewAccountRepository(db)
	enemy, ok := game.FindEnemy("stoneback-colossus")
	if !ok {
		t.Fatal("Stoneback Colossus is missing from the public enemy catalog")
	}

	account, _, err := repo.Register(ctx,
		fmt.Sprintf("stoneback-reward-%d@example.test", now.UnixNano()),
		"StrongTestPassword!2026", now)
	if err != nil {
		t.Fatal(err)
	}
	playerID := game.PlayerID(fmt.Sprintf("stoneback-reward-player-%d", now.UnixNano()))
	player, err := game.NewPlayer(playerID, "StonebackHunter", now)
	if err != nil {
		t.Fatal(err)
	}
	player.Position = enemy.Position
	player.Inventory["training_sword"] = 1
	if err := player.EquipItem("training_sword"); err != nil {
		t.Fatalf("equip test weapon: %v", err)
	}
	if err := repo.CreateCharacter(ctx, account.ID, player); err != nil {
		t.Fatal(err)
	}

	const encounterID = "stoneback-reward-claim-cycle"
	_, _ = db.ExecContext(ctx, "DELETE FROM enemy_encounter_reward_claims WHERE enemy_id=$1", enemy.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM enemy_encounter_settlements WHERE enemy_id=$1", enemy.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM enemy_encounter_participants WHERE enemy_id=$1", enemy.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM enemy_spawn_state WHERE enemy_id=$1", enemy.ID)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_encounter_reward_claims WHERE enemy_id=$1", enemy.ID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_encounter_settlements WHERE enemy_id=$1", enemy.ID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_encounter_participants WHERE enemy_id=$1", enemy.ID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_spawn_state WHERE enemy_id=$1", enemy.ID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM players WHERE id=$1", string(playerID))
		_, _ = db.ExecContext(context.Background(), "DELETE FROM accounts WHERE id=$1", account.ID)
	})

	if _, err := db.ExecContext(ctx, `INSERT INTO enemy_spawn_state(
		enemy_id,health,initialized,behavior,position_zone_id,position_x,position_z,encounter_id)
		VALUES($1,1,TRUE,'idle',$2,$3,$4,$5)`,
		enemy.ID, enemy.Position.ZoneID, enemy.Position.X, enemy.Position.Z, encounterID); err != nil {
		t.Fatalf("seed Stoneback spawn state: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO enemy_encounter_participants(
		encounter_id,enemy_id,account_id,character_id,damage_dealt,damage_received,
		action_count,first_contribution_at,last_contribution_at,experience_skill_id)
		VALUES($1,$2,$3,$4,10,0,1,$5,$5,'blades')`,
		encounterID, enemy.ID, account.ID, string(playerID), now); err != nil {
		t.Fatalf("seed eligible boss contribution: %v", err)
	}

	updated, result, err := repo.AttackEnemyForAccount(ctx, account.ID, playerID, enemy.ID, now.Add(2*time.Second))
	if err != nil {
		t.Fatalf("defeat Stoneback Colossus: %v", err)
	}
	if !result.EnemyDefeated || result.Experience != 0 || result.LootItemID != "" {
		t.Fatalf("boss combat result bypassed durable rewards: %+v", result)
	}
	if updated.Health <= 0 {
		t.Fatalf("boss ordinary attack channel unexpectedly damaged player: %d", updated.Health)
	}

	var defeated bool
	if err := db.QueryRowContext(ctx, `SELECT defeated FROM boss_encounter_states
		WHERE enemy_id=$1 AND encounter_id=$2`, enemy.ID, encounterID).Scan(&defeated); err != nil {
		t.Fatalf("read terminal boss state: %v", err)
	}
	if !defeated {
		t.Fatal("player-initiated boss defeat did not persist terminal boss state")
	}

	claims, err := repo.ListEncounterRewardClaimsForAccount(ctx, account.ID, playerID)
	if err != nil {
		t.Fatal(err)
	}
	if len(claims) != 1 || claims[0].Status != "pending" {
		t.Fatalf("Stoneback reward claims=%+v, want one pending claim", claims)
	}
	scaleQuantity := 0
	for _, item := range claims[0].LootItems {
		switch item.ItemID {
		case "stoneback_scale":
			scaleQuantity = item.Quantity
		case "stoneback_plate", "stoneback_core":
			if item.Quantity <= 0 {
				t.Fatalf("invalid Stoneback bonus item quantity: %+v", item)
			}
		default:
			t.Fatalf("unexpected Stoneback loot item: %+v", item)
		}
	}
	if scaleQuantity < 2 || scaleQuantity > 4 {
		t.Fatalf("guaranteed Stoneback Scale quantity=%d, want 2..4; items=%+v", scaleQuantity, claims[0].LootItems)
	}

	claimedPlayer, claim, err := repo.ClaimEncounterRewardForAccount(ctx, account.ID, playerID, enemy.ID, encounterID, now.Add(3*time.Second))
	if err != nil {
		t.Fatalf("claim Stoneback multi-item reward: %v", err)
	}
	if claim.Status != "claimed" || len(claim.LootItems) != len(claims[0].LootItems) ||
		claimedPlayer.Inventory["stoneback_scale"] != scaleQuantity ||
		claimedPlayer.Skills["blades"].Experience != 500 ||
		claimedPlayer.Skills["melee"].Experience != 125 {
		t.Fatalf("claimed Stoneback reward state: claim=%+v inventory=%v blades=%d melee=%d",
			claim, claimedPlayer.Inventory, claimedPlayer.Skills["blades"].Experience, claimedPlayer.Skills["melee"].Experience)
	}

	replayedPlayer, replayedClaim, err := repo.ClaimEncounterRewardForAccount(ctx, account.ID, playerID, enemy.ID, encounterID, now.Add(4*time.Second))
	if err != nil {
		t.Fatalf("retry Stoneback reward claim: %v", err)
	}
	if replayedClaim.Status != "claimed" ||
		replayedPlayer.Inventory["stoneback_scale"] != scaleQuantity ||
		replayedPlayer.Skills["blades"].Experience != 500 ||
		replayedPlayer.Skills["melee"].Experience != 125 {
		t.Fatalf("Stoneback reward retry duplicated items or XP: claim=%+v inventory=%v", replayedClaim, replayedPlayer.Inventory)
	}
}
