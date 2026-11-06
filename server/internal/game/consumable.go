package game

import "errors"

var (
	ErrItemNotUsable = errors.New("item has no use effect")
	ErrHealthFull = errors.New("health is already full")
)

func (p *Player) UseItem(itemID string)(int,error){
	if err:=p.CanAct();err!=nil{return 0,err}
	item,ok:=FindItem(itemID)
	if !ok||!item.Consumable||item.HealAmount<=0{return 0,ErrItemNotUsable}
	if p.Inventory[itemID]<1{return 0,ErrInsufficientItems}
	if p.Health>=MaxPlayerHealth{return 0,ErrHealthFull}
	healed:=item.HealAmount
	if healed>MaxPlayerHealth-p.Health{healed=MaxPlayerHealth-p.Health}
	p.Inventory[itemID]--
	if p.Inventory[itemID]==0{delete(p.Inventory,itemID)}
	p.Health+=healed
	return healed,nil
}
