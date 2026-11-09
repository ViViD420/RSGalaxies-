package game

import (
	"errors"
	"testing"
)

func TestRollLootTableUsesWeightedBoundaries(t *testing.T) {
	tests := []struct {
		roll int
		want string
	}{
		{0, "rat_pelt"},
		{6999, "rat_pelt"},
		{7000, "rat_tooth"},
		{8999, "rat_tooth"},
	}
	for _, tt := range tests {
		got, ok, err := RollLootTable("frontier-rat-common", tt.roll, 0)
		if err != nil || !ok || got.ItemID != tt.want || got.Quantity != 1 {
			t.Fatalf("roll %d = %+v ok=%v err=%v, want %s x1", tt.roll, got, ok, err, tt.want)
		}
	}
	for _, roll := range []int{9000, 9999} {
		if _, ok, err := RollLootTable("frontier-rat-common", roll, 0); err != nil || ok {
			t.Fatalf("no-drop roll %d = ok:%v err:%v, want no drop", roll, ok, err)
		}
	}
}

func TestRollLootTableRejectsUnknownTablesAndOutOfRangeRolls(t *testing.T) {
	if _, _, err := RollLootTable("missing-table", 0, 0); !errors.Is(err, ErrUnknownLootTable) {
		t.Fatalf("unknown table error=%v", err)
	}
	for _, roll := range []int{-1, LootRollRange} {
		if _, _, err := RollLootTable("frontier-rat-common", roll, 0); !errors.Is(err, ErrInvalidLootRoll) {
			t.Fatalf("roll %d error=%v, want ErrInvalidLootRoll", roll, err)
		}
	}
	if _, _, err := RollLootTable("frontier-rat-common", 0, LootRollRange); !errors.Is(err, ErrInvalidLootRoll) {
		t.Fatalf("quantity roll error=%v, want ErrInvalidLootRoll", err)
	}
}

func TestLootTableCatalogIsValidAndDefensivelyCopied(t *testing.T) {
	tables := LootTableDefinitions()
	if len(tables) != 2 {
		t.Fatalf("loot table count=%d, want 2", len(tables))
	}
	for _, table := range tables {
		total := 0
		seen := map[string]bool{}
		for _, drop := range table.Drops {
			if seen[drop.ItemID] {
				t.Fatalf("duplicate item %s in %s", drop.ItemID, table.ID)
			}
			seen[drop.ItemID] = true
			total += drop.Weight
			if _, ok := FindItem(drop.ItemID); !ok {
				t.Fatalf("unknown item %s in %s", drop.ItemID, table.ID)
			}
		}
		if total <= 0 || total > LootRollRange {
			t.Fatalf("table %s weights=%d, want 1..%d", table.ID, total, LootRollRange)
		}
	}
	tables[0].Drops[0].ItemID = "corrupted"
	again, _ := FindLootTable("frontier-rat-common")
	if again.Drops[0].ItemID != "rat_pelt" {
		t.Fatal("loot table catalog exposed mutable internal slices")
	}
}

func TestRollLootTableDropsResolvesGuaranteedAndIndependentBonusDrops(t *testing.T) {
	drops,err:=RollLootTableDrops("frontier-rat-harvest",[]LootRollInput{
		{DropRoll:9999,QuantityRoll:0},
		{DropRoll:0,QuantityRoll:9999},
	})
	if err!=nil{t.Fatal(err)}
	if len(drops)!=2 || drops[0]!=(ItemQuantity{ItemID:"rat_pelt",Quantity:1}) ||
		drops[1]!=(ItemQuantity{ItemID:"rat_tooth",Quantity:2}) {
		t.Fatalf("multi-drop result=%+v, want guaranteed pelt plus 2-tooth bonus",drops)
	}
}

func TestRollLootTableDropsSupportsIndependentNoDropOutcomes(t *testing.T) {
	drops,err:=RollLootTableDrops("frontier-boar-harvest",[]LootRollInput{
		{DropRoll:0,QuantityRoll:0},
		{DropRoll:9999,QuantityRoll:0},
	})
	if err!=nil{t.Fatal(err)}
	if len(drops)!=1 || drops[0]!=(ItemQuantity{ItemID:"boar_hide",Quantity:1}) {
		t.Fatalf("no-bonus result=%+v, want guaranteed hide only",drops)
	}
}

func TestRollLootTableDropsValidatesRollCountAndAggregatedStackLimit(t *testing.T) {
	if _,err:=RollLootTableDrops("frontier-rat-harvest",nil);!errors.Is(err,ErrInvalidLootRoll){
		t.Fatalf("missing rolls error=%v, want ErrInvalidLootRoll",err)
	}
	if _,err:=RollLootTableDrops("frontier-rat-harvest",[]LootRollInput{
		{DropRoll:0,QuantityRoll:0},{DropRoll:-1,QuantityRoll:0},
	});!errors.Is(err,ErrInvalidLootRoll){
		t.Fatalf("negative drop roll error=%v, want ErrInvalidLootRoll",err)
	}
}

func TestLootTableCatalogCopiesNestedMultiDropSlices(t *testing.T) {
	tables:=LootTableDefinitions()
	if len(tables)!=4{t.Fatalf("loot table count=%d, want 4",len(tables))}
	tables[2].GuaranteedDrops[0].ItemID="corrupted"
	tables[2].BonusRolls[0].Drops[0].ItemID="corrupted"
	again,ok:=FindLootTable("frontier-rat-harvest")
	if !ok || again.GuaranteedDrops[0].ItemID!="rat_pelt" ||
		again.BonusRolls[0].Drops[0].ItemID!="rat_tooth" {
		t.Fatal("multi-drop loot catalog exposed mutable nested slices")
	}
}
