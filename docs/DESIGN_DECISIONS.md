# RS Galaxies — Design Decision Log

Last updated: 2026-11-09

This log records decisions made from the supplied SWG Pre-CU, OSRS and RS3 reference JSONs, existing code inspection and explicit user direction. Source sections are referenced precisely in `docs/DESIGN_SYNTHESIS.md`.

Statuses:
- **Accepted:** approved direction for RS Galaxies.
- **Adapted:** underlying pattern is useful, but implementation must be original.
- **Rejected:** do not use the referenced mechanic as a direct design.
- **Deferred:** potentially useful, but dependencies or timing are not ready.
- **Open:** requires a separate design decision; do not guess.

| ID | Decision | Source attribution | Status | Rationale / implementation boundary |
|---|---|---|---|---|
| D-001 | Characters may train many or all skills; no arbitrary cap on skills learned. | User's explicit product decision; contrasts SWG `section_2_skill_point_matrix.system_constraints.total_skill_points_budget` | **Accepted** | Skill access is not a zero-sum skill-point budget. Preserve broad progression. |
| D-002 | Specialization comes from gear, tools, recipes, materials, stations, access and investment. | User's explicit product decision; informed by SWG `section_2_skill_point_matrix` and `section_3_sandbox_economy` | **Accepted** | Build meaningful interdependence without forbidding other skill training. |
| D-003 | XP should be connected to the activity that earns it. | SWG `section_2_skill_point_matrix.experience_distribution`; OSRS `skills_deep_dive` | **Adapted** | Existing server action-based XP remains the model; do not reproduce profession XP labels or exact lists. |
| D-004 | Keep the current 18-skill, level 1–99, 200M cap foundation while improving skill utility. | OSRS `stat_and_experience_matrices`; existing RS Galaxies roster and user direction | **Accepted** | The current catalog is implemented. Revisit cap only with explicit design rationale, not because RS3 has selected level-120 skills. |
| D-005 | Do not reproduce SWG's 250-point skill budget or forced box-dropping. | SWG `section_2_skill_point_matrix`, `section_6_jedi_unlock` | **Rejected** | Conflicts with broad progression and creates avoidable respec friction. |
| D-006 | Optional mastery/specialization branches may exist without restricting broad skill learning. | SWG `section_2_skill_point_matrix`; RS3 `combat_evolution_engine_eoc` | **Adapted** | Mastery is additive, optional and equipment/recipe-oriented; exact profession grids are not copied. |
| D-007 | Resource properties should be introduced only when specific recipes use them. | SWG `section_3_sandbox_economy.resource_attributes_matrix` | **Adapted** | Avoid a large unused attribute matrix. Start with a small set of meaningful, UI-explained properties. |
| D-008 | Regional resource lifecycles and industrial harvesting are later systems. | SWG `section_3_sandbox_economy` | **Deferred** | Requires resource catalog, economy telemetry, persistence and anti-exploit controls. |
| D-009 | Crafting experimentation may be added with bounded, explainable outcomes. | SWG `section_3_sandbox_economy.experimental_crafting_phase` | **Adapted** | Do not copy random hex formula, destructive critical failure or Mind shock. Preserve core ingredients on failure unless a future user-approved risk model says otherwise. |
| D-010 | Slicing-style item modification is not a current feature. | SWG `section_3_sandbox_economy.smuggler_slicing` | **Rejected** | The exact buffs and permanent lockout/explosion risk are not appropriate as a direct port. Future enhancement systems need original rules. |
| D-011 | Player housing/settlements are a long-term pillar, not a near-term code priority. | SWG `section_4_planetary_colonization` | **Deferred** | First implement groups, ownership, permissions, placement, construction and storage. Population tiers/taxes/elections are not copied. |
| D-012 | Faction PvP must be designed as an explicit, readable consent/status system. | SWG `section_5_galactic_civil_war`; OSRS `economy_and_item_systems.death_mechanics` | **Deferred** | No faction PvP system now. Do not copy exact flags, durations, travel restrictions or wilderness drop rules. |
| D-013 | Jedi-like mastery is a future original system; six combat specializations are only candidates. | SWG `section_6_jedi_unlock`; user decision in existing handoff | **Deferred** | Reject random holocron profession grind, visibility formula and permadeath. Requires original lore and explicit design. |
| D-014 | Keep continuous third-person 3D movement, not tile-grid movement. | OSRS `core_mechanics.grid_and_movement`; RS3 `game_engine_architecture.tick_and_grid_system` | **Accepted** | Existing Godot client is a 3D CharacterBody controller. Tile coordinates and exact grid movement do not fit the current product. |
| D-015 | Use deterministic server-side action ordering and testable cooldowns, not a copied 600 ms tick. | OSRS `core_mechanics.game_tick_system`; RS3 `game_engine_architecture.tick_and_grid_system` | **Adapted** | Server currently has a 250 ms enemy worker and HTTP requests. A general player action scheduler remains a future architecture choice. |
| D-016 | Tick-manipulation techniques are not intended mechanics. | OSRS `core_mechanics.game_tick_system.tick_manipulation_techniques` | **Rejected** | Do not reward input exploits or hidden action-order loopholes. |
| D-017 | Combat resolution stays server-authoritative and readable. | OSRS `combat_formulas_and_stat_engine`; RS3 `combat_evolution_engine_eoc` | **Accepted** | Do not copy formulas or constants. Improve server-derived hit/miss and damage feedback with tests. |
| D-018 | A future ability system may use a small, understandable tier structure and cooldown/resource feedback. | RS3 `combat_evolution_engine_eoc.ability_classification`; `game_engine_architecture.buff_debuff_and_tick_systems` | **Adapted** | No exact basic/threshold/ultimate labels, adrenaline values or 3-tick GCD. Keep touch UI manageable. |
| D-019 | A secondary combat resource is not yet approved. | SWG `section_1_character_engine.ham_system`; RS3 `buff_debuff_and_tick_systems.adrenaline_system` | **Open** | First polish current health, combat and consumable loop; add a resource only if it creates useful choices. |
| D-020 | Item salvage/recycling is a promising future crafting/economy loop. | RS3 `progression_and_economy_metrics.elite_skill_invention_architecture` | **Adapted** | Begin with transparent component yields, not probabilistic gizmos/perks. Depends on richer item/component definitions. |
| D-021 | Market taxes and currency sinks require telemetry and a ledger first. | OSRS `economy_and_item_systems.grand_exchange`; RS3 `progression_and_economy_metrics.economy_sinks_and_death_systems` | **Deferred** | No copied Grand Exchange or tax rates. Build ledger, monitor sources/sinks, then tune original mechanisms. |
| D-022 | Death and item recovery rules need a separate player-centered decision. | OSRS `economy_and_item_systems.death_mechanics`; RS3 `progression_and_economy_metrics.economy_sinks_and_death_systems` | **Open** | Do not import gravestone timers, top-item protection or Death Office pricing by default. |
| D-023 | Modern usability, responsive UI and mobile ergonomics are product goals, but the supplied JSONs do not provide detailed UI evidence. | User product direction; source coverage gap in all three JSONs | **Accepted** | Improve existing Godot panels and test on real devices; do not claim this came from the RS3 JSON. |
| D-024 | Quests, exploration, terrain storytelling and environmental dressing require original design research. | Source coverage gap in all three JSONs | **Deferred** | Current terrain is a deterministic prototype; build original landmarks, activities and zone identity incrementally. |
| D-025 | Never trade correctness for speed; all errors must be diagnosed and fixed before unrelated work. | Explicit user requirement and `docs/DEVELOPMENT_RULES.md` | **Accepted** | Preserve failed runs and fixes; do not suppress logs or weaken checks. |

## Decision review rule

When a new proposal conflicts with an accepted decision, identify the conflict and explain the trade-off. Do not silently reverse an accepted decision because a reference game did something else. New decisions should include source section, rationale, alternatives, affected systems, dependencies, tests and status.
