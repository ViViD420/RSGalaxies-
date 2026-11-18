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

func TestForagingHerbPatchAwardsForagingExperience(t *testing.T) {
	now:=time.Date(2026,11,6,12,0,0,0,time.UTC)
	player,err:=NewPlayer("forager-1","HerbGatherer",now);if err!=nil{t.Fatal(err)}
	player.Position.X=8
	player.Position.Z=1
	node,ok:=FindGatheringNode("herb-001");if !ok{t.Fatal("herb patch missing")}
	result,state,err:=player.Gather(node,GatheringNodeState{},now);if err!=nil{t.Fatal(err)}
	if result.ItemID!="wild_herb"||result.SkillID!="foraging"||result.Experience!=25{t.Fatalf("foraging result=%+v",result)}
	if player.Inventory["wild_herb"]!=1||player.Skills["foraging"].Experience!=25||state.Charges!=3{t.Fatalf("foraging state player=%+v node=%+v",player,state)}
}

func TestFarmingAndScavengingNodesAwardExpectedResourcesAndXP(t *testing.T) {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	player, err := NewPlayer("frontier-gatherer", "FrontierGatherer", now)
	if err != nil {
		t.Fatal(err)
	}

	cropNode, ok := FindGatheringNode("crop-001")
	if !ok {
		t.Fatal("crop-001 node missing")
	}
	player.Position = cropNode.Position
	cropResult, cropState, err := player.Gather(cropNode, GatheringNodeState{}, now)
	if err != nil {
		t.Fatalf("gather crop-001: %v", err)
	}
	if cropResult.ItemID != "frontier_grain" || cropResult.SkillID != "farming" || cropResult.Experience != 25 || player.Inventory["frontier_grain"] != 1 || player.Skills["farming"].Experience != 25 || cropState.Charges != 3 {
		t.Fatalf("unexpected farming result=%+v state=%+v player=%+v", cropResult, cropState, player)
	}

	scrapNode, ok := FindGatheringNode("scrap-001")
	if !ok {
		t.Fatal("scrap-001 node missing")
	}
	player.Position = scrapNode.Position
	scrapResult, scrapState, err := player.Gather(scrapNode, GatheringNodeState{}, now)
	if err != nil {
		t.Fatalf("gather scrap-001: %v", err)
	}
	if scrapResult.ItemID != "salvaged_scrap" || scrapResult.SkillID != "scavenging" || scrapResult.Experience != 30 || player.Inventory["salvaged_scrap"] != 1 || player.Skills["scavenging"].Experience != 30 || scrapState.Charges != 2 {
		t.Fatalf("unexpected scavenging result=%+v state=%+v player=%+v", scrapResult, scrapState, player)
	}
}

func TestHigherTierMiningNodesEnforceLevelAndGrantOreAndCrystal(t *testing.T) {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	player, err := NewPlayer("master-miner", "MasterMiner", now)
	if err != nil {
		t.Fatal(err)
	}

	ironNode, ok := FindGatheringNode("iron-001")
	if !ok {
		t.Fatal("iron-001 node missing")
	}
	player.Position = ironNode.Position
	if _, _, err := player.Gather(ironNode, GatheringNodeState{}, now); !errors.Is(err, ErrInsufficientSkillLevel) {
		t.Fatalf("gather iron-001 at mining level 1 = %v, want ErrInsufficientSkillLevel", err)
	}
	if _, err := player.AwardExperience("mining", 100); err != nil {
		t.Fatal(err)
	}
	if level := LevelForExperience(player.Skills["mining"].Experience); level != 2 {
		t.Fatalf("mining level after 100 XP = %d, want 2", level)
	}

	// Check the Level 3 titanium gate while still at Level 2. The iron and crystal
	// gathers below grant 100 XP, which crosses the Level 3 threshold (174 XP).
	titaniumNode, ok := FindGatheringNode("titanium-001")
	if !ok {
		t.Fatal("titanium-001 node missing")
	}
	player.Position = titaniumNode.Position
	if _, _, err := player.Gather(titaniumNode, GatheringNodeState{}, now); !errors.Is(err, ErrInsufficientSkillLevel) {
		t.Fatalf("gather titanium-001 below mining level 3 = %v, want ErrInsufficientSkillLevel", err)
	}
	if player.Inventory["titanium_ore"] != 0 || player.Skills["mining"].Experience != 100 {
		t.Fatalf("rejected titanium gather mutated player: inventory=%v miningXP=%d", player.Inventory, player.Skills["mining"].Experience)
	}

	player.Position = ironNode.Position
	ironRes, _, err := player.Gather(ironNode, GatheringNodeState{}, now)
	if err != nil || ironRes.ItemID != "iron_ore" || ironRes.Experience != 45 {
		t.Fatalf("gather iron-001 at mining level 2: res=%+v err=%v", ironRes, err)
	}

	crystalNode, ok := FindGatheringNode("crystal-001")
	if !ok {
		t.Fatal("crystal-001 node missing")
	}
	player.Position = crystalNode.Position
	crystalRes, _, err := player.Gather(crystalNode, GatheringNodeState{}, now)
	if err != nil || crystalRes.ItemID != "aether_crystal" || crystalRes.Experience != 55 {
		t.Fatalf("gather crystal-001 at mining level 2: res=%+v err=%v", crystalRes, err)
	}
	if level := LevelForExperience(player.Skills["mining"].Experience); level != 3 {
		t.Fatalf("mining level after 200 XP = %d, want 3", level)
	}

	player.Position = titaniumNode.Position
	titaniumRes, _, err := player.Gather(titaniumNode, GatheringNodeState{}, now)
	if err != nil || titaniumRes.ItemID != "titanium_ore" || titaniumRes.Experience != 65 {
		t.Fatalf("gather titanium-001 at mining level 3: res=%+v err=%v", titaniumRes, err)
	}
}
