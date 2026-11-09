package postgres

import (
	"context"
	"errors"
	"database/sql"
	"fmt"
	"math"
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
	var targetCharacter sql.NullString
	if err := db.QueryRowContext(ctx, "SELECT target_character_id FROM enemy_spawn_state WHERE enemy_id=$1", enemyID).Scan(&targetCharacter); err != nil { t.Fatal(err) }
	if !targetCharacter.Valid || targetCharacter.String != string(playerID) {
		t.Fatalf("player-initiated combat target = %+v, want %s", targetCharacter, playerID)
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
	var activeEncounterID string
	if err := db.QueryRowContext(ctx, "SELECT encounter_id FROM enemy_spawn_state WHERE enemy_id=$1", enemyID).Scan(&activeEncounterID); err != nil { t.Fatal(err) }
	contributions, err := accountRepo.ListEncounterParticipants(ctx, enemyID, activeEncounterID)
	if err != nil { t.Fatalf("list live encounter contributions: %v", err) }
	if len(contributions) != 1 || contributions[0].DamageDealt != 2 ||
		contributions[0].DamageReceived != 4 || contributions[0].ActionCount != 2 {
		t.Fatalf("live encounter contribution ledger = %+v", contributions)
	}

	snapshots, err := accountRepo.ListEnemySnapshotsForAccount(ctx, account.ID, playerID, now.Add(4*time.Second))
	if err != nil { t.Fatalf("list enemy snapshots: %v", err) }
	if len(snapshots) != 3 { t.Fatalf("snapshot count = %d, want three zone enemies including Stoneback", len(snapshots)) }
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
	if err := db.QueryRowContext(ctx, "SELECT target_character_id FROM enemy_spawn_state WHERE enemy_id=$1", enemyID).Scan(&targetCharacter); err != nil { t.Fatal(err) }
	if targetCharacter.Valid { t.Fatalf("enemy retained stale target after leash return: %s", targetCharacter.String) }

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

 
func TestPostgresEnemyAITargetRemainsStickyWhilePresenceIsActive(t *testing.T) {
	databaseURL := os.Getenv("RS_GALAXIES_TEST_DATABASE_URL")
	if databaseURL == "" { t.Skip("RS_GALAXIES_TEST_DATABASE_URL is not configured") }
	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil { t.Fatal(err) }
	defer db.Close()
	db.SetMaxOpenConns(8)
	if err := db.PingContext(ctx); err != nil { t.Fatalf("connect test database: %v", err) }
	if err := Migrate(ctx, db); err != nil { t.Fatalf("migrate test database: %v", err) }
	now := time.Now().UTC().Truncate(time.Second)
	accountRepo := NewAccountRepository(db)
	email := fmt.Sprintf("enemy-sticky-%d@example.test", now.UnixNano())
	account, _, err := accountRepo.Register(ctx, email, "StrongTestPassword!2026", now)
	if err != nil { t.Fatalf("register test account: %v", err) }
	firstID := game.PlayerID(fmt.Sprintf("enemy-sticky-first-%d", now.UnixNano()))
	stickyID := game.PlayerID(fmt.Sprintf("enemy-sticky-current-%d", now.UnixNano()))
	first, err := game.NewPlayer(firstID, "FirstTarget", now)
	if err != nil { t.Fatal(err) }
	sticky, err := game.NewPlayer(stickyID, "StickyTarget", now)
	if err != nil { t.Fatal(err) }
	sticky.Position.X = 2
	sticky.Position.Z = 4
	if err := accountRepo.CreateCharacter(ctx, account.ID, first); err != nil { t.Fatalf("create first character: %v", err) }
	if err := accountRepo.CreateCharacter(ctx, account.ID, sticky); err != nil { t.Fatalf("create sticky character: %v", err) }
	const enemyID = "frontier-rat-001"
	const testZoneID = "sticky-target-test-zone"
	if _, err := db.ExecContext(ctx, "UPDATE players SET zone_id=$3 WHERE id IN ($1,$2)", string(firstID), string(stickyID), testZoneID); err != nil { t.Fatal(err) }
	if _, err := db.ExecContext(ctx, `INSERT INTO enemy_spawn_state(enemy_id,health,initialized,behavior,position_zone_id,position_x,position_z)
		VALUES($1,12,TRUE,'idle',$2,8,0)
		ON CONFLICT(enemy_id) DO UPDATE SET health=12,initialized=TRUE,respawn_at=NULL,behavior='idle',
		position_zone_id=$2,position_x=8,position_z=0,target_character_id=NULL,
		telegraph_until=NULL,next_attack_at=NULL,last_ai_updated_at=NULL`, enemyID, testZoneID); err != nil { t.Fatal(err) }
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM zone_presence WHERE character_id IN ($1,$2)", string(firstID), string(stickyID))
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_spawn_state WHERE enemy_id=$1", enemyID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM players WHERE id IN ($1,$2)", string(firstID), string(stickyID))
		_, _ = db.ExecContext(context.Background(), "DELETE FROM accounts WHERE id=$1", account.ID)
	})
	if err := accountRepo.HeartbeatPresenceForAccount(ctx, account.ID, firstID, now); err != nil { t.Fatal(err) }
	if err := accountRepo.HeartbeatPresenceForAccount(ctx, account.ID, stickyID, now); err != nil { t.Fatal(err) }

	if err := accountRepo.TickEnemyAI(ctx, enemyID, now); err != nil { t.Fatalf("initial target selection: %v", err) }
	var targetID string
	if err := db.QueryRowContext(ctx, "SELECT target_character_id FROM enemy_spawn_state WHERE enemy_id=$1", enemyID).Scan(&targetID); err != nil { t.Fatal(err) }
	if targetID != string(stickyID) { t.Fatalf("initial target = %s, want nearest %s", targetID, stickyID) }

	// The first character moves closer, but the enemy must keep its existing
	// live target rather than oscillating between nearby players each tick.
	if _, err := db.ExecContext(ctx, "UPDATE players SET position_x=7,last_move_at=$2 WHERE id=$1", string(firstID), now.Add(time.Second)); err != nil { t.Fatal(err) }
	if err := accountRepo.HeartbeatPresenceForAccount(ctx, account.ID, firstID, now.Add(time.Second)); err != nil { t.Fatal(err) }
	if err := accountRepo.TickEnemyAI(ctx, enemyID, now.Add(2*time.Second)); err != nil { t.Fatalf("sticky target tick: %v", err) }
	var x, z float64
	if err := db.QueryRowContext(ctx, "SELECT target_character_id,position_x,position_z FROM enemy_spawn_state WHERE enemy_id=$1", enemyID).Scan(&targetID,&x,&z); err != nil { t.Fatal(err) }
	if targetID != string(stickyID) { t.Fatalf("target switched to %s, want sticky target %s", targetID, stickyID) }
	wantX := 8.0 - 4.0*6.0/math.Sqrt(52.0)
	wantZ := 4.0*4.0/math.Sqrt(52.0)
	if math.Abs(x-wantX) > 0.001 || math.Abs(z-wantZ) > 0.001 {
		t.Fatalf("enemy moved toward a different target: position=(%v,%v), want=(%v,%v)", x,z,wantX,wantZ)
	}

	// If the sticky target's lease disappears, the next eligible character
	// can acquire the enemy rather than leaving it bound to stale identity.
	if _, err := db.ExecContext(ctx, "DELETE FROM zone_presence WHERE character_id=$1", string(stickyID)); err != nil { t.Fatal(err) }
	if err := accountRepo.TickEnemyAI(ctx, enemyID, now.Add(3*time.Second)); err != nil { t.Fatalf("switch after sticky target disappears: %v", err) }
	if err := db.QueryRowContext(ctx, "SELECT target_character_id FROM enemy_spawn_state WHERE enemy_id=$1", enemyID).Scan(&targetID); err != nil { t.Fatal(err) }
	if targetID != string(firstID) { t.Fatalf("target after previous lease expires = %s, want %s", targetID, firstID) }
}

 
func TestPostgresEncounterContributionsPersistAcrossEnemyRespawn(t *testing.T) {
	databaseURL := os.Getenv("RS_GALAXIES_TEST_DATABASE_URL")
	if databaseURL == "" { t.Skip("RS_GALAXIES_TEST_DATABASE_URL is not configured") }
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil { t.Fatal(err) }
	defer db.Close()
	db.SetMaxOpenConns(6)
	if err := db.PingContext(ctx); err != nil { t.Fatalf("connect test database: %v", err) }
	if err := Migrate(ctx, db); err != nil { t.Fatalf("migrate test database: %v", err) }

	now := time.Now().UTC().Truncate(time.Second)
	accountRepo := NewAccountRepository(db)
	email := fmt.Sprintf("enemy-ledger-%d@example.test", now.UnixNano())
	account, _, err := accountRepo.Register(ctx, email, "StrongTestPassword!2026", now)
	if err != nil { t.Fatalf("register test account: %v", err) }
	playerID := game.PlayerID(fmt.Sprintf("enemy-ledger-player-%d", now.UnixNano()))
	player, err := game.NewPlayer(playerID, "LedgerWayfarer", now)
	if err != nil { t.Fatal(err) }
	player.Position.X = 8 // Frontier Rat's catalog spawn position.
	if err := accountRepo.CreateCharacter(ctx, account.ID, player); err != nil { t.Fatalf("create ledger character: %v", err) }
	const enemyID = "frontier-rat-001"
	if _, err := db.ExecContext(ctx, "DELETE FROM enemy_encounter_participants WHERE enemy_id=$1", enemyID); err != nil { t.Fatal(err) }
	if _, err := db.ExecContext(ctx, "DELETE FROM enemy_spawn_state WHERE enemy_id=$1", enemyID); err != nil { t.Fatal(err) }
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_encounter_participants WHERE enemy_id=$1", enemyID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_spawn_state WHERE enemy_id=$1", enemyID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM players WHERE id=$1", string(playerID))
		_, _ = db.ExecContext(context.Background(), "DELETE FROM accounts WHERE id=$1", account.ID)
	})

	_, first, err := accountRepo.AttackEnemyForAccount(ctx, account.ID, playerID, enemyID, now.Add(2*time.Second))
	if err != nil { t.Fatalf("first encounter attack: %v", err) }
	if first.DamageDealt != 2 || first.DamageReceived != 2 { t.Fatalf("first combat result = %+v", first) }
	var firstEncounterID string
	if err := db.QueryRowContext(ctx, "SELECT encounter_id FROM enemy_spawn_state WHERE enemy_id=$1", enemyID).Scan(&firstEncounterID); err != nil { t.Fatal(err) }
	firstParticipants, err := accountRepo.ListEncounterParticipants(ctx, enemyID, firstEncounterID)
	if err != nil { t.Fatalf("list first encounter participants: %v", err) }
	if len(firstParticipants) != 1 || firstParticipants[0].CharacterID != playerID ||
		firstParticipants[0].DamageDealt != 2 || firstParticipants[0].DamageReceived != 2 ||
		firstParticipants[0].ActionCount != 1 {
		t.Fatalf("first encounter contribution = %+v", firstParticipants)
	}

	if _, err := db.ExecContext(ctx, "UPDATE enemy_spawn_state SET health=0,respawn_at=$2 WHERE enemy_id=$1", enemyID, now.Add(10*time.Second)); err != nil { t.Fatal(err) }
	if _, _, err := accountRepo.AttackEnemyForAccount(ctx, account.ID, playerID, enemyID, now.Add(9*time.Second)); !errors.Is(err, game.ErrEnemyRespawning) {
		t.Fatalf("attack during encounter respawn error = %v, want ErrEnemyRespawning", err)
	}
	_, second, err := accountRepo.AttackEnemyForAccount(ctx, account.ID, playerID, enemyID, now.Add(11*time.Second))
	if err != nil { t.Fatalf("attack new encounter after respawn: %v", err) }
	if second.DamageDealt != 2 { t.Fatalf("second encounter combat result = %+v", second) }
	var secondEncounterID string
	if err := db.QueryRowContext(ctx, "SELECT encounter_id FROM enemy_spawn_state WHERE enemy_id=$1", enemyID).Scan(&secondEncounterID); err != nil { t.Fatal(err) }
	if secondEncounterID == firstEncounterID || len(secondEncounterID) != 32 {
		t.Fatalf("respawn did not create a new encounter cycle: first=%q second=%q", firstEncounterID, secondEncounterID)
	}
	secondParticipants, err := accountRepo.ListEncounterParticipants(ctx, enemyID, secondEncounterID)
	if err != nil { t.Fatalf("list second encounter participants: %v", err) }
	if len(secondParticipants) != 1 || secondParticipants[0].CharacterID != playerID ||
		secondParticipants[0].DamageDealt != 2 || secondParticipants[0].DamageReceived != 2 ||
		secondParticipants[0].ActionCount != 1 {
		t.Fatalf("second encounter contribution = %+v", secondParticipants)
	}
}


func TestPostgresEnemyTargetSelectionPrefersEncounterThreat(t *testing.T) {
	databaseURL := os.Getenv("RS_GALAXIES_TEST_DATABASE_URL")
	if databaseURL == "" { t.Skip("RS_GALAXIES_TEST_DATABASE_URL is not configured") }
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil { t.Fatal(err) }
	defer db.Close()
	db.SetMaxOpenConns(6)
	if err := db.PingContext(ctx); err != nil { t.Fatal(err) }
	if err := Migrate(ctx, db); err != nil { t.Fatal(err) }

	now := time.Now().UTC().Truncate(time.Second)
	accountRepo := NewAccountRepository(db)
	account, _, err := accountRepo.Register(ctx, fmt.Sprintf("enemy-threat-%d@example.test", now.UnixNano()), "StrongTestPassword!2026", now)
	if err != nil { t.Fatal(err) }
	threatID := game.PlayerID(fmt.Sprintf("enemy-threat-leader-%d", now.UnixNano()))
	nearID := game.PlayerID(fmt.Sprintf("enemy-threat-near-%d", now.UnixNano()))
	threatPlayer, err := game.NewPlayer(threatID, "ThreatLeader", now)
	if err != nil { t.Fatal(err) }
	nearPlayer, err := game.NewPlayer(nearID, "NearbyPlayer", now)
	if err != nil { t.Fatal(err) }
	threatPlayer.Position = game.WorldPosition{ZoneID:"threat-test-zone", X:7, Z:0}
	nearPlayer.Position = game.WorldPosition{ZoneID:"threat-test-zone", X:2, Z:0}
	if err := accountRepo.CreateCharacter(ctx, account.ID, threatPlayer); err != nil { t.Fatal(err) }
	if err := accountRepo.CreateCharacter(ctx, account.ID, nearPlayer); err != nil { t.Fatal(err) }

	const enemyID = "frontier-rat-001"
	const encounterID = "threat-test-encounter"
	_, err = db.ExecContext(ctx, `INSERT INTO enemy_spawn_state(
		enemy_id,health,initialized,behavior,position_zone_id,position_x,position_z,encounter_id,target_character_id)
		VALUES($1,12,TRUE,'idle',$2,8,0,$3,NULL)
		ON CONFLICT(enemy_id) DO UPDATE SET health=12,initialized=TRUE,respawn_at=NULL,
		behavior='idle',position_zone_id=$2,position_x=8,position_z=0,
		encounter_id=$3,target_character_id=NULL,telegraph_until=NULL,next_attack_at=NULL,last_ai_updated_at=NULL`,
		enemyID,"threat-test-zone",encounterID)
	if err != nil { t.Fatal(err) }
	_, err = db.ExecContext(ctx, `INSERT INTO enemy_encounter_participants(
		encounter_id,enemy_id,account_id,character_id,damage_dealt,damage_received,action_count,first_contribution_at,last_contribution_at)
		VALUES($1,$2,$3,$4,30,0,3,$5,$5)`, encounterID,enemyID,account.ID,string(threatID),now)
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM zone_presence WHERE character_id IN ($1,$2)", string(threatID),string(nearID))
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_encounter_participants WHERE enemy_id=$1", enemyID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM enemy_spawn_state WHERE enemy_id=$1", enemyID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM players WHERE id IN ($1,$2)", string(threatID),string(nearID))
		_, _ = db.ExecContext(context.Background(), "DELETE FROM accounts WHERE id=$1", account.ID)
	})
	if err := accountRepo.HeartbeatPresenceForAccount(ctx,account.ID,threatID,now); err != nil { t.Fatal(err) }
	if err := accountRepo.HeartbeatPresenceForAccount(ctx,account.ID,nearID,now); err != nil { t.Fatal(err) }

	if err := accountRepo.TickEnemyAI(ctx,enemyID,now); err != nil { t.Fatalf("tick enemy with threat candidates: %v",err) }
	var targetID string
	if err := db.QueryRowContext(ctx,"SELECT target_character_id FROM enemy_spawn_state WHERE enemy_id=$1",enemyID).Scan(&targetID); err != nil { t.Fatal(err) }
	if targetID != string(threatID) {
		t.Fatalf("enemy target = %s, want higher-threat participant %s despite nearer bystander %s",targetID,threatID,nearID)
	}
}
