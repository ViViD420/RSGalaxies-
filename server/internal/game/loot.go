package game

import "errors"

const LootRollRange = 10000

var (
	ErrUnknownLootTable = errors.New("unknown loot table")
	ErrInvalidLootRoll = errors.New("invalid loot roll")
)

type LootTableDrop struct {
	ItemID string
	Weight int
	MinQuantity int
	MaxQuantity int
}

type LootTableDefinition struct {
	ID string
	Drops []LootTableDrop
}

var lootTableDefinitions = []LootTableDefinition{
	{ID:"frontier-rat-common",Drops:[]LootTableDrop{
		{ItemID:"rat_pelt",Weight:7500,MinQuantity:1,MaxQuantity:1},
		{ItemID:"rat_tooth",Weight:2500,MinQuantity:1,MaxQuantity:1},
	}},
	{ID:"frontier-boar-common",Drops:[]LootTableDrop{
		{ItemID:"boar_hide",Weight:8000,MinQuantity:1,MaxQuantity:1},
		{ItemID:"boar_tusk",Weight:2000,MinQuantity:1,MaxQuantity:1},
	}},
}

func LootTableDefinitions() []LootTableDefinition {
	result := make([]LootTableDefinition,len(lootTableDefinitions))
	for index, table := range lootTableDefinitions {
		result[index] = table
		result[index].Drops = append([]LootTableDrop(nil),table.Drops...)
	}
	return result
}

func FindLootTable(id string) (LootTableDefinition,bool) {
	for _, table := range lootTableDefinitions {
		if table.ID == id {
			copy := table
			copy.Drops = append([]LootTableDrop(nil),table.Drops...)
			return copy,true
		}
	}
	return LootTableDefinition{},false
}

// RollLootTable maps a server-generated basis-point roll to one weighted drop.
// quantityRoll is a second independent server-generated value in [0,9999].
func RollLootTable(tableID string, roll, quantityRoll int) (ItemQuantity,bool,error) {
	table,ok := FindLootTable(tableID)
	if !ok { return ItemQuantity{},false,ErrUnknownLootTable }
	if roll < 0 || roll >= LootRollRange || quantityRoll < 0 || quantityRoll >= LootRollRange {
		return ItemQuantity{},false,ErrInvalidLootRoll
	}
	totalWeight := 0
	for _, drop := range table.Drops {
		if drop.Weight <= 0 || drop.MinQuantity <= 0 || drop.MaxQuantity < drop.MinQuantity {
			return ItemQuantity{},false,ErrUnknownLootTable
		}
		item,exists := FindItem(drop.ItemID)
		if !exists || drop.MaxQuantity > item.MaxStack {
			return ItemQuantity{},false,ErrUnknownLootTable
		}
		totalWeight += drop.Weight
	}
	if totalWeight != LootRollRange { return ItemQuantity{},false,ErrUnknownLootTable }
	cursor := 0
	for _, drop := range table.Drops {
		cursor += drop.Weight
		if roll < cursor {
			quantityRange := drop.MaxQuantity-drop.MinQuantity+1
			quantity := drop.MinQuantity + quantityRoll*quantityRange/LootRollRange
			return ItemQuantity{ItemID:drop.ItemID,Quantity:quantity},true,nil
		}
	}
	return ItemQuantity{},false,ErrUnknownLootTable
}
