package game

type ItemDefinition struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	MaxStack            int    `json:"maxStack"`
	BuyPrice            int64  `json:"buyPrice"`
	SellPrice           int64  `json:"sellPrice"`
	EquipmentSlot       string `json:"equipmentSlot,omitempty"`
	RequiredSkillID     string `json:"requiredSkillId,omitempty"`
	RequiredSkillLevel  int    `json:"requiredSkillLevel,omitempty"`
	AttackBonus         int    `json:"attackBonus,omitempty"`
	DefenseBonus        int    `json:"defenseBonus,omitempty"`
	ConstitutionBonus   int    `json:"constitutionBonus,omitempty"`
	StaminaBonus        int    `json:"staminaBonus,omitempty"`
	WillpowerBonus      int    `json:"willpowerBonus,omitempty"`
	Consumable          bool   `json:"consumable,omitempty"`
	HealAmount          int    `json:"healAmount,omitempty"`
	ActionRestoreAmount int    `json:"actionRestoreAmount,omitempty"`
	MindRestoreAmount   int    `json:"mindRestoreAmount,omitempty"`
	BuffHealthAmount    int    `json:"buffHealthAmount,omitempty"`
	BuffActionAmount    int    `json:"buffActionAmount,omitempty"`
	BuffMindAmount      int    `json:"buffMindAmount,omitempty"`
	BuffDurationSeconds int    `json:"buffDurationSeconds,omitempty"`
}

var itemDefinitions = []ItemDefinition{
	{ID: "softwood_log", Name: "Softwood Log", Description: "A freshly harvested log suitable for basic woodworking.", MaxStack: 100, BuyPrice: 5, SellPrice: 2},
	{ID: "rough_stone", Name: "Rough Stone", Description: "An uncut stone gathered from exposed deposits.", MaxStack: 100, BuyPrice: 3, SellPrice: 1},
	{ID: "wooden_plank", Name: "Wooden Plank", Description: "A simple processed timber component for basic construction.", MaxStack: 100, BuyPrice: 12, SellPrice: 5},
	{ID: "training_sword", Name: "Training Sword", Description: "A basic practice blade for new adventurers.", MaxStack: 1, BuyPrice: 50, SellPrice: 20, EquipmentSlot: "main_hand", RequiredSkillID: "blades", RequiredSkillLevel: 1, AttackBonus: 2},
	{ID: "leather_cap", Name: "Leather Cap", Description: "A simple leather cap offering light protection.", MaxStack: 1, BuyPrice: 30, SellPrice: 12, EquipmentSlot: "head", RequiredSkillID: "defense", RequiredSkillLevel: 1, DefenseBonus: 1, ConstitutionBonus: 1},
	{ID: "leather_tunic", Name: "Leather Tunic", Description: "A fitted leather tunic for frontier travel.", MaxStack: 1, BuyPrice: 60, SellPrice: 25, EquipmentSlot: "body", RequiredSkillID: "defense", RequiredSkillLevel: 1, DefenseBonus: 2, ConstitutionBonus: 2},
	{ID: "leather_boots", Name: "Leather Boots", Description: "Sturdy boots for long journeys.", MaxStack: 1, BuyPrice: 25, SellPrice: 10, EquipmentSlot: "feet", RequiredSkillID: "defense", RequiredSkillLevel: 1, DefenseBonus: 1, StaminaBonus: 1},
	{ID: "copper_ore", Name: "Copper Ore", Description: "A rough copper-bearing mineral from frontier deposits.", MaxStack: 100, BuyPrice: 3, SellPrice: 1},
	{ID: "copper_bar", Name: "Copper Bar", Description: "A refined copper ingot ready for smithing.", MaxStack: 100, BuyPrice: 10, SellPrice: 4},
	{ID: "rat_pelt", Name: "Rat Pelt", Description: "A small hide recovered from a frontier rat.", MaxStack: 100, BuyPrice: 6, SellPrice: 2},
	{ID: "rat_tooth", Name: "Rat Tooth", Description: "A sharp tooth salvaged from a frontier rat.", MaxStack: 100, BuyPrice: 0, SellPrice: 1},
	{ID: "rat_tooth_blade", Name: "Rat-Tooth Blade", Description: "A forged frontier blade reinforced with sharpened rat tooth.", MaxStack: 1, BuyPrice: 0, SellPrice: 35, EquipmentSlot: "main_hand", RequiredSkillID: "blades", RequiredSkillLevel: 2, AttackBonus: 3},
	{ID: "boar_tusk_spear", Name: "Boar-Tusk Spear", Description: "A balanced polearm with a hardened boar tusk head, forged for frontier hunters.", MaxStack: 1, BuyPrice: 0, SellPrice: 55, EquipmentSlot: "main_hand", RequiredSkillID: "polearms", RequiredSkillLevel: 2, AttackBonus: 4},
	{ID: "boar_hide", Name: "Boar Hide", Description: "A tough hide taken from a wild boar.", MaxStack: 100, BuyPrice: 15, SellPrice: 5},
	{ID: "boar_tusk", Name: "Boar Tusk", Description: "A heavy tusk recovered from a wild boar.", MaxStack: 100, BuyPrice: 0, SellPrice: 4},
	{ID: "minor_healing_potion", Name: "Minor Healing Potion", Description: "Restores up to 25 health when used.", MaxStack: 10, BuyPrice: 20, SellPrice: 5, Consumable: true, HealAmount: 25},
	{ID: "wild_herb", Name: "Wild Herb", Description: "A hardy medicinal plant gathered from the frontier.", MaxStack: 100, BuyPrice: 4, SellPrice: 2},
	{ID: "herbal_bundle", Name: "Herbal Bundle", Description: "A carefully packed bundle of medicinal plants.", MaxStack: 100, BuyPrice: 18, SellPrice: 8},
	{ID: "herbal_tea", Name: "Herbal Tea", Description: "A warm restorative drink brewed from frontier herbs. Restores up to 15 health.", MaxStack: 10, BuyPrice: 15, SellPrice: 4, Consumable: true, HealAmount: 15, MindRestoreAmount: 10},
	{ID: "stoneback_scale", Name: "Stoneback Scale", Description: "A dense mineralized scale shed by the Stoneback Colossus, useful in advanced armor crafting.", MaxStack: 100, BuyPrice: 0, SellPrice: 8},
	{ID: "stoneback_plate", Name: "Stoneback Plate", Description: "A broad armored plate prized by frontier armorers for durable protective equipment.", MaxStack: 100, BuyPrice: 0, SellPrice: 18},
	{ID: "stoneback_core", Name: "Stoneback Core", Description: "A rare hardened core recovered from the Stoneback Colossus and suited to specialized high-tier crafting.", MaxStack: 10, BuyPrice: 0, SellPrice: 75},
	{ID: "frontier_grain", Name: "Frontier Grain", Description: "Hardy cultivated grain harvested from frontier crop plots.", MaxStack: 100, BuyPrice: 4, SellPrice: 2},
	{ID: "frontier_ration", Name: "Frontier Ration", Description: "A hearty baked field ration prepared from frontier grain and herbs. Restores up to 20 health and 15 action.", MaxStack: 10, BuyPrice: 18, SellPrice: 6, Consumable: true, HealAmount: 20, ActionRestoreAmount: 15},
	{ID: "salvaged_scrap", Name: "Salvaged Scrap", Description: "Recovered alloy fittings and wiring salvaged from frontier wreckage.", MaxStack: 100, BuyPrice: 6, SellPrice: 2},
	{ID: "salvaged_power_cell", Name: "Salvaged Power Cell", Description: "A stabilized energy cell assembled from salvaged frontier scrap and refined copper.", MaxStack: 100, BuyPrice: 22, SellPrice: 9},
	{ID: "frontier_pistol", Name: "Frontier Pistol", Description: "A compact sidearm crafted from salvaged alloy and copper, suited to close ranged skirmishes.", MaxStack: 1, BuyPrice: 0, SellPrice: 40, EquipmentSlot: "main_hand", RequiredSkillID: "pistols", RequiredSkillLevel: 1, AttackBonus: 2},
	{ID: "frontier_rifle", Name: "Frontier Carbine", Description: "A long-barreled frontier rifle built for steady ranged marksmanship.", MaxStack: 1, BuyPrice: 0, SellPrice: 60, EquipmentSlot: "main_hand", RequiredSkillID: "rifles", RequiredSkillLevel: 2, AttackBonus: 3},
	{ID: "resonance_focus", Name: "Resonance Focus Staff", Description: "A channeler's focus staff tuned with a salvaged power cell and wild herbs.", MaxStack: 1, BuyPrice: 0, SellPrice: 45, EquipmentSlot: "main_hand", RequiredSkillID: "magic", RequiredSkillLevel: 1, AttackBonus: 2, WillpowerBonus: 3},
	// 8 Mining Ores, Crystals, and Refined Alloys
	{ID: "iron_ore", Name: "Iron Ore", Description: "Dense iron-bearing ore mined from deep frontier seams.", MaxStack: 100, BuyPrice: 8, SellPrice: 3},
	{ID: "iron_bar", Name: "Iron Bar", Description: "A strong smelted iron ingot used for heavy weapons and armor.", MaxStack: 100, BuyPrice: 20, SellPrice: 8},
	{ID: "cobalt_ore", Name: "Cobalt Ore", Description: "A blue-tinted magnetic mineral mined from cobalt ridges.", MaxStack: 100, BuyPrice: 11, SellPrice: 4},
	{ID: "cobalt_bar", Name: "Cobalt Bar", Description: "A refined cobalt ingot prized for conductive weaponry and focus housings.", MaxStack: 100, BuyPrice: 26, SellPrice: 10},
	{ID: "titanium_ore", Name: "Titanium Ore", Description: "High-grade metallic ore prized for lightweight, high-strength alloy forging.", MaxStack: 100, BuyPrice: 16, SellPrice: 6},
	{ID: "titanium_bar", Name: "Titanium Bar", Description: "A refined titanium alloy ingot for masterwork frontier weaponry and plating.", MaxStack: 100, BuyPrice: 38, SellPrice: 15},
	{ID: "aether_crystal", Name: "Aether Crystal", Description: "A luminescent crystalline mineral that resonates with bio-electric and mental energy.", MaxStack: 100, BuyPrice: 24, SellPrice: 10},
	{ID: "durasteel_ore", Name: "Durasteel Ore", Description: "Carbon-dense metallic ore extracted from deep tectonic fissures.", MaxStack: 100, BuyPrice: 28, SellPrice: 11},
	{ID: "durasteel_bar", Name: "Durasteel Bar", Description: "An ultra-dense durasteel ingot forged for heavy frontier arms and armor.", MaxStack: 100, BuyPrice: 62, SellPrice: 25},
	{ID: "luminite_gem", Name: "Luminite Gem", Description: "A radiant prismatic crystal formed in high-energy mineral seams.", MaxStack: 100, BuyPrice: 42, SellPrice: 16},
	{ID: "mythril_ore", Name: "Mythril Ore", Description: "Rare silvery star-metal ore resonating with harmonic energy.", MaxStack: 100, BuyPrice: 52, SellPrice: 20},
	{ID: "mythril_bar", Name: "Mythril Bar", Description: "A gleaming mythril ingot combining featherweight balance with unmatched strength.", MaxStack: 100, BuyPrice: 115, SellPrice: 45},
	// Medical Packs & Attribute Buff Enhancers (Health, Action, Mind, Constitution, Stamina, Willpower)
	{ID: "bacta_gel", Name: "Bio-Gel Compound", Description: "A sterile regenerative compound distilled from wild herbs and mineral salts.", MaxStack: 100, BuyPrice: 16, SellPrice: 6},
	{ID: "field_medpack", Name: "Field Medical Pack", Description: "A compact field medpack that rapidly restores 45 Health.", MaxStack: 10, BuyPrice: 35, SellPrice: 12, Consumable: true, HealAmount: 45},
	{ID: "advanced_medpack", Name: "Advanced Trauma Medpack", Description: "A high-potency surgical medpack that restores 80 Health and 25 Action.", MaxStack: 10, BuyPrice: 65, SellPrice: 24, Consumable: true, HealAmount: 80, ActionRestoreAmount: 25},
	{ID: "surgical_trauma_kit", Name: "Surgical Trauma Kit", Description: "A masterwork emergency medpack that restores 120 Health, 40 Action, and 40 Mind.", MaxStack: 10, BuyPrice: 110, SellPrice: 42, Consumable: true, HealAmount: 120, ActionRestoreAmount: 40, MindRestoreAmount: 40},
	{ID: "vitality_stim_pack", Name: "Constitution Vitality Stim", Description: "An attribute enhancer that boosts Max Health by +35 and Constitution by +5 (increasing Health regeneration) for 10 minutes.", MaxStack: 10, BuyPrice: 55, SellPrice: 20, Consumable: true, HealAmount: 35, BuffHealthAmount: 35, ConstitutionBonus: 5, BuffDurationSeconds: 600},
	{ID: "adrenal_action_booster", Name: "Stamina Action Booster", Description: "A kinetic metabolic booster that increases Max Action by +35 and Stamina by +5 (increasing Action regeneration) for 10 minutes.", MaxStack: 10, BuyPrice: 55, SellPrice: 20, Consumable: true, ActionRestoreAmount: 35, BuffActionAmount: 35, StaminaBonus: 5, BuffDurationSeconds: 600},
	{ID: "neural_mind_tonic", Name: "Willpower Clarity Tonic", Description: "An aether-infused neural tonic that boosts Max Mind by +35 and Willpower by +5 (increasing Mind regeneration) for 10 minutes.", MaxStack: 10, BuyPrice: 55, SellPrice: 20, Consumable: true, MindRestoreAmount: 35, BuffMindAmount: 35, WillpowerBonus: 5, BuffDurationSeconds: 600},
	{ID: "overdrive_omni_stim", Name: "Tri-Stat Overdrive Stim", Description: "A supreme synthetic enhancer that boosts Max Health/Action/Mind by +50 and Constitution/Stamina/Willpower by +8 for 15 minutes.", MaxStack: 10, BuyPrice: 140, SellPrice: 55, Consumable: true, HealAmount: 50, ActionRestoreAmount: 50, MindRestoreAmount: 50, BuffHealthAmount: 50, BuffActionAmount: 50, BuffMindAmount: 50, ConstitutionBonus: 8, StaminaBonus: 8, WillpowerBonus: 8, BuffDurationSeconds: 900},
	// Monster Drops & High-Tier Crafted Weapons/Armor
	{ID: "stalker_chitin", Name: "Stalker Chitin", Description: "A hardened carapace plate harvested from a Canyon Stalker.", MaxStack: 100, BuyPrice: 0, SellPrice: 6},
	{ID: "stalker_venom_sac", Name: "Stalker Venom Sac", Description: "A potent bio-chemical gland extracted from a Canyon Stalker, prized for advanced medpacks and stims.", MaxStack: 100, BuyPrice: 0, SellPrice: 12},
	{ID: "hound_fang", Name: "Ash Hound Fang", Description: "A serrated crystalline fang recovered from an Ash Hound.", MaxStack: 100, BuyPrice: 0, SellPrice: 9},
	{ID: "sentinel_core", Name: "Sentinel Alloy Core", Description: "An intact power regulator salvaged from a Ruin Sentinel.", MaxStack: 100, BuyPrice: 0, SellPrice: 18},
	{ID: "drake_scale", Name: "Mire Drake Scale", Description: "An iridescent amphibious scale shed by a Mire Drake.", MaxStack: 100, BuyPrice: 0, SellPrice: 11},
	{ID: "drake_bile", Name: "Mire Drake Acid Gland", Description: "A volatile enzymatic gland harvested from a Mire Drake.", MaxStack: 100, BuyPrice: 0, SellPrice: 16},
	{ID: "prowler_claw", Name: "Frost Prowler Claw", Description: "A razor-sharp cryo-hardened claw taken from a Frost Prowler.", MaxStack: 100, BuyPrice: 0, SellPrice: 14},
	{ID: "prowler_pelt", Name: "Frost Prowler Pelt", Description: "Dense insulated fur harvested from a Frost Prowler.", MaxStack: 100, BuyPrice: 0, SellPrice: 10},
	{ID: "wraith_essence", Name: "Aether Wraith Essence", Description: "Condensed phase-matter left behind by an Aether Wraith.", MaxStack: 100, BuyPrice: 0, SellPrice: 22},
	{ID: "iron_greatsword", Name: "Iron Broadsword", Description: "A heavy forged iron blade with superior reach and cutting power.", MaxStack: 1, BuyPrice: 0, SellPrice: 70, EquipmentSlot: "main_hand", RequiredSkillID: "blades", RequiredSkillLevel: 3, AttackBonus: 5, ConstitutionBonus: 2},
	{ID: "cobalt_repeater", Name: "Cobalt Blaster Pistol", Description: "A rapid-cycling sidearm built with a cobalt magnetic barrel.", MaxStack: 1, BuyPrice: 0, SellPrice: 75, EquipmentSlot: "main_hand", RequiredSkillID: "pistols", RequiredSkillLevel: 3, AttackBonus: 5, StaminaBonus: 2},
	{ID: "titanium_carbine", Name: "Titanium Pulse Rifle", Description: "A precision titanium-framed rifle powered by a sentinel core.", MaxStack: 1, BuyPrice: 0, SellPrice: 110, EquipmentSlot: "main_hand", RequiredSkillID: "rifles", RequiredSkillLevel: 4, AttackBonus: 6, StaminaBonus: 3},
	{ID: "durasteel_halberd", Name: "Durasteel Halberd", Description: "An armor-piercing polearm forged from durasteel and frost prowler claws.", MaxStack: 1, BuyPrice: 0, SellPrice: 125, EquipmentSlot: "main_hand", RequiredSkillID: "polearms", RequiredSkillLevel: 4, AttackBonus: 7, ConstitutionBonus: 3},
	{ID: "aether_scepter", Name: "Luminite Aether Staff", Description: "A resonant focus staff crowned with a luminite gem and wraith essence.", MaxStack: 1, BuyPrice: 0, SellPrice: 130, EquipmentSlot: "main_hand", RequiredSkillID: "magic", RequiredSkillLevel: 4, AttackBonus: 6, WillpowerBonus: 6},
	{ID: "mythril_saber", Name: "Mythril Edge", Description: "A masterwork star-metal blade humming with harmonic energy.", MaxStack: 1, BuyPrice: 0, SellPrice: 160, EquipmentSlot: "main_hand", RequiredSkillID: "blades", RequiredSkillLevel: 5, AttackBonus: 8, ConstitutionBonus: 3, StaminaBonus: 3},
	{ID: "chitin_cuirass", Name: "Stalker Chitin Cuirass", Description: "Reinforced chest armor laminated with Canyon Stalker chitin and iron.", MaxStack: 1, BuyPrice: 0, SellPrice: 85, EquipmentSlot: "body", RequiredSkillID: "defense", RequiredSkillLevel: 3, DefenseBonus: 4, ConstitutionBonus: 4},
	{ID: "durasteel_helm", Name: "Durasteel Visor Helm", Description: "Heavy protective headgear forged from durasteel and mire drake scales.", MaxStack: 1, BuyPrice: 0, SellPrice: 95, EquipmentSlot: "head", RequiredSkillID: "defense", RequiredSkillLevel: 4, DefenseBonus: 3, ConstitutionBonus: 3, WillpowerBonus: 2},
	{ID: "prowler_striders", Name: "Frost Prowler Greaves", Description: "Insulated titanium-lined boots that boost mobility and stamina.", MaxStack: 1, BuyPrice: 0, SellPrice: 90, EquipmentSlot: "feet", RequiredSkillID: "defense", RequiredSkillLevel: 3, DefenseBonus: 3, StaminaBonus: 4},
	// Tier 6-8 Mining Ores, Crystals & Refined Alloys (+8 items)
	{ID: "orichalcum_ore", Name: "Orichalcum Ore", Description: "Resonant golden-red metallic ore mined from deep volcanic vents.", MaxStack: 100, BuyPrice: 68, SellPrice: 26},
	{ID: "orichalcum_bar", Name: "Orichalcum Bar", Description: "A gleaming orichalcum alloy bar prized for high-energy conductive weaponry.", MaxStack: 100, BuyPrice: 150, SellPrice: 58},
	{ID: "obsidian_shard", Name: "Volcanic Obsidian Shard", Description: "A razor-edged volcanic glass shard with natural thermal insulation.", MaxStack: 100, BuyPrice: 58, SellPrice: 22},
	{ID: "neutronium_ore", Name: "Neutronium Ore", Description: "Super-dense gravitic ore extracted from collapsed crater seams.", MaxStack: 100, BuyPrice: 88, SellPrice: 34},
	{ID: "neutronium_bar", Name: "Neutronium Bar", Description: "An ultra-heavy neutronium ingot capable of withstanding extreme kinetic impact.", MaxStack: 100, BuyPrice: 195, SellPrice: 76},
	{ID: "void_quartz", Name: "Void Quartz Crystal", Description: "A dark violet phase-crystal that amplifies mental focus and energy regeneration.", MaxStack: 100, BuyPrice: 78, SellPrice: 30},
	{ID: "starsteel_ore", Name: "Star-Metal Ore", Description: "Meteoritic alloy oreinfused with cosmic radiation from fallen sky-stones.", MaxStack: 100, BuyPrice: 115, SellPrice: 45},
	{ID: "starsteel_bar", Name: "Star-Metal Bar", Description: "A pinnacle forged star-metal ingot of unmatched durability and edge retention.", MaxStack: 100, BuyPrice: 250, SellPrice: 98},
	// Additional Monster Materials (+5 items)
	{ID: "scavenger_servo", Name: "Scavenger Servo Unit", Description: "A high-torque mechanical actuator salvaged from a Feral Scavenger Drone.", MaxStack: 100, BuyPrice: 0, SellPrice: 15},
	{ID: "wyrm_carapace", Name: "Obsidian Wyrm Carapace", Description: "A heat-tempered crystalline shell plate shed by an Obsidian Wyrm.", MaxStack: 100, BuyPrice: 0, SellPrice: 24},
	{ID: "wyrm_heartstone", Name: "Wyrm Molten Core", Description: "An incandescent mineral organ recovered from an Obsidian Wyrm.", MaxStack: 100, BuyPrice: 0, SellPrice: 38},
	{ID: "behemoth_horn", Name: "Thornback Behemoth Horn", Description: "A massive iron-hard horn taken from a Thornback Behemoth.", MaxStack: 100, BuyPrice: 0, SellPrice: 28},
	{ID: "reaver_plasma_coil", Name: "Void Reaver Plasma Coil", Description: "An overcharged phase-coil salvaged from a Void Reaver.", MaxStack: 100, BuyPrice: 0, SellPrice: 32},
	// Additional Medical Packs, Stims & Masterwork Gear (+6 items)
	{ID: "nanite_repair_injector", Name: "Nanite Trauma Injector", Description: "An emergency medical injector that restores 150 Health and 60 Action.", MaxStack: 10, BuyPrice: 165, SellPrice: 64, Consumable: true, HealAmount: 150, ActionRestoreAmount: 60},
	{ID: "titan_constitution_elixir", Name: "Titan Constitution Elixir", Description: "A high-potency bio-elixir that boosts Max Health by +75 and Constitution by +10 for 15 minutes.", MaxStack: 10, BuyPrice: 185, SellPrice: 70, Consumable: true, HealAmount: 75, BuffHealthAmount: 75, ConstitutionBonus: 10, BuffDurationSeconds: 900},
	{ID: "orichalcum_glaive", Name: "Orichalcum Sonic Glaive", Description: "A resonant long-reach polearm forged from orichalcum and behemoth horn.", MaxStack: 1, BuyPrice: 0, SellPrice: 175, EquipmentSlot: "main_hand", RequiredSkillID: "polearms", RequiredSkillLevel: 5, AttackBonus: 9, ConstitutionBonus: 4, StaminaBonus: 3},
	{ID: "neutronium_platebody", Name: "Neutronium Vanguard Cuirass", Description: "Heavy frontier battle armor laminated with neutronium and obsidian wyrm carapace.", MaxStack: 1, BuyPrice: 0, SellPrice: 210, EquipmentSlot: "body", RequiredSkillID: "defense", RequiredSkillLevel: 5, DefenseBonus: 7, ConstitutionBonus: 6, StaminaBonus: 2},
	{ID: "void_psionic_crown", Name: "Void Quartz Circlet", Description: "A neural-amplifier helm inlaid with void quartz and aether crystal.", MaxStack: 1, BuyPrice: 0, SellPrice: 180, EquipmentSlot: "head", RequiredSkillID: "defense", RequiredSkillLevel: 5, DefenseBonus: 4, ConstitutionBonus: 3, WillpowerBonus: 7},
	{ID: "starsteel_longblade", Name: "Star-Metal Vanguard Blade", Description: "A masterwork meteoritic longblade balanced with a void reaver plasma coil.", MaxStack: 1, BuyPrice: 0, SellPrice: 240, EquipmentSlot: "main_hand", RequiredSkillID: "blades", RequiredSkillLevel: 6, AttackBonus: 10, ConstitutionBonus: 4, StaminaBonus: 4},
	// Tier 7-10 Mining Ores, Crystals & Refined Alloys (+8 items)
	{ID: "electrum_ore", Name: "Electrum Ore", Description: "A natural gold-silver conductive ore mined from high-altitude frontier veins.", MaxStack: 100, BuyPrice: 95, SellPrice: 38},
	{ID: "electrum_bar", Name: "Electrum Bar", Description: "A refined electrum alloy bar prized for precision circuit-forged weaponry.", MaxStack: 100, BuyPrice: 210, SellPrice: 82},
	{ID: "chronite_ore", Name: "Chronite Ore", Description: "Time-crystallized metallic ore extracted from deep temporal rifts.", MaxStack: 100, BuyPrice: 135, SellPrice: 52},
	{ID: "chronite_bar", Name: "Chronite Bar", Description: "A stabilized chronite ingot that hums with kinetic acceleration.", MaxStack: 100, BuyPrice: 290, SellPrice: 115},
	{ID: "solarite_crystal", Name: "Solarite Crystal", Description: "A blazing amber crystal saturated with concentrated thermal energy.", MaxStack: 100, BuyPrice: 120, SellPrice: 48},
	{ID: "abyssal_ore", Name: "Abyssal Adamant Ore", Description: "Ultra-dense dark mineral recovered from deep subterranean chasms.", MaxStack: 100, BuyPrice: 165, SellPrice: 65},
	{ID: "abyssal_bar", Name: "Abyssal Adamant Bar", Description: "An virtually indestructible abyssal alloy bar for endgame frontier armor.", MaxStack: 100, BuyPrice: 350, SellPrice: 140},
	{ID: "nova_prism", Name: "Nova Core Prism", Description: "A stellar-grade refractive crystal that channels immense mental and bio-energy.", MaxStack: 100, BuyPrice: 180, SellPrice: 72},
	// Additional Monster Materials (+5 items)
	{ID: "scorpion_stinger", Name: "Dune Scorpion Venom Stinger", Description: "A barbed metallic stinger harvested from an Armored Dune Scorpion.", MaxStack: 100, BuyPrice: 0, SellPrice: 26},
	{ID: "mantis_scythe", Name: "Razor Mantis Foreblade", Description: "A monomolecular chitin scythe taken from a Crystalline Mantis.", MaxStack: 100, BuyPrice: 0, SellPrice: 30},
	{ID: "colossus_crab_shell", Name: "Magma Carapace Shell", Description: "A heat-shielded mineral shell pried from a Magma Crusher.", MaxStack: 100, BuyPrice: 0, SellPrice: 35},
	{ID: "specter_shroud", Name: "Rift Specter Shroud", Description: "A gossamer phase-cloth remnant condensed from a Rift Specter.", MaxStack: 100, BuyPrice: 0, SellPrice: 40},
	{ID: "leviathan_fang", Name: "Apex Devourer Fang", Description: "A colossal serrated tooth recovered from an Apex Rift Devourer.", MaxStack: 100, BuyPrice: 0, SellPrice: 55},
	// Additional Medical Packs, Stims & Masterwork Weapons/Armor (+5 items)
	{ID: "phoenix_resuscitation_pack", Name: "Phoenix Resuscitation Medpack", Description: "A pinnacle emergency medical pack that restores 180 Health, 75 Action, and 75 Mind.", MaxStack: 10, BuyPrice: 240, SellPrice: 95, Consumable: true, HealAmount: 180, ActionRestoreAmount: 75, MindRestoreAmount: 75},
	{ID: "celestial_harmony_draught", Name: "Celestial Tri-Stat Draught", Description: "Boosts Max Health/Action/Mind by +90 and Constitution/Stamina/Willpower by +12 for 20 minutes.", MaxStack: 10, BuyPrice: 280, SellPrice: 110, Consumable: true, HealAmount: 90, ActionRestoreAmount: 90, MindRestoreAmount: 90, BuffHealthAmount: 90, BuffActionAmount: 90, BuffMindAmount: 90, ConstitutionBonus: 12, StaminaBonus: 12, WillpowerBonus: 12, BuffDurationSeconds: 1200},
	{ID: "electrum_knuckles", Name: "Electrum Shock Cestus", Description: "Conductive hand-to-hand combat gauntlets that amplify unarmed strikes.", MaxStack: 1, BuyPrice: 0, SellPrice: 195, EquipmentSlot: "main_hand", RequiredSkillID: "unarmed", RequiredSkillLevel: 5, AttackBonus: 9, ConstitutionBonus: 4, StaminaBonus: 5},
	{ID: "chronite_sniper_rifle", Name: "Chronite Phase Longrifle", Description: "A hyper-velocity longrifle built with a chronite barrel and solarite crystal lens.", MaxStack: 1, BuyPrice: 0, SellPrice: 265, EquipmentSlot: "main_hand", RequiredSkillID: "rifles", RequiredSkillLevel: 6, AttackBonus: 11, StaminaBonus: 6, WillpowerBonus: 3},
	{ID: "abyssal_bulwark_armor", Name: "Abyssal Adamant Bulwark", Description: "Masterwork heavy chest armor forged from abyssal adamant and magma carapace.", MaxStack: 1, BuyPrice: 0, SellPrice: 295, EquipmentSlot: "body", RequiredSkillID: "defense", RequiredSkillLevel: 7, DefenseBonus: 10, ConstitutionBonus: 9, StaminaBonus: 4},
	// Multi-Skill Gathering Resources: Woodcutting, Foraging, Farming & Scavenging Tiers 2-4 (+8 items)
	{ID: "ironwood_log", Name: "Ironwood Log", Description: "Dense timber cut from mature frontier ironwood trees.", MaxStack: 100, BuyPrice: 14, SellPrice: 5},
	{ID: "petrified_starwood", Name: "Petrified Starwood Log", Description: "Mineralized ancient timber prized for high-impact weapon hafts and staves.", MaxStack: 100, BuyPrice: 36, SellPrice: 14},
	{ID: "glowcap_fungus", Name: "Bioluminescent Glowcap", Description: "A glowing cavern mushroom rich in restorative bio-alkaloids.", MaxStack: 100, BuyPrice: 12, SellPrice: 5},
	{ID: "sunblossom_petal", Name: "Solaris Sunblossom", Description: "A radiant desert bloom that concentrates thermal and mental vitality.", MaxStack: 100, BuyPrice: 32, SellPrice: 12},
	{ID: "tuber_root", Name: "Highland Heart-Tuber", Description: "A nutrient-dense tuber cultivated in irrigated frontier terraces.", MaxStack: 100, BuyPrice: 12, SellPrice: 5},
	{ID: "starberry_cluster", Name: "Star-Orchard Berry Cluster", Description: "High-energy fruit harvested from hydroponic starberry arbors.", MaxStack: 100, BuyPrice: 34, SellPrice: 13},
	{ID: "circuit_board", Name: "Intact Logic Board", Description: "A recoverable targeting and telemetry circuit board pulled from frontier relays.", MaxStack: 100, BuyPrice: 18, SellPrice: 7},
	{ID: "quantum_capacitor", Name: "Quantum Flux Capacitor", Description: "A high-density energy discharge module salvaged from crashed orbital pods.", MaxStack: 100, BuyPrice: 48, SellPrice: 18},
	// Parent Combat, Co-Skill & Specialization Milestone Equipment (+12 items)
	{ID: "brawler_handwraps", Name: "Frontier Brawler Handwraps", Description: "Reinforced leather and boar-hide striking wraps for close-quarters brawlers.", MaxStack: 1, BuyPrice: 35, SellPrice: 14, EquipmentSlot: "main_hand", RequiredSkillID: "unarmed", RequiredSkillLevel: 1, AttackBonus: 2, StaminaBonus: 1},
	{ID: "spiked_knuckledusters", Name: "Iron Knuckledusters", Description: "Heavy iron-studded cestus gauntlets weighted with stalker chitin.", MaxStack: 1, BuyPrice: 0, SellPrice: 68, EquipmentSlot: "main_hand", RequiredSkillID: "unarmed", RequiredSkillLevel: 3, AttackBonus: 5, ConstitutionBonus: 2, StaminaBonus: 2},
	{ID: "crystal_channeled_rod", Name: "Aether Channeling Rod", Description: "An ironwood focus rod tipped with an aether crystal prism.", MaxStack: 1, BuyPrice: 0, SellPrice: 62, EquipmentSlot: "main_hand", RequiredSkillID: "magic", RequiredSkillLevel: 2, AttackBonus: 4, WillpowerBonus: 4},
	{ID: "void_harmonic_staff", Name: "Void Harmonic Grand Staff", Description: "A petrified starwood staff resonating with void quartz and quantum flux.", MaxStack: 1, BuyPrice: 0, SellPrice: 235, EquipmentSlot: "main_hand", RequiredSkillID: "magic", RequiredSkillLevel: 6, AttackBonus: 10, ConstitutionBonus: 3, WillpowerBonus: 8},
	{ID: "pulse_needler_pistol", Name: "Quantum Pulse Sidearm", Description: "A high-cadence sidearm overcharged by a quantum flux capacitor.", MaxStack: 1, BuyPrice: 0, SellPrice: 185, EquipmentSlot: "main_hand", RequiredSkillID: "pistols", RequiredSkillLevel: 5, AttackBonus: 8, StaminaBonus: 4, WillpowerBonus: 2},
	{ID: "vanguard_cleaver", Name: "Melee Vanguard Cleaver", Description: "A balanced close-combat cleaver unlocking for seasoned Melee fighters.", MaxStack: 1, BuyPrice: 0, SellPrice: 52, EquipmentSlot: "main_hand", RequiredSkillID: "melee", RequiredSkillLevel: 2, AttackBonus: 4, ConstitutionBonus: 2},
	{ID: "warlord_shock_maul", Name: "Melee Warlord Shock Maul", Description: "A durasteel-headed shock maul wielded by veteran Melee combatants.", MaxStack: 1, BuyPrice: 0, SellPrice: 135, EquipmentSlot: "main_hand", RequiredSkillID: "melee", RequiredSkillLevel: 4, AttackBonus: 7, ConstitutionBonus: 4, StaminaBonus: 2},
	{ID: "skirmisher_repeater_bow", Name: "Ranged Skirmisher Bowcaster", Description: "A tensioned ironwood and circuit-assisted bowcaster for Ranged specialists.", MaxStack: 1, BuyPrice: 0, SellPrice: 54, EquipmentSlot: "main_hand", RequiredSkillID: "ranged", RequiredSkillLevel: 2, AttackBonus: 4, StaminaBonus: 2},
	{ID: "deadeye_rail_cannon", Name: "Ranged Deadeye Rail-Arbalest", Description: "A magnetic rail-arbalest unlocking at Ranged Level 4.", MaxStack: 1, BuyPrice: 0, SellPrice: 140, EquipmentSlot: "main_hand", RequiredSkillID: "ranged", RequiredSkillLevel: 4, AttackBonus: 7, StaminaBonus: 4, WillpowerBonus: 2},
	{ID: "colossus_breaker_greataxe", Name: "Strength Breaker Greataxe", Description: "A massive two-handed iron and boar-tusk greataxe requiring brute Strength to swing.", MaxStack: 1, BuyPrice: 0, SellPrice: 64, EquipmentSlot: "main_hand", RequiredSkillID: "strength", RequiredSkillLevel: 2, AttackBonus: 5, ConstitutionBonus: 3},
	{ID: "titan_harness", Name: "Titan Strength Harness", Description: "Weighted durasteel and behemoth-horn chest harness that channels raw physical power.", MaxStack: 1, BuyPrice: 0, SellPrice: 150, EquipmentSlot: "body", RequiredSkillID: "strength", RequiredSkillLevel: 4, AttackBonus: 3, DefenseBonus: 5, ConstitutionBonus: 5, StaminaBonus: 3},
	{ID: "studded_frontier_vest", Name: "Studded Boar-Hide Vest", Description: "Copper-riveted boar hide armor offering dependable early Defense protection.", MaxStack: 1, BuyPrice: 0, SellPrice: 48, EquipmentSlot: "body", RequiredSkillID: "defense", RequiredSkillLevel: 2, DefenseBonus: 3, ConstitutionBonus: 3},
	// Culinary Progression Meals (+2 items)
	{ID: "hearty_tuber_stew", Name: "Hearty Tuber & Glowcap Stew", Description: "A simmering frontier stew that restores 65 Health and 35 Action.", MaxStack: 10, BuyPrice: 45, SellPrice: 16, Consumable: true, HealAmount: 65, ActionRestoreAmount: 35},
	{ID: "starberry_ambrosia", Name: "Solaris Starberry Banquet", Description: "A revitalizing gourmet feast that boosts Max Health/Action/Mind by +25 and restores 95 Health, 50 Action, and 50 Mind.", MaxStack: 10, BuyPrice: 95, SellPrice: 36, Consumable: true, HealAmount: 95, ActionRestoreAmount: 50, MindRestoreAmount: 50, BuffHealthAmount: 25, BuffActionAmount: 25, BuffMindAmount: 25, ConstitutionBonus: 3, StaminaBonus: 3, WillpowerBonus: 3, BuffDurationSeconds: 600},
	// Tier 6-8 Woodcutting, Foraging, Farming & Scavenging Resources (+8 items)
	{ID: "aether_bough_timber", Name: "Aether-Bough Timber", Description: "Resonant silver-grained timber harvested from ancient aether-bough trees.", MaxStack: 100, BuyPrice: 85, SellPrice: 34},
	{ID: "voidbark_heartwood", Name: "Abyssal Voidbark Heartwood", Description: "Ultra-dense abyssal heartwood impervious to warp and plasma heat.", MaxStack: 100, BuyPrice: 145, SellPrice: 58},
	{ID: "cryo_lotus_petal", Name: "Cryo-Lotus Petal", Description: "A frost-rimed alpine blossom prized for thermal stabilization and high-grade tonics.", MaxStack: 100, BuyPrice: 82, SellPrice: 32},
	{ID: "astral_mycelium", Name: "Astral Spore Mycelium", Description: "Luminous starlight-fed mycelium filament that amplifies psionic and cellular recovery.", MaxStack: 100, BuyPrice: 140, SellPrice: 56},
	{ID: "auric_grain_sheaf", Name: "Auric Grain Sheaf", Description: "Golden sun-ripened grain cultivated in high-altitude frontier terraces.", MaxStack: 100, BuyPrice: 80, SellPrice: 32},
	{ID: "celestial_nectar_fruit", Name: "Celestial Nectar Fruit", Description: "A radiant orchard fruit overflowing with bio-restorative nectar.", MaxStack: 100, BuyPrice: 142, SellPrice: 56},
	{ID: "titan_actuator_coil", Name: "Titan Actuator Servo-Coil", Description: "A heavy-duty kinetic actuator coil salvaged from derelict mech chassis.", MaxStack: 100, BuyPrice: 92, SellPrice: 36},
	{ID: "singularity_core_cell", Name: "Singularity Micro-Core", Description: "A stabilized gravimetric power core recovered from shattered warp wreckage.", MaxStack: 100, BuyPrice: 155, SellPrice: 62},
	// Tier 6-8 Monster Materials (+5 items)
	{ID: "ancient_heartwood_sap", Name: "Ironbark Amber Resin", Description: "Crystallized metallic amber sap harvested from an Ironbark Ancient.", MaxStack: 100, BuyPrice: 0, SellPrice: 34},
	{ID: "glacial_core_shard", Name: "Glacial Warden Core", Description: "A sub-zero crystalline core pried from a Glacial Frost-Warden.", MaxStack: 100, BuyPrice: 0, SellPrice: 38},
	{ID: "solaris_fire_gland", Name: "Solaris Ignition Gland", Description: "A superheated plasma organ recovered from a Solaris Sun-Drake.", MaxStack: 100, BuyPrice: 0, SellPrice: 44},
	{ID: "temporal_phase_tendril", Name: "Temporal Phase Tendril", Description: "A chrono-shifted sinew filament harvested from a Temporal Phase-Hound.", MaxStack: 100, BuyPrice: 0, SellPrice: 50},
	{ID: "abyssal_behemoth_plate", Name: "Abyssal Goliath Carapace", Description: "An interlocking chasm-forged armor plate taken from an Abyssal Chasm Goliath.", MaxStack: 100, BuyPrice: 0, SellPrice: 62},
	// Tier 6-8 Masterwork Combat Weapons & Armor Across All 10 Combat Skills (+10 items)
	{ID: "overlord_siege_maul", Name: "Melee Overlord Siege Maul", Description: "A gravimetric siege maul forged from neutronium and a titan actuator coil for veteran Melee fighters.", MaxStack: 1, BuyPrice: 0, SellPrice: 255, EquipmentSlot: "main_hand", RequiredSkillID: "melee", RequiredSkillLevel: 7, AttackBonus: 11, ConstitutionBonus: 6, StaminaBonus: 4},
	{ID: "abyssal_artillery_cannon", Name: "Ranged Abyssal Siege Arbalest", Description: "A heavy singularity-drawn siege arbalest unlocking at Ranged Level 7.", MaxStack: 1, BuyPrice: 0, SellPrice: 260, EquipmentSlot: "main_hand", RequiredSkillID: "ranged", RequiredSkillLevel: 7, AttackBonus: 11, StaminaBonus: 6, WillpowerBonus: 4},
	{ID: "juggernaut_war_plate", Name: "Strength Juggernaut War-Harness", Description: "An actuator-assisted neutronium war-harness that magnifies raw physical Strength.", MaxStack: 1, BuyPrice: 0, SellPrice: 270, EquipmentSlot: "body", RequiredSkillID: "strength", RequiredSkillLevel: 7, AttackBonus: 5, DefenseBonus: 8, ConstitutionBonus: 8, StaminaBonus: 5},
	{ID: "dreadnought_aegis_helm", Name: "Defense Dreadnought Aegis Greathelm", Description: "An abyssal adamant and goliath-plate greathelm unlocking at Defense Level 8.", MaxStack: 1, BuyPrice: 0, SellPrice: 285, EquipmentSlot: "head", RequiredSkillID: "defense", RequiredSkillLevel: 8, DefenseBonus: 8, ConstitutionBonus: 8, WillpowerBonus: 4},
	{ID: "sovereign_greatsword", Name: "Chronite Sovereign Greatsword", Description: "A phase-honed chronite greatsword bound with ironbark amber and voidbark heartwood.", MaxStack: 1, BuyPrice: 0, SellPrice: 310, EquipmentSlot: "main_hand", RequiredSkillID: "blades", RequiredSkillLevel: 8, AttackBonus: 13, ConstitutionBonus: 5, StaminaBonus: 5},
	{ID: "solaris_halberd", Name: "Solaris Thermal Halberd", Description: "A star-metal polearm superheated by a Solaris ignition gland.", MaxStack: 1, BuyPrice: 0, SellPrice: 265, EquipmentSlot: "main_hand", RequiredSkillID: "polearms", RequiredSkillLevel: 7, AttackBonus: 11, ConstitutionBonus: 5, StaminaBonus: 4},
	{ID: "chronite_kinetic_gauntlets", Name: "Chronite Kinetic Gauntlets", Description: "Time-accelerated striking cestus powered by a singularity micro-core.", MaxStack: 1, BuyPrice: 0, SellPrice: 295, EquipmentSlot: "main_hand", RequiredSkillID: "unarmed", RequiredSkillLevel: 8, AttackBonus: 12, ConstitutionBonus: 5, StaminaBonus: 6},
	{ID: "cryo_pulse_blaster", Name: "Glacial Cryo-Pulse Repeater", Description: "A precision electrum sidearm cooled by a glacial warden core and cryo-lotus extract.", MaxStack: 1, BuyPrice: 0, SellPrice: 250, EquipmentSlot: "main_hand", RequiredSkillID: "pistols", RequiredSkillLevel: 7, AttackBonus: 10, StaminaBonus: 5, WillpowerBonus: 4},
	{ID: "elderwood_longbow", Name: "Voidbark Singularity Longrifle", Description: "A masterwork longrifle stock-carved from abyssal voidbark and powered by a singularity micro-core.", MaxStack: 1, BuyPrice: 0, SellPrice: 315, EquipmentSlot: "main_hand", RequiredSkillID: "rifles", RequiredSkillLevel: 8, AttackBonus: 13, StaminaBonus: 7, WillpowerBonus: 4},
	{ID: "sunfire_focus_scepter", Name: "Astral Nova Scepter", Description: "An aether-bough and astral mycelium scepter crowned with a Solaris ignition focus.", MaxStack: 1, BuyPrice: 0, SellPrice: 305, EquipmentSlot: "main_hand", RequiredSkillID: "magic", RequiredSkillLevel: 8, AttackBonus: 12, ConstitutionBonus: 4, WillpowerBonus: 10},
	// Tier 6-8 Culinary Feasts (+2 items -> 151 total)
	{ID: "auric_harvest_feast", Name: "Auric Harvest Roast", Description: "A golden grain and cryo-lotus banquet that restores 135 Health, 70 Action, and 70 Mind while boosting Max pools by +45 and attributes by +6 for 15 minutes.", MaxStack: 10, BuyPrice: 175, SellPrice: 68, Consumable: true, HealAmount: 135, ActionRestoreAmount: 70, MindRestoreAmount: 70, BuffHealthAmount: 45, BuffActionAmount: 45, BuffMindAmount: 45, ConstitutionBonus: 6, StaminaBonus: 6, WillpowerBonus: 6, BuffDurationSeconds: 900},
	{ID: "astral_ambrosia_elixir", Name: "Celestial Astral Nectar", Description: "A pinnacle culinary nectar that restores 175 Health, 95 Action, and 95 Mind while boosting Max pools by +70 and attributes by +9 for 20 minutes.", MaxStack: 10, BuyPrice: 260, SellPrice: 102, Consumable: true, HealAmount: 175, ActionRestoreAmount: 95, MindRestoreAmount: 95, BuffHealthAmount: 70, BuffActionAmount: 70, BuffMindAmount: 70, ConstitutionBonus: 9, StaminaBonus: 9, WillpowerBonus: 9, BuffDurationSeconds: 1200},
	// Tier 10 Capstone Gathering & Smelting Resources (+6 items)
	{ID: "runic_banite_ore", Name: "Runic Banite Ore", Description: "A cyan-veined runic mineral fused with dense banite metal from ancient frontier monoliths.", MaxStack: 100, BuyPrice: 195, SellPrice: 78},
	{ID: "runic_banite_bar", Name: "Runic Banite Alloy Bar", Description: "A supreme runic banite ingot tempered with nova prism energy for Level 10 apex arms and armor.", MaxStack: 100, BuyPrice: 420, SellPrice: 168},
	{ID: "elder_wroshyr_timber", Name: "Elder Wroshyr Heart-Timber", Description: "Immense spiral-grained heartwood harvested from centuries-old Elder Wroshyr great-trees.", MaxStack: 100, BuyPrice: 185, SellPrice: 74},
	{ID: "sovereign_torstol_leaf", Name: "Sovereign Torstol-Bacta Leaf", Description: "An aromatic apex herb saturated with bio-regenerative torstol alkaloids and bacta sap.", MaxStack: 100, BuyPrice: 185, SellPrice: 74},
	{ID: "spirit_muja_fruit", Name: "Sacred Spirit-Muja Fruit", Description: "A golden spirit-tree fruit prized by master chefs for supreme restorative banquets.", MaxStack: 100, BuyPrice: 185, SellPrice: 74},
	{ID: "hypergate_relic_core", Name: "Hyper-Gate Relic Core", Description: "An ancient self-sustaining hyper-gate power matrix recovered from precursor ruins.", MaxStack: 100, BuyPrice: 195, SellPrice: 78},
	// Tier 9-10 Elite Creature & Level 10 Apex Boss Trophy Materials (+6 items, loot-only)
	{ID: "krayt_wyrm_pearl", Name: "Flawless Krayt-Wyrm Pearl", Description: "A lustrous crystalline gizzard pearl recovered from a Canyon Krayt-Wyrm or the Ancient Krayt-Jad Sovereign.", MaxStack: 100, BuyPrice: 0, SellPrice: 95},
	{ID: "ichor_spirit_sigil", Name: "Shadow-Ichor Spirit Sigil", Description: "An arcane spirit-ward sigil condensed from Night-Coven Hexweaver ichor.", MaxStack: 100, BuyPrice: 0, SellPrice: 95},
	{ID: "noxious_reaper_limb", Name: "Noxious Acklay Scythe-Limb", Description: "A venom-honed monomolecular forelimb harvested from a Geonosian Chitin-Reaper.", MaxStack: 100, BuyPrice: 0, SellPrice: 98},
	{ID: "bandosian_gorax_tasset", Name: "Crag-Gorax War-Tasset Plate", Description: "A colossal iron-bound war-plate pried from an Endor Crag-Gorax.", MaxStack: 100, BuyPrice: 0, SellPrice: 105},
	{ID: "sovereign_jad_fire_core", Name: "Sovereign Jad-Fire Core", Description: "The blazing volcanic heart-core of the Ancient Krayt-Jad Sovereign, radiating cataclysmic power.", MaxStack: 50, BuyPrice: 0, SellPrice: 160},
	{ID: "ancient_krayt_dragon_scale", Name: "Ancient Krayt-Sovereign Scale", Description: "An impenetrable draconic crest-scale shed by the Ancient Krayt-Jad Sovereign.", MaxStack: 100, BuyPrice: 0, SellPrice: 115},
	// Tier 10 Apex Weapons & Armor Across All 10 Combat Skills (+11 items)
	{ID: "gorax_titan_maul", Name: "Elder Gorax-Maul", Description: "A seismic two-handed war-maul forged from runic banite and Crag-Gorax war-plates for Level 10 Melee masters.", MaxStack: 1, BuyPrice: 0, SellPrice: 410, EquipmentSlot: "main_hand", RequiredSkillID: "melee", RequiredSkillLevel: 10, AttackBonus: 16, ConstitutionBonus: 8, StaminaBonus: 5},
	{ID: "zaryte_bowcaster_cannon", Name: "Zaryte-Wroshyr Heavy Bowcaster", Description: "An ancient hyper-gate powered heavy bowcaster carved from elder wroshyr heartwood for Level 10 Ranged masters.", MaxStack: 1, BuyPrice: 0, SellPrice: 415, EquipmentSlot: "main_hand", RequiredSkillID: "ranged", RequiredSkillLevel: 10, AttackBonus: 16, StaminaBonus: 8, WillpowerBonus: 5},
	{ID: "torva_gorax_warplate", Name: "Torva-Gorax Sovereign Harness", Description: "Apex strength-amplifying battle armor laminated with runic banite and Crag-Gorax tasset plates.", MaxStack: 1, BuyPrice: 0, SellPrice: 430, EquipmentSlot: "body", RequiredSkillID: "strength", RequiredSkillLevel: 10, AttackBonus: 7, DefenseBonus: 11, ConstitutionBonus: 10, StaminaBonus: 6},
	{ID: "elysian_spirit_cuirass", Name: "Elysian Krayt-Scale Cuirass", Description: "An impenetrable spirit-sigil and Krayt-Sovereign scale chestplate unlocking at Defense Level 10.", MaxStack: 1, BuyPrice: 0, SellPrice: 445, EquipmentSlot: "body", RequiredSkillID: "defense", RequiredSkillLevel: 10, DefenseBonus: 14, ConstitutionBonus: 12, WillpowerBonus: 6},
	{ID: "infernal_jad_greathelm", Name: "Infernal Jad-Crest Greathelm", Description: "A molten-crested greathelm infused with a Sovereign Jad-Fire Core for Level 10 Defense juggernauts.", MaxStack: 1, BuyPrice: 0, SellPrice: 425, EquipmentSlot: "head", RequiredSkillID: "defense", RequiredSkillLevel: 10, DefenseBonus: 11, ConstitutionBonus: 10, StaminaBonus: 6, WillpowerBonus: 6},
	{ID: "krayt_fang_whip_blade", Name: "Abyssal Krayt-Pearl Saber", Description: "A segmented runic saber focused through a flawless Krayt-Wyrm pearl for Level 10 Blades grandmasters.", MaxStack: 1, BuyPrice: 0, SellPrice: 450, EquipmentSlot: "main_hand", RequiredSkillID: "blades", RequiredSkillLevel: 10, AttackBonus: 17, ConstitutionBonus: 7, StaminaBonus: 7},
	{ID: "noxious_acklay_halberd", Name: "Noxious Acklay-Scythe Halberd", Description: "A sweeping venom-crested polearm forged from an Acklay scythe-limb and runic banite.", MaxStack: 1, BuyPrice: 0, SellPrice: 440, EquipmentSlot: "main_hand", RequiredSkillID: "polearms", RequiredSkillLevel: 10, AttackBonus: 17, ConstitutionBonus: 8, StaminaBonus: 6},
	{ID: "beskar_dragon_claws", Name: "Runic Dragon-Talon Vibroclaws", Description: "Twin hyper-resonant vibro-talons forged from runic banite and Krayt-Sovereign scales for Level 10 Unarmed masters.", MaxStack: 1, BuyPrice: 0, SellPrice: 435, EquipmentSlot: "main_hand", RequiredSkillID: "unarmed", RequiredSkillLevel: 10, AttackBonus: 17, ConstitutionBonus: 6, StaminaBonus: 8},
	{ID: "westar_armadyl_blaster", Name: "Armadyl-Nova Dual Repeater", Description: "A high-velocity precision sidearm tuned with a Krayt-Wyrm pearl and hyper-gate relic core.", MaxStack: 1, BuyPrice: 0, SellPrice: 420, EquipmentSlot: "main_hand", RequiredSkillID: "pistols", RequiredSkillLevel: 10, AttackBonus: 16, StaminaBonus: 8, WillpowerBonus: 6},
	{ID: "twisted_wroshyr_longbow", Name: "Twisted Wroshyr-Heart Longrifle", Description: "A legendary long-range marksmanship rifle carved from elder wroshyr timber and a Sovereign Jad-Fire Core.", MaxStack: 1, BuyPrice: 0, SellPrice: 455, EquipmentSlot: "main_hand", RequiredSkillID: "rifles", RequiredSkillLevel: 10, AttackBonus: 17, StaminaBonus: 9, WillpowerBonus: 5},
	{ID: "tumeken_ichor_scepter", Name: "Shadow-Ichor Sovereign Staff", Description: "An apex channeler's grand staff crowned with a Shadow-Ichor Spirit Sigil and Sovereign Jad-Fire Core.", MaxStack: 1, BuyPrice: 0, SellPrice: 450, EquipmentSlot: "main_hand", RequiredSkillID: "magic", RequiredSkillLevel: 10, AttackBonus: 17, ConstitutionBonus: 6, WillpowerBonus: 14},
	// Tier 9-10 Capstone Consumables (+3 items -> 177 total)
	{ID: "rakatan_stim_infuser", Name: "Hyper-Gate Nanite Medpack", Description: "An apex trauma infuser that immediately restores 210 Health, 95 Action, and 95 Mind.", MaxStack: 10, BuyPrice: 320, SellPrice: 128, Consumable: true, HealAmount: 210, ActionRestoreAmount: 95, MindRestoreAmount: 95},
	{ID: "saradomin_muja_banquet", Name: "Sovereign Muja-Brew Feast", Description: "A golden spirit-muja and torstol banquet that restores 195 Health, 105 Action, and 105 Mind while boosting Max pools by +85 and attributes by +11 for 20 minutes.", MaxStack: 10, BuyPrice: 340, SellPrice: 136, Consumable: true, HealAmount: 195, ActionRestoreAmount: 105, MindRestoreAmount: 105, BuffHealthAmount: 85, BuffActionAmount: 85, BuffMindAmount: 85, ConstitutionBonus: 11, StaminaBonus: 11, WillpowerBonus: 11, BuffDurationSeconds: 1200},
	{ID: "overload_bacta_flask", Name: "Supreme Overload Bacta-Flask", Description: "The ultimate alchemical infusion: restores 220 Health, 120 Action, and 120 Mind while boosting Max Health/Action/Mind by +120 and Constitution/Stamina/Willpower by +15 for 25 minutes.", MaxStack: 10, BuyPrice: 450, SellPrice: 180, Consumable: true, HealAmount: 220, ActionRestoreAmount: 120, MindRestoreAmount: 120, BuffHealthAmount: 120, BuffActionAmount: 120, BuffMindAmount: 120, ConstitutionBonus: 15, StaminaBonus: 15, WillpowerBonus: 15, BuffDurationSeconds: 1500},
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
