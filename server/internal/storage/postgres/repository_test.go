package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func integrationDatabase(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("RS_GALAXIES_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("RS_GALAXIES_TEST_DATABASE_URL is not configured")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil { t.Fatalf("ping integration database: %v", err) }
	if err := Migrate(ctx, db); err != nil { t.Fatalf("migrate integration database: %v", err) }
	return db
}

func TestPostgresRepositoryPersistsPlayerAndCraftingTransaction(t *testing.T) {
	db := integrationDatabase(t)
	repository := NewPlayerRepository(db)
	ctx := context.Background()
	id := game.PlayerID(fmt.Sprintf("integration-%d", time.Now().UnixNano()))
	player, err := game.NewPlayer(id, "PersistedWayfarer", time.Now().UTC())
	if err != nil { t.Fatal(err) }
	player.Position.Z = 2
	player.Inventory["softwood_log"] = 2
	player.Bank["rough_stone"] = 4
	if err := repository.Create(ctx, player); err != nil { t.Fatalf("create player: %v", err) }
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM players WHERE id=$1", string(id)) })

	loaded, err := repository.Get(ctx, id)
	if err != nil { t.Fatalf("get player: %v", err) }
	if loaded.Inventory["softwood_log"] != 2 || loaded.Bank["rough_stone"] != 4 || loaded.Position.Z != 2 {
		t.Fatalf("loaded player state mismatch: %+v", loaded)
	}

	craftedPlayer, result, err := repository.Craft(ctx, id, "basic-plank", time.Now())
	if err != nil { t.Fatalf("craft: %v", err) }
	if result.Experience != 30 || craftedPlayer.Inventory["softwood_log"] != 0 || craftedPlayer.Inventory["wooden_plank"] != 1 {
		t.Fatalf("craft result mismatch: result=%+v inventory=%v", result, craftedPlayer.Inventory)
	}

	persisted, err := repository.Get(ctx, id)
	if err != nil { t.Fatalf("reload crafted player: %v", err) }
	if persisted.Skills["crafting"].Experience != 30 || persisted.Inventory["wooden_plank"] != 1 {
		t.Fatalf("craft transaction was not persisted: %+v", persisted)
	}

	// PostgreSQL TIMESTAMPTZ stores microsecond precision; normalize the
	// test value so a reload can be compared without nanosecond truncation.
	moveAt := time.Now().UTC().Add(100 * time.Millisecond).Truncate(time.Microsecond)
	moved, err := repository.MovePlayer(ctx, id, 0.1, 2, moveAt)
	if err != nil { t.Fatalf("move player: %v", err) }
	if moved.Position.X != 0.1 || !moved.LastMoveAt.Equal(moveAt) {
		t.Fatalf("movement state mismatch: position=%+v lastMoveAt=%s", moved.Position, moved.LastMoveAt)
	}
	reloaded, err := repository.Get(ctx, id)
	if err != nil { t.Fatalf("reload moved player: %v", err) }
	if reloaded.Position.X != 0.1 || !reloaded.LastMoveAt.Equal(moveAt) {
		t.Fatalf("movement timestamp/position not persisted: position=%+v lastMoveAt=%s", reloaded.Position, reloaded.LastMoveAt)
	}
}

func TestPostgresRepositoryRollsBackInvalidCraft(t *testing.T) {
	db := integrationDatabase(t)
	repository := NewPlayerRepository(db)
	ctx := context.Background()
	id := game.PlayerID(fmt.Sprintf("rollback-%d", time.Now().UnixNano()))
	player, err := game.NewPlayer(id, "RollbackWayfarer", time.Now().UTC())
	if err != nil { t.Fatal(err) }
	player.Position.Z = 2
	if err := repository.Create(ctx, player); err != nil { t.Fatalf("create player: %v", err) }
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM players WHERE id=$1", string(id)) })

	if _, _, err := repository.Craft(ctx, id, "basic-plank", time.Now()); err == nil {
		t.Fatal("craft without ingredients unexpectedly succeeded")
	}
	persisted, err := repository.Get(ctx, id)
	if err != nil { t.Fatal(err) }
	if persisted.Skills["crafting"].Experience != 0 || persisted.Inventory["wooden_plank"] != 0 {
		t.Fatalf("failed craft changed persisted state: %+v", persisted)
	}
}

func TestPostgresGatheringPersistsNodeAndPlayerAtomically(t *testing.T) {
	db := integrationDatabase(t)
	repository := NewPlayerRepository(db)
	ctx := context.Background()
	id := game.PlayerID(fmt.Sprintf("gather-%d", time.Now().UnixNano()))
	player, err := game.NewPlayer(id, "GatheringWayfarer", time.Now().UTC())
	if err != nil { t.Fatal(err) }
	player.Position.X = 2
	if err := repository.Create(ctx, player); err != nil { t.Fatalf("create player: %v", err) }
	_, _ = db.ExecContext(ctx, "DELETE FROM gathering_node_state WHERE node_id=$1", "tree-001")
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM players WHERE id=$1", string(id))
		_, _ = db.ExecContext(context.Background(), "DELETE FROM gathering_node_state WHERE node_id=$1", "tree-001")
	})

	gathered, result, err := repository.Gather(ctx, id, "tree-001", time.Now())
	if err != nil { t.Fatalf("gather: %v", err) }
	if result.ItemID != "softwood_log" || gathered.Inventory["softwood_log"] != 1 || gathered.Skills["woodcutting"].Experience != 25 {
		t.Fatalf("gather result mismatch: result=%+v player=%+v", result, gathered)
	}
	var charges int
	var initialized bool
	if err := db.QueryRowContext(ctx, "SELECT charges, initialized FROM gathering_node_state WHERE node_id=$1", "tree-001").Scan(&charges, &initialized); err != nil {
		t.Fatalf("read persisted node state: %v", err)
	}
	if !initialized || charges != 2 { t.Fatalf("persisted node state = initialized:%v charges:%d", initialized, charges) }
	reloaded, err := repository.Get(ctx, id)
	if err != nil { t.Fatal(err) }
	if reloaded.Inventory["softwood_log"] != 1 || reloaded.Skills["woodcutting"].Experience != 25 {
		t.Fatalf("player gather reward not persisted: %+v", reloaded)
	}
}
