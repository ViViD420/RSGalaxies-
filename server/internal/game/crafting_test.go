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
	if !ok || again.Ingredients[0].Quantity != 2 || again.Outputs[0].Quantity != 1 {
		t.Fatal("recipe definitions were mutated through returned slice")
	}
}

func TestCopperSmithingRecipesProduceEquipmentMaterials(t *testing.T) {
 player,err:=NewPlayer("smith","CopperSmith",time.Now());if err!=nil{t.Fatal(err)}
 player.Position.Z=8
 player.Inventory["copper_ore"]=2
 barRecipe,ok:=FindCraftingRecipe("copper-bar");if !ok{t.Fatal("copper bar recipe missing")}
 forge,ok:=FindCraftingStation("frontier-forge");if !ok{t.Fatal("frontier forge missing")}
 result,err:=player.Craft(barRecipe,forge,time.Now());if err!=nil{t.Fatalf("smelt copper bar: %v",err)}
 if result.Experience!=35||player.Inventory["copper_ore"]!=0||player.Inventory["copper_bar"]!=1{t.Fatalf("smelting result=%+v inventory=%v",result,player.Inventory)}
 player.Inventory["copper_bar"]=3
 player.Inventory["wooden_plank"]=1
 swordRecipe,ok:=FindCraftingRecipe("copper-sword");if !ok{t.Fatal("copper sword recipe missing")}
 sword,err:=player.Craft(swordRecipe,forge,time.Now());if err!=nil{t.Fatalf("forge training sword: %v",err)}
 if sword.Experience!=60||player.Inventory["copper_bar"]!=0||player.Inventory["wooden_plank"]!=0||player.Inventory["training_sword"]!=1{t.Fatalf("sword crafting result=%+v inventory=%v",sword,player.Inventory)}
 if player.Skills["smithing"].Experience!=95{t.Fatalf("smithing XP=%d want=95",player.Skills["smithing"].Experience)}
}

func TestHerbalBundleRecipeConsumesForagedMaterials(t *testing.T) {
	player,err:=NewPlayer("herbal-crafter","HerbalCrafter",time.Now());if err!=nil{t.Fatal(err)}
	player.Position.Z=2
	player.Inventory["wild_herb"]=3
	player.Inventory["wooden_plank"]=1
	recipe,ok:=FindCraftingRecipe("herbal-bundle");if !ok{t.Fatal("herbal bundle recipe missing")}
	station,ok:=FindCraftingStation(recipe.StationID);if !ok{t.Fatal("herbal bundle station missing")}
	result,err:=player.Craft(recipe,station,time.Now());if err!=nil{t.Fatalf("craft herbal bundle: %v",err)}
	if result.Experience!=45||result.SkillID!="crafting"||player.Inventory["wild_herb"]!=0||player.Inventory["wooden_plank"]!=0||player.Inventory["herbal_bundle"]!=1{t.Fatalf("herbal bundle result=%+v inventory=%v",result,player.Inventory)}
}
