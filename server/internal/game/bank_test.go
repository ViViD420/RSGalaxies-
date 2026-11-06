package game

import (
	"errors"
	"testing"
	"time"
)

func TestPlayerBankDepositAndWithdrawal(t *testing.T) {
	player,err:=NewPlayer("bank-player","BankWayfarer",time.Now())
	if err!=nil{t.Fatal(err)}
	player.Inventory["softwood_log"]=3
	if err:=player.DepositItem("softwood_log",2);err!=nil{t.Fatalf("deposit: %v",err)}
	if player.Inventory["softwood_log"]!=1||player.Bank["softwood_log"]!=2{t.Fatalf("deposit state inventory=%v bank=%v",player.Inventory,player.Bank)}
	if err:=player.WithdrawItem("softwood_log",1);err!=nil{t.Fatalf("withdraw: %v",err)}
	if player.Inventory["softwood_log"]!=2||player.Bank["softwood_log"]!=1{t.Fatalf("withdraw state inventory=%v bank=%v",player.Inventory,player.Bank)}
}

func TestPlayerBankTransfersRejectInvalidOperationsWithoutMutation(t *testing.T) {
	player,err:=NewPlayer("bank-player-2","BankWayfarer",time.Now())
	if err!=nil{t.Fatal(err)}
	player.Inventory["softwood_log"]=5
	player.Bank["rough_stone"]=4
	tests:=[]struct{name string;operation func()error;want error}{
		{"zero deposit",func()error{return player.DepositItem("softwood_log",0)},ErrInvalidItemQuantity},
		{"unknown deposit",func()error{return player.DepositItem("unknown",1)},ErrUnknownBankItem},
		{"insufficient deposit",func()error{return player.DepositItem("softwood_log",6)},ErrInsufficientItems},
		{"insufficient withdrawal",func()error{return player.WithdrawItem("rough_stone",5)},ErrInsufficientItems},
		{"full bank stack",func()error{player.Bank["softwood_log"]=99;return player.DepositItem("softwood_log",2)},ErrBankStackFull},
		{"full inventory stack",func()error{player.Inventory["rough_stone"]=99;return player.WithdrawItem("rough_stone",2)},ErrInventoryStackFull},
	}
	for _,test:=range tests{t.Run(test.name,func(t *testing.T){beforeInventory:=clonePlayer(player).Inventory;beforeBank:=clonePlayer(player).Bank
		if err:=test.operation();!errors.Is(err,test.want){t.Fatalf("error=%v want=%v",err,test.want)}
		if test.name!="full bank stack"&&test.name!="full inventory stack"{
			for id,count:=range beforeInventory{if player.Inventory[id]!=count{t.Fatalf("inventory mutated after rejected transfer: %v",player.Inventory)}}
			for id,count:=range beforeBank{if player.Bank[id]!=count{t.Fatalf("bank mutated after rejected transfer: %v",player.Bank)}}
		}
	})}
}
