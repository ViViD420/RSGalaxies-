package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresEnemyAIPersistsAcrossRepositoryRestart(t *testing.T) {
	databaseURL := os.Getenv("RS_GALAXIES_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("RS_GALAXIES_TEST_DATABASE_URL is not configured")
	}
	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil { t.Fatal(err) }
	defer db.Close()
	if err := db.PingContext(ctx); err != nil { t.Fatalf("connect test database: %v", err) }
	if err := Migrate(ctx, db); err != nil { t.Fatalf("migrate test database: %v", err) }

	const enemyID = "frontier-rat-001"
	if _, err := db.ExecContext(ctx, "DELETE FROM enemy_spawn_state WHERE enemy_id=$1", enemyID); err != nil { t.Fatal(err) }
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_spawn_state WHERE enemy_id=$1", enemyID) })

	start := time.Now().UTC().Truncate(time.Second)
	player := game.WorldPosition{ZoneID: game.DefaultWorldZoneID, X: 8, Z: 0}
	firstRepository := NewPlayerRepository(db)
	first, err := firstRepository.AdvanceEnemyAIForPlayer(ctx, enemyID, player, start)
	if err != nil { t.Fatalf("first AI tick: %v", err) }
	if first.State.Behavior != game.EnemyChasing { t.Fatalf("first behavior = %q, want chasing", first.State.Behavior) }

	// A new repository instance simulates a process restart: behavior and
	// telegraph timing must be restored from PostgreSQL, not process memory.
	restartedRepository := NewPlayerRepository(db)
	second, err := restartedRepository.AdvanceEnemyAIForPlayer(ctx, enemyID, player, start.Add(time.Second))
	if err != nil { t.Fatalf("second AI tick after restart: %v", err) }
	if second.State.Behavior != game.EnemyTelegraphing { t.Fatalf("second behavior = %q, want telegraphing", second.State.Behavior) }
	if second.State.TelegraphUntil.IsZero() { t.Fatal("telegraph deadline was not persisted") }

	third, err := restartedRepository.AdvanceEnemyAIForPlayer(ctx, enemyID, player, start.Add(2*time.Second))
	if err != nil { t.Fatalf("third AI tick: %v", err) }
	if !third.AttackRequested { t.Fatal("completed telegraph did not request an attack") }

	var behavior string
	var x, z float64
	var lastUpdated time.Time
	if err := db.QueryRowContext(ctx, `SELECT behavior, position_x, position_z, last_ai_updated_at
		FROM enemy_spawn_state WHERE enemy_id=$1`, enemyID).Scan(&behavior, &x, &z, &lastUpdated); err != nil {
		t.Fatalf("read persisted enemy state: %v", err)
	}
	if behavior != string(game.EnemyChasing) || x != 8 || z != 0 || !lastUpdated.Equal(start.Add(2*time.Second)) {
		t.Fatalf("persisted AI state = behavior:%s position:(%v,%v) updated:%s", behavior, x, z, lastUpdated)
	}
}
