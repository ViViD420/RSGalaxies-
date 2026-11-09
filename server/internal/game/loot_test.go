package game

import (
	"errors"
	"testing"
)

func TestRollLootTableUsesWeightedBoundaries(t *testing.T) {
	tests := []struct{roll int; want string}{
		{0,"rat_pelt"},{7499,"rat_pelt"},{7500,"rat_tooth"},{9999,"rat_tooth"},
	}
	for _,tt := range tests {
		got,ok,err := RollLootTable("frontier-rat-common",tt.roll,0)
		if err != nil || !ok || got.ItemID != tt.want || got.Quantity != 1 {
			t.Fatalf("roll %d = %+v ok=%v err=%v, want %s x1",tt.roll,got,ok,err,tt.want)
		}
	}
}

func TestRollLootTableRejectsUnknownTablesAndOutOfRangeRolls(t *testing.T) {
	if _,_,err := RollLootTable("missing-table",0,0); !errors.Is(err,ErrUnknownLootTable) {
		t.Fatalf("unknown table error=%v",err)
	}
	for _,roll := range []int{-1,LootRollRange} {
		if _,_,err := RollLootTable("frontier-rat-common",roll,0); !errors.Is(err,ErrInvalidLootRoll) {
			t.Fatalf("roll %d error=%v, want ErrInvalidLootRoll",roll,err)
		}
	}
	if _,_,err := RollLootTable("frontier-rat-common",0,LootRollRange); !errors.Is(err,ErrInvalidLootRoll) {
		t.Fatalf("quantity roll error=%v, want ErrInvalidLootRoll",err)
	}
}

func TestLootTableCatalogIsValidAndDefensivelyCopied(t *testing.T) {
	tables := LootTableDefinitions()
	if len(tables) != 2 { t.Fatalf("loot table count=%d, want 2",len(tables)) }
	for _,table := range tables {
		total := 0
		seen := map[string]bool{}
		for _,drop := range table.Drops {
			if seen[drop.ItemID] { t.Fatalf("duplicate item %s in %s",drop.ItemID,table.ID) }
			seen[drop.ItemID] = true
			total += drop.Weight
			if _,ok := FindItem(drop.ItemID); !ok { t.Fatalf("unknown item %s in %s",drop.ItemID,table.ID) }
		}
		if total != LootRollRange { t.Fatalf("table %s weights=%d, want %d",table.ID,total,LootRollRange) }
	}
	tables[0].Drops[0].ItemID = "corrupted"
	again,_ := FindLootTable("frontier-rat-common")
	if again.Drops[0].ItemID != "rat_pelt" { t.Fatal("loot table catalog exposed mutable internal slices") }
}
