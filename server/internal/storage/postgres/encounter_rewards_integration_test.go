package postgres

import (
	"context"
	"errors"
	"fmt"
	"database/sql"
	"testing"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestPostgresEncounterRewardClaimsAreDurableAndIdempotencyScoped(t *testing.T) {
	db := integrationDatabase(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	repo := NewAccountRepository(db)
	account, _, err := repo.Register(ctx, fmt.Sprintf("reward-claims-%d@example.test", now.UnixNano()), "correct-horse-battery-staple", now)
	if err != nil { t.Fatal(err) }
	playerID := game.PlayerID(fmt.Sprintf("reward-claim-player-%d", now.UnixNano()))
	player, err := game.NewPlayer(playerID, "RewardClaim", now)
	if err != nil { t.Fatal(err) }
	if err := repo.CreateCharacter(ctx, account.ID, player); err != nil { t.Fatal(err) }
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_encounter_settlements WHERE enemy_id=$1", "reward-claim-test-enemy")
		cleanupTestAccounts(db, account.ID)
	})

	const enemyID = "reward-claim-test-enemy"
	const encounterID = "reward-claim-test-cycle"
	if _, err := db.ExecContext(ctx, `INSERT INTO enemy_encounter_settlements(enemy_id,encounter_id,settled_at,reward_policy_version)
		VALUES($1,$2,$3,1)`, enemyID, encounterID, now); err != nil { t.Fatal(err) }
		if _, err := db.ExecContext(ctx, `INSERT INTO enemy_encounter_reward_claims(
			enemy_id,encounter_id,account_id,character_id,experience_skill_id,experience_amount,
			parent_skill_id,parent_experience_amount,loot_item_id,loot_quantity,status,created_at)
			VALUES($1,$2,$3,$4,'blades',25,'melee',6,'rat_pelt',1,'pending',$5)`,
			enemyID, encounterID, account.ID, string(playerID), now); err != nil { t.Fatal(err) }

	var status string
	var claimedAt sql.NullTime
	if err := db.QueryRowContext(ctx, `SELECT status,claimed_at FROM enemy_encounter_reward_claims
		WHERE enemy_id=$1 AND encounter_id=$2 AND character_id=$3`, enemyID, encounterID, string(playerID)).Scan(&status, &claimedAt); err != nil { t.Fatal(err) }
	if status != "pending" || claimedAt.Valid { t.Fatalf("new reward claim state = %q, claimedAt valid=%v", status, claimedAt.Valid) }

	// Replaying the same settlement key is safe and cannot create a second claim.
	if _, err := db.ExecContext(ctx, `INSERT INTO enemy_encounter_settlements(enemy_id,encounter_id,settled_at,reward_policy_version)
		VALUES($1,$2,$3,1) ON CONFLICT(enemy_id,encounter_id) DO NOTHING`, enemyID, encounterID, now.Add(time.Second)); err != nil { t.Fatal(err) }
	var settlementCount, claimCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM enemy_encounter_settlements WHERE enemy_id=$1 AND encounter_id=$2", enemyID, encounterID).Scan(&settlementCount); err != nil { t.Fatal(err) }
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM enemy_encounter_reward_claims WHERE enemy_id=$1 AND encounter_id=$2", enemyID, encounterID).Scan(&claimCount); err != nil { t.Fatal(err) }
	if settlementCount != 1 || claimCount != 1 { t.Fatalf("settlements=%d claims=%d, want exactly one each", settlementCount, claimCount) }

	_, err = db.ExecContext(ctx, `INSERT INTO enemy_encounter_reward_claims(
		enemy_id,encounter_id,account_id,character_id,experience_skill_id,created_at)
		VALUES($1,$2,$3,$4,'blades',$5)`, enemyID, encounterID, account.ID, string(playerID), now)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" { t.Fatalf("duplicate claim error=%v, want unique violation", err) }

	_, err = db.ExecContext(ctx, `UPDATE enemy_encounter_reward_claims SET status='claimed' WHERE enemy_id=$1 AND encounter_id=$2 AND character_id=$3`, enemyID, encounterID, string(playerID))
	if err == nil { t.Fatal("claimed state without claimed_at should violate invariant") }
}
