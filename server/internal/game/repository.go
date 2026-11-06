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
	AwardExperience(ctx context.Context, id PlayerID, skillID string, amount int) (Player, bool, error)
}

type MemoryPlayerRepository struct { mu sync.RWMutex; players map[PlayerID]Player }

func NewMemoryPlayerRepository() *MemoryPlayerRepository {
	return &MemoryPlayerRepository{players:make(map[PlayerID]Player)}
}

func (r *MemoryPlayerRepository) Create(_ context.Context, player Player) error {
	r.mu.Lock(); defer r.mu.Unlock()
	if _, ok := r.players[player.ID]; ok { return ErrPlayerAlreadyExists }
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
	sort.Slice(players, func(i,j int) bool { return players[i].ID < players[j].ID })
	return players, nil
}

// AwardExperience is atomic here; persistent implementations must use a database transaction.
func (r *MemoryPlayerRepository) AwardExperience(_ context.Context, id PlayerID, skillID string, amount int) (Player, bool, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	player, ok := r.players[id]
	if !ok { return Player{}, false, ErrPlayerNotFound }
	leveled, err := player.AwardExperience(skillID, amount)
	if err != nil { return Player{}, false, err }
	r.players[id] = clonePlayer(player)
	return clonePlayer(player), leveled, nil
}

func clonePlayer(player Player) Player {
	cloned := player
	cloned.Skills = make(map[string]SkillProgress, len(player.Skills))
	for id, progress := range player.Skills { cloned.Skills[id] = progress }
	return cloned
}
