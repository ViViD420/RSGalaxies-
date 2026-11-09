package game

import (
	"errors"
	"testing"
	"time"
)

func TestReceiveEnemyAttackUsesLivePositionAndDefense(t *testing.T) {
	now := time.Now().UTC()
	player, err := NewPlayer("enemy-hit-target", "TargetWayfarer", now.Add(-time.Second))
	if err != nil { t.Fatal(err) }
	enemy, ok := FindEnemy("frontier-rat-001")
	if !ok { t.Fatal("frontier rat missing") }
	livePosition := WorldPosition{ZoneID: player.Position.ZoneID, X: player.Position.X + 2, Z: player.Position.Z}
	result, err := player.ReceiveEnemyAttack(enemy, livePosition, now)
	if err != nil { t.Fatal(err) }
	if result.DamageReceived != 2 || result.PlayerDefeated {
		t.Fatalf("enemy attack result = %+v, want 2 damage and alive", result)
	}
	if player.Health != MaxPlayerHealth-2 || player.Skills["defense"].Experience != 1 {
		t.Fatalf("player health/Defense XP = %d/%d", player.Health, player.Skills["defense"].Experience)
	}
	if !player.LastCombatAt.Equal(now) { t.Fatalf("last combat at = %s, want %s", player.LastCombatAt, now) }
}

func TestReceiveEnemyAttackRejectsInvalidTargetsAndDefeat(t *testing.T) {
	now := time.Now().UTC()
	player, err := NewPlayer("enemy-hit-invalid", "TargetWayfarer", now)
	if err != nil { t.Fatal(err) }
	enemy, _ := FindEnemy("frontier-rat-001")
	if _, err := player.ReceiveEnemyAttack(enemy, WorldPosition{ZoneID: "other-zone"}, now); !errors.Is(err, ErrEnemyOutOfRange) {
		t.Fatalf("cross-zone attack error = %v", err)
	}
	player.Health = 1
	result, err := player.ReceiveEnemyAttack(enemy, player.Position, now)
	if err != nil { t.Fatal(err) }
	if result.DamageReceived != 1 || !result.PlayerDefeated || player.Health != 0 {
		t.Fatalf("lethal enemy attack result = %+v, health=%d", result, player.Health)
	}
	if _, err := player.ReceiveEnemyAttack(enemy, player.Position, now.Add(time.Second)); !errors.Is(err, ErrPlayerDefeated) {
		t.Fatalf("attack on defeated player error = %v", err)
	}
}
