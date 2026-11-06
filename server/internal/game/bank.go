package game

import "errors"

const MaxBankSlots = 100

var (
	ErrUnknownBankItem = errors.New("unknown item")
	ErrInvalidItemQuantity = errors.New("item quantity must be positive")
	ErrInsufficientItems = errors.New("insufficient items")
	ErrBankFull = errors.New("bank has no free slots")
	ErrBankStackFull = errors.New("bank item stack is full")
	ErrInventoryStackFull = errors.New("inventory item stack is full")
)

// DepositItem moves a stack from carried inventory into the account character's bank.
// Both sides are validated before either map is changed.
func (p *Player) DepositItem(itemID string, quantity int) error {
	if err:=p.CanAct();err!=nil{return err}
	if quantity <= 0 { return ErrInvalidItemQuantity }
	item, ok := FindItem(itemID)
	if !ok { return ErrUnknownBankItem }
	if p.Inventory[itemID] < quantity { return ErrInsufficientItems }
	if p.Bank == nil { p.Bank = make(map[string]int) }
	current, exists := p.Bank[itemID]
	if !exists && len(p.Bank) >= MaxBankSlots { return ErrBankFull }
	if current > item.MaxStack-quantity { return ErrBankStackFull }
	p.Inventory[itemID] -= quantity
	if p.Inventory[itemID] == 0 { delete(p.Inventory,itemID) }
	p.Bank[itemID] = current+quantity
	return nil
}

// WithdrawItem moves a stack from the bank into carried inventory.
func (p *Player) WithdrawItem(itemID string, quantity int) error {
	if err:=p.CanAct();err!=nil{return err}
	if quantity <= 0 { return ErrInvalidItemQuantity }
	item, ok := FindItem(itemID)
	if !ok { return ErrUnknownBankItem }
	if p.Bank[itemID] < quantity { return ErrInsufficientItems }
	if p.Inventory == nil { p.Inventory = make(map[string]int) }
	if p.Inventory[itemID] > item.MaxStack-quantity { return ErrInventoryStackFull }
	p.Bank[itemID] -= quantity
	if p.Bank[itemID] == 0 { delete(p.Bank,itemID) }
	p.Inventory[itemID] += quantity
	return nil
}
