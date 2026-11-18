package game

import "errors"

const LootRollRange = 10000

var (
	ErrUnknownLootTable = errors.New("unknown loot table")
	ErrInvalidLootRoll  = errors.New("invalid loot roll")
)

type LootTableDrop struct {
	ItemID      string
	Weight      int
	MinQuantity int
	MaxQuantity int
}

type LootTableRoll struct {
	Drops []LootTableDrop
}

type LootRollInput struct {
	DropRoll     int
	QuantityRoll int
}

type LootTableDefinition struct {
	ID      string
	Version int
	// Drops is the original single weighted roll format. It remains supported
	// for existing tables while new tables use guaranteed and independent bonus rolls.
	Drops           []LootTableDrop
	GuaranteedDrops []LootTableDrop
	BonusRolls      []LootTableRoll
}

var lootTableDefinitions = []LootTableDefinition{
	{ID: "frontier-rat-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "rat_pelt", MinQuantity: 1, MaxQuantity: 1},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "rat_tooth", Weight: 2000, MinQuantity: 1, MaxQuantity: 2}}},
	}},
	{ID: "frontier-boar-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "boar_hide", MinQuantity: 1, MaxQuantity: 1},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "boar_tusk", Weight: 1500, MinQuantity: 1, MaxQuantity: 2}, {ItemID: "studded_frontier_vest", Weight: 800, MinQuantity: 1, MaxQuantity: 1}}},
	}},
	{ID: "frontier-rat-common", Version: 1, Drops: []LootTableDrop{
		{ItemID: "rat_pelt", Weight: 7000, MinQuantity: 1, MaxQuantity: 1},
		{ItemID: "rat_tooth", Weight: 2000, MinQuantity: 1, MaxQuantity: 1},
	}},
	{ID: "frontier-boar-common", Version: 1, Drops: []LootTableDrop{
		{ItemID: "boar_hide", Weight: 7500, MinQuantity: 1, MaxQuantity: 1},
		{ItemID: "boar_tusk", Weight: 1500, MinQuantity: 1, MaxQuantity: 1},
	}},
	{ID: "stoneback-colossus-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "stoneback_scale", MinQuantity: 2, MaxQuantity: 4},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "stoneback_core", Weight: 300, MinQuantity: 1, MaxQuantity: 1}}},
		{Drops: []LootTableDrop{{ItemID: "stoneback_plate", Weight: 1500, MinQuantity: 1, MaxQuantity: 2}}},
	}},
	{ID: "canyon-stalker-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "stalker_chitin", MinQuantity: 1, MaxQuantity: 2},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "stalker_venom_sac", Weight: 2500, MinQuantity: 1, MaxQuantity: 2}, {ItemID: "vanguard_cleaver", Weight: 900, MinQuantity: 1, MaxQuantity: 1}}},
	}},
	{ID: "ash-hound-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "hound_fang", MinQuantity: 1, MaxQuantity: 2},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "aether_crystal", Weight: 2000, MinQuantity: 1, MaxQuantity: 1}, {ItemID: "colossus_breaker_greataxe", Weight: 900, MinQuantity: 1, MaxQuantity: 1}}},
	}},
	{ID: "ruin-sentinel-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "salvaged_scrap", MinQuantity: 1, MaxQuantity: 3},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "sentinel_core", Weight: 2200, MinQuantity: 1, MaxQuantity: 1}, {ItemID: "skirmisher_repeater_bow", Weight: 1000, MinQuantity: 1, MaxQuantity: 1}}},
	}},
	{ID: "mire-drake-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "drake_scale", MinQuantity: 1, MaxQuantity: 2},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "drake_bile", Weight: 2400, MinQuantity: 1, MaxQuantity: 2}}},
	}},
	{ID: "frost-prowler-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "prowler_pelt", MinQuantity: 1, MaxQuantity: 2},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "prowler_claw", Weight: 2200, MinQuantity: 1, MaxQuantity: 2}}},
	}},
	{ID: "aether-wraith-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "wraith_essence", MinQuantity: 1, MaxQuantity: 2},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "luminite_gem", Weight: 1800, MinQuantity: 1, MaxQuantity: 1}}},
	}},
	{ID: "feral-drone-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "salvaged_scrap", MinQuantity: 2, MaxQuantity: 3},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "scavenger_servo", Weight: 2500, MinQuantity: 1, MaxQuantity: 2}, {ItemID: "circuit_board", Weight: 2200, MinQuantity: 1, MaxQuantity: 2}}},
	}},
	{ID: "obsidian-wyrm-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "wyrm_carapace", MinQuantity: 1, MaxQuantity: 2},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "wyrm_heartstone", Weight: 1800, MinQuantity: 1, MaxQuantity: 1}}},
	}},
	{ID: "thornback-behemoth-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "behemoth_horn", MinQuantity: 1, MaxQuantity: 2},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "orichalcum_ore", Weight: 2200, MinQuantity: 1, MaxQuantity: 2}, {ItemID: "titan_harness", Weight: 850, MinQuantity: 1, MaxQuantity: 1}}},
	}},
	{ID: "void-reaver-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "reaver_plasma_coil", MinQuantity: 1, MaxQuantity: 2},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "void_quartz", Weight: 2000, MinQuantity: 1, MaxQuantity: 1}, {ItemID: "pulse_needler_pistol", Weight: 900, MinQuantity: 1, MaxQuantity: 1}}},
	}},
	{ID: "star-golem-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "neutronium_ore", MinQuantity: 1, MaxQuantity: 2},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "starsteel_ore", Weight: 1600, MinQuantity: 1, MaxQuantity: 1}}},
	}},
	{ID: "dune-scorpion-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "scorpion_stinger", MinQuantity: 1, MaxQuantity: 2},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "electrum_ore", Weight: 2200, MinQuantity: 1, MaxQuantity: 2}}},
	}},
	{ID: "crystalline-mantis-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "mantis_scythe", MinQuantity: 1, MaxQuantity: 2},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "solarite_crystal", Weight: 2000, MinQuantity: 1, MaxQuantity: 1}}},
	}},
	{ID: "magma-crusher-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "colossus_crab_shell", MinQuantity: 1, MaxQuantity: 2},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "obsidian_shard", Weight: 2500, MinQuantity: 1, MaxQuantity: 2}}},
	}},
	{ID: "rift-specter-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "specter_shroud", MinQuantity: 1, MaxQuantity: 2},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "chronite_ore", Weight: 1800, MinQuantity: 1, MaxQuantity: 1}}},
	}},
	{ID: "apex-devourer-harvest", Version: 1, GuaranteedDrops: []LootTableDrop{
		{ItemID: "leviathan_fang", MinQuantity: 1, MaxQuantity: 2},
	}, BonusRolls: []LootTableRoll{
		{Drops: []LootTableDrop{{ItemID: "nova_prism", Weight: 1500, MinQuantity: 1, MaxQuantity: 1}}},
	}},
}

func LootTableDefinitions() []LootTableDefinition {
	result := make([]LootTableDefinition, len(lootTableDefinitions))
	for index, table := range lootTableDefinitions {
		result[index] = table
		result[index].Drops = append([]LootTableDrop(nil), table.Drops...)
		result[index].GuaranteedDrops = append([]LootTableDrop(nil), table.GuaranteedDrops...)
		result[index].BonusRolls = make([]LootTableRoll, len(table.BonusRolls))
		for rollIndex, roll := range table.BonusRolls {
			result[index].BonusRolls[rollIndex].Drops = append([]LootTableDrop(nil), roll.Drops...)
		}
	}
	return result
}

func FindLootTable(id string) (LootTableDefinition, bool) {
	for _, table := range lootTableDefinitions {
		if table.ID == id {
			copy := table
			copy.Drops = append([]LootTableDrop(nil), table.Drops...)
			copy.GuaranteedDrops = append([]LootTableDrop(nil), table.GuaranteedDrops...)
			copy.BonusRolls = make([]LootTableRoll, len(table.BonusRolls))
			for rollIndex, roll := range table.BonusRolls {
				copy.BonusRolls[rollIndex].Drops = append([]LootTableDrop(nil), roll.Drops...)
			}
			return copy, true
		}
	}
	return LootTableDefinition{}, false
}

// validateLootDrop verifies an item and its quantity bounds before any rolls are applied.
func validateLootDrop(drop LootTableDrop, weighted bool) (ItemDefinition, error) {
	item, exists := FindItem(drop.ItemID)
	if !exists || drop.MinQuantity <= 0 || drop.MaxQuantity < drop.MinQuantity ||
		drop.MaxQuantity > item.MaxStack || (weighted && drop.Weight <= 0) {
		return ItemDefinition{}, ErrUnknownLootTable
	}
	return item, nil
}

func quantityForRoll(drop LootTableDrop, roll int) int {
	quantityRange := drop.MaxQuantity - drop.MinQuantity + 1
	return drop.MinQuantity + roll*quantityRange/LootRollRange
}

// RollLootTableDrops resolves guaranteed drops plus each independent weighted
// bonus roll. A weighted roll whose total is below 10,000 includes an implicit
// no-drop interval. The caller supplies one cryptographically generated pair
// for every guaranteed drop and bonus roll; no client-supplied values are used.
func RollLootTableDrops(tableID string, rolls []LootRollInput) ([]ItemQuantity, error) {
	table, ok := FindLootTable(tableID)
	if !ok {
		return nil, ErrUnknownLootTable
	}
	if len(table.Drops) > 0 && (len(table.GuaranteedDrops) > 0 || len(table.BonusRolls) > 0) {
		return nil, ErrUnknownLootTable
	}
	rollCount := len(table.GuaranteedDrops) + len(table.BonusRolls)
	if len(table.Drops) > 0 {
		rollCount = 1
	}
	if rollCount == 0 || len(rolls) != rollCount {
		return nil, ErrInvalidLootRoll
	}

	result := make([]ItemQuantity, 0, rollCount)
	indexes := make(map[string]int, rollCount)
	appendDrop := func(drop LootTableDrop, quantityRoll int) error {
		item, err := validateLootDrop(drop, false)
		if err != nil {
			return err
		}
		quantity := quantityForRoll(drop, quantityRoll)
		if index, exists := indexes[drop.ItemID]; exists {
			if result[index].Quantity > item.MaxStack-quantity {
				return ErrUnknownLootTable
			}
			result[index].Quantity += quantity
		} else {
			indexes[drop.ItemID] = len(result)
			result = append(result, ItemQuantity{ItemID: drop.ItemID, Quantity: quantity})
		}
		return nil
	}
	validRoll := func(input LootRollInput) bool {
		return input.DropRoll >= 0 && input.DropRoll < LootRollRange &&
			input.QuantityRoll >= 0 && input.QuantityRoll < LootRollRange
	}
	rollIndex := 0
	for _, drop := range table.GuaranteedDrops {
		input := rolls[rollIndex]
		rollIndex++
		if !validRoll(input) {
			return nil, ErrInvalidLootRoll
		}
		if _, err := validateLootDrop(drop, false); err != nil {
			return nil, err
		}
		if err := appendDrop(drop, input.QuantityRoll); err != nil {
			return nil, err
		}
	}
	weightedRolls := table.BonusRolls
	if len(table.Drops) > 0 {
		weightedRolls = []LootTableRoll{{Drops: table.Drops}}
	}
	for _, weighted := range weightedRolls {
		input := rolls[rollIndex]
		rollIndex++
		if !validRoll(input) {
			return nil, ErrInvalidLootRoll
		}
		totalWeight := 0
		for _, drop := range weighted.Drops {
			if _, err := validateLootDrop(drop, true); err != nil {
				return nil, err
			}
			if totalWeight > LootRollRange-drop.Weight {
				return nil, ErrUnknownLootTable
			}
			totalWeight += drop.Weight
		}
		if totalWeight <= 0 {
			return nil, ErrUnknownLootTable
		}
		if input.DropRoll >= totalWeight {
			continue
		}
		cursor := 0
		for _, drop := range weighted.Drops {
			cursor += drop.Weight
			if input.DropRoll < cursor {
				if err := appendDrop(drop, input.QuantityRoll); err != nil {
					return nil, err
				}
				break
			}
		}
	}
	return result, nil
}

// RollLootTable is the legacy one-result adapter. New settlement code should
// use RollLootTableDrops so every guaranteed and independent bonus drop survives.
func RollLootTable(tableID string, roll, quantityRoll int) (ItemQuantity, bool, error) {
	drops, err := RollLootTableDrops(tableID, []LootRollInput{{DropRoll: roll, QuantityRoll: quantityRoll}})
	if err != nil {
		return ItemQuantity{}, false, err
	}
	if len(drops) == 0 {
		return ItemQuantity{}, false, nil
	}
	return drops[0], true, nil
}
