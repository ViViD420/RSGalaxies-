package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ViViD420/RSGalaxies-/server/internal/game"
)

type PlayerRepository struct { db *sql.DB }

var _ game.PlayerRepository = (*PlayerRepository)(nil)

func NewPlayerRepository(db *sql.DB) *PlayerRepository { return &PlayerRepository{db: db} }

type scanner interface { Scan(dest ...any) error }
type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

const playerColumns = "id, name, created_at, skills, inventory, bank, zone_id, position_x, position_z, last_move_at"

func scanPlayer(row scanner) (game.Player, error) {
	var id, name, zoneID string
	var createdAt, lastMoveAt time.Time
	var skillsJSON, inventoryJSON, bankJSON []byte
	var x, z float64
	var coins int64
	if err := row.Scan(&id, &name, &createdAt, &skillsJSON, &inventoryJSON, &bankJSON, &coins, &zoneID, &x, &z, &lastMoveAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) { return game.Player{}, game.ErrPlayerNotFound }
		return game.Player{}, fmt.Errorf("scan player: %w", err)
	}
	player := game.Player{
		ID: game.PlayerID(id), Name: name, CreatedAt: createdAt.UTC(),
		Position: game.WorldPosition{ZoneID: zoneID, X: x, Z: z},
		LastMoveAt: lastMoveAt.UTC(), Coins: coins,
	}
	if err := json.Unmarshal(skillsJSON, &player.Skills); err != nil { return game.Player{}, fmt.Errorf("decode player skills: %w", err) }
	if err := json.Unmarshal(inventoryJSON, &player.Inventory); err != nil { return game.Player{}, fmt.Errorf("decode player inventory: %w", err) }
	if err := json.Unmarshal(bankJSON, &player.Bank); err != nil { return game.Player{}, fmt.Errorf("decode player bank: %w", err) }
	if player.Skills == nil { player.Skills = make(map[string]game.SkillProgress) }
	if player.Inventory == nil { player.Inventory = make(map[string]int) }
	if player.Bank == nil { player.Bank = make(map[string]int) }
	return player, nil
}

func (r *PlayerRepository) Create(ctx context.Context, player game.Player) error {
	skills, err := json.Marshal(player.Skills)
	if err != nil { return fmt.Errorf("encode skills: %w", err) }
	inventory, err := json.Marshal(player.Inventory)
	if err != nil { return fmt.Errorf("encode inventory: %w", err) }
	bank, err := json.Marshal(player.Bank)
	if err != nil { return fmt.Errorf("encode bank: %w", err) }
	_, err = r.db.ExecContext(ctx, `INSERT INTO players
		(id, name, created_at, skills, inventory, bank, coins, zone_id, position_x, position_z, last_move_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		string(player.ID), player.Name, player.CreatedAt, string(skills), string(inventory), string(bank), player.Coins,
		player.Position.ZoneID, player.Position.X, player.Position.Z, player.LastMoveAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { return game.ErrPlayerAlreadyExists }
		return fmt.Errorf("create player: %w", err)
	}
	return nil
}

func (r *PlayerRepository) Get(ctx context.Context, id game.PlayerID) (game.Player, error) {
	query := "SELECT " + playerColumns + " FROM players WHERE id = $1"
	return scanPlayer(r.db.QueryRowContext(ctx, query, string(id)))
}

func (r *PlayerRepository) List(ctx context.Context) ([]game.Player, error) {
	query := "SELECT " + playerColumns + " FROM players ORDER BY id"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil { return nil, fmt.Errorf("list players: %w", err) }
	defer rows.Close()
	players := make([]game.Player, 0)
	for rows.Next() {
		player, err := scanPlayer(rows)
		if err != nil { return nil, err }
		players = append(players, player)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate players: %w", err) }
	return players, nil
}

func updatePlayer(ctx context.Context, exec executor, player game.Player) error {
	skills, err := json.Marshal(player.Skills)
	if err != nil { return fmt.Errorf("encode skills: %w", err) }
	inventory, err := json.Marshal(player.Inventory)
	if err != nil { return fmt.Errorf("encode inventory: %w", err) }
	bank, err := json.Marshal(player.Bank)
	if err != nil { return fmt.Errorf("encode bank: %w", err) }
	_, err = exec.ExecContext(ctx, `UPDATE players SET skills=$2, inventory=$3, bank=$4, coins=$5, zone_id=$6,
		position_x=$7, position_z=$8, last_move_at=$9 WHERE id=$1`,
		string(player.ID), string(skills), string(inventory), string(bank), player.Coins, player.Position.ZoneID,
		player.Position.X, player.Position.Z, player.LastMoveAt)
	if err != nil { return fmt.Errorf("update player: %w", err) }
	return nil
}

func (r *PlayerRepository) lockPlayer(ctx context.Context, tx *sql.Tx, id game.PlayerID) (game.Player, error) {
	query := "SELECT " + playerColumns + " FROM players WHERE id = $1 FOR UPDATE"
	return scanPlayer(tx.QueryRowContext(ctx, query, string(id)))
}

func (r *PlayerRepository) AwardExperience(ctx context.Context, id game.PlayerID, skillID string, amount int) (game.Player, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return game.Player{}, false, err }
	defer tx.Rollback()
	player, err := r.lockPlayer(ctx, tx, id)
	if err != nil { return game.Player{}, false, err }
	leveled, err := player.AwardExperience(skillID, amount)
	if err != nil { return game.Player{}, false, err }
	if err := updatePlayer(ctx, tx, player); err != nil { return game.Player{}, false, err }
	if err := tx.Commit(); err != nil { return game.Player{}, false, fmt.Errorf("commit XP award: %w", err) }
	return player, leveled, nil
}

func (r *PlayerRepository) MovePlayer(ctx context.Context, id game.PlayerID, x, z float64, now time.Time) (game.Player, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return game.Player{}, err }
	defer tx.Rollback()
	player, err := r.lockPlayer(ctx, tx, id)
	if err != nil { return game.Player{}, err }
	if err := player.MoveTo(x, z, now); err != nil { return player, err }
	if err := updatePlayer(ctx, tx, player); err != nil { return game.Player{}, err }
	if err := tx.Commit(); err != nil { return game.Player{}, fmt.Errorf("commit movement: %w", err) }
	return player, nil
}

func (r *PlayerRepository) Gather(ctx context.Context, id game.PlayerID, nodeID string, now time.Time) (game.Player, game.GatheringResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return game.Player{}, game.GatheringResult{}, err }
	defer tx.Rollback()
	player, err := r.lockPlayer(ctx, tx, id)
	if err != nil { return game.Player{}, game.GatheringResult{}, err }
	node, ok := game.FindGatheringNode(nodeID)
	if !ok { return player, game.GatheringResult{}, game.ErrUnknownGatheringNode }
	if _, err := tx.ExecContext(ctx, `INSERT INTO gathering_node_state (node_id, charges, initialized)
		VALUES ($1, 0, FALSE) ON CONFLICT (node_id) DO NOTHING`, nodeID); err != nil {
		return game.Player{}, game.GatheringResult{}, fmt.Errorf("initialize gathering node: %w", err)
	}
	var state game.GatheringNodeState
	var respawnAt sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT charges, respawn_at, initialized
		FROM gathering_node_state WHERE node_id=$1 FOR UPDATE`, nodeID).Scan(&state.Charges, &respawnAt, &state.Initialized); err != nil {
		return game.Player{}, game.GatheringResult{}, fmt.Errorf("lock gathering node: %w", err)
	}
	if respawnAt.Valid { state.RespawnAt = respawnAt.Time }
	result, nextState, err := player.Gather(node, state, now)
	if err != nil { return player, game.GatheringResult{}, err }
	var respawnValue any
	if !nextState.RespawnAt.IsZero() { respawnValue = nextState.RespawnAt }
	if _, err := tx.ExecContext(ctx, `UPDATE gathering_node_state SET charges=$2, respawn_at=$3, initialized=$4 WHERE node_id=$1`,
		nodeID, nextState.Charges, respawnValue, nextState.Initialized); err != nil {
		return game.Player{}, game.GatheringResult{}, fmt.Errorf("update gathering node: %w", err)
	}
	if err := updatePlayer(ctx, tx, player); err != nil { return game.Player{}, game.GatheringResult{}, err }
	if err := tx.Commit(); err != nil { return game.Player{}, game.GatheringResult{}, fmt.Errorf("commit gathering: %w", err) }
	return player, result, nil
}

func (r *PlayerRepository) Craft(ctx context.Context, id game.PlayerID, recipeID string, now time.Time) (game.Player, game.CraftingResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return game.Player{}, game.CraftingResult{}, err }
	defer tx.Rollback()
	player, err := r.lockPlayer(ctx, tx, id)
	if err != nil { return game.Player{}, game.CraftingResult{}, err }
	recipe, ok := game.FindCraftingRecipe(recipeID)
	if !ok { return player, game.CraftingResult{}, game.ErrUnknownCraftingRecipe }
	station, ok := game.FindCraftingStation(recipe.StationID)
	if !ok { return player, game.CraftingResult{}, game.ErrUnknownCraftingStation }
	result, err := player.Craft(recipe, station, now)
	if err != nil { return player, game.CraftingResult{}, err }
	if err := updatePlayer(ctx, tx, player); err != nil { return game.Player{}, game.CraftingResult{}, err }
	if err := tx.Commit(); err != nil { return game.Player{}, game.CraftingResult{}, fmt.Errorf("commit crafting: %w", err) }
	return player, result, nil
}
