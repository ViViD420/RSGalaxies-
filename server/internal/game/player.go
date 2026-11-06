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

type SkillProgress struct {
	Experience int `json:"experience"`
	Level      int `json:"level"`
}

// Player is the server-owned character identity and progression state.
type Player struct {
	ID        PlayerID                `json:"id"`
	Name      string                  `json:"name"`
	CreatedAt time.Time               `json:"createdAt"`
	Skills    map[string]SkillProgress `json:"skills"`
}

func NewPlayer(id PlayerID, name string, now time.Time) (Player, error) {
	name = strings.TrimSpace(name)
	if len([]rune(name)) < 3 || len([]rune(name)) > 20 { return Player{}, ErrInvalidPlayerName }
	if id == "" { return Player{}, errors.New("player id is required") }
	return Player{ID: id, Name: name, CreatedAt: now.UTC(), Skills: make(map[string]SkillProgress, len(skillDefinitions))}, nil
}

// AwardExperience applies a validated server-side award and reports a level-up.
func (p *Player) AwardExperience(skillID string, amount int) (bool, error) {
	skillID = strings.TrimSpace(strings.ToLower(skillID))
	if !IsKnownSkill(skillID) { return false, ErrUnknownSkill }
	if amount <= 0 { return false, ErrInvalidExperience }
	current := p.Skills[skillID]
	if current.Experience > MaxSkillExperience-amount { return false, ErrExperienceOverflow }
	nextXP := current.Experience + amount
	nextLevel := LevelForExperience(nextXP)
	leveled := nextLevel > current.Level
	p.Skills[skillID] = SkillProgress{Experience: nextXP, Level: nextLevel}
	return leveled, nil
}
