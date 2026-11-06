package game

import (
	"errors"
	"testing"
	"time"
)

func TestCraftingConsumesIngredientsAndCreatesOutput(t *testing.T) {
	player, err := NewPlayer("p-1", "Wayfarer", time.Now())
	if err != nil { t.Fatal(err) }
	player.Position.Z = 2
	player.Inventory["softwood_log"] = 2
	recipe, ok := FindCraftingRecipe("basic-plank")
	if !ok { t.Fatal("recipe missing") }
	station, ok := FindCraftingStation(recipe.StationID)
	if !ok { t.Fatal("station missing") }

	result, err := player.Craft(recipe, station, time.Now())
	if err != nil { t.Fatal(err) }
	if player.Inventory["softwood_log"] != 0 || player.Inventory["wooden_plank"] != 1 {
		t.Fatalf("unexpected inventory after craft: %v", player.Inventory)
	}
	if player.Skills["crafting"].Experience != 30 || result.Experience != 30 {
		t.Fatalf("craft XP not awarded: %+v", result)
	}
}

func TestCraftingRejectsMissingIngredientsAndOutOfRange(t *testing.T) {
	player, _ := NewPlayer("p-1", "Wayfarer", time.Now())
	recipe, _ := FindCraftingRecipe("basic-plank")
	station, _ := FindCraftingStation(recipe.StationID)
	if _, err := player.Craft(recipe, station, time.Now()); !errors.Is(err, ErrCraftingOutOfRange) {
		t.Fatalf("out-of-range error = %v", err)
	}
	player.Position.Z = 2
	if _, err := player.Craft(recipe, station, time.Now()); !errors.Is(err, ErrInsufficientIngredients) {
		t.Fatalf("missing ingredient error = %v", err)
	}
	if player.Skills["crafting"].Experience != 0 {
		t.Fatal("failed craft awarded XP")
	}
}

func TestCraftingDefinitionsAreDefensive(t *testing.T) {
	recipe, ok := FindCraftingRecipe("basic-plank")
	if !ok { t.Fatal("recipe missing") }
	recipe.Ingredients[0].Quantity = 999
	recipes := CraftingRecipes()
	recipes[0].Outputs[0].Quantity = 999
	again, ok := FindCraftingRecipe("basic-plank")
	if !ok || again.Ingredients[0].Quantity != 2 {
		t.Fatal("recipe definitions were mutated through returned slice")
	}
}
