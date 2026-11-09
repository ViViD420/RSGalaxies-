package game

import "time"

// EncounterParticipant is the server-recorded contribution summary for one
// character during one spawn life. Client-reported contribution is never used.
type EncounterParticipant struct {
	EncounterID string
	EnemyID string
	AccountID string
	CharacterID PlayerID
	DamageDealt int64
	DamageReceived int64
	ActionCount int64
	FirstContributionAt time.Time
	LastContributionAt time.Time
}
