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
	if err != nil {
		t.Fatal(err)
	}

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

func TestPlayerMovementAllowsSerializedClientLatencyWithinWindow(t *testing.T) {
	start := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	player, err := NewPlayer("p-latency", "Wayfarer", start)
	if err != nil {
		t.Fatal(err)
	}

	// The client sends movement every 200 ms but permits only one HTTP
	// request in flight. A delayed acknowledgement can therefore produce
	// a 400 ms interval between server-accepted positions.
	elapsed := 400 * time.Millisecond
	distance := MaxPlayerMoveSpeed * elapsed.Seconds()
	if err := player.MoveTo(distance, 0, start.Add(elapsed)); err != nil {
		t.Fatalf("legitimate movement across a delayed 400 ms update was rejected: %v", err)
	}
	if math.Abs(player.Position.X-distance) > 0.0001 {
		t.Fatalf("accepted position = %v, want %v", player.Position.X, distance)
	}
}

func TestPlayerMovementStillBoundsLongIdleCatchup(t *testing.T) {
	start := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	player, err := NewPlayer("p-idle", "Wayfarer", start)
	if err != nil {
		t.Fatal(err)
	}

	// A long idle interval is capped to MaxMovementWindow rather than
	// allowing an unbounded distance to be accumulated in one request.
	maxDistance := MaxPlayerMoveSpeed*MaxMovementWindow.Seconds() + MovementTolerance
	if err := player.MoveTo(maxDistance+0.01, 0, start.Add(5*time.Second)); !errors.Is(err, ErrMovementTooFast) {
		t.Fatalf("long-idle movement error = %v, want ErrMovementTooFast", err)
	}
}

func TestWorldZoneDefinitionsAreDefensive(t *testing.T) {
	zones := WorldZones()
	if len(zones) == 0 {
		t.Fatal("expected at least one world zone")
	}
	zones[0].Name = "mutated"
	again, ok := FindWorldZone(DefaultWorldZoneID)
	if !ok || again.Name == "mutated" {
		t.Fatal("world zone definitions were mutated")
	}
}

func TestMovementRejectsNonFiniteCoordinates(t *testing.T) {
	now := time.Now()
	player, err := NewPlayer("p-1", "Wayfarer", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := player.MoveTo(math.Inf(1), 0, now.Add(100*time.Millisecond)); !errors.Is(err, ErrInvalidMovement) {
		t.Fatalf("infinite coordinate error = %v", err)
	}
}
