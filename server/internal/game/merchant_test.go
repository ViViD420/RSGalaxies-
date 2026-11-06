package game

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestMerchantBuyAndSellItems(t *testing.T) {
	player,err:=NewPlayer("merchant-player","MerchantWayfarer",time.Now());if err!=nil{t.Fatal(err)}
	cost,err:=player.BuyItem("softwood_log",4)
	if err!=nil||cost!=20||player.Coins!=StartingCoins-20||player.Inventory["softwood_log"]!=4{t.Fatalf("buy cost=%d player=%+v err=%v",cost,player,err)}
	revenue,err:=player.SellItem("softwood_log",3)
	if err!=nil||revenue!=6||player.Coins!=StartingCoins-14||player.Inventory["softwood_log"]!=1{t.Fatalf("sell revenue=%d player=%+v err=%v",revenue,player,err)}
}

func TestMerchantRejectsInvalidTradesWithoutChangingBalances(t *testing.T) {
	player,err:=NewPlayer("merchant-player-2","MerchantWayfarer",time.Now());if err!=nil{t.Fatal(err)}
	player.Coins=10
	player.Inventory["rough_stone"]=2
	tests:=[]struct{name string;run func()error;want error}{
		{"insufficient coins",func()error{_,err:=player.BuyItem("wooden_plank",1);return err},ErrInsufficientCurrency},
		{"insufficient items",func()error{_,err:=player.SellItem("softwood_log",1);return err},ErrInsufficientItems},
		{"unknown item",func()error{_,err:=player.BuyItem("unknown",1);return err},ErrUnknownMerchantItem},
		{"zero quantity",func()error{_,err:=player.SellItem("rough_stone",0);return err},ErrInvalidItemQuantity},
		{"inventory full",func()error{player.Coins=100;player.Inventory["wooden_plank"]=99;_,err:=player.BuyItem("wooden_plank",2);return err},ErrInventoryStackFull},
	}
	for _,test:=range tests{t.Run(test.name,func(t *testing.T){coins:=player.Coins;inventory:=clonePlayer(player).Inventory
		if err:=test.run();!errors.Is(err,test.want){t.Fatalf("error=%v want=%v",err,test.want)}
		if test.name!="inventory full"&&(player.Coins!=coins||player.Inventory["rough_stone"]!=inventory["rough_stone"]){t.Fatal("rejected trade changed player balances")}
	})}
}

func TestMerchantSaleRejectsCurrencyOverflowWithoutConsumingItems(t *testing.T) {
 player,err:=NewPlayer("merchant-player-3","MerchantWayfarer",time.Now());if err!=nil{t.Fatal(err)}
 player.Coins=math.MaxInt64
 player.Inventory["softwood_log"]=2
 if _,err:=player.SellItem("softwood_log",1);!errors.Is(err,ErrCurrencyOverflow){t.Fatalf("sale overflow error=%v",err)}
 if player.Coins!=math.MaxInt64||player.Inventory["softwood_log"]!=2{t.Fatalf("overflow sale mutated player: coins=%d inventory=%v",player.Coins,player.Inventory)}
}
