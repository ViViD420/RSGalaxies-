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
	if _,err:=player.UseItem("minor_healing_potion");!errors.Is(err,ErrHealthFull){t.Fatalf("full health error=%v",err)}
	if player.Inventory["minor_healing_potion"]!=1{t.Fatal("failed potion use consumed item")}
}
