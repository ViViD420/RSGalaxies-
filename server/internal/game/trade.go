package game

import (
	"errors"
	"math"
	"time"
)

const (
	TradeLifetime = 5 * time.Minute
	TradeInteractionRange = 6.0
)

type TradeStatus string

const (
	TradePending TradeStatus = "pending"
	TradeOpen TradeStatus = "open"
	TradeCompleted TradeStatus = "completed"
	TradeCancelled TradeStatus = "cancelled"
	TradeExpired TradeStatus = "expired"
	TradeStatusPending = TradePending
	TradeStatusOpen = TradeOpen
	TradeStatusCompleted = TradeCompleted
	TradeStatusCancelled = TradeCancelled
	TradeStatusExpired = TradeExpired
)

type TradeParticipant struct {
	CharacterID PlayerID `json:"characterId"`
	CharacterName string `json:"characterName"`
	Role string `json:"role"`
	Offer TradeOffer `json:"offer"`
	Confirmed bool `json:"confirmed"`
}

type TradeSnapshot struct {
	ID string `json:"id"`
	Status TradeStatus `json:"status"`
	Revision int64 `json:"revision"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	Participants []TradeParticipant `json:"participants"`
}

var (
	ErrTradeNotFound = errors.New("trade not found")
	ErrTradeNotParticipant = errors.New("character is not a trade participant")
	ErrTradeInvalidState = errors.New("trade is not in the required state")
	ErrTradeExpired = errors.New("trade has expired")
	ErrTradeOutOfRange = errors.New("trade participants are too far apart")
	ErrTradeAlreadyActive = errors.New("character already has an active trade")
	ErrTradeStaleRevision = errors.New("trade revision is stale")
	ErrTradeSelf = errors.New("cannot trade with the same character")
	ErrTradeDifferentZone = errors.New("trade partners are in different zones")
	ErrTradeParticipantUnavailable = errors.New("trade partner is unavailable")
	ErrTradeNotAccepted = ErrTradeInvalidState
	ErrTradeRevisionConflict = ErrTradeStaleRevision
	ErrInvalidTradeOffer = errors.New("invalid trade offer")
	ErrTradeItemEquipped = errors.New("equipped items cannot be traded")
	ErrTradeOfferCapacity = errors.New("trade offer exceeds inventory or currency capacity")
)

// TradeOffer is the server-validated snapshot of the items and coins escrowed
// from one character. The repository persists this snapshot transactionally.
type TradeOffer struct {
	Items map[string]int `json:"items"`
	Coins int64 `json:"coins"`
}

func (o TradeOffer) Clone() TradeOffer {
	cloned := TradeOffer{Coins:o.Coins,Items:make(map[string]int,len(o.Items))}
	for itemID,quantity:=range o.Items { cloned.Items[itemID]=quantity }
	return cloned
}

// ValidateTradeProximity requires distinct, active characters in the same zone
// and within the configured interaction radius.
func ValidateTradeProximity(first,second Player) error {
	if first.ID==second.ID{return ErrTradeSelf}
	if err:=first.CanAct();err!=nil{return err}
	if err:=second.CanAct();err!=nil{return err}
	if first.Position.ZoneID!=second.Position.ZoneID{return ErrTradeDifferentZone}
	if math.Hypot(first.Position.X-second.Position.X,first.Position.Z-second.Position.Z)>TradeInteractionRange{return ErrTradeOutOfRange}
	return nil
}

func ValidateTradeInvite(initiator,target Player) error {
	if initiator.ID==""||target.ID==""{return ErrTradeParticipantUnavailable}
	if err:=ValidateTradeProximity(initiator,target);err!=nil{
		if errors.Is(err,ErrPlayerDefeated){return ErrTradeParticipantUnavailable}
		return err
	}
	return nil
}

// ValidateTradeOffer checks an offer against the character's current, unescrowed
// inventory and wallet. Item metadata is always resolved from the server catalog.
func ValidateTradeOffer(player Player,offer TradeOffer)error{
	if err:=player.CanAct();err!=nil{return err}
	if offer.Coins<0||offer.Coins>player.Coins{return ErrInsufficientCurrency}
	for itemID,quantity:=range offer.Items{
		item,ok:=FindItem(itemID)
		if !ok||quantity<=0||quantity>item.MaxStack{return ErrInvalidTradeOffer}
		if player.Inventory[itemID]<quantity{return ErrInsufficientItems}
		for _,equippedItemID:=range player.Equipment{if equippedItemID==itemID{return ErrTradeItemEquipped}}
	}
	return nil
}

// EscrowTradeOffer removes a validated offer. Call only in the transaction that
// writes the matching durable escrow rows.
func EscrowTradeOffer(player *Player,offer TradeOffer)error{
	if player==nil{return ErrInvalidTradeOffer}
	if err:=ValidateTradeOffer(*player,offer);err!=nil{return err}
	for itemID,quantity:=range offer.Items{
		player.Inventory[itemID]-=quantity
		if player.Inventory[itemID]==0{delete(player.Inventory,itemID)}
	}
	player.Coins-=offer.Coins
	return nil
}

// ApplyTradeOffer validates every destination before mutating the player.
func ApplyTradeOffer(player *Player,offer TradeOffer)error{
	if player==nil||player.Coins<0||offer.Coins<0||offer.Coins>math.MaxInt64-player.Coins{return ErrTradeOfferCapacity}
	for itemID,quantity:=range offer.Items{
		item,ok:=FindItem(itemID)
		if !ok||quantity<=0||quantity>item.MaxStack||player.Inventory[itemID]<0||player.Inventory[itemID]>item.MaxStack-quantity{return ErrTradeOfferCapacity}
	}
	if player.Inventory==nil{player.Inventory=make(map[string]int)}
	for itemID,quantity:=range offer.Items{player.Inventory[itemID]+=quantity}
	player.Coins+=offer.Coins
	return nil
}
