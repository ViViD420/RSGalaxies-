package game

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestEscrowTradeOfferValidatesBeforeMutating(t *testing.T) {
	now := time.Now().UTC()
	player, err := NewPlayer("trade-offer-owner", "OfferOwner", now)
	if err != nil { t.Fatal(err) }
	player.Inventory["softwood_log"] = 5
	offer := TradeOffer{Items: map[string]int{"softwood_log": 3}, Coins: 40}
	if err := EscrowTradeOffer(&player, offer); err != nil { t.Fatalf("escrow valid offer: %v", err) }
	if player.Inventory["softwood_log"] != 2 || player.Coins != StartingCoins-40 {
		t.Fatalf("escrowed state inventory=%v coins=%d", player.Inventory, player.Coins)
	}
	offer.Items["softwood_log"] = 4
	offer.Coins = 1
	before := clonePlayer(player)
	if err := EscrowTradeOffer(&player, offer); !errors.Is(err, ErrInsufficientItems) {
		t.Fatalf("insufficient item error=%v", err)
	}
	if player.Inventory["softwood_log"] != before.Inventory["softwood_log"] || player.Coins != before.Coins {
		t.Fatal("failed escrow partially mutated the character")
	}
}

func TestValidateTradeOfferRejectsEquippedUnknownAndInvalidItems(t *testing.T) {
	player, err := NewPlayer("trade-offer-equipped", "EquippedOwner", time.Now())
	if err != nil { t.Fatal(err) }
	player.Inventory["training_sword"] = 1
	player.Equipment["main_hand"] = "training_sword"
	cases := []struct {
		name string
		offer TradeOffer
		want error
	}{
		{"equipped", TradeOffer{Items: map[string]int{"training_sword": 1}}, ErrTradeItemEquipped},
		{"unknown", TradeOffer{Items: map[string]int{"unknown_item": 1}}, ErrInvalidTradeOffer},
		{"zero quantity", TradeOffer{Items: map[string]int{"softwood_log": 0}}, ErrInvalidTradeOffer},
		{"negative quantity", TradeOffer{Items: map[string]int{"softwood_log": -1}}, ErrInvalidTradeOffer},
		{"negative coins", TradeOffer{Coins: -1}, ErrInsufficientCurrency},
		{"excess coins", TradeOffer{Coins: player.Coins + 1}, ErrInsufficientCurrency},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateTradeOffer(player, test.offer); !errors.Is(err, test.want) {
				t.Fatalf("validation error=%v want=%v", err, test.want)
			}
		})
	}
}

func TestApplyTradeOfferChecksAllCapacityBeforeMutation(t *testing.T) {
	player, err := NewPlayer("trade-offer-recipient", "OfferRecipient", time.Now())
	if err != nil { t.Fatal(err) }
	player.Inventory["softwood_log"] = 99
	offer := TradeOffer{Items: map[string]int{"softwood_log": 2}, Coins: 10}
	if err := ApplyTradeOffer(&player, offer); !errors.Is(err, ErrTradeOfferCapacity) {
		t.Fatalf("full stack error=%v", err)
	}
	if player.Inventory["softwood_log"] != 99 || player.Coins != StartingCoins {
		t.Fatal("capacity failure partially mutated the recipient")
	}
	player.Inventory["softwood_log"] = 0
	player.Coins = math.MaxInt64 - 5
	if err := ApplyTradeOffer(&player, offer); !errors.Is(err, ErrTradeOfferCapacity) {
		t.Fatalf("currency overflow error=%v", err)
	}
	if player.Inventory["softwood_log"] != 0 || player.Coins != math.MaxInt64-5 {
		t.Fatal("currency overflow partially mutated the recipient")
	}
	player.Coins = 0
	if err := ApplyTradeOffer(&player, offer); err != nil { t.Fatalf("apply valid offer: %v", err) }
	if player.Inventory["softwood_log"] != 2 || player.Coins != 10 {
		t.Fatalf("applied offer state inventory=%v coins=%d", player.Inventory, player.Coins)
	}
}

func TestTradeOfferCloneOwnsItsItemMap(t *testing.T) {
	original := TradeOffer{Items: map[string]int{"wild_herb": 2}, Coins: 5}
	cloned := original.Clone()
	cloned.Items["wild_herb"] = 9
	cloned.Items["softwood_log"] = 1
	if original.Items["wild_herb"] != 2 || original.Items["softwood_log"] != 0 {
		t.Fatalf("clone mutated original offer: %+v", original)
	}
}

func TestTradeInviteRequiresAvailableNearbyCharactersInSameZone(t *testing.T) {
	now:=time.Now().UTC()
	initiator,err:=NewPlayer("invite-a","InviteA",now);if err!=nil{t.Fatal(err)}
	target,err:=NewPlayer("invite-b","InviteB",now);if err!=nil{t.Fatal(err)}
	if err:=ValidateTradeInvite(initiator,target);err!=nil{t.Fatalf("valid invite rejected: %v",err)}
	if err:=ValidateTradeInvite(initiator,initiator);!errors.Is(err,ErrTradeSelf){t.Fatalf("self-trade error=%v",err)}
	target.Position.X=TradeInteractionRange+1
	if err:=ValidateTradeInvite(initiator,target);!errors.Is(err,ErrTradeOutOfRange){t.Fatalf("out-of-range invite error=%v",err)}
	target.Position=initiator.Position
	target.Position.ZoneID="another-zone"
	if err:=ValidateTradeInvite(initiator,target);!errors.Is(err,ErrTradeDifferentZone){t.Fatalf("cross-zone invite error=%v",err)}
	target.Position=initiator.Position
	target.Health=0
	if err:=ValidateTradeInvite(initiator,target);!errors.Is(err,ErrTradeParticipantUnavailable){t.Fatalf("defeated target invite error=%v",err)}
}


func TestValidateTradeProximityUsesServerZoneAndRange(t *testing.T) {
	now := time.Now().UTC()
	first, err := NewPlayer("trade-near-a", "NearTraderA", now)
	if err != nil { t.Fatal(err) }
	second, err := NewPlayer("trade-near-b", "NearTraderB", now)
	if err != nil { t.Fatal(err) }
	second.Position.X = TradeInteractionRange
	if err := ValidateTradeProximity(first, second); err != nil {
		t.Fatalf("characters at interaction boundary should trade: %v", err)
	}
	second.Position.X += 0.01
	if err := ValidateTradeProximity(first, second); !errors.Is(err, ErrTradeOutOfRange) {
		t.Fatalf("out-of-range error=%v", err)
	}
	second.Position.X = 0
	second.Position.ZoneID = "another-zone"
	if err := ValidateTradeProximity(first, second); !errors.Is(err, ErrTradeOutOfRange) {
		t.Fatalf("cross-zone error=%v", err)
	}
	second.Position.ZoneID = first.Position.ZoneID
	second.ID = first.ID
	if err := ValidateTradeProximity(first, second); !errors.Is(err, ErrTradeSelf) {
		t.Fatalf("self-trade error=%v", err)
	}
}
