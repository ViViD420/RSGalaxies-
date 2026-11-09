package game

type ItemDefinition struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxStack    int    `json:"maxStack"`
	BuyPrice int64 `json:"buyPrice"`
	SellPrice int64 `json:"sellPrice"`
	EquipmentSlot string `json:"equipmentSlot,omitempty"`
	RequiredSkillID string `json:"requiredSkillId,omitempty"`
	RequiredSkillLevel int `json:"requiredSkillLevel,omitempty"`
	AttackBonus int `json:"attackBonus,omitempty"`
	DefenseBonus int `json:"defenseBonus,omitempty"`
	Consumable bool `json:"consumable,omitempty"`
	HealAmount int `json:"healAmount,omitempty"`
}

var itemDefinitions = []ItemDefinition{
	{ID: "softwood_log", Name: "Softwood Log", Description: "A freshly harvested log suitable for basic woodworking.", MaxStack: 100, BuyPrice: 5, SellPrice: 2},
	{ID: "rough_stone", Name: "Rough Stone", Description: "An uncut stone gathered from exposed deposits.", MaxStack: 100, BuyPrice: 3, SellPrice: 1},
	{ID: "wooden_plank", Name: "Wooden Plank", Description: "A simple processed timber component for basic construction.", MaxStack: 100, BuyPrice: 12, SellPrice: 5},
	{ID: "training_sword", Name: "Training Sword", Description: "A basic practice blade for new adventurers.", MaxStack: 1, BuyPrice: 50, SellPrice: 20, EquipmentSlot: "main_hand", RequiredSkillID: "blades", RequiredSkillLevel: 1, AttackBonus: 2},
	{ID: "leather_cap", Name: "Leather Cap", Description: "A simple leather cap offering light protection.", MaxStack: 1, BuyPrice: 30, SellPrice: 12, EquipmentSlot: "head", RequiredSkillID: "defense", RequiredSkillLevel: 1, DefenseBonus: 1},
	{ID: "leather_tunic", Name: "Leather Tunic", Description: "A fitted leather tunic for frontier travel.", MaxStack: 1, BuyPrice: 60, SellPrice: 25, EquipmentSlot: "body", RequiredSkillID: "defense", RequiredSkillLevel: 1, DefenseBonus: 2},
	{ID: "leather_boots", Name: "Leather Boots", Description: "Sturdy boots for long journeys.", MaxStack: 1, BuyPrice: 25, SellPrice: 10, EquipmentSlot: "feet", RequiredSkillID: "defense", RequiredSkillLevel: 1, DefenseBonus: 1},
	{ID: "copper_ore", Name: "Copper Ore", Description: "A rough copper-bearing mineral from frontier deposits.", MaxStack: 100, BuyPrice: 3, SellPrice: 1},
	{ID: "copper_bar", Name: "Copper Bar", Description: "A refined copper ingot ready for smithing.", MaxStack: 100, BuyPrice: 10, SellPrice: 4},
	{ID: "rat_pelt", Name: "Rat Pelt", Description: "A small hide recovered from a frontier rat.", MaxStack: 100, BuyPrice: 6, SellPrice: 2},
	{ID: "rat_tooth", Name: "Rat Tooth", Description: "A sharp tooth salvaged from a frontier rat.", MaxStack: 100, BuyPrice: 0, SellPrice: 1},
	{ID: "boar_hide", Name: "Boar Hide", Description: "A tough hide taken from a wild boar.", MaxStack: 100, BuyPrice: 15, SellPrice: 5},
	{ID: "boar_tusk", Name: "Boar Tusk", Description: "A heavy tusk recovered from a wild boar.", MaxStack: 100, BuyPrice: 0, SellPrice: 4},
	{ID: "minor_healing_potion", Name: "Minor Healing Potion", Description: "Restores up to 25 health when used.", MaxStack: 10, BuyPrice: 20, SellPrice: 5, Consumable: true, HealAmount: 25},
	{ID: "wild_herb", Name: "Wild Herb", Description: "A hardy medicinal plant gathered from the frontier.", MaxStack: 100, BuyPrice: 4, SellPrice: 2},
	{ID: "herbal_bundle", Name: "Herbal Bundle", Description: "A carefully packed bundle of medicinal plants.", MaxStack: 100, BuyPrice: 18, SellPrice: 8},
	{ID: "herbal_tea", Name: "Herbal Tea", Description: "A warm restorative drink brewed from frontier herbs. Restores up to 15 health.", MaxStack: 10, BuyPrice: 15, SellPrice: 4, Consumable: true, HealAmount: 15},
}

func ItemDefinitions() []ItemDefinition {
	items := make([]ItemDefinition, len(itemDefinitions))
	copy(items, itemDefinitions)
	return items
}

func FindItem(id string) (ItemDefinition, bool) {
	for _, item := range itemDefinitions {
		if item.ID == id {
			return item, true
		}
	}
	return ItemDefinition{}, false
}
