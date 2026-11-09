package game

import "time"

// EncounterRewardContributionWindow bounds stale participation. A character
// must have dealt meaningful damage recently; merely being nearby or receiving
// an enemy attack is not enough to qualify for a kill reward.
const EncounterRewardContributionWindow = 2 * time.Minute

// EncounterRewardDamagePercent is the minimum share of enemy maximum health
// that a participant must damage to qualify for encounter XP.
const EncounterRewardDamagePercent = 10

// IsEligibleForEncounterReward applies the server-owned anti-leech policy.
// Damage received and action count are intentionally not sufficient on their
// own: until support actions have explicit contribution types, reward
// eligibility is based on verified damage dealt plus recent participation.
func IsEligibleForEncounterReward(participant EncounterParticipant, enemy EnemyDefinition, now time.Time) bool {
	if enemy.MaxHealth <= 0 || participant.EnemyID != enemy.ID ||
		participant.EncounterID == "" || participant.CharacterID == "" ||
		participant.AccountID == "" || participant.ActionCount <= 0 ||
		participant.DamageDealt < 0 || participant.LastContributionAt.IsZero() {
		return false
	}

	minimumDamage := (enemy.MaxHealth*EncounterRewardDamagePercent + 99) / 100
	if minimumDamage < 1 {
		minimumDamage = 1
	}
	if participant.DamageDealt < int64(minimumDamage) {
		return false
	}

	age := now.UTC().Sub(participant.LastContributionAt.UTC())
	return age >= 0 && age <= EncounterRewardContributionWindow
}
