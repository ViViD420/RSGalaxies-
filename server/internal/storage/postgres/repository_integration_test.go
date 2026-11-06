package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresRepositoryGameplayTransaction(t *testing.T) {
	databaseURL := os.Getenv("RS_GALAXIES_TEST_DATABASE_URL")
	if databaseURL == "" { t.Skip("RS_GALAXIES_TEST_DATABASE_URL is not configured") }

	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil { t.Fatal(err) }
	defer db.Close()
	db.SetMaxOpenConns(5)
	if err := db.PingContext(ctx); err != nil { t.Fatalf("connect test database: %v", err) }
	if err := Migrate(ctx, db); err != nil { t.Fatalf("migrate test database: %v", err) }
	if err := Migrate(ctx, db); err != nil { t.Fatalf("second migration should be idempotent: %v", err) }

	if _, err := db.ExecContext(ctx, "DELETE FROM gathering_node_state WHERE node_id = $1", "tree-001"); err != nil { t.Fatal(err) }
	now := time.Now().UTC()
	playerID := game.PlayerID(fmt.Sprintf("integration-%d", now.UnixNano()))
	player, err := game.NewPlayer(playerID, "DBTester", now)
	if err != nil { t.Fatal(err) }
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(new(int)); err != nil { t.Fatal(err) }

	repository := NewPlayerRepository(db)
	if err := repository.Create(ctx, player); err != nil { t.Fatalf("create player: %v", err) }
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM players WHERE id = $1", string(playerID))
		_, _ = db.ExecContext(context.Background(), "DELETE FROM gathering_node_state WHERE node_id = $1", "tree-001")
	})
	if err := repository.Create(ctx, player); !errors.Is(err, game.ErrPlayerAlreadyExists) {
		t.Fatalf("duplicate create error = %v, want ErrPlayerAlreadyExists", err)
	}

	loaded, err := repository.Get(ctx, playerID)
	if err != nil || loaded.ID != playerID { t.Fatalf("get player: %+v, %v", loaded, err) }
	if _, leveled, err := repository.AwardExperience(ctx, playerID, "mining", 100); err != nil || !leveled {
		t.Fatalf("award mining XP: leveled=%v err=%v", leveled, err)
	}
	if _, err := repository.MovePlayer(ctx, playerID, 0, 1.5, now.Add(200*time.Millisecond)); err != nil {
		t.Fatalf("move player: %v", err)
	}
	if _, _, err := repository.Gather(ctx, playerID, "tree-001", now.Add(250*time.Millisecond)); err != nil {
		t.Fatalf("first gather: %v", err)
	}
	if _, _, err := repository.Gather(ctx, playerID, "tree-001", now.Add(300*time.Millisecond)); err != nil {
		t.Fatalf("second gather: %v", err)
	}
	if _, _, err := repository.Craft(ctx, playerID, "basic-plank", now.Add(350*time.Millisecond)); err != nil {
		t.Fatalf("craft plank: %v", err)
	}

	updated, err := repository.Get(ctx, playerID)
	if err != nil { t.Fatal(err) }
	if updated.Inventory["softwood_log"] != 0 || updated.Inventory["wooden_plank"] != 1 {
		t.Fatalf("unexpected persisted inventory: %v", updated.Inventory)
	}
	if updated.Skills["woodcutting"].Experience != 50 || updated.Skills["crafting"].Experience != 30 || updated.Skills["mining"].Experience != 100 {
		t.Fatalf("unexpected persisted skills: %+v", updated.Skills)
	}
	players, err := repository.List(ctx)
	if err != nil { t.Fatal(err) }
	found := false
	for _, listed := range players { if listed.ID == playerID { found = true; break } }
	if !found { t.Fatal("created player missing from list") }
}
