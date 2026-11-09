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

func TestPostgresBossWorldTickResolvesAbilityExactlyOnce(t *testing.T) {
	databaseURL := os.Getenv("RS_GALAXIES_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("RS_GALAXIES_TEST_DATABASE_URL is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(10)
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	definition, ok := game.FindBossEncounterForEnemy("stoneback-colossus")
	if !ok {
		t.Fatal("Stoneback Colossus encounter definition missing")
	}
	enemy := game.EnemyDefinition{
		ID: definition.EnemyID,
		Name: "Stoneback Colossus",
		MaxHealth: 100,
		Position: game.WorldPosition{ZoneID: game.DefaultWorldZoneID, X: 20, Z: 0},
		InteractionRange: 3,
		RespawnSeconds: 120,
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	accountRepo := NewAccountRepository(db)
	account, _, err := accountRepo.Register(ctx,
		fmt.Sprintf("boss-world-%d@example.test", now.UnixNano()),
		"StrongTestPassword!2026", now)
	if err != nil {
		t.Fatalf("register boss test account: %v", err)
	}
	playerID := game.PlayerID(fmt.Sprintf("boss-world-player-%d", now.UnixNano()))
	player, err := game.NewPlayer(playerID, "BossWorldTester", now)
	if err != nil {
		t.Fatal(err)
	}
	player.Position = game.WorldPosition{ZoneID: game.DefaultWorldZoneID, X: 22, Z: 0}
	if err := accountRepo.CreateCharacter(ctx, account.ID, player); err != nil {
		t.Fatalf("create boss test character: %v", err)
	}

	if _, err := db.ExecContext(ctx, "DELETE FROM enemy_spawn_state WHERE enemy_id=$1", enemy.ID); err != nil {
		t.Fatalf("clear isolated boss spawn: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM zone_presence WHERE character_id=$1", string(playerID))
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_spawn_state WHERE enemy_id=$1", enemy.ID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM players WHERE id=$1", string(playerID))
		_, _ = db.ExecContext(context.Background(), "DELETE FROM accounts WHERE id=$1", account.ID)
	})
	if err := accountRepo.HeartbeatPresenceForAccount(ctx, account.ID, playerID, now); err != nil {
		t.Fatalf("heartbeat boss test character: %v", err)
	}

	// The first tick must start a durable telegraph without applying damage.
	if err := accountRepo.tickEnemyAIForDefinition(ctx, enemy, now); err != nil {
		t.Fatalf("start boss telegraph: %v", err)
	}
	var activeAbility string
	var telegraphStarted, telegraphUntil time.Time
	if err := db.QueryRowContext(ctx, `SELECT active_ability_id,telegraph_started_at,telegraph_until
		FROM boss_encounter_states WHERE enemy_id=$1`, enemy.ID).Scan(
		&activeAbility, &telegraphStarted, &telegraphUntil); err != nil {
		t.Fatalf("read started boss telegraph: %v", err)
	}
	if activeAbility != "stone-fist" ||
		!telegraphStarted.Equal(now) ||
		!telegraphUntil.Equal(now.Add(1500*time.Millisecond)) {
		t.Fatalf("boss telegraph = %s [%s,%s]", activeAbility, telegraphStarted, telegraphUntil)
	}

	// Concurrent workers reaching the same deadline must resolve the active
	// ability once. The player row is locked before the enemy row in both paths.
	var workers sync.WaitGroup
	errors := make(chan error, 2)
	for i := 0; i < 2; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			errors <- accountRepo.tickEnemyAIForDefinition(ctx, enemy, now.Add(1500*time.Millisecond))
		}()
	}
	workers.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatalf("concurrent boss tick: %v", err)
		}
	}

	updated, err := accountRepo.GetCharacter(ctx, account.ID, playerID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Health != game.MaxPlayerHealth-8 {
		t.Fatalf("player health=%d, want %d (one Stone Fist hit)", updated.Health, game.MaxPlayerHealth-8)
	}
	if updated.Skills["defense"].Experience != 1 {
		t.Fatalf("Defense XP=%d, want 1 for exactly one resolved ability", updated.Skills["defense"].Experience)
	}

	var clearedAbility sql.NullString
	var defeated bool
	if err := db.QueryRowContext(ctx, `SELECT active_ability_id,defeated
		FROM boss_encounter_states WHERE enemy_id=$1`, enemy.ID).Scan(&clearedAbility, &defeated); err != nil {
		t.Fatal(err)
	}
	if clearedAbility.Valid || defeated {
		t.Fatalf("resolved boss state retained an active ability or was defeated: ability=%+v defeated=%v", clearedAbility, defeated)
	}
}
