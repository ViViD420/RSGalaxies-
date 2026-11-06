package game

import "time"

// RemotePlayer is the minimal public state needed to render another active
// character. Account identifiers and private character data are never exposed.
type RemotePlayer struct {
	CharacterID PlayerID `json:"characterId"`
	Name string `json:"name"`
	ZoneID string `json:"zoneId"`
	X float64 `json:"x"`
	Z float64 `json:"z"`
	LastSeenAt time.Time `json:"lastSeenAt"`
}

const (
	PresenceLease = 20 * time.Second
	MaxZonePresence = 100
)
