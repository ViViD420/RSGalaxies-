package game

import (
	"errors"
	"testing"
	"time"
)

func TestGatheringAwardsInventoryAndExperience(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	player, err := NewPlayer("p-1", "Wayfarer", now)
	if err != nil { t.Fatal(err) }
	player.Position.X = 3
	node, ok := FindGatheringNode("tree-001")
	if !ok { t.Fatal("tree node missing") }

	result, state, err := player.Gather(node, GatheringNodeState{}, now)
	if err != nil { t.Fatal(err) }
	if result.ItemID != "softwood_log" || result.Experience != 25 || result.Quantity != 1 {
		t.Fatalf("unexpected gathering result: %+v", result)
	}
	if player.Inventory["softwood_log"] != 1 || player.Skills["woodcutting"].Experience != 25 {
		t.Fatalf("gather did not atomically update player: inventory=%v skill=%+v", player.Inventory, player.Skills["woodcutting"])
	}
	if state.Charges != 2 { t.Fatalf("remaining node charges = %d, want 2", state.Charges) }
}

func TestGatheringNodeDepletionAndRespawn(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	player, _ := NewPlayer("p-1", "Wayfarer", now)
	player.Position.X = 3
	node, _ := FindGatheringNode("tree-001")
	state := GatheringNodeState{}
	for i := 0; i < node.MaxCharges; i++ {
		_, next, err := player.Gather(node, state, now)
		if err != nil { t.Fatalf("gather %d: %v", i, err) }
		state = next
	}
	if _, _, err := player.Gather(node, state, now.Add(time.Second)); !errors.Is(err, ErrGatheringNodeDepleted) {
		t.Fatalf("depleted node error = %v", err)
	}
	_, state, err := player.Gather(node, state, now.Add(31*time.Second))
	if err != nil { t.Fatalf("gather after respawn: %v", err) }
	if state.Charges != node.MaxCharges-1 { t.Fatalf("charges after respawn = %d", state.Charges) }
}

func TestGatheringRejectsOutOfRangeAndFullInventory(t *testing.T) {
	now := time.Now()
	player, _ := NewPlayer("p-1", "Wayfarer", now)
	node, _ := FindGatheringNode("tree-001")
	player.Position.X = -20
	if _, _, err := player.Gather(node, GatheringNodeState{}, now); !errors.Is(err, ErrGatheringOutOfRange) {
		t.Fatalf("out-of-range error = %v", err)
	}
	player.Position.X = 3
	player.Inventory["softwood_log"] = 100
	if _, _, err := player.Gather(node, GatheringNodeState{}, now); !errors.Is(err, ErrInventoryFull) {
		t.Fatalf("full inventory error = %v", err)
	}
}

func TestCopperVeinAwardsMiningExperienceAndOre(t *testing.T) {
 now:=time.Now().UTC()
 player,err:=NewPlayer("copper-miner","CopperMiner",now);if err!=nil{t.Fatal(err)}
 player.Position.Z=-5
 node,ok:=FindGatheringNode("copper-001");if !ok{t.Fatal("copper vein missing")}
 result,state,err:=player.Gather(node,GatheringNodeState{},now);if err!=nil{t.Fatal(err)}
 if result.ItemID!="copper_ore"||result.Experience!=30||player.Inventory["copper_ore"]!=1||player.Skills["mining"].Experience!=30{t.Fatalf("copper gather result=%+v player=%+v",result,player)}
 if state.Charges!=1{t.Fatalf("copper vein charges=%d want=1",state.Charges)}
}
