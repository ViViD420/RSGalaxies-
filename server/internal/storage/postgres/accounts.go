package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ViViD420/RSGalaxies-/server/internal/auth"
	"github.com/ViViD420/RSGalaxies-/server/internal/game"
)

const (
	SessionLifetime = 30 * 24 * time.Hour
	MaxActiveSessionsPerAccount = 10
)

var (
	ErrAccountExists = auth.ErrAccountExists
	ErrAccountNotFound = auth.ErrAccountNotFound
	ErrInvalidCredentials = auth.ErrInvalidCredentials
	ErrInvalidEmail = auth.ErrInvalidEmail
	ErrSessionNotFound = auth.ErrSessionNotFound
)

type AccountRepository struct { db *sql.DB }

var _ auth.AccountStore = (*AccountRepository)(nil)

func NewAccountRepository(db *sql.DB) *AccountRepository { return &AccountRepository{db: db} }

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if len(email) < 3 || len(email) > 320 { return "", ErrInvalidEmail }
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || !strings.Contains(email, "@") {
		return "", ErrInvalidEmail
	}
	return email, nil
}

func newAccountID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil { return "", err }
	return hex.EncodeToString(raw[:]), nil
}

func (r *AccountRepository) Register(ctx context.Context, email, password string, now time.Time) (auth.Account, string, error) {
	email, err := normalizeEmail(email)
	if err != nil { return auth.Account{}, "", err }
	passwordHash, err := auth.HashPassword(password)
	if err != nil { return auth.Account{}, "", err }
	id, err := newAccountID()
	if err != nil { return auth.Account{}, "", fmt.Errorf("generate account id: %w", err) }
	token, tokenHash, err := auth.NewSessionToken()
	if err != nil { return auth.Account{}, "", err }
	now = now.UTC()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return auth.Account{}, "", fmt.Errorf("begin account registration: %w", err) }
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO accounts (id,email,password_hash,created_at) VALUES ($1,$2,$3,$4)`, id, email, passwordHash, now)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { return auth.Account{}, "", ErrAccountExists }
		return auth.Account{}, "", fmt.Errorf("insert account: %w", err)
	}
	if err := createSession(ctx, tx, id, tokenHash[:], now); err != nil {
		return auth.Account{}, "", fmt.Errorf("create initial session: %w", err)
	}
	if err := tx.Commit(); err != nil { return auth.Account{}, "", fmt.Errorf("commit account registration: %w", err) }
	return auth.Account{ID:id, Email:email, CreatedAt:now}, token, nil
}

func (r *AccountRepository) Login(ctx context.Context, email, password string, now time.Time) (auth.Account, string, error) {
	email, err := normalizeEmail(email)
	if err != nil { return auth.Account{}, "", ErrInvalidCredentials }
	var account auth.Account
	var passwordHash string
	var disabledAt sql.NullTime
	err = r.db.QueryRowContext(ctx, `SELECT id,email,password_hash,created_at,disabled_at FROM accounts WHERE LOWER(email)=$1`, email).
		Scan(&account.ID,&account.Email,&passwordHash,&account.CreatedAt,&disabledAt)
	if errors.Is(err, sql.ErrNoRows) { auth.BurnPasswordVerification(password); return auth.Account{}, "", ErrInvalidCredentials }
	if err != nil { return auth.Account{}, "", fmt.Errorf("load account for login: %w", err) }
	if disabledAt.Valid { auth.BurnPasswordVerification(password); return auth.Account{}, "", ErrInvalidCredentials }
	valid, err := auth.VerifyPassword(passwordHash,password)
	if err != nil || !valid { if err != nil { auth.BurnPasswordVerification(password) }; return auth.Account{}, "", ErrInvalidCredentials }
	token, tokenHash, err := auth.NewSessionToken()
	if err != nil { return auth.Account{}, "", err }
	now = now.UTC()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return auth.Account{}, "", fmt.Errorf("begin login session: %w", err) }
	defer tx.Rollback()
	var lockedAccountID string
	if err := tx.QueryRowContext(ctx, "SELECT id FROM accounts WHERE id=$1 FOR UPDATE", account.ID).Scan(&lockedAccountID); err != nil {
		return auth.Account{}, "", fmt.Errorf("lock account for session: %w", err)
	}
	if err := createSession(ctx, tx, account.ID, tokenHash[:], now); err != nil {
		return auth.Account{}, "", fmt.Errorf("create login session: %w", err)
	}
	if err := tx.Commit(); err != nil { return auth.Account{}, "", fmt.Errorf("commit login session: %w", err) }
	account.CreatedAt = account.CreatedAt.UTC()
	return account, token, nil
}

func (r *AccountRepository) AuthenticateSession(ctx context.Context, token string, now time.Time) (auth.Account, error) {
	tokenHash, err := auth.HashSessionToken(token)
	if err != nil { return auth.Account{}, ErrSessionNotFound }
	var account auth.Account
	err = r.db.QueryRowContext(ctx, `UPDATE account_sessions s SET last_used_at=$2
		FROM accounts a
		WHERE s.token_hash=$1 AND s.account_id=a.id AND s.revoked_at IS NULL
		AND s.expires_at > $2 AND a.disabled_at IS NULL
		RETURNING a.id,a.email,a.created_at`, tokenHash[:], now.UTC()).
		Scan(&account.ID,&account.Email,&account.CreatedAt)
	if errors.Is(err,sql.ErrNoRows) { return auth.Account{}, ErrSessionNotFound }
	if err != nil { return auth.Account{}, fmt.Errorf("authenticate session: %w",err) }
	account.CreatedAt = account.CreatedAt.UTC()
	return account,nil
}

func (r *AccountRepository) RevokeSession(ctx context.Context, token string, now time.Time) error {
	tokenHash, err := auth.HashSessionToken(token)
	if err != nil { return nil }
	_, err = r.db.ExecContext(ctx,`UPDATE account_sessions SET revoked_at=$2 WHERE token_hash=$1 AND revoked_at IS NULL`,tokenHash[:],now.UTC())
	if err != nil { return fmt.Errorf("revoke session: %w",err) }
	return nil
}

// CreateCharacter atomically binds a newly-created character to its owning account.
func (r *AccountRepository) CreateCharacter(ctx context.Context, accountID string, player game.Player) error {
	skills, err := json.Marshal(player.Skills)
	if err != nil { return fmt.Errorf("encode character skills: %w", err) }
	inventory, err := json.Marshal(player.Inventory)
	if err != nil { return fmt.Errorf("encode character inventory: %w", err) }
	bank, err := json.Marshal(player.Bank)
	if err != nil { return fmt.Errorf("encode character bank: %w", err) }
	_, err = r.db.ExecContext(ctx,`INSERT INTO players
		(id,account_id,name,created_at,skills,inventory,bank,coins,zone_id,position_x,position_z,last_move_at)
		VALUES ($1,$2,$3,$4,$5::jsonb,$6::jsonb,$7::jsonb,$8,$9,$10,$11,$12)`,
		string(player.ID),accountID,player.Name,player.CreatedAt,string(skills),string(inventory),string(bank),player.Coins,
		player.Position.ZoneID,player.Position.X,player.Position.Z,player.LastMoveAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err,&pgErr) {
			switch pgErr.Code {
			case "23505": return game.ErrPlayerAlreadyExists
			case "23503": return ErrAccountNotFound
			}
		}
		return fmt.Errorf("create owned character: %w",err)
	}
	return nil
}

func (r *AccountRepository) GetCharacter(ctx context.Context, accountID string, playerID game.PlayerID) (game.Player,error) {
	query := "SELECT "+playerColumns+" FROM players WHERE account_id=$1 AND id=$2"
	return scanPlayer(r.db.QueryRowContext(ctx,query,accountID,string(playerID)))
}

func (r *AccountRepository) ListCharacters(ctx context.Context, accountID string) ([]game.Player,error) {
	rows,err:=r.db.QueryContext(ctx,"SELECT "+playerColumns+" FROM players WHERE account_id=$1 ORDER BY id",accountID)
	if err!=nil{return nil,fmt.Errorf("list owned characters: %w",err)}
	defer rows.Close()
	players:=make([]game.Player,0)
	for rows.Next(){player,err:=scanPlayer(rows);if err!=nil{return nil,err};players=append(players,player)}
	if err:=rows.Err();err!=nil{return nil,fmt.Errorf("iterate owned characters: %w",err)}
	return players,nil
}



func (r *AccountRepository) lockOwnedPlayer(ctx context.Context, tx *sql.Tx, accountID string, id game.PlayerID) (game.Player,error) {
	query := "SELECT "+playerColumns+" FROM players WHERE account_id=$1 AND id=$2 FOR UPDATE"
	return scanPlayer(tx.QueryRowContext(ctx,query,accountID,string(id)))
}

func updateOwnedPlayer(ctx context.Context, tx *sql.Tx, accountID string, player game.Player) error {
	skills,err:=json.Marshal(player.Skills)
	if err!=nil{return fmt.Errorf("encode owned player skills: %w",err)}
	inventory,err:=json.Marshal(player.Inventory)
	if err!=nil{return fmt.Errorf("encode owned player inventory: %w",err)}
	bank,err:=json.Marshal(player.Bank)
	if err!=nil{return fmt.Errorf("encode owned player bank: %w",err)}
	result,err:=tx.ExecContext(ctx,`UPDATE players SET skills=$3,inventory=$4,bank=$5,coins=$6,zone_id=$7,position_x=$8,position_z=$9,last_move_at=$10 WHERE account_id=$1 AND id=$2`,
		accountID,string(player.ID),string(skills),string(inventory),string(bank),player.Coins,player.Position.ZoneID,player.Position.X,player.Position.Z,player.LastMoveAt)
	if err!=nil{return fmt.Errorf("update owned player: %w",err)}
	affected,err:=result.RowsAffected()
	if err!=nil{return fmt.Errorf("count updated owned player: %w",err)}
	if affected!=1{return game.ErrPlayerNotFound}
	return nil
}

func (r *AccountRepository) AwardExperienceForAccount(ctx context.Context,accountID string,id game.PlayerID,skillID string,amount int)(game.Player,bool,error){
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.Player{},false,err};defer tx.Rollback()
	player,err:=r.lockOwnedPlayer(ctx,tx,accountID,id);if err!=nil{return game.Player{},false,err}
	leveled,err:=player.AwardExperience(skillID,amount);if err!=nil{return game.Player{},false,err}
	if err:=updateOwnedPlayer(ctx,tx,accountID,player);err!=nil{return game.Player{},false,err}
	if err:=tx.Commit();err!=nil{return game.Player{},false,fmt.Errorf("commit owned XP award: %w",err)}
	return player,leveled,nil
}

func (r *AccountRepository) MovePlayerForAccount(ctx context.Context,accountID string,id game.PlayerID,x,z float64,now time.Time)(game.Player,error){
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.Player{},err};defer tx.Rollback()
	player,err:=r.lockOwnedPlayer(ctx,tx,accountID,id);if err!=nil{return game.Player{},err}
	if err:=player.MoveTo(x,z,now);err!=nil{return player,err}
	if err:=updateOwnedPlayer(ctx,tx,accountID,player);err!=nil{return game.Player{},err}
	if err:=tx.Commit();err!=nil{return game.Player{},fmt.Errorf("commit owned movement: %w",err)}
	return player,nil
}

func (r *AccountRepository) GatherForAccount(ctx context.Context,accountID string,id game.PlayerID,nodeID string,now time.Time)(game.Player,game.GatheringResult,error){
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.Player{},game.GatheringResult{},err};defer tx.Rollback()
	player,err:=r.lockOwnedPlayer(ctx,tx,accountID,id);if err!=nil{return game.Player{},game.GatheringResult{},err}
	node,ok:=game.FindGatheringNode(nodeID);if !ok{return player,game.GatheringResult{},game.ErrUnknownGatheringNode}
	if _,err:=tx.ExecContext(ctx,`INSERT INTO gathering_node_state (node_id,charges,initialized) VALUES ($1,0,FALSE) ON CONFLICT (node_id) DO NOTHING`,nodeID);err!=nil{return game.Player{},game.GatheringResult{},fmt.Errorf("initialize gathering node: %w",err)}
	var state game.GatheringNodeState
	var respawnAt sql.NullTime
	if err:=tx.QueryRowContext(ctx,`SELECT charges,respawn_at,initialized FROM gathering_node_state WHERE node_id=$1 FOR UPDATE`,nodeID).Scan(&state.Charges,&respawnAt,&state.Initialized);err!=nil{return game.Player{},game.GatheringResult{},fmt.Errorf("lock gathering node: %w",err)}
	if respawnAt.Valid{state.RespawnAt=respawnAt.Time}
	result,nextState,err:=player.Gather(node,state,now);if err!=nil{return player,game.GatheringResult{},err}
	var respawnValue any
	if !nextState.RespawnAt.IsZero(){respawnValue=nextState.RespawnAt}
	if _,err:=tx.ExecContext(ctx,`UPDATE gathering_node_state SET charges=$2,respawn_at=$3,initialized=$4 WHERE node_id=$1`,nodeID,nextState.Charges,respawnValue,nextState.Initialized);err!=nil{return game.Player{},game.GatheringResult{},fmt.Errorf("update gathering node: %w",err)}
	if err:=updateOwnedPlayer(ctx,tx,accountID,player);err!=nil{return game.Player{},game.GatheringResult{},err}
	if err:=tx.Commit();err!=nil{return game.Player{},game.GatheringResult{},fmt.Errorf("commit owned gathering: %w",err)}
	return player,result,nil
}

func (r *AccountRepository) CraftForAccount(ctx context.Context,accountID string,id game.PlayerID,recipeID string,now time.Time)(game.Player,game.CraftingResult,error){
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.Player{},game.CraftingResult{},err};defer tx.Rollback()
	player,err:=r.lockOwnedPlayer(ctx,tx,accountID,id);if err!=nil{return game.Player{},game.CraftingResult{},err}
	recipe,ok:=game.FindCraftingRecipe(recipeID);if !ok{return player,game.CraftingResult{},game.ErrUnknownCraftingRecipe}
	station,ok:=game.FindCraftingStation(recipe.StationID);if !ok{return player,game.CraftingResult{},game.ErrUnknownCraftingStation}
	result,err:=player.Craft(recipe,station,now);if err!=nil{return player,game.CraftingResult{},err}
	if err:=updateOwnedPlayer(ctx,tx,accountID,player);err!=nil{return game.Player{},game.CraftingResult{},err}
	if err:=tx.Commit();err!=nil{return game.Player{},game.CraftingResult{},fmt.Errorf("commit owned crafting: %w",err)}
	return player,result,nil
}

func createSession(ctx context.Context,tx *sql.Tx,accountID string,tokenHash []byte,now time.Time) error {
	if _,err:=tx.ExecContext(ctx,`INSERT INTO account_sessions (token_hash,account_id,created_at,expires_at,last_used_at) VALUES ($1,$2,$3,$4,$3)`,tokenHash,accountID,now,now.Add(SessionLifetime));err!=nil{return fmt.Errorf("insert account session: %w",err)}
	if _,err:=tx.ExecContext(ctx,`DELETE FROM account_sessions WHERE account_id=$1 AND token_hash IN (
		SELECT token_hash FROM account_sessions WHERE account_id=$1 AND revoked_at IS NULL AND expires_at>$2
		ORDER BY created_at DESC, token_hash DESC OFFSET $3
	)`,accountID,now,MaxActiveSessionsPerAccount);err!=nil{return fmt.Errorf("prune excess account sessions: %w",err)}
	if _,err:=tx.ExecContext(ctx,`DELETE FROM account_sessions WHERE account_id=$1 AND (expires_at<=$2 OR revoked_at<$2-INTERVAL '7 days')`,accountID,now);err!=nil{return fmt.Errorf("prune expired account sessions: %w",err)}
	return nil
}

func (r *AccountRepository) TransferBankItemForAccount(ctx context.Context, accountID string, id game.PlayerID, itemID string, quantity int, deposit bool) (game.Player, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return game.Player{}, err }
	defer tx.Rollback()
	player, err := r.lockOwnedPlayer(ctx, tx, accountID, id)
	if err != nil { return game.Player{}, err }
	if deposit { err = player.DepositItem(itemID, quantity) } else { err = player.WithdrawItem(itemID, quantity) }
	if err != nil { return player, err }
	if err := updateOwnedPlayer(ctx, tx, accountID, player); err != nil { return game.Player{}, err }
	if err := tx.Commit(); err != nil { return game.Player{}, fmt.Errorf("commit owned bank transfer: %w", err) }
	return player, nil
}

func (r *AccountRepository) TradeItemForAccount(ctx context.Context,accountID string,id game.PlayerID,itemID string,quantity int,buy bool)(game.Player,int64,error){
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.Player{},0,err};defer tx.Rollback()
	player,err:=r.lockOwnedPlayer(ctx,tx,accountID,id);if err!=nil{return game.Player{},0,err}
	var amount int64
	if buy { amount,err=player.BuyItem(itemID,quantity) } else { amount,err=player.SellItem(itemID,quantity) }
	if err!=nil{return player,0,err}
	if err:=updateOwnedPlayer(ctx,tx,accountID,player);err!=nil{return game.Player{},0,err}
	if err:=tx.Commit();err!=nil{return game.Player{},0,fmt.Errorf("commit owned market trade: %w",err)}
	return player,amount,nil
}
