package game

import (
	"errors"
	"math"
	"time"
)

const (
	MaxPlayerMoveSpeed = 7.0
	MaxMovementWindow = 250 * time.Millisecond
	MovementTolerance = 0.15
)

var (
	ErrInvalidMovement = errors.New("invalid movement timing")
	ErrOutsideWorld = errors.New("destination is outside the world zone")
	ErrMovementTooFast = errors.New("movement exceeds maximum speed")
	ErrUnknownWorldZone = errors.New("unknown world zone")
)

// MoveTo validates a requested destination against server-measured elapsed time.
func (p *Player) MoveTo(x, z float64, now time.Time) error {
	if err:=p.CanAct();err!=nil{return err}
	zone, ok := FindWorldZone(p.Position.ZoneID)
	if !ok {
		return ErrUnknownWorldZone
	}
	if math.IsNaN(x) || math.IsNaN(z) || math.IsInf(x, 0) || math.IsInf(z, 0) {
		return ErrInvalidMovement
	}
	if x < zone.Bounds.MinX || x > zone.Bounds.MaxX || z < zone.Bounds.MinZ || z > zone.Bounds.MaxZ {
		return ErrOutsideWorld
	}
	elapsed := now.Sub(p.LastMoveAt)
	if elapsed <= 0 {
		return ErrInvalidMovement
	}
	if elapsed > MaxMovementWindow {
		elapsed = MaxMovementWindow
	}
	distance := math.Hypot(x-p.Position.X, z-p.Position.Z)
	maxDistance := MaxPlayerMoveSpeed*elapsed.Seconds() + MovementTolerance
	if distance > maxDistance {
		return ErrMovementTooFast
	}
	p.Position.X = x
	p.Position.Z = z
	p.LastMoveAt = now
	return nil
}
