package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresEnemyWorldTickCommitsAttackExactlyOnce(t *testing.T) {
	databaseURL := os.Getenv("RS_GALAXIES_TEST_DATABASE_URL")
	if databaseURL == "" { t.Skip("RS_GALAXIES_TEST_DATABASE_URL is not configured") }
	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil { t.Fatal(err) }
	defer db.Close()
	db.SetMaxOpenConns(10)
	if err := db.PingContext(ctx); err != nil { t.Fatalf("connect test database: %v", err) }
	if err := Migrate(ctx, db); err != nil { t.Fatalf("migrate test database: %v", err) }

	now := time.Now().UTC().Truncate(time.Second)
	accountRepo := NewAccountRepository(db)
	email := fmt.Sprintf("enemy-ai-%d@example.test", now.UnixNano())
	account, _, err := accountRepo.Register(ctx, email, "StrongTestPassword!2026", now)
	if err != nil { t.Fatalf("register test account: %v", err) }
	playerID := game.PlayerID(fmt.Sprintf("enemy-ai-player-%d", now.UnixNano()))
	player, err := game.NewPlayer(playerID, "EnemyTickTester", now)
	if err != nil { t.Fatal(err) }
	if err := accountRepo.CreateCharacter(ctx, account.ID, player); err != nil { t.Fatalf("create test character: %v", err) }
	const enemyID = "frontier-rat-001"
	if _, err := db.ExecContext(ctx, "DELETE FROM enemy_spawn_state WHERE enemy_id=$1", enemyID); err != nil { t.Fatal(err) }
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM zone_presence WHERE character_id=$1", string(playerID))
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_spawn_state WHERE enemy_id=$1", enemyID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM players WHERE id=$1", string(playerID))
		_, _ = db.ExecContext(context.Background(), "DELETE FROM accounts WHERE id=$1", account.ID)
	})
	if err := accountRepo.HeartbeatPresenceForAccount(ctx, account.ID, playerID, now); err != nil { t.Fatalf("heartbeat test character: %v", err) }

	for _, offset := range []time.Duration{0, time.Second, 2*time.Second, 3*time.Second} {
		if err := accountRepo.TickEnemyAI(ctx, enemyID, now.Add(offset)); err != nil {
			t.Fatalf("advance enemy at +%s: %v", offset, err)
		}
	}

	// Two world workers may reach the same deadline concurrently. The player
	// then enemy row-lock order must serialize them and allow only one hit.
	var wait sync.WaitGroup
	errors := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			errors <- accountRepo.TickEnemyAI(ctx, enemyID, now.Add(4*time.Second))
		}()
	}
	wait.Wait()
	close(errors)
	for err := range errors { if err != nil { t.Fatalf("concurrent enemy tick: %v", err) } }

	updated, err := accountRepo.GetCharacter(ctx, account.ID, playerID)
	if err != nil { t.Fatal(err) }
	if updated.Health != game.MaxPlayerHealth-2 {
		t.Fatalf("player health = %d, want %d (exactly one enemy hit)", updated.Health, game.MaxPlayerHealth-2)
	}
	if updated.Skills["defense"].Experience != 1 {
		t.Fatalf("Defense XP = %d, want 1 (exactly one award)", updated.Skills["defense"].Experience)
	}

	// Player-initiated combat must also use the moving instance position,
	// rather than the static catalog spawn coordinates.
	_, combat, err := accountRepo.AttackEnemyForAccount(ctx, account.ID, playerID, enemyID, now.Add(6*time.Second))
	if err != nil { t.Fatalf("attack moving enemy at persisted position: %v", err) }
	if combat.DamageDealt <= 0 || combat.EnemyHealth != 10 {
		t.Fatalf("attack against moving enemy result = %+v", combat)
	}

	snapshots, err := accountRepo.ListEnemySnapshotsForAccount(ctx, account.ID, playerID, now.Add(4*time.Second))
	if err != nil { t.Fatalf("list enemy snapshots: %v", err) }
	if len(snapshots) != 2 { t.Fatalf("snapshot count = %d, want two zone enemies", len(snapshots)) }
	found := false
	for _, snapshot := range snapshots {
		if snapshot.EnemyID == enemyID {
			found = true
			if snapshot.Behavior != game.EnemyChasing || snapshot.X != 0 || snapshot.Health != 10 {
				t.Fatalf("unexpected rat snapshot after telegraph: %+v", snapshot)
			}
		}
	}
	if !found { t.Fatal("frontier rat missing from zone enemy snapshots") }

	// Once the presence lease expires, the worker must stop attacking and
	// return the enemy to its catalogued spawn instead of chasing stale data.
	if err := accountRepo.TickEnemyAI(ctx, enemyID, now.Add(30*time.Second)); err != nil {
		t.Fatalf("tick after target presence expires: %v", err)
	}
	if err := accountRepo.TickEnemyAI(ctx, enemyID, now.Add(31*time.Second)); err != nil {
		t.Fatalf("complete enemy leash return: %v", err)
	}
	snapshots, err = accountRepo.ListEnemySnapshotsForAccount(ctx, account.ID, playerID, now.Add(31*time.Second))
	if err != nil { t.Fatalf("list snapshots after leash return: %v", err) }
	for _, snapshot := range snapshots {
		if snapshot.EnemyID == enemyID {
			if snapshot.Behavior != game.EnemyIdle || snapshot.X != 8 || snapshot.Z != 0 {
				t.Fatalf("enemy did not return to spawn after stale target: %+v", snapshot)
			}
			return
		}
	}
	t.Fatal("frontier rat missing after leash return")
}
