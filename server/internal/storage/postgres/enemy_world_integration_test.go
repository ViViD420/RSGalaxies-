package postgres

import (
	"context"
	"errors"
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
	tickErrors := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			tickErrors <- accountRepo.TickEnemyAI(ctx, enemyID, now.Add(4*time.Second))
		}()
	}
	wait.Wait()
	close(tickErrors)
	for err := range tickErrors { if err != nil { t.Fatalf("concurrent enemy tick: %v", err) } }

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

	// The immediate retaliation above must reset the autonomous cooldown and
	// cancel any pending telegraph; otherwise the worker could double-strike.
	for _, offset := range []time.Duration{7*time.Second, 8*time.Second} {
		if err := accountRepo.TickEnemyAI(ctx, enemyID, now.Add(offset)); err != nil { t.Fatalf("tick during shared enemy cooldown: %v", err) }
	}
	var healthAfterCooldown int
	if err := db.QueryRowContext(ctx, "SELECT health FROM players WHERE id=$1", string(playerID)).Scan(&healthAfterCooldown); err != nil { t.Fatal(err) }
	if healthAfterCooldown != 96 { t.Fatalf("enemy attacked during shared cooldown: health=%d, want 96", healthAfterCooldown) }
	if err := accountRepo.TickEnemyAI(ctx, enemyID, now.Add(9*time.Second)); err != nil { t.Fatal(err) }
	if err := accountRepo.TickEnemyAI(ctx, enemyID, now.Add(10*time.Second)); err != nil { t.Fatal(err) }
	if err := db.QueryRowContext(ctx, "SELECT health FROM players WHERE id=$1", string(playerID)).Scan(&healthAfterCooldown); err != nil { t.Fatal(err) }
	if healthAfterCooldown != 94 { t.Fatalf("enemy did not resume after shared cooldown/telegraph: health=%d, want 94", healthAfterCooldown) }

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
	found = false
	for _, snapshot := range snapshots {
		if snapshot.EnemyID == enemyID {
			found = true
			if snapshot.Behavior != game.EnemyIdle || snapshot.X != 8 || snapshot.Z != 0 {
				t.Fatalf("enemy did not return to spawn after stale target: %+v", snapshot)
			}
			break
		}
	}
	if !found { t.Fatal("frontier rat missing after leash return") }

	// Even a still-fresh presence must not make a defeated character an AI
	// target. Force a telegraph at the character and verify the tick returns
	// the enemy without applying a second attack.
	if _, err := db.ExecContext(ctx, "UPDATE players SET health=0 WHERE id=$1", string(playerID)); err != nil { t.Fatal(err) }
	if _, err := db.ExecContext(ctx, `UPDATE zone_presence SET zone_id=$2,position_x=0,position_z=0,last_seen_at=$3 WHERE character_id=$1`,
		string(playerID), game.DefaultWorldZoneID, now.Add(31*time.Second)); err != nil { t.Fatal(err) }
	if _, err := db.ExecContext(ctx, `UPDATE enemy_spawn_state SET behavior='telegraphing',position_zone_id=$2,
		position_x=0,position_z=0,telegraph_until=$3,last_ai_updated_at=$4 WHERE enemy_id=$1`,
		enemyID, game.DefaultWorldZoneID, now.Add(31*time.Second), now.Add(31*time.Second)); err != nil { t.Fatal(err) }
	if err := accountRepo.TickEnemyAI(ctx, enemyID, now.Add(32*time.Second)); err != nil {
		t.Fatalf("tick with defeated active presence: %v", err)
	}
	var defeatedHealth int
	if err := db.QueryRowContext(ctx, "SELECT health FROM players WHERE id=$1", string(playerID)).Scan(&defeatedHealth); err != nil { t.Fatal(err) }
	if defeatedHealth != 0 { t.Fatalf("defeated character health changed to %d", defeatedHealth) }
	var behavior string
	if err := db.QueryRowContext(ctx, "SELECT behavior FROM enemy_spawn_state WHERE enemy_id=$1", enemyID).Scan(&behavior); err != nil { t.Fatal(err) }
	if behavior != string(game.EnemyReturning) { t.Fatalf("enemy behavior with defeated target = %q, want returning", behavior) }

	// A fresh presence in another zone is not a valid target for this spawn.
	if _, err := db.ExecContext(ctx, "UPDATE players SET health=100,zone_id='another-zone' WHERE id=$1", string(playerID)); err != nil { t.Fatal(err) }
	if _, err := db.ExecContext(ctx, "UPDATE zone_presence SET zone_id='another-zone',last_seen_at=$2 WHERE character_id=$1", string(playerID), now.Add(32*time.Second)); err != nil { t.Fatal(err) }
	if err := accountRepo.TickEnemyAI(ctx, enemyID, now.Add(33*time.Second)); err != nil {
		t.Fatalf("tick after target changes zone: %v", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT health FROM players WHERE id=$1", string(playerID)).Scan(&defeatedHealth); err != nil { t.Fatal(err) }
	if defeatedHealth != 100 { t.Fatalf("cross-zone target took damage: health=%d", defeatedHealth) }
	if err := db.QueryRowContext(ctx, "SELECT behavior FROM enemy_spawn_state WHERE enemy_id=$1", enemyID).Scan(&behavior); err != nil { t.Fatal(err) }
	if behavior != string(game.EnemyIdle) { t.Fatalf("enemy behavior after zone change = %q, want idle", behavior) }

	// A dead spawn rejects attacks until its respawn deadline, then returns to
	// its catalog spawn with all transient AI timers cleared.
	if _, err := db.ExecContext(ctx, "UPDATE players SET health=100,zone_id=$2,position_x=8,last_combat_at=$3 WHERE id=$1",
		string(playerID), game.DefaultWorldZoneID, now.Add(33*time.Second)); err != nil { t.Fatal(err) }
	if _, err := db.ExecContext(ctx, "UPDATE zone_presence SET zone_id=$2,position_x=8,last_seen_at=$3 WHERE character_id=$1",
		string(playerID), game.DefaultWorldZoneID, now.Add(33*time.Second)); err != nil { t.Fatal(err) }
	if _, err := db.ExecContext(ctx, `UPDATE enemy_spawn_state SET health=0,respawn_at=$2,behavior='returning',
		position_zone_id=$3,position_x=0,position_z=0,telegraph_until=$4,last_ai_updated_at=$4 WHERE enemy_id=$1`,
		enemyID, now.Add(35*time.Second), game.DefaultWorldZoneID, now.Add(34*time.Second)); err != nil { t.Fatal(err) }
	if _, _, err := accountRepo.AttackEnemyForAccount(ctx, account.ID, playerID, enemyID, now.Add(34*time.Second)); !errors.Is(err, game.ErrEnemyRespawning) {
		t.Fatalf("attack before respawn deadline error = %v, want ErrEnemyRespawning", err)
	}
	_, respawnCombat, err := accountRepo.AttackEnemyForAccount(ctx, account.ID, playerID, enemyID, now.Add(36*time.Second))
	if err != nil { t.Fatalf("attack respawned enemy at spawn: %v", err) }
	if respawnCombat.EnemyHealth != 10 { t.Fatalf("respawned enemy health = %d, want 10", respawnCombat.EnemyHealth) }
	var spawnX float64
	var telegraph, lastUpdated, nextAttack sql.NullTime
	if err := db.QueryRowContext(ctx, `SELECT position_x,telegraph_until,last_ai_updated_at,next_attack_at FROM enemy_spawn_state WHERE enemy_id=$1`, enemyID).Scan(&spawnX,&telegraph,&lastUpdated,&nextAttack); err != nil { t.Fatal(err) }
	if spawnX != 8 || telegraph.Valid || !lastUpdated.Valid || !lastUpdated.Time.Equal(now.Add(36*time.Second)) ||
		!nextAttack.Valid || !nextAttack.Time.Equal(now.Add(39*time.Second)) {
		t.Fatalf("respawn did not reset AI position/timers: x=%v telegraph=%v updated=%v nextAttack=%v", spawnX, telegraph, lastUpdated, nextAttack)
	}
}
