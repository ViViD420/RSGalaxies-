package game

import (
	"errors"
	"testing"
	"time"
)

func TestPlayerEquipAndUnequipItems(t *testing.T) {
	player,err:=NewPlayer("equipment-player","EquipmentWayfarer",time.Now());if err!=nil{t.Fatal(err)}
	player.Inventory["training_sword"]=1
	player.Inventory["leather_cap"]=1
	player.Inventory["leather_tunic"]=1
	player.Inventory["leather_boots"]=1
	for _,itemID:=range []string{"training_sword","leather_cap","leather_tunic","leather_boots"}{if err:=player.EquipItem(itemID);err!=nil{t.Fatalf("equip %s: %v",itemID,err)}}
	attack,defense:=player.EquipmentBonuses()
	if attack!=2||defense!=4{t.Fatalf("equipment bonuses attack=%d defense=%d",attack,defense)}
	if len(player.Equipment)!=4||player.Inventory["training_sword"]!=0{t.Fatalf("equipment state=%v inventory=%v",player.Equipment,player.Inventory)}
	if err:=player.UnequipItem("HEAD");err!=nil{t.Fatalf("unequip head: %v",err)}
	if player.Inventory["leather_cap"]!=1{t.Fatalf("unequipped item not returned: %v",player.Inventory)}
	attack,defense=player.EquipmentBonuses()
	if attack!=2||defense!=3{t.Fatalf("bonuses after unequip attack=%d defense=%d",attack,defense)}
}

func TestEquipmentRejectsInvalidOperationsWithoutLosingItems(t *testing.T) {
	player,err:=NewPlayer("equipment-player-2","EquipmentWayfarer",time.Now());if err!=nil{t.Fatal(err)}
	player.Inventory["softwood_log"]=1
	if err:=player.EquipItem("softwood_log");!errors.Is(err,ErrItemNotEquippable){t.Fatalf("non-equipment error=%v",err)}
	if err:=player.UnequipItem("unknown");!errors.Is(err,ErrEquipmentSlotEmpty){t.Fatalf("empty slot error=%v",err)}
	if err:=player.UnequipItem("");!errors.Is(err,ErrInvalidEquipmentSlot){t.Fatalf("invalid slot error=%v",err)}
	player.Inventory["leather_cap"]=1
	if err:=player.EquipItem("leather_cap");err!=nil{t.Fatal(err)}
	player.Inventory["leather_cap"]=1
	if err:=player.UnequipItem("head");!errors.Is(err,ErrInventoryStackFull){t.Fatalf("full inventory slot error=%v",err)}
	if player.Equipment["head"]!="leather_cap"||player.Inventory["leather_cap"]!=1{t.Fatal("failed unequip changed equipment or inventory")}
}
