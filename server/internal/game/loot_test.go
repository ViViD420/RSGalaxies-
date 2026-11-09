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
	if len(tables) != 5 {
		t.Fatalf("loot table count=%d, want 4", len(tables))
	}
	for _, table := range tables {
		if table.ID == "" || table.Version < 1 {
			t.Fatalf("loot table has invalid identity/version: %+v", table)
		}
		seen := map[string]bool{}
		validate := func(drop LootTableDrop, weighted bool) {
			t.Helper()
			if seen[drop.ItemID] {
				t.Fatalf("duplicate item %s in %s", drop.ItemID, table.ID)
			}
			seen[drop.ItemID] = true
			item, ok := FindItem(drop.ItemID)
			if !ok || drop.MinQuantity <= 0 || drop.MaxQuantity < drop.MinQuantity ||
				drop.MaxQuantity > item.MaxStack || (weighted && drop.Weight <= 0) {
				t.Fatalf("invalid loot drop %+v in %s", drop, table.ID)
			}
		}
		for _, drop := range table.GuaranteedDrops {
			validate(drop, false)
		}
		weightedRolls := table.BonusRolls
		if len(table.Drops) > 0 {
			if len(table.GuaranteedDrops) > 0 || len(table.BonusRolls) > 0 {
				t.Fatalf("legacy and multi-drop formats mixed in %s", table.ID)
			}
			weightedRolls = []LootTableRoll{{Drops: table.Drops}}
		}
		for _, roll := range weightedRolls {
			total := 0
			for _, drop := range roll.Drops {
				validate(drop, true)
				if total > LootRollRange-drop.Weight {
					t.Fatalf("weight overflow in %s", table.ID)
				}
				total += drop.Weight
			}
			if total <= 0 || total > LootRollRange {
				t.Fatalf("table %s weighted roll=%d, want 1..%d", table.ID, total, LootRollRange)
			}
		}
		if len(table.Drops) == 0 && len(table.GuaranteedDrops)+len(table.BonusRolls) == 0 {
			t.Fatalf("table %s has no outcomes", table.ID)
		}
	}
	for index := range tables {
		if tables[index].ID == "frontier-rat-common" {
			tables[index].Drops[0].ItemID = "corrupted"
			break
		}
	}
	again, _ := FindLootTable("frontier-rat-common")
	if again.Drops[0].ItemID != "rat_pelt" {
		t.Fatal("legacy loot table catalog exposed mutable internal slices")
	}
	for index := range tables {
		if tables[index].ID == "frontier-rat-harvest" {
			tables[index].GuaranteedDrops[0].ItemID = "corrupted"
			tables[index].BonusRolls[0].Drops[0].ItemID = "corrupted"
			break
		}
	}
	again, _ = FindLootTable("frontier-rat-harvest")
	if again.GuaranteedDrops[0].ItemID != "rat_pelt" || again.BonusRolls[0].Drops[0].ItemID != "rat_tooth" {
		t.Fatal("multi-drop loot table catalog exposed mutable nested slices")
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

func TestRollLootTableDropsValidatesRollCountAndBounds(t *testing.T) {
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
	if len(tables)!=5{t.Fatalf("loot table count=%d, want 4",len(tables))}
	for index := range tables {
		if tables[index].ID == "frontier-rat-harvest" {
			tables[index].GuaranteedDrops[0].ItemID = "corrupted"
			tables[index].BonusRolls[0].Drops[0].ItemID = "corrupted"
			break
		}
	}
	again,ok:=FindLootTable("frontier-rat-harvest")
	if !ok || again.GuaranteedDrops[0].ItemID!="rat_pelt" ||
		again.BonusRolls[0].Drops[0].ItemID!="rat_tooth" {
		t.Fatal("multi-drop loot catalog exposed mutable nested slices")
	}
}

func TestStonebackColossusLootProducesMultipleCraftingMaterials(t *testing.T) {
	drops, err := RollLootTableDrops("stoneback-colossus-harvest", []LootRollInput{
		{DropRoll: 9999, QuantityRoll: 9999},
		{DropRoll: 0, QuantityRoll: 0},
		{DropRoll: 0, QuantityRoll: 9999},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []ItemQuantity{
		{ItemID: "stoneback_scale", Quantity: 4},
		{ItemID: "stoneback_core", Quantity: 1},
		{ItemID: "stoneback_plate", Quantity: 2},
	}
	if len(drops) != len(want) {
		t.Fatalf("Stoneback loot drops=%+v, want %+v", drops, want)
	}
	for i := range want {
		if drops[i] != want[i] {
			t.Fatalf("Stoneback loot drop[%d]=%+v, want %+v", i, drops[i], want[i])
		}
	}
}

func TestStonebackLootTableSupportsGuaranteedOnlyOutcome(t *testing.T) {
	drops, err := RollLootTableDrops("stoneback-colossus-harvest", []LootRollInput{
		{DropRoll: 0, QuantityRoll: 0},
		{DropRoll: 9999, QuantityRoll: 0},
		{DropRoll: 9999, QuantityRoll: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(drops) != 1 || drops[0] != (ItemQuantity{ItemID: "stoneback_scale", Quantity: 2}) {
		t.Fatalf("guaranteed Stoneback result=%+v", drops)
	}
}

