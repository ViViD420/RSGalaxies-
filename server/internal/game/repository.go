package game

import (
	"context"
	"sort"
	"sync"
)

type PlayerRepository interface {
	Create(ctx context.Context, player Player) error
	Get(ctx context.Context, id PlayerID) (Player, error)
	List(ctx context.Context) ([]Player, error)
	Update(ctx context.Context, player Player) error
}

// MemoryPlayerRepository is for local development and tests only.
type MemoryPlayerRepository struct {
	mu sync.RWMutex
	players map[PlayerID]Player
}

func NewMemoryPlayerRepository() *MemoryPlayerRepository {
	return &MemoryPlayerRepository{players: make(map[PlayerID]Player)}
}

func (r *MemoryPlayerRepository) Create(_ context.Context, player Player) error {
	r.mu.Lock(); defer r.mu.Unlock()
	if _, exists := r.players[player.ID]; exists { return ErrPlayerAlreadyExists }
	r.players[player.ID] = clonePlayer(player)
	return nil
}

func (r *MemoryPlayerRepository) Get(_ context.Context, id PlayerID) (Player, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	player, ok := r.players[id]
	if !ok { return Player{}, ErrPlayerNotFound }
	return clonePlayer(player), nil
}

func (r *MemoryPlayerRepository) List(_ context.Context) ([]Player, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	players := make([]Player, 0, len(r.players))
	for _, player := range r.players { players = append(players, clonePlayer(player)) }
	sort.Slice(players, func(i, j int) bool { return players[i].ID < players[j].ID })
	return players, nil
}

func (r *MemoryPlayerRepository) Update(_ context.Context, player Player) error {
	r.mu.Lock(); defer r.mu.Unlock()
	if _, exists := r.players[player.ID]; !exists { return ErrPlayerNotFound }
	r.players[player.ID] = clonePlayer(player)
	return nil
}

func clonePlayer(player Player) Player {
	cloned := player
	cloned.Skills = make(map[string]SkillProgress, len(player.Skills))
	for id, progress := range player.Skills { cloned.Skills[id] = progress }
	return cloned
}
