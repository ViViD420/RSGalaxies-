package game

import (
	"errors"
	"math"
	"time"
)

var (
	ErrUnknownCraftingRecipe = errors.New("unknown crafting recipe")
	ErrUnknownCraftingStation = errors.New("unknown crafting station")
	ErrCraftingOutOfRange = errors.New("crafting station is out of range")
	ErrInsufficientIngredients = errors.New("missing crafting ingredients")
	ErrCraftingSkillTooLow = errors.New("crafting skill level is too low")
	ErrCraftingOutputFull = errors.New("crafting output stack is full")
)

type ItemQuantity struct {
	ItemID string `json:"itemId"`
	Quantity int `json:"quantity"`
}

type CraftingStationDefinition struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Position WorldPosition `json:"position"`
	InteractionRange float64 `json:"interactionRange"`
}

type CraftingRecipeDefinition struct {
	ID string `json:"id"`
	Name string `json:"name"`
	SkillID string `json:"skillId"`
	RequiredLevel int `json:"requiredLevel"`
	StationID string `json:"stationId"`
	Ingredients []ItemQuantity `json:"ingredients"`
	Outputs []ItemQuantity `json:"outputs"`
	Experience int `json:"experience"`
}

type CraftingResult struct {
	RecipeID string `json:"recipeId"`
	Outputs []ItemQuantity `json:"outputs"`
	SkillID string `json:"skillId"`
	Experience int `json:"experience"`
	Skill SkillProgress `json:"skill"`
	LeveledUp bool `json:"leveledUp"`
}

var craftingStations = []CraftingStationDefinition{
	{ID: "frontier-workbench", Name: "Frontier Workbench", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 0, Z: 4}, InteractionRange: 3},
	{ID: "frontier-forge", Name: "Frontier Forge", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 0, Z: 8}, InteractionRange: 3},
}

var craftingRecipes = []CraftingRecipeDefinition{
	{ID: "basic-plank", Name: "Basic Wooden Plank", SkillID: "crafting", RequiredLevel: 1, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "softwood_log", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "wooden_plank", Quantity: 1}}, Experience: 30},
	{ID: "copper-bar", Name: "Smelt Copper Bar", SkillID: "smithing", RequiredLevel: 1, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "copper_ore", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "copper_bar", Quantity: 1}}, Experience: 35},
	{ID: "copper-sword", Name: "Forge Training Sword", SkillID: "smithing", RequiredLevel: 1, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "copper_bar", Quantity: 3}, {ItemID: "wooden_plank", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "training_sword", Quantity: 1}}, Experience: 60},
}

func CraftingStations() []CraftingStationDefinition {
	result := make([]CraftingStationDefinition, len(craftingStations))
	copy(result, craftingStations)
	return result
}

func CraftingRecipes() []CraftingRecipeDefinition {
	result := make([]CraftingRecipeDefinition, len(craftingRecipes))
	for i, recipe := range craftingRecipes {
		result[i] = recipe
		result[i].Ingredients = append([]ItemQuantity(nil), recipe.Ingredients...)
		result[i].Outputs = append([]ItemQuantity(nil), recipe.Outputs...)
	}
	return result
}

func FindCraftingStation(id string) (CraftingStationDefinition, bool) {
	for _, station := range craftingStations {
		if station.ID == id { return station, true }
	}
	return CraftingStationDefinition{}, false
}

func FindCraftingRecipe(id string) (CraftingRecipeDefinition, bool) {
	for _, recipe := range craftingRecipes {
		if recipe.ID == id {
			copy := recipe
			copy.Ingredients = append([]ItemQuantity(nil), recipe.Ingredients...)
			copy.Outputs = append([]ItemQuantity(nil), recipe.Outputs...)
			return copy, true
		}
	}
	return CraftingRecipeDefinition{}, false
}

func (p *Player) Craft(recipe CraftingRecipeDefinition, station CraftingStationDefinition, _ time.Time) (CraftingResult, error) {
	if err:=p.CanAct();err!=nil{return CraftingResult{},err}
	if p.Position.ZoneID != station.Position.ZoneID || math.Hypot(p.Position.X-station.Position.X, p.Position.Z-station.Position.Z) > station.InteractionRange {
		return CraftingResult{}, ErrCraftingOutOfRange
	}
	if p.Skills[recipe.SkillID].Level < recipe.RequiredLevel {
		return CraftingResult{}, ErrCraftingSkillTooLow
	}
	if p.Inventory == nil { p.Inventory = make(map[string]int) }
	for _, ingredient := range recipe.Ingredients {
		if p.Inventory[ingredient.ItemID] < ingredient.Quantity {
			return CraftingResult{}, ErrInsufficientIngredients
		}
	}
	for _, output := range recipe.Outputs {
		item, ok := FindItem(output.ItemID)
		if !ok { return CraftingResult{}, ErrUnknownCraftingRecipe }
		if p.Inventory[output.ItemID]+output.Quantity > item.MaxStack {
			return CraftingResult{}, ErrCraftingOutputFull
		}
	}
	leveled, err := p.AwardExperience(recipe.SkillID, recipe.Experience)
	if err != nil { return CraftingResult{}, err }
	for _, ingredient := range recipe.Ingredients { p.Inventory[ingredient.ItemID] -= ingredient.Quantity }
	for _, output := range recipe.Outputs { p.Inventory[output.ItemID] += output.Quantity }
	return CraftingResult{
		RecipeID: recipe.ID, Outputs: append([]ItemQuantity(nil), recipe.Outputs...),
		SkillID: recipe.SkillID, Experience: recipe.Experience,
		Skill: p.Skills[recipe.SkillID], LeveledUp: leveled,
	}, nil
}
