package game

import (
	"context"
	"sync"
)

type PlayerRepository interface {
	Create(ctx context.Context, player Player) error
	Get(ctx context.Context, id PlayerID) (Player, error)
	List(ctx context.Context) ([]Player, error)
}

// MemoryPlayerRepository is for local development and tests only.
// Replace it with a transactional PostgreSQL implementation before persistent alpha.
type MemoryPlayerRepository struct {
	mu      sync.RWMutex
	players map[PlayerID]Player
}

func NewMemoryPlayerRepository() *MemoryPlayerRepository {
	return &MemoryPlayerRepository{players: make(map[PlayerID]Player)}
}

func (r *MemoryPlayerRepository) Create(_ context.Context, player Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.players[player.ID]; exists {
		return ErrPlayerAlreadyExists
	}
	r.players[player.ID] = player
	return nil
}

func (r *MemoryPlayerRepository) Get(_ context.Context, id PlayerID) (Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	player, ok := r.players[id]
	if !ok {
		return Player{}, ErrPlayerNotFound
	}
	return player, nil
}

func (r *MemoryPlayerRepository) List(_ context.Context) ([]Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	players := make([]Player, 0, len(r.players))
	for _, player := range r.players {
		players = append(players, player)
	}
	return players, nil
}
