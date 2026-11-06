package postgres

import (
	"context"
	"database/sql"
	"errors"
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ViViD420/RSGalaxies-/server/internal/auth"
	"github.com/ViViD420/RSGalaxies-/server/internal/game"
)

const SessionLifetime = 30 * 24 * time.Hour

var (
	ErrAccountExists = errors.New("account already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidEmail = errors.New("invalid email address")
	ErrSessionNotFound = errors.New("session not found or expired")
)

type Account struct {
	ID string
	Email string
	CreatedAt time.Time
}

type AccountRepository struct { db *sql.DB }

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
	id, err := newRandomID()
	if err != nil { return "", err }
	return id, nil
}

func (r *AccountRepository) Register(ctx context.Context, email, password string, now time.Time) (Account, string, error) {
	email, err := normalizeEmail(email)
	if err != nil { return Account{}, "", err }
	passwordHash, err := auth.HashPassword(password)
	if err != nil { return Account{}, "", err }
	id, err := newAccountID()
	if err != nil { return Account{}, "", fmt.Errorf("generate account id: %w", err) }
	token, tokenHash, err := auth.NewSessionToken()
	if err != nil { return Account{}, "", err }
	now = now.UTC()
	expires := now.Add(SessionLifetime)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return Account{}, "", fmt.Errorf("begin account registration: %w", err) }
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO accounts (id,email,password_hash,created_at) VALUES ($1,$2,$3,$4)`, id, email, passwordHash, now)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { return Account{}, "", ErrAccountExists }
		return Account{}, "", fmt.Errorf("insert account: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO account_sessions (token_hash,account_id,created_at,expires_at,last_used_at) VALUES ($1,$2,$3,$4,$3)`, tokenHash[:], id, now, expires); err != nil {
		return Account{}, "", fmt.Errorf("create initial session: %w", err)
	}
	if err := tx.Commit(); err != nil { return Account{}, "", fmt.Errorf("commit account registration: %w", err) }
	return Account{ID:id, Email:email, CreatedAt:now}, token, nil
}

func (r *AccountRepository) Login(ctx context.Context, email, password string, now time.Time) (Account, string, error) {
	email, err := normalizeEmail(email)
	if err != nil { return Account{}, "", ErrInvalidCredentials }
	var account Account
	var passwordHash string
	var disabledAt sql.NullTime
	err = r.db.QueryRowContext(ctx, `SELECT id,email,password_hash,created_at,disabled_at FROM accounts WHERE LOWER(email)=$1`, email).
		Scan(&account.ID,&account.Email,&passwordHash,&account.CreatedAt,&disabledAt)
	if err != nil { return Account{}, "", ErrInvalidCredentials }
	if disabledAt.Valid { return Account{}, "", ErrInvalidCredentials }
	valid, err := auth.VerifyPassword(passwordHash,password)
	if err != nil || !valid { return Account{}, "", ErrInvalidCredentials }
	token, tokenHash, err := auth.NewSessionToken()
	if err != nil { return Account{}, "", err }
	now = now.UTC()
	_, err = r.db.ExecContext(ctx, `INSERT INTO account_sessions (token_hash,account_id,created_at,expires_at,last_used_at) VALUES ($1,$2,$3,$4,$3)`,
		tokenHash[:],account.ID,now,now.Add(SessionLifetime))
	if err != nil { return Account{}, "", fmt.Errorf("create login session: %w", err) }
	account.CreatedAt = account.CreatedAt.UTC()
	return account, token, nil
}

func (r *AccountRepository) AuthenticateSession(ctx context.Context, token string, now time.Time) (Account, error) {
	tokenHash, err := auth.HashSessionToken(token)
	if err != nil { return Account{}, ErrSessionNotFound }
	var account Account
	err = r.db.QueryRowContext(ctx, `UPDATE account_sessions s SET last_used_at=$2
		FROM accounts a
		WHERE s.token_hash=$1 AND s.account_id=a.id AND s.revoked_at IS NULL
		AND s.expires_at > $2 AND a.disabled_at IS NULL
		RETURNING a.id,a.email,a.created_at`, tokenHash[:], now.UTC()).
		Scan(&account.ID,&account.Email,&account.CreatedAt)
	if errors.Is(err,sql.ErrNoRows) { return Account{}, ErrSessionNotFound }
	if err != nil { return Account{}, fmt.Errorf("authenticate session: %w",err) }
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
	_, err := r.db.ExecContext(ctx,`INSERT INTO players
		(id,account_id,name,created_at,skills,inventory,zone_id,position_x,position_z,last_move_at)
		VALUES ($1,$2,$3,$4,$5::jsonb,$6::jsonb,$7,$8,$9,$10)`,
		string(player.ID),accountID,player.Name,player.CreatedAt,marshalJSON(player.Skills),marshalJSON(player.Inventory),
		player.Position.ZoneID,player.Position.X,player.Position.Z,player.LastMoveAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err,&pgErr) {
			switch pgErr.Code {
			case "23505": return game.ErrPlayerAlreadyExists
			case "23503": return ErrSessionNotFound
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

func marshalJSON(value any) string {
	encoded,err:=jsonMarshal(value)
	if err!=nil { panic("JSON encoding of validated game state failed: "+err.Error()) }
	return string(encoded)
}

var jsonMarshal = func(value any) ([]byte,error) { return json.Marshal(value) }

var _ = sql.ErrNoRows
