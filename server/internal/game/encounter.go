package game

import (
	"errors"
	"time"
)

var (
	ErrRewardClaimNotFound = errors.New("encounter reward claim not found")
	ErrRewardInventoryFull = errors.New("inventory stack cannot accept encounter reward")
)

// EncounterParticipant is the server-recorded contribution summary for one
// character during one spawn life. Client-reported contribution is never used.
type EncounterParticipant struct {
	EncounterID string
	EnemyID string
	AccountID string
	CharacterID PlayerID
	DamageDealt int64
	ExperienceSkillID string
	DamageReceived int64
	ActionCount int64
	FirstContributionAt time.Time
	LastContributionAt time.Time
}

 
// EncounterRewardClaim is a durable, character-owned reward entitlement.
// Pending claims survive disconnects and full inventories; claiming is a
// separate transaction and is idempotent by enemy/encounter/character.
type EncounterRewardClaim struct {
	EnemyID string `json:"enemyId"`
	EncounterID string `json:"encounterId"`
	CharacterID PlayerID `json:"characterId"`
	ExperienceSkillID string `json:"experienceSkillId"`
	ExperienceAmount int `json:"experienceAmount"`
	ParentSkillID string `json:"parentSkillId,omitempty"`
	ParentExperienceAmount int `json:"parentExperienceAmount,omitempty"`
	LootItemID string `json:"lootItemId,omitempty"`
	LootQuantity int `json:"lootQuantity,omitempty"`
	Status string `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	ClaimedAt time.Time `json:"claimedAt,omitempty"`
}
