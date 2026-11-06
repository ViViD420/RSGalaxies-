// Package auth contains credential and bearer-session primitives.
// Database persistence and HTTP policy are deliberately kept in separate packages.
package auth

import (
	"context"
	"time"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

const (
	passwordIterations = 600_000
	passwordSaltBytes  = 16
	passwordKeyBytes   = 32
	sessionTokenBytes  = 32
	maxPasswordBytes   = 1024
)

var (
	ErrAccountExists = errors.New("account already exists")
	ErrAccountNotFound = errors.New("account not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidEmail = errors.New("invalid email address")
	ErrSessionNotFound = errors.New("session not found or expired")
	ErrInvalidPassword = errors.New("password must contain 12-1024 bytes")
	ErrInvalidHash     = errors.New("invalid password hash")
)

// HashPassword uses PBKDF2-HMAC-SHA-256 with a fresh random salt.
// The encoded value is safe to store; the original password is never retained.
func HashPassword(password string) (string, error) {
	if len(password) < 12 || len(password) > maxPasswordBytes {
		return "", ErrInvalidPassword
	}
	salt := make([]byte, passwordSaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	key := pbkdf2SHA256([]byte(password), salt, passwordIterations, passwordKeyBytes)
	return strings.Join([]string{
		"rsg", "pbkdf2-sha256", strconv.Itoa(passwordIterations),
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	}, "$"), nil
}

// VerifyPassword validates a password hash in constant time.
func VerifyPassword(encoded, password string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 || parts[0] != "rsg" || parts[1] != "pbkdf2-sha256" {
		return false, ErrInvalidHash
	}
	iterations, err := strconv.Atoi(parts[2])
	if err != nil || iterations < 100_000 || iterations > 2_000_000 {
		return false, ErrInvalidHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(salt) != passwordSaltBytes {
		return false, ErrInvalidHash
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(expected) != passwordKeyBytes {
		return false, ErrInvalidHash
	}
	actual := pbkdf2SHA256([]byte(password), salt, iterations, len(expected))
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

// NewSessionToken returns a high-entropy bearer token and its SHA-256 digest.
// Persist only the digest; return the raw token to the client once.
func NewSessionToken() (token string, tokenHash [32]byte, err error) {
	raw := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", tokenHash, fmt.Errorf("generate session token: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	tokenHash = sha256.Sum256([]byte(token))
	return token, tokenHash, nil
}

// HashSessionToken computes the lookup digest for a presented bearer token.
func HashSessionToken(token string) ([32]byte, error) {
	var zero [32]byte
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != sessionTokenBytes {
		return zero, errors.New("invalid session token")
	}
	return sha256.Sum256([]byte(token)), nil
}

func pbkdf2SHA256(password, salt []byte, iterations, keyLength int) []byte {
	const hashLength = sha256.Size
	blocks := (keyLength + hashLength - 1) / hashLength
	result := make([]byte, 0, blocks*hashLength)
	var counter [4]byte
	for block := 1; block <= blocks; block++ {
		counter[0] = byte(block >> 24)
		counter[1] = byte(block >> 16)
		counter[2] = byte(block >> 8)
		counter[3] = byte(block)
		mac := hmac.New(sha256.New, password)
		_, _ = mac.Write(salt)
		_, _ = mac.Write(counter[:])
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)
		for i := 1; i < iterations; i++ {
			mac = hmac.New(sha256.New, password)
			_, _ = mac.Write(u)
			u = mac.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		result = append(result, t...)
	}
	return result[:keyLength]
}


type Account struct {
	ID string `json:"id"`
	Email string `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type AccountStore interface {
	Register(ctx context.Context, email, password string, now time.Time) (Account, string, error)
	Login(ctx context.Context, email, password string, now time.Time) (Account, string, error)
	AuthenticateSession(ctx context.Context, token string, now time.Time) (Account, error)
	RevokeSession(ctx context.Context, token string, now time.Time) error
}
