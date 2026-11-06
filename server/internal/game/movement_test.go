package game

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestPlayerMovementUsesServerTimeAndZoneBounds(t *testing.T) {
	start := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	player, err := NewPlayer("p-1", "Wayfarer", start)
	if err != nil { t.Fatal(err) }

	if err := player.MoveTo(1.0, 0, start.Add(200*time.Millisecond)); err != nil {
		t.Fatalf("valid movement rejected: %v", err)
	}
	if player.Position.ZoneID != DefaultWorldZoneID || player.Position.X != 1.0 {
		t.Fatalf("unexpected position: %+v", player.Position)
	}
	if err := player.MoveTo(10, 0, start.Add(400*time.Millisecond)); !errors.Is(err, ErrMovementTooFast) {
		t.Fatalf("fast movement error = %v, want ErrMovementTooFast", err)
	}
	if err := player.MoveTo(121, 0, start.Add(400*time.Millisecond)); !errors.Is(err, ErrOutsideWorld) {
		t.Fatalf("out-of-bounds error = %v, want ErrOutsideWorld", err)
	}
	if err := player.MoveTo(1, 0, start.Add(200*time.Millisecond)); !errors.Is(err, ErrInvalidMovement) {
		t.Fatalf("non-increasing time error = %v, want ErrInvalidMovement", err)
	}
}

func TestWorldZoneDefinitionsAreDefensive(t *testing.T) {
	zones := WorldZones()
	if len(zones) == 0 { t.Fatal("expected at least one world zone") }
	zones[0].Name = "mutated"
	again, ok := FindWorldZone(DefaultWorldZoneID)
	if !ok || again.Name == "mutated" { t.Fatal("world zone definitions were mutated") }
}

func TestMovementRejectsNonFiniteCoordinates(t *testing.T) {
	now := time.Now()
	player, err := NewPlayer("p-1", "Wayfarer", now)
	if err != nil { t.Fatal(err) }
	if err := player.MoveTo(math.Inf(1), 0, now.Add(100*time.Millisecond)); !errors.Is(err, ErrInvalidMovement) {
		t.Fatalf("infinite coordinate error = %v", err)
	}
}
