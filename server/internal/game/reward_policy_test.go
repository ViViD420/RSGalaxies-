package game

import (
	"testing"
	"time"
)

func TestIsEligibleForEncounterReward(t *testing.T) {
	enemy, ok := FindEnemy("frontier-boar-001")
	if !ok { t.Fatal("catalogued enemy missing") }
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	base := EncounterParticipant{
		EncounterID: "encounter-1", EnemyID: enemy.ID, AccountID: "account-1",
		CharacterID: PlayerID("character-1"), DamageDealt: 4, DamageReceived: 20,
		ActionCount: 2, FirstContributionAt: now.Add(-time.Minute),
		LastContributionAt: now.Add(-time.Minute),
	}
	if !IsEligibleForEncounterReward(base, enemy, now) {
		t.Fatal("meaningful recent damage should qualify")
	}

	tests := []struct {
		name string
		mutate func(*EncounterParticipant)
	}{
		{"damage below threshold", func(p *EncounterParticipant) { p.DamageDealt = 3 }},
		{"stale contribution", func(p *EncounterParticipant) { p.LastContributionAt = now.Add(-EncounterRewardContributionWindow-time.Nanosecond) }},
		{"future contribution", func(p *EncounterParticipant) { p.LastContributionAt = now.Add(time.Second) }},
		{"no actions", func(p *EncounterParticipant) { p.ActionCount = 0 }},
		{"wrong enemy", func(p *EncounterParticipant) { p.EnemyID = "other-enemy" }},
		{"missing account", func(p *EncounterParticipant) { p.AccountID = "" }},
		{"missing character", func(p *EncounterParticipant) { p.CharacterID = "" }},
		{"missing encounter", func(p *EncounterParticipant) { p.EncounterID = "" }},
		{"negative damage", func(p *EncounterParticipant) { p.DamageDealt = -1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			participant := base
			tt.mutate(&participant)
			if IsEligibleForEncounterReward(participant, enemy, now) {
				t.Fatal("invalid contribution unexpectedly qualified")
			}
		})
	}
}

func TestIsEligibleForEncounterRewardDoesNotRewardPassiveDamageReceived(t *testing.T) {
	enemy, _ := FindEnemy("frontier-rat-001")
	now := time.Now().UTC()
	participant := EncounterParticipant{
		EncounterID: "encounter-2", EnemyID: enemy.ID, AccountID: "account-2",
		CharacterID: "character-2", DamageReceived: 100, ActionCount: 10,
		LastContributionAt: now,
	}
	if IsEligibleForEncounterReward(participant, enemy, now) {
		t.Fatal("passive damage received must not qualify for reward")
	}
}
