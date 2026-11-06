package game

import (
	"errors"
	"math"
	"strings"
)

var (
	ErrItemNotEquippable = errors.New("item cannot be equipped")
	ErrEquipmentSlotEmpty = errors.New("equipment slot is empty")
	ErrInvalidEquipmentSlot = errors.New("invalid equipment slot")
)

func (p *Player) EquipItem(itemID string) error {
	item,ok:=FindItem(itemID)
	if !ok||item.EquipmentSlot==""{return ErrItemNotEquippable}
	if p.Inventory[itemID]<1{return ErrInsufficientItems}
	if p.Equipment==nil{p.Equipment=make(map[string]string)}
	oldItemID:=p.Equipment[item.EquipmentSlot]
	if oldItemID!="" {
		oldItem,ok:=FindItem(oldItemID)
		if !ok{return ErrItemNotEquippable}
		if p.Inventory[oldItemID]>=oldItem.MaxStack{return ErrInventoryStackFull}
	}
	p.Inventory[itemID]--
	if p.Inventory[itemID]==0{delete(p.Inventory,itemID)}
	if oldItemID!=""{p.Inventory[oldItemID]++}
	p.Equipment[item.EquipmentSlot]=itemID
	return nil
}

func (p *Player) UnequipItem(slot string) error {
	slot=strings.ToLower(strings.TrimSpace(slot))
	if slot!="main_hand"&&slot!="head"&&slot!="body"&&slot!="feet"{return ErrInvalidEquipmentSlot}
	itemID,ok:=p.Equipment[slot]
	if !ok||itemID==""{return ErrEquipmentSlotEmpty}
	item,ok:=FindItem(itemID)
	if !ok{return ErrItemNotEquippable}
	if p.Inventory[itemID]>=item.MaxStack{return ErrInventoryStackFull}
	if p.Inventory==nil{p.Inventory=make(map[string]int)}
	p.Inventory[itemID]++
	delete(p.Equipment,slot)
	return nil
}

func (p Player) EquipmentBonuses() (attack,defense int) {
	for _,itemID:=range p.Equipment {
		item,ok:=FindItem(itemID);if !ok{continue}
		if item.AttackBonus>0&&attack<=math.MaxInt-item.AttackBonus{attack+=item.AttackBonus}
		if item.DefenseBonus>0&&defense<=math.MaxInt-item.DefenseBonus{defense+=item.DefenseBonus}
	}
	return attack,defense
}
