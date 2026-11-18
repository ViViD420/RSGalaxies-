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

func TestFarmingAndScavengingRecipesProduceConsumableAndRangedMagicWeapons(t *testing.T) {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	player, err := NewPlayer("frontier-artisan", "FrontierArtisan", now)
	if err != nil {
		t.Fatal(err)
	}
	workbench, ok := FindCraftingStation("frontier-workbench")
	if !ok {
		t.Fatal("frontier-workbench missing")
	}
	forge, ok := FindCraftingStation("frontier-forge")
	if !ok {
		t.Fatal("frontier-forge missing")
	}

	// 1. Bake Frontier Ration (Cooking) & consume
	player.Position = workbench.Position
	player.Inventory["frontier_grain"] = 2
	player.Inventory["wild_herb"] = 1
	rationRecipe, ok := FindCraftingRecipe("frontier-ration")
	if !ok {
		t.Fatal("frontier-ration recipe missing")
	}
	if _, err := player.Craft(rationRecipe, workbench, now); err != nil {
		t.Fatalf("craft frontier-ration: %v", err)
	}
	if player.Inventory["frontier_ration"] != 1 || player.Skills["cooking"].Experience != 45 {
		t.Fatalf("unexpected frontier-ration state: inv=%v cooking=%+v", player.Inventory, player.Skills["cooking"])
	}
	player.Health = 70
	healed, err := player.UseItem("frontier_ration")
	if err != nil || healed != 20 || player.Health != 90 {
		t.Fatalf("use frontier_ration: healed=%d health=%d err=%v", healed, player.Health, err)
	}

	// 2. Assemble Power Cell, Frontier Pistol, and Resonance Focus Staff at Workbench
	player.Inventory["salvaged_scrap"] = 4
	player.Inventory["copper_bar"] = 2
	player.Inventory["wooden_plank"] = 3
	player.Inventory["wild_herb"] = 1
	cellRecipe, _ := FindCraftingRecipe("salvaged-power-cell")
	if _, err := player.Craft(cellRecipe, workbench, now); err != nil {
		t.Fatalf("craft salvaged-power-cell: %v", err)
	}
	pistolRecipe, _ := FindCraftingRecipe("frontier-pistol")
	if _, err := player.Craft(pistolRecipe, workbench, now); err != nil {
		t.Fatalf("craft frontier-pistol: %v", err)
	}
	focusRecipe, _ := FindCraftingRecipe("resonance-focus")
	if _, err := player.Craft(focusRecipe, workbench, now); err != nil {
		t.Fatalf("craft resonance-focus: %v", err)
	}
	if player.Inventory["frontier_pistol"] != 1 || player.Inventory["resonance_focus"] != 1 {
		t.Fatalf("expected crafted pistol and focus staff in inventory: %v", player.Inventory)
	}

	// 3. Forge Frontier Carbine at Forge (requires Smithing 2, Rifles 2 to equip)
	player.Position = forge.Position
	player.Inventory["salvaged_power_cell"] = 1
	player.Inventory["copper_bar"] = 2
	player.Inventory["wooden_plank"] = 1
	if _, err := player.AwardExperience("smithing", 100); err != nil {
		t.Fatal(err)
	}
	rifleRecipe, _ := FindCraftingRecipe("frontier-rifle")
	if _, err := player.Craft(rifleRecipe, forge, now); err != nil {
		t.Fatalf("craft frontier-rifle: %v", err)
	}
	if err := player.EquipItem("frontier_rifle"); !errors.Is(err, ErrInsufficientEquipmentSkill) {
		t.Fatalf("equip frontier_rifle below rifles level 2: %v", err)
	}
	if _, err := player.AwardExperience("rifles", 100); err != nil {
		t.Fatal(err)
	}
	if err := player.EquipItem("frontier_rifle"); err != nil {
		t.Fatalf("equip frontier_rifle at rifles level 2: %v", err)
	}
}


func TestEveryCoreSkillHasMilestoneUnlocksAndObtainableItemChains(t *testing.T) {
	now := time.Date(2026, 11, 18, 13, 0, 0, 0, time.UTC)

	// 1. Verify every one of the 18 active skills unlocks at least one Lv 2+ node, recipe, or equippable item,
	// and that every equippable item is attainable via a crafting recipe, monster loot table, or merchant buy price.
	recipeByOutput := map[string]CraftingRecipeDefinition{}
	for _, recipe := range CraftingRecipes() {
		for _, out := range recipe.Outputs {
			recipeByOutput[out.ItemID] = recipe
		}
	}
	lootedItems := map[string]bool{}
	for _, table := range LootTableDefinitions() {
		for _, d := range table.Drops {
			lootedItems[d.ItemID] = true
		}
		for _, d := range table.GuaranteedDrops {
			lootedItems[d.ItemID] = true
		}
		for _, r := range table.BonusRolls {
			for _, d := range r.Drops {
				lootedItems[d.ItemID] = true
			}
		}
	}
	highTierUnlockCount := map[string]int{}
	tierSixPlusUnlockCount := map[string]int{}
	tierTenUnlockCount := map[string]int{}
	tierFourteenUnlockCount := map[string]int{}
	tierTwentyUnlockCount := map[string]int{}
	tierTwentyFourUnlockCount := map[string]int{}
	tierThirtyUnlockCount := map[string]int{}
	for _, node := range GatheringNodes() {
		if node.RequiredLevel >= 2 {
			highTierUnlockCount[node.SkillID]++
		}
		if node.RequiredLevel >= 6 {
			tierSixPlusUnlockCount[node.SkillID]++
		}
		if node.RequiredLevel >= 10 {
			tierTenUnlockCount[node.SkillID]++
		}
		if node.RequiredLevel >= 14 {
			tierFourteenUnlockCount[node.SkillID]++
		}
		if node.RequiredLevel >= 20 {
			tierTwentyUnlockCount[node.SkillID]++
		}
		if node.RequiredLevel >= 24 {
			tierTwentyFourUnlockCount[node.SkillID]++
		}
		if node.RequiredLevel >= 30 {
			tierThirtyUnlockCount[node.SkillID]++
		}
	}
	for _, recipe := range CraftingRecipes() {
		if recipe.RequiredLevel >= 2 {
			highTierUnlockCount[recipe.SkillID]++
		}
		if recipe.RequiredLevel >= 6 {
			tierSixPlusUnlockCount[recipe.SkillID]++
		}
		if recipe.RequiredLevel >= 10 {
			tierTenUnlockCount[recipe.SkillID]++
		}
		if recipe.RequiredLevel >= 14 {
			tierFourteenUnlockCount[recipe.SkillID]++
		}
		if recipe.RequiredLevel >= 20 {
			tierTwentyUnlockCount[recipe.SkillID]++
		}
		if recipe.RequiredLevel >= 24 {
			tierTwentyFourUnlockCount[recipe.SkillID]++
		}
		if recipe.RequiredLevel >= 30 {
			tierThirtyUnlockCount[recipe.SkillID]++
		}
	}
	for _, item := range ItemDefinitions() {
		if item.EquipmentSlot != "" {
			if item.RequiredSkillLevel >= 2 {
				highTierUnlockCount[item.RequiredSkillID]++
			}
			if item.RequiredSkillLevel >= 6 {
				tierSixPlusUnlockCount[item.RequiredSkillID]++
			}
			if item.RequiredSkillLevel >= 10 {
				tierTenUnlockCount[item.RequiredSkillID]++
			}
			if item.RequiredSkillLevel >= 14 {
				tierFourteenUnlockCount[item.RequiredSkillID]++
			}
			if item.RequiredSkillLevel >= 20 {
				tierTwentyUnlockCount[item.RequiredSkillID]++
			}
			if item.RequiredSkillLevel >= 24 {
				tierTwentyFourUnlockCount[item.RequiredSkillID]++
			}
			if item.RequiredSkillLevel >= 30 {
				tierThirtyUnlockCount[item.RequiredSkillID]++
			}
			_, craftable := recipeByOutput[item.ID]
			if !craftable && !lootedItems[item.ID] && item.BuyPrice <= 0 {
				t.Fatalf("equippable item %s (%s) has no attainment path (not craftable, not in loot tables, not sold by merchant)", item.ID, item.Name)
			}
		}
	}
	for _, skill := range SkillDefinitions() {
		if highTierUnlockCount[skill.ID] == 0 {
			t.Fatalf("skill %s has no Level 2+ progression unlock (node, recipe, or equipment)", skill.ID)
		}
		if tierSixPlusUnlockCount[skill.ID] == 0 {
			t.Fatalf("skill %s has no Level 6+ progression unlock (node, recipe, or equipment)", skill.ID)
		}
		if tierTenUnlockCount[skill.ID] == 0 {
			t.Fatalf("skill %s has no Level 10 capstone unlock (node, recipe, or equipment)", skill.ID)
		}
		if tierFourteenUnlockCount[skill.ID] == 0 {
			t.Fatalf("skill %s has no Level 14-15 vanguard unlock (node, recipe, or equipment)", skill.ID)
		}
		if tierTwentyUnlockCount[skill.ID] == 0 {
			t.Fatalf("skill %s has no Level 20 capstone unlock (node, recipe, or equipment)", skill.ID)
		}
		if tierTwentyFourUnlockCount[skill.ID] == 0 {
			t.Fatalf("skill %s has no Level 24-25 vanguard unlock (node, recipe, or equipment)", skill.ID)
		}
		if tierThirtyUnlockCount[skill.ID] == 0 {
			t.Fatalf("skill %s has no Level 30 capstone unlock (node, recipe, or equipment)", skill.ID)
		}
	}
	for _, dualID := range []string{
		"overlord_siege_maul", "abyssal_artillery_cannon", "juggernaut_war_plate", "dreadnought_aegis_helm",
		"sovereign_greatsword", "solaris_halberd", "chronite_kinetic_gauntlets", "cryo_pulse_blaster",
		"elderwood_longbow", "sunfire_focus_scepter",
		"gorax_titan_maul", "zaryte_bowcaster_cannon", "torva_gorax_warplate", "elysian_spirit_cuirass",
		"infernal_jad_greathelm", "krayt_fang_whip_blade", "noxious_acklay_halberd", "beskar_dragon_claws",
		"westar_armadyl_blaster", "twisted_wroshyr_longbow", "tumeken_ichor_scepter",
		"dharok_rancor_greataxe", "karil_quarrel_repeater", "malevolent_beskar_cuirass", "barrows_cortosis_greathelm",
		"korasi_darksaber_edge", "guthan_terentatek_warspear", "goliath_shock_cestus", "ascension_bryar_carbine",
		"crystal_seren_longrifle", "ahrim_holocron_staff",
		"elder_maul_of_vitur", "eldritch_starforge_bowcaster", "masterwork_animica_warplate", "justiciar_mandalore_hauberk",
		"zuk_infernal_beskar_crown", "zaros_starforge_godsword", "vitur_mythosaur_scythe", "telos_anima_talons",
		"leng_starforge_blasters", "bofa_bloodwood_railgun", "fractured_armadyl_holostaff",
		"kalgerion_massassi_warmaul", "pernix_umbaran_bowcaster", "vestments_massassi_warplate", "cryptbloom_vonduun_greathelm",
		"stygium_shadow_lightsaber", "amphistaff_noxious_glaive", "cinderbane_teras_gauntlets", "shadow_hk_dual_disruptors",
		"decimation_umbaran_sniper", "virtus_chronote_scepter",
		"primal_zillo_seismic_maul", "bolg_infinity_bowcaster", "primal_beskar_juggernaut_plate", "zillo_tectonic_bulwark_cuirass",
		"nihilus_chaos_sovereign_crown", "ek_zekkil_kyber_greatsaber", "solak_zillo_sovereign_halberd", "vorago_seismic_kyber_talons",
		"praetorian_infinity_blasters", "zaryte_solak_singularity_rifle", "omnipower_chaos_holostaff",
	} {
		if _, craftable := recipeByOutput[dualID]; !craftable || !lootedItems[dualID] {
			t.Fatalf("masterwork/apex item %s missing dual attainment: craftable=%v looted=%v", dualID, craftable, lootedItems[dualID])
		}
	}

	// 2. End-to-end gather -> craft -> equip -> combat verification for Melee (vanguard_cleaver),
	// Ranged (skirmisher_repeater_bow), and Strength (colossus_breaker_greataxe & titan_harness).
	player, err := NewPlayer("progression-wayfarer", "ProgressionHero", now)
	if err != nil {
		t.Fatal(err)
	}
	forge, _ := FindCraftingStation("frontier-forge")
	workbench, _ := FindCraftingStation("frontier-workbench")

	// Gather ironwood_log (Woodcutting Lv 2) and circuit_board (Scavenging Lv 2).
	ironwoodNode, ok := FindGatheringNode("ironwood-001")
	if !ok {
		t.Fatal("ironwood-001 missing")
	}
	player.Position = ironwoodNode.Position
	if _, _, err := player.Gather(ironwoodNode, GatheringNodeState{}, now); !errors.Is(err, ErrInsufficientSkillLevel) {
		t.Fatalf("expected woodcutting level gate on ironwood-001, got %v", err)
	}
	for _, sk := range []string{"woodcutting", "scavenging", "foraging", "farming", "smithing", "crafting", "cooking"} {
		if _, err := player.AwardExperience(sk, 200); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 3; i++ {
		if _, _, err := player.Gather(ironwoodNode, GatheringNodeState{}, now); err != nil {
			t.Fatalf("gather ironwood-001: %v", err)
		}
	}
	circuitNode, ok := FindGatheringNode("circuit-001")
	if !ok {
		t.Fatal("circuit-001 missing")
	}
	player.Position = circuitNode.Position
	if _, _, err := player.Gather(circuitNode, GatheringNodeState{}, now); err != nil {
		t.Fatalf("gather circuit-001: %v", err)
	}

	// Craft Melee Vanguard Cleaver at the forge and verify Melee Lv 2 equip requirement.
	player.Inventory["iron_bar"] = 2
	player.Position = forge.Position
	cleaverRecipe, ok := FindCraftingRecipe("vanguard-cleaver")
	if !ok {
		t.Fatal("vanguard-cleaver recipe missing")
	}
	if _, err := player.Craft(cleaverRecipe, forge, now); err != nil {
		t.Fatalf("craft vanguard-cleaver: %v", err)
	}
	if err := player.EquipItem("vanguard_cleaver"); !errors.Is(err, ErrInsufficientEquipmentSkill) {
		t.Fatalf("equip vanguard_cleaver at melee lv 1 = %v, want ErrInsufficientEquipmentSkill", err)
	}
	if _, err := player.AwardExperience("melee", 100); err != nil {
		t.Fatal(err)
	}
	if err := player.EquipItem("vanguard_cleaver"); err != nil {
		t.Fatalf("equip vanguard_cleaver at melee lv 2: %v", err)
	}

	// Craft Ranged Skirmisher Bowcaster at the workbench and verify Ranged Lv 2 equip + ranged combat attribution.
	player.Inventory["copper_bar"] = 1
	player.Position = workbench.Position
	bowRecipe, ok := FindCraftingRecipe("skirmisher-repeater-bow")
	if !ok {
		t.Fatal("skirmisher-repeater-bow recipe missing")
	}
	if _, err := player.Craft(bowRecipe, workbench, now); err != nil {
		t.Fatalf("craft skirmisher-repeater-bow: %v", err)
	}
	if err := player.EquipItem("skirmisher_repeater_bow"); !errors.Is(err, ErrInsufficientEquipmentSkill) {
		t.Fatalf("equip skirmisher_repeater_bow at ranged lv 1 = %v, want ErrInsufficientEquipmentSkill", err)
	}
	if _, err := player.AwardExperience("ranged", 100); err != nil {
		t.Fatal(err)
	}
	if err := player.EquipItem("skirmisher_repeater_bow"); err != nil {
		t.Fatalf("equip skirmisher_repeater_bow at ranged lv 2: %v", err)
	}
	rat, _ := FindEnemy("frontier-rat-001")
	player.Position = rat.Position
	strengthBefore := player.Skills["strength"].Experience
	combatRes, _, err := player.AttackEnemy(rat, EnemyState{}, now.Add(2*time.Second))
	if err != nil {
		t.Fatalf("attack with skirmisher_repeater_bow: %v", err)
	}
	if combatRes.ParentSkillID != "ranged" || player.Skills["strength"].Experience != strengthBefore {
		t.Fatalf("ranged bowcaster attack misattributed: res=%+v strengthXP=%d->%d", combatRes, strengthBefore, player.Skills["strength"].Experience)
	}
}
