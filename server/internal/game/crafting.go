package game

import (
	"errors"
	"math"
	"time"
)

var (
	ErrUnknownCraftingRecipe   = errors.New("unknown crafting recipe")
	ErrUnknownCraftingStation  = errors.New("unknown crafting station")
	ErrCraftingOutOfRange      = errors.New("crafting station is out of range")
	ErrInsufficientIngredients = errors.New("missing crafting ingredients")
	ErrCraftingSkillTooLow     = errors.New("crafting skill level is too low")
	ErrCraftingOutputFull      = errors.New("crafting output stack is full")
)

type ItemQuantity struct {
	ItemID   string `json:"itemId"`
	Quantity int    `json:"quantity"`
}

type CraftingStationDefinition struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Position         WorldPosition `json:"position"`
	InteractionRange float64       `json:"interactionRange"`
}

type CraftingRecipeDefinition struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	SkillID       string         `json:"skillId"`
	RequiredLevel int            `json:"requiredLevel"`
	StationID     string         `json:"stationId"`
	Ingredients   []ItemQuantity `json:"ingredients"`
	Outputs       []ItemQuantity `json:"outputs"`
	Experience    int            `json:"experience"`
}

type CraftingResult struct {
	RecipeID   string         `json:"recipeId"`
	Outputs    []ItemQuantity `json:"outputs"`
	SkillID    string         `json:"skillId"`
	Experience int            `json:"experience"`
	Skill      SkillProgress  `json:"skill"`
	LeveledUp  bool           `json:"leveledUp"`
}

var craftingStations = []CraftingStationDefinition{
	{ID: "frontier-workbench", Name: "Frontier Workbench", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 0, Z: 4}, InteractionRange: 3},
	{ID: "frontier-forge", Name: "Frontier Forge", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 0, Z: 8}, InteractionRange: 3},
}

var craftingRecipes = []CraftingRecipeDefinition{
	{ID: "basic-plank", Name: "Basic Wooden Plank", SkillID: "crafting", RequiredLevel: 1, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "softwood_log", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "wooden_plank", Quantity: 1}}, Experience: 30},
	{ID: "copper-bar", Name: "Smelt Copper Bar", SkillID: "smithing", RequiredLevel: 1, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "copper_ore", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "copper_bar", Quantity: 1}}, Experience: 35},
	{ID: "copper-sword", Name: "Forge Training Sword", SkillID: "smithing", RequiredLevel: 1, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "copper_bar", Quantity: 3}, {ItemID: "wooden_plank", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "training_sword", Quantity: 1}}, Experience: 60},
	{ID: "rat-tooth-blade", Name: "Forge Rat-Tooth Blade", SkillID: "smithing", RequiredLevel: 2, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "copper_bar", Quantity: 2}, {ItemID: "rat_tooth", Quantity: 2}, {ItemID: "wooden_plank", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "rat_tooth_blade", Quantity: 1}}, Experience: 90},
	{ID: "boar-tusk-spear", Name: "Forge Boar-Tusk Spear", SkillID: "smithing", RequiredLevel: 3, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "copper_bar", Quantity: 2}, {ItemID: "boar_tusk", Quantity: 2}, {ItemID: "wooden_plank", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "boar_tusk_spear", Quantity: 1}}, Experience: 120},
	{ID: "herbal-bundle", Name: "Pack Herbal Bundle", SkillID: "crafting", RequiredLevel: 1, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "wild_herb", Quantity: 3}, {ItemID: "wooden_plank", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "herbal_bundle", Quantity: 1}}, Experience: 45},
	{ID: "herbal-tea", Name: "Brew Herbal Tea", SkillID: "cooking", RequiredLevel: 1, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "herbal_bundle", Quantity: 1}, {ItemID: "wild_herb", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "herbal_tea", Quantity: 1}}, Experience: 35},
	{ID: "frontier-ration", Name: "Bake Frontier Ration", SkillID: "cooking", RequiredLevel: 1, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "frontier_grain", Quantity: 2}, {ItemID: "wild_herb", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "frontier_ration", Quantity: 1}}, Experience: 45},
	{ID: "salvaged-power-cell", Name: "Assemble Power Cell", SkillID: "crafting", RequiredLevel: 1, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "salvaged_scrap", Quantity: 2}, {ItemID: "copper_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "salvaged_power_cell", Quantity: 1}}, Experience: 50},
	{ID: "frontier-pistol", Name: "Assemble Frontier Pistol", SkillID: "crafting", RequiredLevel: 1, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "salvaged_scrap", Quantity: 2}, {ItemID: "copper_bar", Quantity: 1}, {ItemID: "wooden_plank", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "frontier_pistol", Quantity: 1}}, Experience: 65},
	{ID: "frontier-rifle", Name: "Forge Frontier Carbine", SkillID: "smithing", RequiredLevel: 2, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "salvaged_power_cell", Quantity: 1}, {ItemID: "copper_bar", Quantity: 2}, {ItemID: "wooden_plank", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "frontier_rifle", Quantity: 1}}, Experience: 95},
	{ID: "resonance-focus", Name: "Craft Resonance Focus Staff", SkillID: "crafting", RequiredLevel: 1, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "salvaged_power_cell", Quantity: 1}, {ItemID: "wooden_plank", Quantity: 2}, {ItemID: "wild_herb", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "resonance_focus", Quantity: 1}}, Experience: 70},
	// Alloy Smelting (5 new alloy recipes)
	{ID: "iron-bar", Name: "Smelt Iron Bar", SkillID: "smithing", RequiredLevel: 2, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "iron_ore", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "iron_bar", Quantity: 1}}, Experience: 55},
	{ID: "cobalt-bar", Name: "Smelt Cobalt Bar", SkillID: "smithing", RequiredLevel: 2, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "cobalt_ore", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "cobalt_bar", Quantity: 1}}, Experience: 65},
	{ID: "titanium-bar", Name: "Smelt Titanium Bar", SkillID: "smithing", RequiredLevel: 3, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "titanium_ore", Quantity: 2}, {ItemID: "iron_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "titanium_bar", Quantity: 1}}, Experience: 85},
	{ID: "durasteel-bar", Name: "Smelt Durasteel Bar", SkillID: "smithing", RequiredLevel: 4, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "durasteel_ore", Quantity: 2}, {ItemID: "titanium_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "durasteel_bar", Quantity: 1}}, Experience: 115},
	{ID: "mythril-bar", Name: "Smelt Mythril Bar", SkillID: "smithing", RequiredLevel: 5, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "mythril_ore", Quantity: 2}, {ItemID: "luminite_gem", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "mythril_bar", Quantity: 1}}, Experience: 155},
	// Medical Packs & Attribute Buff Enhancers (8 medical/stim recipes)
	{ID: "bacta-gel", Name: "Distill Bio-Gel Compound", SkillID: "cooking", RequiredLevel: 2, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "wild_herb", Quantity: 2}, {ItemID: "rough_stone", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "bacta_gel", Quantity: 1}}, Experience: 50},
	{ID: "field-medpack", Name: "Assemble Field Medical Pack", SkillID: "crafting", RequiredLevel: 2, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "bacta_gel", Quantity: 1}, {ItemID: "herbal_bundle", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "field_medpack", Quantity: 1}}, Experience: 65},
	{ID: "advanced-medpack", Name: "Synthesize Advanced Trauma Medpack", SkillID: "crafting", RequiredLevel: 3, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "bacta_gel", Quantity: 2}, {ItemID: "stalker_venom_sac", Quantity: 1}, {ItemID: "aether_crystal", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "advanced_medpack", Quantity: 1}}, Experience: 110},
	{ID: "surgical-trauma-kit", Name: "Assemble Surgical Trauma Kit", SkillID: "crafting", RequiredLevel: 4, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "bacta_gel", Quantity: 2}, {ItemID: "drake_bile", Quantity: 1}, {ItemID: "luminite_gem", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "surgical_trauma_kit", Quantity: 1}}, Experience: 145},
	{ID: "vitality-stim-pack", Name: "Synthesize Constitution Vitality Stim", SkillID: "cooking", RequiredLevel: 2, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "bacta_gel", Quantity: 1}, {ItemID: "frontier_grain", Quantity: 2}, {ItemID: "aether_crystal", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "vitality_stim_pack", Quantity: 1}}, Experience: 80},
	{ID: "adrenal-action-booster", Name: "Synthesize Stamina Action Booster", SkillID: "cooking", RequiredLevel: 2, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "bacta_gel", Quantity: 1}, {ItemID: "hound_fang", Quantity: 1}, {ItemID: "wild_herb", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "adrenal_action_booster", Quantity: 1}}, Experience: 85},
	{ID: "neural-mind-tonic", Name: "Brew Willpower Clarity Tonic", SkillID: "cooking", RequiredLevel: 2, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "aether_crystal", Quantity: 1}, {ItemID: "herbal_bundle", Quantity: 1}, {ItemID: "wild_herb", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "neural_mind_tonic", Quantity: 1}}, Experience: 85},
	{ID: "overdrive-omni-stim", Name: "Synthesize Tri-Stat Overdrive Stim", SkillID: "cooking", RequiredLevel: 4, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "bacta_gel", Quantity: 2}, {ItemID: "wraith_essence", Quantity: 1}, {ItemID: "luminite_gem", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "overdrive_omni_stim", Quantity: 1}}, Experience: 160},
	// High-Tier Weapons & Armor (9 equipment recipes)
	{ID: "iron-greatsword", Name: "Forge Iron Broadsword", SkillID: "smithing", RequiredLevel: 3, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "iron_bar", Quantity: 3}, {ItemID: "hound_fang", Quantity: 2}, {ItemID: "wooden_plank", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "iron_greatsword", Quantity: 1}}, Experience: 135},
	{ID: "cobalt-repeater", Name: "Assemble Cobalt Blaster Pistol", SkillID: "crafting", RequiredLevel: 3, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "cobalt_bar", Quantity: 2}, {ItemID: "salvaged_power_cell", Quantity: 1}, {ItemID: "wooden_plank", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "cobalt_repeater", Quantity: 1}}, Experience: 125},
	{ID: "titanium-carbine", Name: "Forge Titanium Pulse Rifle", SkillID: "smithing", RequiredLevel: 4, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "titanium_bar", Quantity: 2}, {ItemID: "sentinel_core", Quantity: 1}, {ItemID: "salvaged_power_cell", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "titanium_carbine", Quantity: 1}}, Experience: 165},
	{ID: "durasteel-halberd", Name: "Forge Durasteel Halberd", SkillID: "smithing", RequiredLevel: 4, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "durasteel_bar", Quantity: 2}, {ItemID: "prowler_claw", Quantity: 2}, {ItemID: "wooden_plank", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "durasteel_halberd", Quantity: 1}}, Experience: 175},
	{ID: "aether-scepter", Name: "Craft Luminite Aether Staff", SkillID: "crafting", RequiredLevel: 4, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "luminite_gem", Quantity: 1}, {ItemID: "wraith_essence", Quantity: 1}, {ItemID: "cobalt_bar", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "aether_scepter", Quantity: 1}}, Experience: 170},
	{ID: "mythril-saber", Name: "Forge Mythril Edge", SkillID: "smithing", RequiredLevel: 5, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "mythril_bar", Quantity: 2}, {ItemID: "luminite_gem", Quantity: 1}, {ItemID: "wooden_plank", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "mythril_saber", Quantity: 1}}, Experience: 210},
	{ID: "chitin-cuirass", Name: "Forge Stalker Chitin Cuirass", SkillID: "smithing", RequiredLevel: 3, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "stalker_chitin", Quantity: 3}, {ItemID: "iron_bar", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "chitin_cuirass", Quantity: 1}}, Experience: 140},
	{ID: "durasteel-helm", Name: "Forge Durasteel Visor Helm", SkillID: "smithing", RequiredLevel: 4, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "durasteel_bar", Quantity: 2}, {ItemID: "drake_scale", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "durasteel_helm", Quantity: 1}}, Experience: 160},
	{ID: "prowler-striders", Name: "Craft Frost Prowler Greaves", SkillID: "crafting", RequiredLevel: 3, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "prowler_pelt", Quantity: 2}, {ItemID: "titanium_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "prowler_striders", Quantity: 1}}, Experience: 145},
	{ID: "orichalcum-bar", Name: "Smelt Orichalcum Bar", SkillID: "smithing", RequiredLevel: 5, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "orichalcum_ore", Quantity: 2}, {ItemID: "obsidian_shard", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "orichalcum_bar", Quantity: 1}}, Experience: 165},
	{ID: "neutronium-bar", Name: "Smelt Neutronium Bar", SkillID: "smithing", RequiredLevel: 6, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "neutronium_ore", Quantity: 2}, {ItemID: "durasteel_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "neutronium_bar", Quantity: 1}}, Experience: 195},
	{ID: "starsteel-bar", Name: "Smelt Star-Metal Bar", SkillID: "smithing", RequiredLevel: 7, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "starsteel_ore", Quantity: 2}, {ItemID: "mythril_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "starsteel_bar", Quantity: 1}}, Experience: 240},
	{ID: "nanite-repair-injector", Name: "Assemble Nanite Trauma Injector", SkillID: "crafting", RequiredLevel: 5, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "bacta_gel", Quantity: 2}, {ItemID: "scavenger_servo", Quantity: 1}, {ItemID: "void_quartz", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "nanite_repair_injector", Quantity: 1}}, Experience: 185},
	{ID: "titan-constitution-elixir", Name: "Distill Titan Constitution Elixir", SkillID: "cooking", RequiredLevel: 5, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "bacta_gel", Quantity: 2}, {ItemID: "wyrm_heartstone", Quantity: 1}, {ItemID: "void_quartz", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "titan_constitution_elixir", Quantity: 1}}, Experience: 195},
	{ID: "orichalcum-glaive", Name: "Forge Orichalcum Sonic Glaive", SkillID: "smithing", RequiredLevel: 5, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "orichalcum_bar", Quantity: 2}, {ItemID: "behemoth_horn", Quantity: 2}, {ItemID: "wooden_plank", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "orichalcum_glaive", Quantity: 1}}, Experience: 225},
	{ID: "neutronium-platebody", Name: "Forge Neutronium Vanguard Cuirass", SkillID: "smithing", RequiredLevel: 6, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "neutronium_bar", Quantity: 2}, {ItemID: "wyrm_carapace", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "neutronium_platebody", Quantity: 1}}, Experience: 250},
	{ID: "void-psionic-crown", Name: "Craft Void Quartz Circlet", SkillID: "crafting", RequiredLevel: 5, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "void_quartz", Quantity: 2}, {ItemID: "orichalcum_bar", Quantity: 1}, {ItemID: "aether_crystal", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "void_psionic_crown", Quantity: 1}}, Experience: 215},
	{ID: "starsteel-longblade", Name: "Forge Star-Metal Vanguard Blade", SkillID: "smithing", RequiredLevel: 7, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "starsteel_bar", Quantity: 2}, {ItemID: "reaver_plasma_coil", Quantity: 1}, {ItemID: "wooden_plank", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "starsteel_longblade", Quantity: 1}}, Experience: 285},
	{ID: "electrum-bar", Name: "Smelt Electrum Bar", SkillID: "smithing", RequiredLevel: 6, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "electrum_ore", Quantity: 2}, {ItemID: "cobalt_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "electrum_bar", Quantity: 1}}, Experience: 180},
	{ID: "chronite-bar", Name: "Smelt Chronite Bar", SkillID: "smithing", RequiredLevel: 8, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "chronite_ore", Quantity: 2}, {ItemID: "solarite_crystal", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "chronite_bar", Quantity: 1}}, Experience: 255},
	{ID: "abyssal-bar", Name: "Smelt Abyssal Adamant Bar", SkillID: "smithing", RequiredLevel: 9, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "abyssal_ore", Quantity: 2}, {ItemID: "neutronium_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "abyssal_bar", Quantity: 1}}, Experience: 310},
	{ID: "phoenix-resuscitation-pack", Name: "Synthesize Phoenix Resuscitation Medpack", SkillID: "crafting", RequiredLevel: 6, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "bacta_gel", Quantity: 3}, {ItemID: "scorpion_stinger", Quantity: 1}, {ItemID: "solarite_crystal", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "phoenix_resuscitation_pack", Quantity: 1}}, Experience: 230},
	{ID: "celestial-harmony-draught", Name: "Brew Celestial Tri-Stat Draught", SkillID: "cooking", RequiredLevel: 7, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "bacta_gel", Quantity: 2}, {ItemID: "specter_shroud", Quantity: 1}, {ItemID: "nova_prism", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "celestial_harmony_draught", Quantity: 1}}, Experience: 275},
	{ID: "electrum-knuckles", Name: "Craft Electrum Shock Cestus", SkillID: "crafting", RequiredLevel: 6, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "electrum_bar", Quantity: 2}, {ItemID: "mantis_scythe", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "electrum_knuckles", Quantity: 1}}, Experience: 220},
	{ID: "chronite-sniper-rifle", Name: "Forge Chronite Phase Longrifle", SkillID: "smithing", RequiredLevel: 8, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "chronite_bar", Quantity: 2}, {ItemID: "solarite_crystal", Quantity: 1}, {ItemID: "leviathan_fang", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "chronite_sniper_rifle", Quantity: 1}}, Experience: 315},
	{ID: "abyssal-bulwark-armor", Name: "Forge Abyssal Adamant Bulwark", SkillID: "smithing", RequiredLevel: 9, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "abyssal_bar", Quantity: 2}, {ItemID: "colossus_crab_shell", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "abyssal_bulwark_armor", Quantity: 1}}, Experience: 350},
	// Multi-Skill Milestone Weapons, Armor & Culinary Recipes (+14 recipes)
	{ID: "brawler-handwraps", Name: "Stitch Frontier Brawler Handwraps", SkillID: "crafting", RequiredLevel: 1, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "rat_pelt", Quantity: 2}, {ItemID: "softwood_log", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "brawler_handwraps", Quantity: 1}}, Experience: 45},
	{ID: "spiked-knuckledusters", Name: "Forge Iron Knuckledusters", SkillID: "smithing", RequiredLevel: 3, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "iron_bar", Quantity: 2}, {ItemID: "stalker_chitin", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "spiked_knuckledusters", Quantity: 1}}, Experience: 120},
	{ID: "crystal-channeled-rod", Name: "Craft Aether Channeling Rod", SkillID: "crafting", RequiredLevel: 2, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "ironwood_log", Quantity: 2}, {ItemID: "aether_crystal", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "crystal_channeled_rod", Quantity: 1}}, Experience: 90},
	{ID: "void-harmonic-staff", Name: "Craft Void Harmonic Grand Staff", SkillID: "crafting", RequiredLevel: 6, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "petrified_starwood", Quantity: 2}, {ItemID: "void_quartz", Quantity: 1}, {ItemID: "quantum_capacitor", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "void_harmonic_staff", Quantity: 1}}, Experience: 235},
	{ID: "pulse-needler-pistol", Name: "Assemble Quantum Pulse Sidearm", SkillID: "crafting", RequiredLevel: 5, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "durasteel_bar", Quantity: 1}, {ItemID: "quantum_capacitor", Quantity: 1}, {ItemID: "circuit_board", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "pulse_needler_pistol", Quantity: 1}}, Experience: 195},
	{ID: "vanguard-cleaver", Name: "Forge Melee Vanguard Cleaver", SkillID: "smithing", RequiredLevel: 2, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "iron_bar", Quantity: 2}, {ItemID: "ironwood_log", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "vanguard_cleaver", Quantity: 1}}, Experience: 85},
	{ID: "warlord-shock-maul", Name: "Forge Melee Warlord Shock Maul", SkillID: "smithing", RequiredLevel: 4, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "durasteel_bar", Quantity: 2}, {ItemID: "petrified_starwood", Quantity: 1}, {ItemID: "circuit_board", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "warlord_shock_maul", Quantity: 1}}, Experience: 170},
	{ID: "skirmisher-repeater-bow", Name: "Assemble Ranged Skirmisher Bowcaster", SkillID: "crafting", RequiredLevel: 2, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "ironwood_log", Quantity: 2}, {ItemID: "circuit_board", Quantity: 1}, {ItemID: "copper_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "skirmisher_repeater_bow", Quantity: 1}}, Experience: 90},
	{ID: "deadeye-rail-cannon", Name: "Assemble Ranged Deadeye Rail-Arbalest", SkillID: "crafting", RequiredLevel: 4, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "petrified_starwood", Quantity: 1}, {ItemID: "titanium_bar", Quantity: 2}, {ItemID: "quantum_capacitor", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "deadeye_rail_cannon", Quantity: 1}}, Experience: 175},
	{ID: "colossus-breaker-greataxe", Name: "Forge Strength Breaker Greataxe", SkillID: "smithing", RequiredLevel: 2, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "iron_bar", Quantity: 2}, {ItemID: "boar_tusk", Quantity: 2}, {ItemID: "ironwood_log", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "colossus_breaker_greataxe", Quantity: 1}}, Experience: 95},
	{ID: "titan-harness", Name: "Forge Titan Strength Harness", SkillID: "smithing", RequiredLevel: 4, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "durasteel_bar", Quantity: 2}, {ItemID: "behemoth_horn", Quantity: 1}, {ItemID: "boar_hide", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "titan_harness", Quantity: 1}}, Experience: 180},
	{ID: "studded-frontier-vest", Name: "Craft Studded Boar-Hide Vest", SkillID: "crafting", RequiredLevel: 2, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "boar_hide", Quantity: 2}, {ItemID: "copper_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "studded_frontier_vest", Quantity: 1}}, Experience: 80},
	{ID: "hearty-tuber-stew", Name: "Simmer Hearty Tuber & Glowcap Stew", SkillID: "cooking", RequiredLevel: 3, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "tuber_root", Quantity: 2}, {ItemID: "glowcap_fungus", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "hearty_tuber_stew", Quantity: 1}}, Experience: 110},
	{ID: "starberry-ambrosia", Name: "Prepare Solaris Starberry Banquet", SkillID: "cooking", RequiredLevel: 4, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "starberry_cluster", Quantity: 2}, {ItemID: "sunblossom_petal", Quantity: 1}, {ItemID: "frontier_grain", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "starberry_ambrosia", Quantity: 1}}, Experience: 150},
	// Tier 6-8 Masterwork Equipment & Culinary Recipes (+12 recipes -> 77 total)
	{ID: "overlord-siege-maul", Name: "Forge Melee Overlord Siege Maul", SkillID: "smithing", RequiredLevel: 7, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "neutronium_bar", Quantity: 2}, {ItemID: "titan_actuator_coil", Quantity: 1}, {ItemID: "aether_bough_timber", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "overlord_siege_maul", Quantity: 1}}, Experience: 270},
	{ID: "abyssal-artillery-cannon", Name: "Assemble Ranged Abyssal Siege Arbalest", SkillID: "crafting", RequiredLevel: 7, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "aether_bough_timber", Quantity: 2}, {ItemID: "singularity_core_cell", Quantity: 1}, {ItemID: "neutronium_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "abyssal_artillery_cannon", Quantity: 1}}, Experience: 275},
	{ID: "juggernaut-war-plate", Name: "Forge Strength Juggernaut War-Harness", SkillID: "smithing", RequiredLevel: 7, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "neutronium_bar", Quantity: 2}, {ItemID: "titan_actuator_coil", Quantity: 1}, {ItemID: "glacial_core_shard", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "juggernaut_war_plate", Quantity: 1}}, Experience: 280},
	{ID: "dreadnought-aegis-helm", Name: "Forge Defense Dreadnought Aegis Greathelm", SkillID: "smithing", RequiredLevel: 8, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "abyssal_bar", Quantity: 1}, {ItemID: "abyssal_behemoth_plate", Quantity: 2}, {ItemID: "chronite_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "dreadnought_aegis_helm", Quantity: 1}}, Experience: 325},
	{ID: "sovereign-greatsword", Name: "Forge Chronite Sovereign Greatsword", SkillID: "smithing", RequiredLevel: 8, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "chronite_bar", Quantity: 2}, {ItemID: "ancient_heartwood_sap", Quantity: 1}, {ItemID: "voidbark_heartwood", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "sovereign_greatsword", Quantity: 1}}, Experience: 330},
	{ID: "solaris-halberd", Name: "Forge Solaris Thermal Halberd", SkillID: "smithing", RequiredLevel: 7, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "starsteel_bar", Quantity: 2}, {ItemID: "solaris_fire_gland", Quantity: 1}, {ItemID: "aether_bough_timber", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "solaris_halberd", Quantity: 1}}, Experience: 285},
	{ID: "chronite-kinetic-gauntlets", Name: "Assemble Chronite Kinetic Gauntlets", SkillID: "crafting", RequiredLevel: 8, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "chronite_bar", Quantity: 1}, {ItemID: "temporal_phase_tendril", Quantity: 2}, {ItemID: "singularity_core_cell", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "chronite_kinetic_gauntlets", Quantity: 1}}, Experience: 320},
	{ID: "cryo-pulse-blaster", Name: "Assemble Glacial Cryo-Pulse Repeater", SkillID: "crafting", RequiredLevel: 7, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "electrum_bar", Quantity: 2}, {ItemID: "glacial_core_shard", Quantity: 1}, {ItemID: "cryo_lotus_petal", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "cryo_pulse_blaster", Quantity: 1}}, Experience: 265},
	{ID: "elderwood-longbow", Name: "Craft Voidbark Singularity Longrifle", SkillID: "crafting", RequiredLevel: 8, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "voidbark_heartwood", Quantity: 2}, {ItemID: "singularity_core_cell", Quantity: 1}, {ItemID: "chronite_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "elderwood_longbow", Quantity: 1}}, Experience: 335},
	{ID: "sunfire-focus-scepter", Name: "Craft Astral Nova Scepter", SkillID: "crafting", RequiredLevel: 8, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "aether_bough_timber", Quantity: 1}, {ItemID: "astral_mycelium", Quantity: 2}, {ItemID: "solaris_fire_gland", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "sunfire_focus_scepter", Quantity: 1}}, Experience: 325},
	{ID: "auric-harvest-feast", Name: "Roast Auric Harvest Banquet", SkillID: "cooking", RequiredLevel: 6, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "auric_grain_sheaf", Quantity: 2}, {ItemID: "cryo_lotus_petal", Quantity: 1}, {ItemID: "bacta_gel", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "auric_harvest_feast", Quantity: 1}}, Experience: 225},
	{ID: "astral-ambrosia-elixir", Name: "Distill Celestial Astral Nectar", SkillID: "cooking", RequiredLevel: 8, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "celestial_nectar_fruit", Quantity: 2}, {ItemID: "astral_mycelium", Quantity: 1}, {ItemID: "ancient_heartwood_sap", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "astral_ambrosia_elixir", Quantity: 1}}, Experience: 310},
	// Tier 9-10 Capstone Smelting, Apex Weapons/Armor & Overload Consumable Recipes (+15 recipes -> 92 total)
	{ID: "runic-banite-bar", Name: "Smelt Runic Banite Alloy Bar", SkillID: "smithing", RequiredLevel: 10, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "runic_banite_ore", Quantity: 2}, {ItemID: "nova_prism", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "runic_banite_bar", Quantity: 1}}, Experience: 380},
	{ID: "gorax-titan-maul", Name: "Forge Elder Gorax-Maul", SkillID: "smithing", RequiredLevel: 10, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "runic_banite_bar", Quantity: 2}, {ItemID: "bandosian_gorax_tasset", Quantity: 1}, {ItemID: "elder_wroshyr_timber", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "gorax_titan_maul", Quantity: 1}}, Experience: 420},
	{ID: "zaryte-bowcaster-cannon", Name: "Assemble Zaryte-Wroshyr Heavy Bowcaster", SkillID: "crafting", RequiredLevel: 10, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "elder_wroshyr_timber", Quantity: 2}, {ItemID: "hypergate_relic_core", Quantity: 1}, {ItemID: "noxious_reaper_limb", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "zaryte_bowcaster_cannon", Quantity: 1}}, Experience: 425},
	{ID: "torva-gorax-warplate", Name: "Forge Torva-Gorax Sovereign Harness", SkillID: "smithing", RequiredLevel: 10, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "runic_banite_bar", Quantity: 2}, {ItemID: "bandosian_gorax_tasset", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "torva_gorax_warplate", Quantity: 1}}, Experience: 435},
	{ID: "elysian-spirit-cuirass", Name: "Forge Elysian Krayt-Scale Cuirass", SkillID: "smithing", RequiredLevel: 10, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "ancient_krayt_dragon_scale", Quantity: 2}, {ItemID: "ichor_spirit_sigil", Quantity: 1}, {ItemID: "runic_banite_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "elysian_spirit_cuirass", Quantity: 1}}, Experience: 445},
	{ID: "infernal-jad-greathelm", Name: "Forge Infernal Jad-Crest Greathelm", SkillID: "smithing", RequiredLevel: 10, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "sovereign_jad_fire_core", Quantity: 1}, {ItemID: "ancient_krayt_dragon_scale", Quantity: 1}, {ItemID: "runic_banite_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "infernal_jad_greathelm", Quantity: 1}}, Experience: 430},
	{ID: "krayt-fang-whip-blade", Name: "Forge Abyssal Krayt-Pearl Saber", SkillID: "smithing", RequiredLevel: 10, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "runic_banite_bar", Quantity: 2}, {ItemID: "krayt_wyrm_pearl", Quantity: 1}, {ItemID: "hypergate_relic_core", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "krayt_fang_whip_blade", Quantity: 1}}, Experience: 440},
	{ID: "noxious-acklay-halberd", Name: "Forge Noxious Acklay-Scythe Halberd", SkillID: "smithing", RequiredLevel: 10, StationID: "frontier-forge", Ingredients: []ItemQuantity{{ItemID: "runic_banite_bar", Quantity: 1}, {ItemID: "noxious_reaper_limb", Quantity: 2}, {ItemID: "elder_wroshyr_timber", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "noxious_acklay_halberd", Quantity: 1}}, Experience: 435},
	{ID: "beskar-dragon-claws", Name: "Craft Runic Dragon-Talon Vibroclaws", SkillID: "crafting", RequiredLevel: 10, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "runic_banite_bar", Quantity: 1}, {ItemID: "ancient_krayt_dragon_scale", Quantity: 2}, {ItemID: "hypergate_relic_core", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "beskar_dragon_claws", Quantity: 1}}, Experience: 430},
	{ID: "westar-armadyl-blaster", Name: "Assemble Armadyl-Nova Dual Repeater", SkillID: "crafting", RequiredLevel: 10, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "hypergate_relic_core", Quantity: 1}, {ItemID: "krayt_wyrm_pearl", Quantity: 1}, {ItemID: "runic_banite_bar", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "westar_armadyl_blaster", Quantity: 1}}, Experience: 420},
	{ID: "twisted-wroshyr-longbow", Name: "Craft Twisted Wroshyr-Heart Longrifle", SkillID: "crafting", RequiredLevel: 10, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "elder_wroshyr_timber", Quantity: 2}, {ItemID: "sovereign_jad_fire_core", Quantity: 1}, {ItemID: "hypergate_relic_core", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "twisted_wroshyr_longbow", Quantity: 1}}, Experience: 450},
	{ID: "tumeken-ichor-scepter", Name: "Craft Shadow-Ichor Sovereign Staff", SkillID: "crafting", RequiredLevel: 10, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "ichor_spirit_sigil", Quantity: 1}, {ItemID: "sovereign_jad_fire_core", Quantity: 1}, {ItemID: "elder_wroshyr_timber", Quantity: 2}}, Outputs: []ItemQuantity{{ItemID: "tumeken_ichor_scepter", Quantity: 1}}, Experience: 445},
	{ID: "rakatan-stim-infuser", Name: "Synthesize Hyper-Gate Nanite Medpack", SkillID: "crafting", RequiredLevel: 9, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "bacta_gel", Quantity: 2}, {ItemID: "hypergate_relic_core", Quantity: 1}, {ItemID: "sovereign_torstol_leaf", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "rakatan_stim_infuser", Quantity: 1}}, Experience: 345},
	{ID: "saradomin-muja-banquet", Name: "Prepare Sovereign Muja-Brew Feast", SkillID: "cooking", RequiredLevel: 9, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "spirit_muja_fruit", Quantity: 2}, {ItemID: "sovereign_torstol_leaf", Quantity: 1}, {ItemID: "bacta_gel", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "saradomin_muja_banquet", Quantity: 1}}, Experience: 350},
	{ID: "overload-bacta-flask", Name: "Distill Supreme Overload Bacta-Flask", SkillID: "cooking", RequiredLevel: 10, StationID: "frontier-workbench", Ingredients: []ItemQuantity{{ItemID: "sovereign_torstol_leaf", Quantity: 2}, {ItemID: "spirit_muja_fruit", Quantity: 1}, {ItemID: "sovereign_jad_fire_core", Quantity: 1}}, Outputs: []ItemQuantity{{ItemID: "overload_bacta_flask", Quantity: 1}}, Experience: 440},
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
		if station.ID == id {
			return station, true
		}
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
	if err := p.CanAct(); err != nil {
		return CraftingResult{}, err
	}
	if p.Position.ZoneID != station.Position.ZoneID || math.Hypot(p.Position.X-station.Position.X, p.Position.Z-station.Position.Z) > station.InteractionRange {
		return CraftingResult{}, ErrCraftingOutOfRange
	}
	if recipe.StationID != station.ID {
		return CraftingResult{}, ErrUnknownCraftingStation
	}
	if recipe.RequiredLevel < 1 || recipe.RequiredLevel > MaxSkillLevel || recipe.Experience <= 0 {
		return CraftingResult{}, ErrUnknownCraftingRecipe
	}
	progress, skillExists := p.Skills[recipe.SkillID]
	if !skillExists || progress.Level < recipe.RequiredLevel {
		return CraftingResult{}, ErrCraftingSkillTooLow
	}

	consumed := make(map[string]int, len(recipe.Ingredients))
	produced := make(map[string]int, len(recipe.Outputs))
	for _, ingredient := range recipe.Ingredients {
		item, exists := FindItem(ingredient.ItemID)
		if !exists || ingredient.Quantity <= 0 || consumed[ingredient.ItemID] > item.MaxStack-ingredient.Quantity {
			return CraftingResult{}, ErrUnknownCraftingRecipe
		}
		consumed[ingredient.ItemID] += ingredient.Quantity
	}
	for _, output := range recipe.Outputs {
		item, exists := FindItem(output.ItemID)
		if !exists || output.Quantity <= 0 || output.Quantity > item.MaxStack {
			return CraftingResult{}, ErrUnknownCraftingRecipe
		}
		if produced[output.ItemID] > item.MaxStack-output.Quantity {
			return CraftingResult{}, ErrCraftingOutputFull
		}
		produced[output.ItemID] += output.Quantity
	}
	for itemID, quantity := range consumed {
		if p.Inventory[itemID] < quantity {
			return CraftingResult{}, ErrInsufficientIngredients
		}
	}

	// Simulate the complete net inventory change before mutating anything.
	// Aggregating duplicate entries also prevents malformed recipes from
	// bypassing stack limits or consuming the same stack more than once.
	nextInventory := make(map[string]int, len(p.Inventory)+len(produced))
	for itemID, quantity := range p.Inventory {
		nextInventory[itemID] = quantity
	}
	for itemID, quantity := range consumed {
		nextInventory[itemID] -= quantity
	}
	for itemID, quantity := range produced {
		item, _ := FindItem(itemID)
		if nextInventory[itemID] > item.MaxStack-quantity {
			return CraftingResult{}, ErrCraftingOutputFull
		}
		nextInventory[itemID] += quantity
	}
	for itemID, quantity := range nextInventory {
		if quantity == 0 {
			delete(nextInventory, itemID)
		}
	}

	leveled, err := p.AwardExperience(recipe.SkillID, recipe.Experience)
	if err != nil {
		return CraftingResult{}, err
	}
	p.Inventory = nextInventory
	return CraftingResult{
		RecipeID: recipe.ID, Outputs: append([]ItemQuantity(nil), recipe.Outputs...),
		SkillID: recipe.SkillID, Experience: recipe.Experience,
		Skill: p.Skills[recipe.SkillID], LeveledUp: leveled,
	}, nil
}
