package game

type ItemDefinition struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxStack    int    `json:"maxStack"`
}

var itemDefinitions = []ItemDefinition{
	{ID: "softwood_log", Name: "Softwood Log", Description: "A freshly harvested log suitable for basic woodworking.", MaxStack: 100},
	{ID: "rough_stone", Name: "Rough Stone", Description: "An uncut stone gathered from exposed deposits.", MaxStack: 100},
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
