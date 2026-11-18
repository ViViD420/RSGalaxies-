package game

import (
	"errors"
	"testing"
	"time"
)

func TestHealingPotionRestoresHealthAndConsumesOne(t *testing.T) {
	player,err:=NewPlayer("potion-player","PotionWayfarer",time.Now());if err!=nil{t.Fatal(err)}
	player.Health=90
	player.Inventory["minor_healing_potion"]=2
	healed,err:=player.UseItem("minor_healing_potion")
	if err!=nil||healed!=10||player.Health!=100||player.Inventory["minor_healing_potion"]!=1{t.Fatalf("potion healed=%d health=%d inventory=%v err=%v",healed,player.Health,player.Inventory,err)}
}

func TestConsumableRejectsInvalidUseWithoutConsumingItem(t *testing.T) {
	player,err:=NewPlayer("potion-player-2","PotionWayfarer",time.Now());if err!=nil{t.Fatal(err)}
	player.Health=50;player.Inventory["minor_healing_potion"]=1
	if _,err:=player.UseItem("softwood_log");!errors.Is(err,ErrItemNotUsable){t.Fatalf("non-consumable error=%v",err)}
	if _,err:=player.UseItem("minor_healing_potion");err!=nil{t.Fatal(err)}
	if _,err:=player.UseItem("minor_healing_potion");!errors.Is(err,ErrInsufficientItems){t.Fatalf("empty potion stack error=%v",err)}
	player.Inventory["minor_healing_potion"]=1
	player.Health=MaxPlayerHealth
	if _,err:=player.UseItem("minor_healing_potion");!errors.Is(err,ErrHealthFull){t.Fatalf("full health error=%v",err)}
	if player.Inventory["minor_healing_potion"]!=1{t.Fatal("failed potion use consumed item")}
}

func TestMedicalPacksAndAttributeBuffEnhancersIncreasePoolsAndConstitutionRegen(t *testing.T) {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	player, err := NewPlayer("med-player", "MedWayfarer", now)
	if err != nil {
		t.Fatal(err)
	}

	// Base Constitution is 10 -> HealthRegenRate = 1 + 10/5 = 3 HP per 5s pulse.
	if player.HealthRegenRate() != 3 || player.EffectiveMaxHealth() != 100 {
		t.Fatalf("unexpected base health regen=%d maxHealth=%d", player.HealthRegenRate(), player.EffectiveMaxHealth())
	}

	// Field Medical Pack restores up to 45 Health.
	player.Health = 40
	player.Inventory["field_medpack"] = 1
	healed, err := player.UseItem("field_medpack")
	if err != nil || healed != 45 || player.Health != 85 {
		t.Fatalf("field_medpack healed=%d health=%d err=%v", healed, player.Health, err)
	}

	// Constitution Vitality Stim works even at full base health, raising MaxHealth (+35 buff + 5*2 Constitution = 145)
	// and boosting Constitution by +5 (HealthRegenRate increases from 3 to 4).
	player.Health = 100
	player.Inventory["vitality_stim_pack"] = 1
	boosted, err := player.UseItem("vitality_stim_pack")
	if err != nil || boosted <= 0 {
		t.Fatalf("vitality_stim_pack failed: boosted=%d err=%v", boosted, err)
	}
	if player.EffectiveMaxHealth() != 145 || player.Health != 135 || player.HealthRegenRate() != 4 {
		t.Fatalf("after vitality_stim_pack: maxHealth=%d health=%d regen=%d", player.EffectiveMaxHealth(), player.Health, player.HealthRegenRate())
	}

	// RegenerateVitals uses the boosted Constitution rate (+4 HP).
	vitals := player.RegenerateVitals(now.Add(5 * time.Second))
	if player.Health != 139 || vitals.HealthRegenPer5s != 4 || vitals.Secondary.Constitution != 15 {
		t.Fatalf("unexpected regenerated vitals: %+v", vitals)
	}
}

// Regression: buff items previously restored HealAmount+BuffHealthAmount (and the
// Action/Mind equivalents), i.e. twice their documented restore amount. Buff
// amounts must only raise the pool caps.
func TestBuffItemsRestoreOnlyTheirDocumentedAmount(t *testing.T) {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		itemID                      string
		wantHealth, wantAct, wantMd int
	}{
		{itemID: "adrenal_action_booster", wantHealth: 20, wantAct: 55, wantMd: 20},
		{itemID: "neural_mind_tonic", wantHealth: 20, wantAct: 20, wantMd: 55},
		{itemID: "overdrive_omni_stim", wantHealth: 70, wantAct: 70, wantMd: 70},
	}
	for _, tc := range cases {
		player, err := NewPlayer(PlayerID("buff-"+tc.itemID), "BuffWayfarer", now)
		if err != nil {
			t.Fatal(err)
		}
		player.Health, player.Action, player.Mind = 20, 20, 20
		player.Inventory[tc.itemID] = 1
		if _, err := player.UseItem(tc.itemID); err != nil {
			t.Fatalf("%s: %v", tc.itemID, err)
		}
		if player.Health != tc.wantHealth || player.Action != tc.wantAct || player.Mind != tc.wantMd {
			t.Fatalf("%s restored health=%d action=%d mind=%d, want %d/%d/%d", tc.itemID, player.Health, player.Action, player.Mind, tc.wantHealth, tc.wantAct, tc.wantMd)
		}
	}
}
