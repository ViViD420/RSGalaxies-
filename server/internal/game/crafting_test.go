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

func TestHerbalTeaCraftingCreatesUsableConsumable(t *testing.T) {
	player,err:=NewPlayer("tea-crafter","TeaCrafter",time.Now());if err!=nil{t.Fatal(err)}
	player.Position.Z=2
	player.Inventory["herbal_bundle"]=1
	player.Inventory["wild_herb"]=1
	recipe,ok:=FindCraftingRecipe("herbal-tea");if !ok{t.Fatal("herbal tea recipe missing")}
	station,_:=FindCraftingStation(recipe.StationID)
	result,err:=player.Craft(recipe,station,time.Now());if err!=nil{t.Fatalf("brew herbal tea: %v",err)}
	if result.SkillID!="cooking"||result.Experience!=35||player.Inventory["herbal_tea"]!=1||player.Inventory["herbal_bundle"]!=0||player.Inventory["wild_herb"]!=0{t.Fatalf("herbal tea craft result=%+v inventory=%v",result,player.Inventory)}
	player.Health=80
	healed,err:=player.UseItem("herbal_tea")
	if err!=nil||healed!=15||player.Health!=95||player.Inventory["herbal_tea"]!=0{t.Fatalf("use herbal tea healed=%d health=%d inventory=%v err=%v",healed,player.Health,player.Inventory,err)}
}

func TestCraftingAggregatesDuplicateIngredientsAndPreservesState(t *testing.T) {
	player,err:=NewPlayer("duplicate-ingredients","RecipeTester",time.Now());if err!=nil{t.Fatal(err)}
	station,_:=FindCraftingStation("frontier-workbench")
	player.Position=station.Position
	player.Inventory["softwood_log"]=3
	recipe:=CraftingRecipeDefinition{ID:"duplicate-inputs",SkillID:"crafting",RequiredLevel:1,StationID:station.ID,Ingredients:[]ItemQuantity{{ItemID:"softwood_log",Quantity:2},{ItemID:"softwood_log",Quantity:2}},Outputs:[]ItemQuantity{{ItemID:"wooden_plank",Quantity:1}},Experience:10}
	if _,err:=player.Craft(recipe,station,time.Now());!errors.Is(err,ErrInsufficientIngredients){t.Fatalf("duplicate ingredient error=%v",err)}
	if player.Inventory["softwood_log"]!=3||player.Inventory["wooden_plank"]!=0||player.Skills["crafting"].Experience!=0{t.Fatalf("failed duplicate-input craft mutated player: %+v",player)}
}

func TestCraftingAppliesNetStackChangeAndAggregatesOutputs(t *testing.T) {
	player,err:=NewPlayer("net-stack","RecipeTester",time.Now());if err!=nil{t.Fatal(err)}
	station,_:=FindCraftingStation("frontier-workbench")
	player.Position=station.Position
	player.Inventory["wild_herb"]=100
	netRecipe:=CraftingRecipeDefinition{ID:"net-herbs",SkillID:"crafting",RequiredLevel:1,StationID:station.ID,Ingredients:[]ItemQuantity{{ItemID:"wild_herb",Quantity:2}},Outputs:[]ItemQuantity{{ItemID:"wild_herb",Quantity:2}},Experience:10}
	if _,err:=player.Craft(netRecipe,station,time.Now());err!=nil{t.Fatalf("net-zero stack recipe should fit: %v",err)}
	if player.Inventory["wild_herb"]!=100{t.Fatalf("net-zero recipe changed stack: %d",player.Inventory["wild_herb"])}
	player.Inventory["softwood_log"]=1
	overflowRecipe:=CraftingRecipeDefinition{ID:"duplicate-outputs",SkillID:"crafting",RequiredLevel:1,StationID:station.ID,Ingredients:[]ItemQuantity{{ItemID:"softwood_log",Quantity:1}},Outputs:[]ItemQuantity{{ItemID:"wild_herb",Quantity:60},{ItemID:"wild_herb",Quantity:60}},Experience:10}
	if _,err:=player.Craft(overflowRecipe,station,time.Now());!errors.Is(err,ErrCraftingOutputFull){t.Fatalf("duplicate output overflow error=%v",err)}
	if player.Inventory["softwood_log"]!=1||player.Inventory["wild_herb"]!=100||player.Skills["crafting"].Experience!=10{t.Fatalf("overflowing craft was not atomic: inventory=%v skills=%v",player.Inventory,player.Skills)}
}

func TestCraftingRejectsWrongStation(t *testing.T) {
	player,err:=NewPlayer("wrong-station","RecipeTester",time.Now());if err!=nil{t.Fatal(err)}
	recipe,_:=FindCraftingRecipe("basic-plank")
	forge,_:=FindCraftingStation("frontier-forge")
	player.Position=forge.Position
	if _,err:=player.Craft(recipe,forge,time.Now());!errors.Is(err,ErrUnknownCraftingStation){t.Fatalf("wrong station error=%v",err)}
}

func TestRatToothBladeCraftingConnectsLootToProfessionSpecificEquipment(t *testing.T) {
	player, err := NewPlayer("rat-tooth-smith", "FrontierSmith", time.Now())
	if err != nil { t.Fatal(err) }
	player.Position.Z = 8
	player.Inventory["copper_bar"] = 2
	player.Inventory["rat_tooth"] = 2
	player.Inventory["wooden_plank"] = 1
	if _, err := player.AwardExperience("smithing", 100); err != nil { t.Fatal(err) }
	recipe, ok := FindCraftingRecipe("rat-tooth-blade")
	if !ok { t.Fatal("rat-tooth blade recipe missing") }
	forge, ok := FindCraftingStation("frontier-forge")
	if !ok { t.Fatal("frontier forge missing") }
	result, err := player.Craft(recipe, forge, time.Now())
	if err != nil { t.Fatalf("craft rat-tooth blade: %v", err) }
	if result.Experience != 90 || player.Inventory["copper_bar"] != 0 ||
		player.Inventory["rat_tooth"] != 0 || player.Inventory["wooden_plank"] != 0 ||
		player.Inventory["rat_tooth_blade"] != 1 {
		t.Fatalf("rat-tooth blade craft result=%+v inventory=%v", result, player.Inventory)
	}
	if err := player.EquipItem("rat_tooth_blade"); !errors.Is(err, ErrInsufficientEquipmentSkill) {
		t.Fatalf("equip below blades level 2 error=%v, want skill requirement", err)
	}
	if _, err := player.AwardExperience("blades", 100); err != nil { t.Fatal(err) }
	if err := player.EquipItem("rat_tooth_blade"); err != nil {
		t.Fatalf("equip rat-tooth blade at blades level 2: %v", err)
	}
	attack, _ := player.EquipmentBonuses()
	if attack != 3 { t.Fatalf("rat-tooth blade attack bonus=%d, want 3", attack) }
}

func TestBoarTuskCraftingConnectsLootToPolearmsEquipment(t *testing.T) {
	player, err := NewPlayer("boar-tusk-smith", "FrontierPolearmSmith", time.Now())
	if err != nil { t.Fatal(err) }
	player.Position.Z = 8
	player.Inventory["copper_bar"] = 2
	player.Inventory["boar_tusk"] = 2
	player.Inventory["wooden_plank"] = 1
	if _, err := player.AwardExperience("smithing", 300); err != nil { t.Fatal(err) }

	recipe, ok := FindCraftingRecipe("boar-tusk-spear")
	if !ok { t.Fatal("boar-tusk spear recipe missing") }
	forge, ok := FindCraftingStation("frontier-forge")
	if !ok { t.Fatal("frontier forge missing") }
	result, err := player.Craft(recipe, forge, time.Now())
	if err != nil { t.Fatalf("craft boar-tusk spear: %v", err) }
	if result.Experience != 120 || result.SkillID != "smithing" ||
		player.Inventory["copper_bar"] != 0 || player.Inventory["boar_tusk"] != 0 ||
		player.Inventory["wooden_plank"] != 0 || player.Inventory["boar_tusk_spear"] != 1 {
		t.Fatalf("boar-tusk spear craft result=%+v inventory=%v", result, player.Inventory)
	}
	if err := player.EquipItem("boar_tusk_spear"); !errors.Is(err, ErrInsufficientEquipmentSkill) {
		t.Fatalf("equip below polearms level 2 error=%v, want skill requirement", err)
	}
	if _, err := player.AwardExperience("polearms", 100); err != nil { t.Fatal(err) }
	if err := player.EquipItem("boar_tusk_spear"); err != nil {
		t.Fatalf("equip boar-tusk spear at polearms level 2: %v", err)
	}
	attack, _ := player.EquipmentBonuses()
	if attack != 4 { t.Fatalf("boar-tusk spear attack bonus=%d, want 4", attack) }
}
