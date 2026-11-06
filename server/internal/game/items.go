package game

type ItemDefinition struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxStack    int    `json:"maxStack"`
	BuyPrice int64 `json:"buyPrice"`
	SellPrice int64 `json:"sellPrice"`
}

var itemDefinitions = []ItemDefinition{
	{ID: "softwood_log", Name: "Softwood Log", Description: "A freshly harvested log suitable for basic woodworking.", MaxStack: 100, BuyPrice: 5, SellPrice: 2},
	{ID: "rough_stone", Name: "Rough Stone", Description: "An uncut stone gathered from exposed deposits.", MaxStack: 100, BuyPrice: 3, SellPrice: 1},
	{ID: "wooden_plank", Name: "Wooden Plank", Description: "A simple processed timber component for basic construction.", MaxStack: 100, BuyPrice: 12, SellPrice: 5},
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
