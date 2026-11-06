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
	if err := repository.Create(ctx, player); err != nil { t.Fatalf("create player: %v", err) }
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM players WHERE id=$1", string(id)) })

	loaded, err := repository.Get(ctx, id)
	if err != nil { t.Fatalf("get player: %v", err) }
	if loaded.Inventory["softwood_log"] != 2 || loaded.Position.Z != 2 {
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
