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
	if err:=player.UnequipItem("unknown");!errors.Is(err,ErrInvalidEquipmentSlot){t.Fatalf("unknown slot error=%v",err)}
	if err:=player.UnequipItem("head");!errors.Is(err,ErrEquipmentSlotEmpty){t.Fatalf("empty valid slot error=%v",err)}
	if err:=player.UnequipItem("");!errors.Is(err,ErrInvalidEquipmentSlot){t.Fatalf("invalid slot error=%v",err)}
	player.Inventory["leather_cap"]=1
	if err:=player.EquipItem("leather_cap");err!=nil{t.Fatal(err)}
	player.Inventory["leather_cap"]=1
	if err:=player.UnequipItem("head");!errors.Is(err,ErrInventoryStackFull){t.Fatalf("full inventory slot error=%v",err)}
	if player.Equipment["head"]!="leather_cap"||player.Inventory["leather_cap"]!=1{t.Fatal("failed unequip changed equipment or inventory")}
}

func TestEquipmentEnforcesServerDefinedSkillRequirements(t *testing.T) {
	player, err := NewPlayer("equipment-skill-player", "SkillWayfarer", time.Now())
	if err != nil { t.Fatal(err) }
	player.Inventory["training_sword"] = 1
	player.Skills["blades"] = SkillProgress{Level: 0}

	if err := player.EquipItem("training_sword"); !errors.Is(err, ErrInsufficientEquipmentSkill) {
		t.Fatalf("under-level weapon equip error = %v, want ErrInsufficientEquipmentSkill", err)
	}
	if player.Inventory["training_sword"] != 1 || player.Equipment["main_hand"] != "" {
		t.Fatal("failed skill requirement changed inventory or equipment")
	}

	player.Skills["blades"] = SkillProgress{Level: 1}
	if err := player.EquipItem("training_sword"); err != nil {
		t.Fatalf("equip at required level: %v", err)
	}

	// A skill change must never trap an already-equipped item: unequip is always allowed.
	player.Skills["blades"] = SkillProgress{Level: 0}
	if err := player.UnequipItem("main_hand"); err != nil {
		t.Fatalf("unequip after skill reduction: %v", err)
	}
	if player.Inventory["training_sword"] != 1 || player.Equipment["main_hand"] != "" {
		t.Fatal("unequip after skill reduction did not return the item")
	}
}

func TestArmorUsesDefenseSkillRequirement(t *testing.T) {
	player, err := NewPlayer("armor-skill-player", "ArmorWayfarer", time.Now())
	if err != nil { t.Fatal(err) }
	player.Inventory["leather_tunic"] = 1
	player.Skills["defense"] = SkillProgress{Level: 0}
	if err := player.EquipItem("leather_tunic"); !errors.Is(err, ErrInsufficientEquipmentSkill) {
		t.Fatalf("under-level armor equip error = %v, want ErrInsufficientEquipmentSkill", err)
	}
	player.Skills["defense"] = SkillProgress{Level: 1}
	if err := player.EquipItem("leather_tunic"); err != nil {
		t.Fatalf("equip armor at required Defense level: %v", err)
	}
}
