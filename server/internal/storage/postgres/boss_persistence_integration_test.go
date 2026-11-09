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

func TestPostgresBossEncounterStatePersistsAndRejectsStaleCycles(t *testing.T) {
	databaseURL := os.Getenv("RS_GALAXIES_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("RS_GALAXIES_TEST_DATABASE_URL is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(4)
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	enemyID := fmt.Sprintf("boss-persistence-%d", now.UnixNano())
	firstCycle := "boss-cycle-first"
	secondCycle := "boss-cycle-second"
	if _, err := db.ExecContext(ctx, `INSERT INTO enemy_spawn_state(enemy_id,health,initialized,encounter_id)
		VALUES($1,100,TRUE,$2)`, enemyID, firstCycle); err != nil {
		t.Fatalf("create boss spawn row: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_spawn_state WHERE enemy_id=$1", enemyID)
	})

	state := game.BossEncounterState{
		EncounterID: firstCycle,
		DefinitionVersion: 1,
		PhaseID: "rampage",
		PhaseStartedAt: now.Add(-time.Minute),
		ActiveAbilityID: "ground-shatter",
		TelegraphStartedAt: now.Add(-time.Second),
		TelegraphUntil: now.Add(time.Second),
		AbilityCooldowns: map[string]time.Time{
			"stone-fist": now.Add(4 * time.Second),
			"ground-shatter": now.Add(7 * time.Second),
		},
		NextAbilityIndex: 1,
		LastUpdatedAt: now,
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := persistBossEncounterState(ctx, tx, enemyID, "stoneback-colossus", state); err != nil {
		_ = tx.Rollback()
		t.Fatalf("persist boss state: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit boss state: %v", err)
	}

	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	loaded, definitionID, found, err := loadBossEncounterStateForUpdate(ctx, tx, enemyID, firstCycle)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("load boss state: %v", err)
	}
	if !found || definitionID != "stoneback-colossus" {
		_ = tx.Rollback()
		t.Fatalf("boss state found=%v definition=%q", found, definitionID)
	}
	if loaded.EncounterID != state.EncounterID || loaded.DefinitionVersion != state.DefinitionVersion ||
		loaded.PhaseID != state.PhaseID || !loaded.PhaseStartedAt.Equal(state.PhaseStartedAt) ||
		loaded.ActiveAbilityID != state.ActiveAbilityID ||
		!loaded.TelegraphStartedAt.Equal(state.TelegraphStartedAt) ||
		!loaded.TelegraphUntil.Equal(state.TelegraphUntil) ||
		loaded.NextAbilityIndex != state.NextAbilityIndex || !loaded.LastUpdatedAt.Equal(state.LastUpdatedAt) ||
		loaded.Defeated || len(loaded.AbilityCooldowns) != 2 ||
		!loaded.AbilityCooldowns["stone-fist"].Equal(state.AbilityCooldowns["stone-fist"]) ||
		!loaded.AbilityCooldowns["ground-shatter"].Equal(state.AbilityCooldowns["ground-shatter"]) {
		_ = tx.Rollback()
		t.Fatalf("round-tripped boss state differs: got %+v want %+v", loaded, state)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	// A respawn changes the authoritative parent cycle. A stale worker from
	// the previous life must not load or overwrite the new life.
	if _, err := db.ExecContext(ctx, "UPDATE enemy_spawn_state SET encounter_id=$2 WHERE enemy_id=$1", enemyID, secondCycle); err != nil {
		t.Fatal(err)
	}
	staleTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := loadBossEncounterStateForUpdate(ctx, staleTx, enemyID, firstCycle); !errors.Is(err, ErrBossEncounterCycleMismatch) {
		_ = staleTx.Rollback()
		t.Fatalf("stale-cycle load error = %v, want ErrBossEncounterCycleMismatch", err)
	}
	_ = staleTx.Rollback()

	staleTx, err = db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := persistBossEncounterState(ctx, staleTx, enemyID, "stoneback-colossus", state); !errors.Is(err, ErrBossEncounterCycleMismatch) {
		_ = staleTx.Rollback()
		t.Fatalf("stale-cycle persist error = %v, want ErrBossEncounterCycleMismatch", err)
	}
	_ = staleTx.Rollback()

	newCycleTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _, found, err = loadBossEncounterStateForUpdate(ctx, newCycleTx, enemyID, secondCycle)
	if err != nil {
		_ = newCycleTx.Rollback()
		t.Fatal(err)
	}
	if found {
		_ = newCycleTx.Rollback()
		t.Fatal("new boss cycle inherited prior cycle state")
	}
	if err := newCycleTx.Commit(); err != nil {
		t.Fatal(err)
	}
}
