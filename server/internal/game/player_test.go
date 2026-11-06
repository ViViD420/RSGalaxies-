package game

import (
	"errors"
	"testing"
	"time"
)

func TestNewPlayer(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.FixedZone("test", -5*60*60))
	player, err := NewPlayer("p-1", "  Wayfarer  ", now)
	if err != nil {
		t.Fatalf("NewPlayer() error = %v", err)
	}
	if player.Name != "Wayfarer" {
		t.Fatalf("Name = %q, want Wayfarer", player.Name)
	}
	if player.CreatedAt.Location() != time.UTC {
		t.Fatalf("CreatedAt location = %v, want UTC", player.CreatedAt.Location())
	}
}

func TestNewPlayerRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		id   PlayerID
		want string
	}{
		{name: "missing id", id: "", want: "Wayfarer"},
		{name: "too short", id: "p-2", want: "ab"},
		{name: "too long", id: "p-3", want: "abcdefghijklmnopqrstu"},
		{name: "whitespace", id: "p-4", want: "   "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPlayer(tt.id, tt.want, time.Now())
			if err == nil {
				t.Fatal("expected validation error")
			}
			if tt.id != "" && !errors.Is(err, ErrInvalidPlayerName) {
				t.Fatalf("error = %v, want ErrInvalidPlayerName", err)
			}
		})
	}
}
