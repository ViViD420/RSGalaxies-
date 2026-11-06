package game

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidPlayerName = errors.New("player name must be 3-20 characters")
	ErrPlayerNotFound    = errors.New("player not found")
)

type PlayerID string

// Player is the server-owned character identity used by the first development slice.
// Gameplay state will be expanded in later milestones.
type Player struct {
	ID        PlayerID  `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewPlayer(id PlayerID, name string, now time.Time) (Player, error) {
	name = strings.TrimSpace(name)
	if len([]rune(name)) < 3 || len([]rune(name)) > 20 {
		return Player{}, ErrInvalidPlayerName
	}
	if id == "" {
		return Player{}, errors.New("player id is required")
	}
	return Player{ID: id, Name: name, CreatedAt: now.UTC()}, nil
}
