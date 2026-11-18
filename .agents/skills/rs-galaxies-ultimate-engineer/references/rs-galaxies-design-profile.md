# RS Galaxies Design Profile

## Vision

RS Galaxies is an original persistent-world science-fiction sandbox MMORPG combining:
- **SWG-inspired sandbox depth:** open-world exploration, interdependent harvesting and crafting professions, primary **HAM** pools (`Health`, `Action`, `Mind`) backed by secondary attributes (`Constitution`, `Stamina`, `Willpower`), medical packs and attribute enhancers, player-driven economy, and social settlements.
- **OSRS-inspired readable progression:** broad level 1–99 skill advancement where actions train the skills they use, clear tier unlocks across gathering nodes, smelted alloys, recipes, equipment, and monsters, and high item utility.
- **RS3-inspired modern usability:** responsive HUD and panels, contextual interaction prompts, clear combat/telegraph feedback, walk/run and autorun controls, first/third-person camera toggle, and mobile-friendly touch ergonomics.

## Active 18-skill roster

- **Combat parent skills (2):** `melee`, `ranged`
- **Combat specializations (6):** `blades`, `polearms`, `unarmed` (Melee parent); `pistols`, `rifles`, `magic` (Ranged parent)
- **Combat co-skills (2):** `defense`, `strength` (`strength` XP and damage bonus apply to `melee` attacks only; never bleed into `ranged` or `magic`)
- **Gathering skills (4):** `mining`, `woodcutting`, `foraging`, `farming`, `scavenging` (note: `farming`, `foraging`, `mining`, `scavenging`, `woodcutting` are 5 gathering skills paired with 3 production skills below to total 8 non-combat skills)
- **Production skills (3):** `cooking`, `crafting`, `smithing`

Characters can train all 18 skills without arbitrary class caps. Specialization comes from tool/weapon requirements, station access, recipe tiers, and economic interdependence. Jedi remains a documented future hidden progression path (`docs/JEDI_PROGRESSION_DESIGN.md`) and is not exposed in active gameplay.

## Current development priority

Boss expansion beyond the verified Stoneback Colossus vertical slice is paused. Priority is on core gameplay quality: movement, deterministic terrain and biome props, 18-skill gathering/crafting/combat loops, HAM + secondary attribute progression, responsive desktop/mobile UI panels, and strict error-free client/server execution.
