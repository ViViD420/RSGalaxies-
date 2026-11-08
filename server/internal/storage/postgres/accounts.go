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
	now = now.UTC()
	// Authentication is checked on every API request, but persisting last_used_at
	// on every request creates avoidable write amplification. Touch a session at
	// most once per minute while still returning the account on every valid call.
	err = r.db.QueryRowContext(ctx, `WITH touched AS (
		UPDATE account_sessions
		SET last_used_at=$2
		WHERE token_hash=$1 AND revoked_at IS NULL AND expires_at>$2
			AND last_used_at <= $2 - INTERVAL '1 minute'
		RETURNING token_hash
	)
	SELECT a.id,a.email,a.created_at
	FROM account_sessions s
	JOIN accounts a ON a.id=s.account_id
	WHERE s.token_hash=$1 AND s.revoked_at IS NULL
		AND s.expires_at>$2 AND a.disabled_at IS NULL`, tokenHash[:], now).
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
	equipment, err := json.Marshal(player.Equipment)
	if err != nil { return fmt.Errorf("encode character equipment: %w", err) }
	_, err = r.db.ExecContext(ctx,`INSERT INTO players
		(id,account_id,name,created_at,skills,inventory,bank,coins,equipment,zone_id,position_x,position_z,last_move_at,health,last_combat_at)
		VALUES ($1,$2,$3,$4,$5::jsonb,$6::jsonb,$7::jsonb,$8,$9::jsonb,$10,$11,$12,$13,$14,$15)`,
		string(player.ID),accountID,player.Name,player.CreatedAt,string(skills),string(inventory),string(bank),player.Coins,string(equipment),
		player.Position.ZoneID,player.Position.X,player.Position.Z,player.LastMoveAt,player.Health,player.LastCombatAt)
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
	if err := validateTradeReservations(ctx, tx, accountID, player); err != nil { return err }
	skills,err:=json.Marshal(player.Skills)
	if err!=nil{return fmt.Errorf("encode owned player skills: %w",err)}
	inventory,err:=json.Marshal(player.Inventory)
	if err!=nil{return fmt.Errorf("encode owned player inventory: %w",err)}
	bank,err:=json.Marshal(player.Bank)
	if err!=nil{return fmt.Errorf("encode owned player bank: %w",err)}
	equipment,err:=json.Marshal(player.Equipment)
	if err!=nil{return fmt.Errorf("encode owned player equipment: %w",err)}
	result,err:=tx.ExecContext(ctx,`UPDATE players SET skills=$3,inventory=$4,bank=$5,coins=$6,equipment=$7,zone_id=$8,position_x=$9,position_z=$10,last_move_at=$11,health=$12,last_combat_at=$13 WHERE account_id=$1 AND id=$2`,
		accountID,string(player.ID),string(skills),string(inventory),string(bank),player.Coins,string(equipment),player.Position.ZoneID,player.Position.X,player.Position.Z,player.LastMoveAt,player.Health,player.LastCombatAt)
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
	if err:=upsertZonePresence(ctx,tx,accountID,player,now);err!=nil{return game.Player{},err}
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

func (r *AccountRepository) EquipItemForAccount(ctx context.Context,accountID string,id game.PlayerID,itemID string)(game.Player,error){
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.Player{},err};defer tx.Rollback()
	player,err:=r.lockOwnedPlayer(ctx,tx,accountID,id);if err!=nil{return game.Player{},err}
	if err:=player.EquipItem(itemID);err!=nil{return player,err}
	if err:=updateOwnedPlayer(ctx,tx,accountID,player);err!=nil{return game.Player{},err}
	if err:=tx.Commit();err!=nil{return game.Player{},fmt.Errorf("commit equipment change: %w",err)}
	return player,nil
}

func (r *AccountRepository) UnequipItemForAccount(ctx context.Context,accountID string,id game.PlayerID,slot string)(game.Player,error){
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.Player{},err};defer tx.Rollback()
	player,err:=r.lockOwnedPlayer(ctx,tx,accountID,id);if err!=nil{return game.Player{},err}
	if err:=player.UnequipItem(slot);err!=nil{return player,err}
	if err:=updateOwnedPlayer(ctx,tx,accountID,player);err!=nil{return game.Player{},err}
	if err:=tx.Commit();err!=nil{return game.Player{},fmt.Errorf("commit equipment change: %w",err)}
	return player,nil
}

func (r *AccountRepository) AttackEnemyForAccount(ctx context.Context,accountID string,id game.PlayerID,enemyID string,now time.Time)(game.Player,game.CombatResult,error){
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.Player{},game.CombatResult{},err};defer tx.Rollback()
	player,err:=r.lockOwnedPlayer(ctx,tx,accountID,id);if err!=nil{return game.Player{},game.CombatResult{},err}
	enemy,ok:=game.FindEnemy(enemyID);if !ok{return player,game.CombatResult{},game.ErrUnknownEnemy}
	if _,err:=tx.ExecContext(ctx,`INSERT INTO enemy_spawn_state(enemy_id,health,initialized) VALUES($1,0,FALSE) ON CONFLICT(enemy_id) DO NOTHING`,enemyID);err!=nil{return game.Player{},game.CombatResult{},fmt.Errorf("initialize enemy state: %w",err)}
	var state game.EnemyState
	var respawnAt sql.NullTime
	if err:=tx.QueryRowContext(ctx,`SELECT health,respawn_at,initialized FROM enemy_spawn_state WHERE enemy_id=$1 FOR UPDATE`,enemyID).Scan(&state.Health,&respawnAt,&state.Initialized);err!=nil{return game.Player{},game.CombatResult{},fmt.Errorf("lock enemy state: %w",err)}
	if respawnAt.Valid{state.RespawnAt=respawnAt.Time}
	result,nextState,err:=player.AttackEnemy(enemy,state,now)
	if err!=nil{return player,game.CombatResult{},err}
	var respawnValue any
	if !nextState.RespawnAt.IsZero(){respawnValue=nextState.RespawnAt}
	if _,err:=tx.ExecContext(ctx,`UPDATE enemy_spawn_state SET health=$2,respawn_at=$3,initialized=$4 WHERE enemy_id=$1`,enemyID,nextState.Health,respawnValue,nextState.Initialized);err!=nil{return game.Player{},game.CombatResult{},fmt.Errorf("update enemy state: %w",err)}
	if err:=updateOwnedPlayer(ctx,tx,accountID,player);err!=nil{return game.Player{},game.CombatResult{},err}
	if err:=tx.Commit();err!=nil{return game.Player{},game.CombatResult{},fmt.Errorf("commit combat action: %w",err)}
	return player,result,nil
}

func (r *AccountRepository) RestForAccount(ctx context.Context,accountID string,id game.PlayerID)(game.Player,error){
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.Player{},err};defer tx.Rollback()
	player,err:=r.lockOwnedPlayer(ctx,tx,accountID,id);if err!=nil{return game.Player{},err}
	if err:=player.Rest();err!=nil{return player,err}
	if err:=updateOwnedPlayer(ctx,tx,accountID,player);err!=nil{return game.Player{},err}
	if err:=tx.Commit();err!=nil{return game.Player{},fmt.Errorf("commit player rest: %w",err)}
	return player,nil
}

func (r *AccountRepository) SendLocalChatForAccount(ctx context.Context,accountID string,id game.PlayerID,message string,now time.Time)(game.ChatMessage,error){
	cleaned,err:=game.ValidateChatMessage(message);if err!=nil{return game.ChatMessage{},err}
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.ChatMessage{},err};defer tx.Rollback()
	player,err:=r.lockOwnedPlayer(ctx,tx,accountID,id);if err!=nil{return game.ChatMessage{},err}
	var lastMessageAt sql.NullTime
	err=tx.QueryRowContext(ctx,`SELECT created_at FROM chat_messages WHERE sender_character_id=$1 ORDER BY id DESC LIMIT 1`,string(player.ID)).Scan(&lastMessageAt)
	if err!=nil&&!errors.Is(err,sql.ErrNoRows){return game.ChatMessage{},fmt.Errorf("check local chat cooldown: %w",err)}
	if lastMessageAt.Valid&&now.Sub(lastMessageAt.Time)<game.ChatCooldown{return game.ChatMessage{},game.ErrChatCooldown}
	var result game.ChatMessage
	err=tx.QueryRowContext(ctx,`INSERT INTO chat_messages(zone_id,sender_account_id,sender_character_id,sender_character_name,message,created_at)
		VALUES($1,$2,$3,$4,$5,$6) RETURNING id,zone_id,sender_character_id,sender_character_name,message,created_at`,
		player.Position.ZoneID,accountID,string(player.ID),player.Name,cleaned,now).Scan(&result.ID,&result.ZoneID,&result.CharacterID,&result.CharacterName,&result.Message,&result.CreatedAt)
	if err!=nil{return game.ChatMessage{},fmt.Errorf("insert local chat message: %w",err)}
	if err:=tx.Commit();err!=nil{return game.ChatMessage{},fmt.Errorf("commit local chat message: %w",err)}
	return result,nil
}

func (r *AccountRepository) ListLocalChatForAccount(ctx context.Context,accountID string,id game.PlayerID,afterID int64,limit int)([]game.ChatMessage,error){
	if afterID<0||limit<1||limit>50{return nil,game.ErrInvalidChatCursor}
	var zoneID string
	err:=r.db.QueryRowContext(ctx,`SELECT zone_id FROM players WHERE account_id=$1 AND id=$2`,accountID,string(id)).Scan(&zoneID)
	if errors.Is(err,sql.ErrNoRows){return nil,game.ErrPlayerNotFound}
	if err!=nil{return nil,fmt.Errorf("resolve chat character zone: %w",err)}
	var rows *sql.Rows
	if afterID==0 {
		// A new client should see the newest bounded history, not the oldest
		// rows in the zone (which could leave it permanently behind the live cursor).
		rows,err=r.db.QueryContext(ctx,`SELECT id,zone_id,sender_character_id,sender_character_name,message,created_at
			FROM (SELECT id,zone_id,sender_character_id,sender_character_name,message,created_at
				FROM chat_messages WHERE zone_id=$1 ORDER BY id DESC LIMIT $2) recent
			ORDER BY id ASC`,zoneID,limit)
	} else {
		rows,err=r.db.QueryContext(ctx,`SELECT id,zone_id,sender_character_id,sender_character_name,message,created_at
			FROM chat_messages WHERE zone_id=$1 AND id>$2 ORDER BY id ASC LIMIT $3`,zoneID,afterID,limit)
	}
	if err!=nil{return nil,fmt.Errorf("list local chat messages: %w",err)}
	defer rows.Close()
	messages:=make([]game.ChatMessage,0)
	for rows.Next(){
		var message game.ChatMessage
		if err:=rows.Scan(&message.ID,&message.ZoneID,&message.CharacterID,&message.CharacterName,&message.Message,&message.CreatedAt);err!=nil{return nil,fmt.Errorf("scan local chat message: %w",err)}
		messages=append(messages,message)
	}
	if err:=rows.Err();err!=nil{return nil,fmt.Errorf("iterate local chat messages: %w",err)}
	return messages,nil
}

func (r *AccountRepository) UseItemForAccount(ctx context.Context,accountID string,id game.PlayerID,itemID string)(game.Player,int,error){
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.Player{},0,err};defer tx.Rollback()
	player,err:=r.lockOwnedPlayer(ctx,tx,accountID,id);if err!=nil{return game.Player{},0,err}
	healed,err:=player.UseItem(itemID);if err!=nil{return player,0,err}
	if err:=updateOwnedPlayer(ctx,tx,accountID,player);err!=nil{return game.Player{},0,err}
	if err:=tx.Commit();err!=nil{return game.Player{},0,fmt.Errorf("commit item use: %w",err)}
	return player,healed,nil
}

func (r *AccountRepository) HeartbeatPresenceForAccount(ctx context.Context,accountID string,id game.PlayerID,now time.Time) error {
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return err};defer tx.Rollback()
	player,err:=r.lockOwnedPlayer(ctx,tx,id);if err!=nil{return err}
	if player.AccountID != accountID { return game.ErrPlayerNotFound }
	if err:=upsertZonePresence(ctx,tx,accountID,player,now);err!=nil{return err}
	if _,err:=tx.ExecContext(ctx,`DELETE FROM zone_presence WHERE last_seen_at < $1`,now.UTC().Add(-5*time.Minute));err!=nil{return fmt.Errorf("prune stale character presence: %w",err)}
	if err:=tx.Commit();err!=nil{return fmt.Errorf("commit character presence: %w",err)}
	return nil
}

func upsertZonePresence(ctx context.Context,tx *sql.Tx,accountID string,player game.Player,now time.Time) error {
	_,err:=tx.ExecContext(ctx,`INSERT INTO zone_presence(character_id,account_id,zone_id,position_x,position_z,last_seen_at)
		VALUES($1,$2,$3,$4,$5,$6)
		ON CONFLICT(character_id) DO UPDATE SET account_id=EXCLUDED.account_id,zone_id=EXCLUDED.zone_id,
		position_x=EXCLUDED.position_x,position_z=EXCLUDED.position_z,last_seen_at=EXCLUDED.last_seen_at
		WHERE zone_presence.last_seen_at <= EXCLUDED.last_seen_at - INTERVAL '1 second'`,
		string(player.ID),accountID,player.Position.ZoneID,player.Position.X,player.Position.Z,now.UTC())
	if err!=nil{return fmt.Errorf("upsert character presence: %w",err)}
	return nil
}

func (r *AccountRepository) ListZonePresenceForAccount(ctx context.Context,accountID string,id game.PlayerID,now time.Time)([]game.RemotePlayer,error){
	var zoneID string
	err:=r.db.QueryRowContext(ctx,`SELECT zone_id FROM players WHERE account_id=$1 AND id=$2`,accountID,string(id)).Scan(&zoneID)
	if errors.Is(err,sql.ErrNoRows){return nil,game.ErrPlayerNotFound}
	if err!=nil{return nil,fmt.Errorf("resolve presence zone: %w",err)}
	rows,err:=r.db.QueryContext(ctx,`SELECT players.id,players.name,zone_presence.zone_id,zone_presence.position_x,zone_presence.position_z,zone_presence.last_seen_at
		FROM players JOIN zone_presence ON zone_presence.character_id=players.id
		WHERE zone_presence.zone_id=$1 AND zone_presence.character_id<>$2
		AND zone_presence.last_seen_at>$3
		ORDER BY zone_presence.last_seen_at DESC,players.id ASC LIMIT $4`,
		zoneID,string(id),now.UTC().Add(-game.PresenceLease),game.MaxZonePresence)
	if err!=nil{return nil,fmt.Errorf("list active zone presence: %w",err)}
	defer rows.Close()
	result:=make([]game.RemotePlayer,0)
	for rows.Next(){
		var remote game.RemotePlayer
		if err:=rows.Scan(&remote.CharacterID,&remote.Name,&remote.ZoneID,&remote.X,&remote.Z,&remote.LastSeenAt);err!=nil{return nil,fmt.Errorf("scan remote character presence: %w",err)}
		remote.LastSeenAt=remote.LastSeenAt.UTC()
		result=append(result,remote)
	}
	if err:=rows.Err();err!=nil{return nil,fmt.Errorf("iterate zone presence: %w",err)}
	return result,nil
}

type tradeQueryer interface {
	QueryRowContext(context.Context,string,...any)*sql.Row
	QueryContext(context.Context,string,...any)(*sql.Rows,error)
}
