package game

import (
	"errors"
	"math"
	"time"
)

var ErrInvalidBossEncounterDefinition = errors.New("invalid boss encounter definition")

// BossEncounterDefinition is server-owned encounter content. Execution is
// shared by every boss; individual encounters supply only phase and ability data.
type BossEncounterDefinition struct {
	ID      string
	Version int
	EnemyID string
	Phases  []BossPhaseDefinition
	Abilities []BossAbilityDefinition
}

type BossPhaseDefinition struct {
	ID                    string
	EnterAtOrBelowPercent int
	AbilityIDs            []string
}

type BossAbilityDefinition struct {
	ID               string
	TelegraphDuration time.Duration
	Cooldown          time.Duration
	MinRange          float64
	MaxRange          float64
	Damage            int
}

// BossEncounterState is the durable state for one boss spawn-life. The
// persistence layer stores it under the encounter-cycle ID so a respawn cannot
// inherit the previous life’s phase, telegraph or cooldowns.
type BossEncounterState struct {
	EncounterID       string
	DefinitionVersion int
	PhaseID           string
	PhaseStartedAt    time.Time
	ActiveAbilityID   string
	TelegraphStartedAt time.Time
	TelegraphUntil    time.Time
	AbilityCooldowns  map[string]time.Time
	NextAbilityIndex  int
	LastUpdatedAt     time.Time
	Defeated          bool
}

type BossEncounterInput struct {
	EncounterID  string
	Health       int
	MaxHealth    int
	BossPosition WorldPosition
	TargetPosition WorldPosition
	Now          time.Time
}

type BossEncounterResult struct {
	State              BossEncounterState
	PhaseChanged       bool
	AbilityStarted     bool
	AbilityInterrupted bool
	AbilityResolved    bool
	AbilityID          string
	Damage             int
}

var bossEncounterDefinitions = []BossEncounterDefinition{
	{
		ID: "stoneback-colossus",
		Version: 1,
		EnemyID: "stoneback-colossus",
		Phases: []BossPhaseDefinition{
			{ID: "guarded", EnterAtOrBelowPercent: 100, AbilityIDs: []string{"stone-fist"}},
			{ID: "rampage", EnterAtOrBelowPercent: 50, AbilityIDs: []string{"stone-fist", "ground-shatter"}},
		},
		Abilities: []BossAbilityDefinition{
			{ID: "stone-fist", TelegraphDuration: 1500 * time.Millisecond, Cooldown: 5 * time.Second, MinRange: 0, MaxRange: 3.5, Damage: 8},
			{ID: "ground-shatter", TelegraphDuration: 2 * time.Second, Cooldown: 8 * time.Second, MinRange: 0, MaxRange: 6, Damage: 12},
		},
	},
}

func BossEncounterDefinitions() []BossEncounterDefinition {
	result := make([]BossEncounterDefinition, len(bossEncounterDefinitions))
	for i, definition := range bossEncounterDefinitions {
		result[i] = cloneBossEncounterDefinition(definition)
	}
	return result
}

func FindBossEncounter(id string) (BossEncounterDefinition, bool) {
	for _, definition := range bossEncounterDefinitions {
		if definition.ID == id {
			return cloneBossEncounterDefinition(definition), true
		}
	}
	return BossEncounterDefinition{}, false
}

func cloneBossEncounterDefinition(definition BossEncounterDefinition) BossEncounterDefinition {
	copy := definition
	copy.Phases = make([]BossPhaseDefinition, len(definition.Phases))
	for i, phase := range definition.Phases {
		copy.Phases[i] = phase
		copy.Phases[i].AbilityIDs = append([]string(nil), phase.AbilityIDs...)
	}
	copy.Abilities = append([]BossAbilityDefinition(nil), definition.Abilities...)
	return copy
}

func ValidateBossEncounterDefinition(definition BossEncounterDefinition) error {
	if definition.ID == "" || definition.EnemyID == "" || definition.Version <= 0 ||
		len(definition.Phases) == 0 || len(definition.Abilities) == 0 {
		return ErrInvalidBossEncounterDefinition
	}

	abilities := make(map[string]BossAbilityDefinition, len(definition.Abilities))
	for _, ability := range definition.Abilities {
		if ability.ID == "" || ability.TelegraphDuration <= 0 || ability.Cooldown <= 0 ||
			ability.MinRange < 0 || ability.MaxRange < ability.MinRange || ability.Damage <= 0 ||
			math.IsNaN(ability.MinRange) || math.IsNaN(ability.MaxRange) ||
			math.IsInf(ability.MinRange, 0) || math.IsInf(ability.MaxRange, 0) {
			return ErrInvalidBossEncounterDefinition
		}
		if _, exists := abilities[ability.ID]; exists {
			return ErrInvalidBossEncounterDefinition
		}
		abilities[ability.ID] = ability
	}

	previousThreshold := 101
	usedAbilities := make(map[string]bool, len(abilities))
	phaseIDs := make(map[string]bool, len(definition.Phases))
	for _, phase := range definition.Phases {
		if phase.ID == "" || phaseIDs[phase.ID] ||
			phase.EnterAtOrBelowPercent < 1 || phase.EnterAtOrBelowPercent > 100 ||
			phase.EnterAtOrBelowPercent >= previousThreshold || len(phase.AbilityIDs) == 0 {
			return ErrInvalidBossEncounterDefinition
		}
		phaseIDs[phase.ID] = true
		previousThreshold = phase.EnterAtOrBelowPercent
		seenInPhase := make(map[string]bool, len(phase.AbilityIDs))
		for _, abilityID := range phase.AbilityIDs {
			if _, exists := abilities[abilityID]; !exists || seenInPhase[abilityID] {
				return ErrInvalidBossEncounterDefinition
			}
			seenInPhase[abilityID] = true
			usedAbilities[abilityID] = true
		}
	}
	if definition.Phases[0].EnterAtOrBelowPercent != 100 || len(usedAbilities) != len(abilities) {
		return ErrInvalidBossEncounterDefinition
	}
	return nil
}

func bossPhaseIndexForHealth(definition BossEncounterDefinition, health, maxHealth int) int {
	percent := float64(health) * 100 / float64(maxHealth)
	selected := 0
	for index, phase := range definition.Phases {
		if percent <= float64(phase.EnterAtOrBelowPercent) {
			selected = index
		}
	}
	return selected
}

func bossAbilityByID(definition BossEncounterDefinition, id string) (BossAbilityDefinition, bool) {
	for _, ability := range definition.Abilities {
		if ability.ID == id {
			return ability, true
		}
	}
	return BossAbilityDefinition{}, false
}

func bossAbilityInRange(ability BossAbilityDefinition, boss, target WorldPosition) bool {
	if boss.ZoneID == "" || boss.ZoneID != target.ZoneID {
		return false
	}
	distance := math.Hypot(target.X-boss.X, target.Z-boss.Z)
	return distance >= ability.MinRange && distance <= ability.MaxRange
}

func cloneAbilityCooldowns(source map[string]time.Time) map[string]time.Time {
	result := make(map[string]time.Time, len(source))
	for id, readyAt := range source {
		result[id] = readyAt
	}
	return result
}

// AdvanceBossEncounter applies one deterministic server tick to a versioned
// boss state. It emits intents; the caller must persist the returned state and
// apply resolved damage in the same database transaction.
func AdvanceBossEncounter(definition BossEncounterDefinition, state BossEncounterState, input BossEncounterInput) (BossEncounterResult, error) {
	if err := ValidateBossEncounterDefinition(definition); err != nil {
		return BossEncounterResult{}, err
	}
	if input.EncounterID == "" || input.MaxHealth <= 0 || input.Health < 0 ||
		input.Health > input.MaxHealth || input.Now.IsZero() || input.BossPosition.ZoneID == "" {
		return BossEncounterResult{}, ErrInvalidBossEncounterDefinition
	}

	reset := state.EncounterID != input.EncounterID || state.DefinitionVersion != definition.Version
	now := input.Now.UTC()
	if !reset && !state.LastUpdatedAt.IsZero() && now.Before(state.LastUpdatedAt) {
		now = state.LastUpdatedAt.UTC()
	}
	state.AbilityCooldowns = cloneAbilityCooldowns(state.AbilityCooldowns)
	if reset {
		state = BossEncounterState{
			EncounterID: input.EncounterID,
			DefinitionVersion: definition.Version,
			AbilityCooldowns: make(map[string]time.Time),
		}
	} else if state.AbilityCooldowns == nil {
		state.AbilityCooldowns = make(map[string]time.Time)
	}

	if input.Health == 0 {
		state.Defeated = true
		state.ActiveAbilityID = ""
		state.TelegraphStartedAt = time.Time{}
		state.TelegraphUntil = time.Time{}
		state.LastUpdatedAt = now
		return BossEncounterResult{State: state}, nil
	}
	if state.Defeated {
		state.LastUpdatedAt = now
		return BossEncounterResult{State: state}, nil
	}

	phaseIndex := bossPhaseIndexForHealth(definition, input.Health, input.MaxHealth)
	if !reset {
		currentIndex := -1
		for index, phase := range definition.Phases {
			if phase.ID == state.PhaseID {
				currentIndex = index
				break
			}
		}
		if currentIndex > phaseIndex {
			phaseIndex = currentIndex
		}
	}
	phase := definition.Phases[phaseIndex]
	phaseChanged := reset || state.PhaseID != phase.ID
	if phaseChanged {
		state.PhaseID = phase.ID
		state.PhaseStartedAt = now
		state.ActiveAbilityID = ""
		state.TelegraphStartedAt = time.Time{}
		state.TelegraphUntil = time.Time{}
		state.NextAbilityIndex = 0
		state.LastUpdatedAt = now
		return BossEncounterResult{State: state, PhaseChanged: !reset}, nil
	}

	if state.ActiveAbilityID != "" {
		ability, ok := bossAbilityByID(definition, state.ActiveAbilityID)
		if !ok {
			return BossEncounterResult{}, ErrInvalidBossEncounterDefinition
		}
		if !bossAbilityInRange(ability, input.BossPosition, input.TargetPosition) {
			abilityID := state.ActiveAbilityID
			state.ActiveAbilityID = ""
			state.TelegraphStartedAt = time.Time{}
			state.TelegraphUntil = time.Time{}
			state.LastUpdatedAt = now
			return BossEncounterResult{State: state, AbilityInterrupted: true, AbilityID: abilityID}, nil
		}
		if now.Before(state.TelegraphUntil) {
			state.LastUpdatedAt = now
			return BossEncounterResult{State: state}, nil
		}
		abilityID := state.ActiveAbilityID
		state.ActiveAbilityID = ""
		state.TelegraphStartedAt = time.Time{}
		state.TelegraphUntil = time.Time{}
		state.LastUpdatedAt = now
		return BossEncounterResult{
			State: state,
			AbilityResolved: true,
			AbilityID: abilityID,
			Damage: ability.Damage,
		}, nil
	}

	if len(phase.AbilityIDs) == 0 {
		return BossEncounterResult{}, ErrInvalidBossEncounterDefinition
	}
	startIndex := state.NextAbilityIndex % len(phase.AbilityIDs)
	for offset := 0; offset < len(phase.AbilityIDs); offset++ {
		index := (startIndex + offset) % len(phase.AbilityIDs)
		abilityID := phase.AbilityIDs[index]
		ability, ok := bossAbilityByID(definition, abilityID)
		if !ok {
			return BossEncounterResult{}, ErrInvalidBossEncounterDefinition
		}
		if readyAt := state.AbilityCooldowns[abilityID]; now.Before(readyAt) {
			continue
		}
		if !bossAbilityInRange(ability, input.BossPosition, input.TargetPosition) {
			continue
		}
		state.ActiveAbilityID = abilityID
		state.TelegraphStartedAt = now
		state.TelegraphUntil = now.Add(ability.TelegraphDuration)
		state.AbilityCooldowns[abilityID] = now.Add(ability.Cooldown)
		state.NextAbilityIndex = (index + 1) % len(phase.AbilityIDs)
		state.LastUpdatedAt = now
		return BossEncounterResult{State: state, AbilityStarted: true, AbilityID: abilityID}, nil
	}

	state.LastUpdatedAt = now
	return BossEncounterResult{State: state}, nil
}
