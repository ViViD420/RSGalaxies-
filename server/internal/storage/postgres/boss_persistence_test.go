package postgres

import (
	"testing"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
)

func TestValidBossTelegraphState(t *testing.T) {
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		state game.BossEncounterState
		want  bool
	}{
		{name: "inactive", want: true},
		{name: "active complete", state: game.BossEncounterState{
			ActiveAbilityID: "ground-shatter",
			TelegraphStartedAt: now,
			TelegraphUntil: now.Add(2 * time.Second),
		}, want: true},
		{name: "active zero start", state: game.BossEncounterState{
			ActiveAbilityID: "ground-shatter",
			TelegraphUntil: now.Add(2 * time.Second),
		}},
		{name: "active zero deadline", state: game.BossEncounterState{
			ActiveAbilityID: "ground-shatter",
			TelegraphStartedAt: now,
		}},
		{name: "active inverted range", state: game.BossEncounterState{
			ActiveAbilityID: "ground-shatter",
			TelegraphStartedAt: now.Add(2 * time.Second),
			TelegraphUntil: now,
		}},
		{name: "inactive with timestamps", state: game.BossEncounterState{
			TelegraphStartedAt: now,
			TelegraphUntil: now.Add(2 * time.Second),
		}},
		{name: "inactive with start only", state: game.BossEncounterState{
			TelegraphStartedAt: now,
		}},
		{name: "inactive with deadline only", state: game.BossEncounterState{
			TelegraphUntil: now.Add(2 * time.Second),
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validBossTelegraphState(test.state); got != test.want {
				t.Fatalf("validBossTelegraphState() = %v, want %v", got, test.want)
			}
		})
	}
}
