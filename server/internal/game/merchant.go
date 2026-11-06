package game

import (
	"errors"
	"math"
)

const StartingCoins int64 = 100

var (
	ErrUnknownMerchantItem = errors.New("merchant does not trade this item")
	ErrInsufficientCurrency = errors.New("not enough coins")
	ErrCurrencyOverflow = errors.New("currency balance limit reached")
	ErrMerchantStackFull = errors.New("merchant transaction exceeds stack capacity")
)

func (p *Player) BuyItem(itemID string, quantity int) (int64,error) {
	if quantity<=0{return 0,ErrInvalidItemQuantity}
	item,ok:=FindItem(itemID);if !ok||item.BuyPrice<=0{return 0,ErrUnknownMerchantItem}
	if quantity>item.MaxStack{return 0,ErrMerchantStackFull}
	if int64(quantity)>math.MaxInt64/item.BuyPrice{return 0,ErrCurrencyOverflow}
	total:=item.BuyPrice*int64(quantity)
	if p.Coins<total{return 0,ErrInsufficientCurrency}
	if p.Inventory==nil{p.Inventory=make(map[string]int)}
	if p.Inventory[itemID]>item.MaxStack-quantity{return 0,ErrInventoryStackFull}
	p.Coins-=total
	p.Inventory[itemID]+=quantity
	return total,nil
}

func (p *Player) SellItem(itemID string, quantity int) (int64,error) {
	if quantity<=0{return 0,ErrInvalidItemQuantity}
	item,ok:=FindItem(itemID);if !ok||item.SellPrice<=0{return 0,ErrUnknownMerchantItem}
	if quantity>item.MaxStack{return 0,ErrMerchantStackFull}
	if p.Inventory[itemID]<quantity{return 0,ErrInsufficientItems}
	if int64(quantity)>math.MaxInt64/item.SellPrice{return 0,ErrCurrencyOverflow}
	total:=item.SellPrice*int64(quantity)
	if p.Coins>math.MaxInt64-total{return 0,ErrCurrencyOverflow}
	p.Inventory[itemID]-=quantity
	if p.Inventory[itemID]==0{delete(p.Inventory,itemID)}
	p.Coins+=total
	return total,nil
}
