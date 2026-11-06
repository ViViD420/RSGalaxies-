package game

import (
	"errors"
	"math"
)

var (
	ErrInvalidTradeOffer = errors.New("invalid trade offer")
	ErrTradeItemEquipped = errors.New("equipped items cannot be traded")
	ErrTradeOfferCapacity = errors.New("trade offer exceeds inventory or currency capacity")
)

// TradeOffer is a server-validated snapshot of the items and coins escrowed
// from one character. The repository persists this snapshot transactionally.
type TradeOffer struct {
	Items map[string]int `json:"items"`
	Coins int64 `json:"coins"`
}

// Clone returns an independent offer map suitable for persistence or response use.
func (o TradeOffer) Clone() TradeOffer {
	cloned := TradeOffer{Coins: o.Coins, Items: make(map[string]int, len(o.Items))}
	for itemID, quantity := range o.Items {
		cloned.Items[itemID] = quantity
	}
	return cloned
}

// ValidateTradeOffer checks an offer against the character's current, unescrowed
// inventory and wallet. Item metadata is always resolved from the server catalog.
func ValidateTradeOffer(player Player, offer TradeOffer) error {
	if err := player.CanAct(); err != nil {
		return err
	}
	if offer.Coins < 0 || offer.Coins > player.Coins {
		return ErrInsufficientCurrency
	}
	for itemID, quantity := range offer.Items {
		item, ok := FindItem(itemID)
		if !ok || quantity <= 0 || quantity > item.MaxStack {
			return ErrInvalidTradeOffer
		}
		if player.Inventory[itemID] < quantity {
			return ErrInsufficientItems
		}
		for _, equippedItemID := range player.Equipment {
			if equippedItemID == itemID {
				return ErrTradeItemEquipped
			}
		}
	}
	return nil
}

// EscrowTradeOffer removes a validated offer from a character. Call this only
// inside the same database transaction that writes the corresponding escrow rows.
func EscrowTradeOffer(player *Player, offer TradeOffer) error {
	if player == nil {
		return ErrInvalidTradeOffer
	}
	if err := ValidateTradeOffer(*player, offer); err != nil {
		return err
	}
	for itemID, quantity := range offer.Items {
		player.Inventory[itemID] -= quantity
		if player.Inventory[itemID] == 0 {
			delete(player.Inventory, itemID)
		}
	}
	player.Coins -= offer.Coins
	return nil
}

// ApplyTradeOffer checks all destination limits before mutating the character.
// This makes it safe to apply the two sides of a trade within one transaction.
func ApplyTradeOffer(player *Player, offer TradeOffer) error {
	if player == nil || player.Coins < 0 || offer.Coins < 0 || offer.Coins > math.MaxInt64-player.Coins {
		return ErrTradeOfferCapacity
	}
	for itemID, quantity := range offer.Items {
		item, ok := FindItem(itemID)
		if !ok || quantity <= 0 || quantity > item.MaxStack || player.Inventory[itemID] < 0 || player.Inventory[itemID] > item.MaxStack-quantity {
			return ErrTradeOfferCapacity
		}
	}
	if player.Inventory == nil {
		player.Inventory = make(map[string]int)
	}
	for itemID, quantity := range offer.Items {
		player.Inventory[itemID] += quantity
	}
	player.Coins += offer.Coins
	return nil
}
