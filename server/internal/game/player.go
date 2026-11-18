package game

import (
	"errors"
	"strings"
	"time"
)

const (
	MaxPlayerAction = 100
	MaxPlayerMind   = 100

	BasePlayerConstitution = 10
	BasePlayerStamina      = 10
	BasePlayerWillpower    = 10
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

// SecondaryAttributes represents the derived secondary statistics governing
// Health, Action, and Mind regeneration rates and pool ceilings:
//   - Constitution governs Health regeneration and maximum Health bonus
//   - Stamina governs Action regeneration and maximum Action bonus
//   - Willpower governs Mind regeneration and maximum Mind bonus
type SecondaryAttributes struct {
	Constitution int `json:"constitution"`
	Stamina      int `json:"stamina"`
	Willpower    int `json:"willpower"`
}

// VitalPools summarizes a character's effective primary pools (Health, Action,
// Mind), their derived SecondaryAttributes, and per-tick regeneration rates.
type VitalPools struct {
	Health           int                 `json:"health"`
	MaxHealth        int                 `json:"maxHealth"`
	Action           int                 `json:"action"`
	MaxAction        int                 `json:"maxAction"`
	Mind             int                 `json:"mind"`
	MaxMind          int                 `json:"maxMind"`
	Secondary        SecondaryAttributes `json:"secondary"`
	HealthRegenPer5s int                 `json:"healthRegenPer5s"`
	ActionRegenPer5s int                 `json:"actionRegenPer5s"`
	MindRegenPer5s   int                 `json:"mindRegenPer5s"`
}

type Player struct {
	ID            PlayerID                 `json:"id"`
	Name          string                   `json:"name"`
	CreatedAt     time.Time                `json:"createdAt"`
	Skills        map[string]SkillProgress `json:"skills"`
	Position      WorldPosition            `json:"position"`
	Inventory     map[string]int           `json:"inventory"`
	Bank          map[string]int           `json:"bank"`
	Coins         int64                    `json:"coins"`
	Equipment     map[string]string        `json:"equipment"`
	Health        int                      `json:"health"`
	Action        int                      `json:"action"`
	Mind          int                      `json:"mind"`
	MaxHealthBuff int                      `json:"maxHealthBuff,omitempty"`
	MaxActionBuff int                      `json:"maxActionBuff,omitempty"`
	MaxMindBuff   int                      `json:"maxMindBuff,omitempty"`
	Constitution  int                      `json:"constitution"`
	Stamina       int                      `json:"stamina"`
	Willpower     int                      `json:"willpower"`
	BuffExpiresAt time.Time                `json:"buffExpiresAt,omitempty"`
	LastCombatAt  time.Time                `json:"-"`
	LastMoveAt    time.Time                `json:"-"`
}

func NewPlayer(id PlayerID, name string, now time.Time) (Player, error) {
	name = strings.TrimSpace(name)
	if len([]rune(name)) < 3 || len([]rune(name)) > 20 {
		return Player{}, ErrInvalidPlayerName
	}
	if id == "" {
		return Player{}, errors.New("player id is required")
	}
	skills := make(map[string]SkillProgress, len(skillDefinitions))
	for _, definition := range skillDefinitions {
		skills[definition.ID] = SkillProgress{Level: 1}
	}
	return Player{
		ID:           id,
		Name:         name,
		CreatedAt:    now.UTC(),
		Skills:       skills,
		Inventory:    make(map[string]int),
		Bank:         make(map[string]int),
		Coins:        StartingCoins,
		Equipment:    make(map[string]string),
		Health:       MaxPlayerHealth,
		Action:       MaxPlayerAction,
		Mind:         MaxPlayerMind,
		Constitution: BasePlayerConstitution,
		Stamina:      BasePlayerStamina,
		Willpower:    BasePlayerWillpower,
		LastCombatAt: now.UTC(),
		Position:     worldZones[0].Spawn,
		LastMoveAt:   now.UTC(),
	}, nil
}

// EnsureVitalDefaults backfills Action, Mind, and SecondaryAttributes on legacy
// persisted characters or test fixtures without altering existing Health.
func (p *Player) EnsureVitalDefaults() {
	if p.Constitution <= 0 {
		p.Constitution = BasePlayerConstitution
	}
	if p.Stamina <= 0 {
		p.Stamina = BasePlayerStamina
	}
	if p.Willpower <= 0 {
		p.Willpower = BasePlayerWillpower
	}
	if p.Action <= 0 && p.Health > 0 {
		p.Action = p.EffectiveMaxAction()
	}
	if p.Mind <= 0 && p.Health > 0 {
		p.Mind = p.EffectiveMaxMind()
	}
}

// ExpireBuffsIfNeeded clears active attribute buffs once BuffExpiresAt has passed
// and clamps pools to their resulting effective maximums.
func (p *Player) ExpireBuffsIfNeeded(now time.Time) {
	if !p.BuffExpiresAt.IsZero() && !now.Before(p.BuffExpiresAt) {
		p.MaxHealthBuff = 0
		p.MaxActionBuff = 0
		p.MaxMindBuff = 0
		p.BuffExpiresAt = time.Time{}
		if maxH := p.EffectiveMaxHealth(); p.Health > maxH {
			p.Health = maxH
		}
		if maxA := p.EffectiveMaxAction(); p.Action > maxA {
			p.Action = maxA
		}
		if maxM := p.EffectiveMaxMind(); p.Mind > maxM {
			p.Mind = maxM
		}
	}
}

// EffectiveSecondaryAttributes computes total Constitution (Health secondary),
// Stamina (Action secondary), and Willpower (Mind secondary) from base stats,
// combat/gathering skill levels, and equipped gear.
func (p Player) EffectiveSecondaryAttributes() SecondaryAttributes {
	constitution := p.Constitution
	if constitution <= 0 {
		constitution = BasePlayerConstitution
	}
	stamina := p.Stamina
	if stamina <= 0 {
		stamina = BasePlayerStamina
	}
	willpower := p.Willpower
	if willpower <= 0 {
		willpower = BasePlayerWillpower
	}

	defenseLevel := p.Skills["defense"].Level
	strengthLevel := p.Skills["strength"].Level
	meleeLevel := p.Skills["melee"].Level
	rangedLevel := p.Skills["ranged"].Level
	magicLevel := p.Skills["magic"].Level

	if defenseLevel > 1 {
		constitution += (defenseLevel - 1) / 5
	}
	if strengthLevel > 1 {
		constitution += (strengthLevel - 1) / 10
	}
	if meleeLevel > 1 {
		stamina += (meleeLevel - 1) / 5
	}
	if rangedLevel > 1 {
		stamina += (rangedLevel - 1) / 5
	}
	if magicLevel > 1 {
		willpower += (magicLevel - 1) / 5
	}

	for _, itemID := range p.Equipment {
		if item, ok := FindItem(itemID); ok {
			constitution += item.ConstitutionBonus
			stamina += item.StaminaBonus
			willpower += item.WillpowerBonus
		}
	}
	return SecondaryAttributes{
		Constitution: constitution,
		Stamina:      stamina,
		Willpower:    willpower,
	}
}

func (p Player) EffectiveMaxHealth() int {
	sec := p.EffectiveSecondaryAttributes()
	extraConstitution := sec.Constitution - BasePlayerConstitution
	if extraConstitution < 0 {
		extraConstitution = 0
	}
	return MaxPlayerHealth + p.MaxHealthBuff + extraConstitution*2
}

func (p Player) EffectiveMaxAction() int {
	sec := p.EffectiveSecondaryAttributes()
	extraStamina := sec.Stamina - BasePlayerStamina
	if extraStamina < 0 {
		extraStamina = 0
	}
	return MaxPlayerAction + p.MaxActionBuff + extraStamina*2
}

func (p Player) EffectiveMaxMind() int {
	sec := p.EffectiveSecondaryAttributes()
	extraWillpower := sec.Willpower - BasePlayerWillpower
	if extraWillpower < 0 {
		extraWillpower = 0
	}
	return MaxPlayerMind + p.MaxMindBuff + extraWillpower*2
}

// HealthRegenRate returns the Health points restored per 5-second regeneration
// pulse based on the player's Constitution secondary attribute.
func (p Player) HealthRegenRate() int {
	sec := p.EffectiveSecondaryAttributes()
	rate := 1 + sec.Constitution/5
	if rate < 1 {
		rate = 1
	}
	return rate
}

// ActionRegenRate returns the Action points restored per 5-second regeneration
// pulse based on the player's Stamina secondary attribute.
func (p Player) ActionRegenRate() int {
	sec := p.EffectiveSecondaryAttributes()
	rate := 2 + sec.Stamina/5
	if rate < 1 {
		rate = 1
	}
	return rate
}

// MindRegenRate returns the Mind points restored per 5-second regeneration
// pulse based on the player's Willpower secondary attribute.
func (p Player) MindRegenRate() int {
	sec := p.EffectiveSecondaryAttributes()
	rate := 2 + sec.Willpower/5
	if rate < 1 {
		rate = 1
	}
	return rate
}

// RegenerateVitals applies one 5-second regeneration tick to Health, Action,
// and Mind using Constitution, Stamina, and Willpower when the player is alive.
func (p *Player) RegenerateVitals(now time.Time) VitalPools {
	p.EnsureVitalDefaults()
	p.ExpireBuffsIfNeeded(now)
	if p.Health > 0 {
		maxH := p.EffectiveMaxHealth()
		if p.Health < maxH {
			p.Health += p.HealthRegenRate()
			if p.Health > maxH {
				p.Health = maxH
			}
		}
		maxA := p.EffectiveMaxAction()
		if p.Action < maxA {
			p.Action += p.ActionRegenRate()
			if p.Action > maxA {
				p.Action = maxA
			}
		}
		maxM := p.EffectiveMaxMind()
		if p.Mind < maxM {
			p.Mind += p.MindRegenRate()
			if p.Mind > maxM {
				p.Mind = maxM
			}
		}
	}
	return p.VitalPoolsSnapshot()
}

func (p Player) VitalPoolsSnapshot() VitalPools {
	return VitalPools{
		Health:           p.Health,
		MaxHealth:        p.EffectiveMaxHealth(),
		Action:           p.Action,
		MaxAction:        p.EffectiveMaxAction(),
		Mind:             p.Mind,
		MaxMind:          p.EffectiveMaxMind(),
		Secondary:        p.EffectiveSecondaryAttributes(),
		HealthRegenPer5s: p.HealthRegenRate(),
		ActionRegenPer5s: p.ActionRegenRate(),
		MindRegenPer5s:   p.MindRegenRate(),
	}
}

func (p *Player) AwardExperience(skillID string, amount int) (bool, error) {
	if err := p.CanAct(); err != nil {
		return false, err
	}
	skillID = strings.TrimSpace(strings.ToLower(skillID))
	if !IsKnownSkill(skillID) {
		return false, ErrUnknownSkill
	}
	if amount <= 0 {
		return false, ErrInvalidExperience
	}
	current := p.Skills[skillID]
	if current.Experience > MaxSkillExperience-amount {
		return false, ErrExperienceOverflow
	}
	p.Skills = EnsureSkillProgress(p.Skills)
	current = p.Skills[skillID]
	if current.Level == 0 {
		current.Level = 1
	}
	nextXP := current.Experience + amount
	nextLevel := LevelForExperience(nextXP)
	leveled := nextLevel > current.Level
	p.Skills[skillID] = SkillProgress{Experience: nextXP, Level: nextLevel}
	return leveled, nil
}
