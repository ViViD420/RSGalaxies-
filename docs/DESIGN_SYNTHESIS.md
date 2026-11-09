# RS Galaxies — Source-Driven Design Synthesis

Last updated: 2026-11-09

## Purpose and evidence rules

This document translates the three user-supplied reference JSON manuals into original RS Galaxies design guidance. It is research synthesis, not a specification to reproduce any reference game.

Every source observation below names its exact file and top-level JSON section. The manuals contain structured JSON values and descriptions; they do not contain embedded Markdown documents. No missing mechanic is filled in from outside knowledge. Player motivations and design strengths/trade-offs are interpretations, not claims explicitly stated by the JSON, and are labelled accordingly.

### Source files inspected

| Reference | Exact supplied file | JSON top-level sections | Validation |
|---|---|---:|---|
| Star Wars Galaxies Pre-CU | `/mnt/data/swg_pre_cu_reference_manual.json` | 7 keys including metadata | Valid JSON; entire document read |
| Old School RuneScape | `/mnt/data/osrs_system_reference_manual.json` | 6 keys including metadata | Valid JSON; entire document read |
| RuneScape 3 | `/mnt/data/rs3_system_reference_manual.json` | 3 keys | Valid JSON; entire document read |

These are conversation/project attachments in the current workspace. They were not found in a Git-tracked `source/` folder in the implementation repository. Do not invent a repository path or imply the raw JSON files have been committed.

### Source coverage limitation

The supplied manuals are narrow technical matrices. They provide useful detail for character resources, skills, combat timing/formulas, crafting/resource attributes, economy and settlements. They do not meaningfully describe quests, exploration content, detailed terrain/world geography, camera/game feel, animation systems, accessibility, mobile controls, UI layout, banking, inventory ergonomics, social chat/groups/guilds, housing decoration, or modern onboarding. For those topics, this document uses existing RS Galaxies implementation facts and the user's explicit product direction; it does not attribute unsupported mechanics to the reference JSONs.

---

## 1. Independent source reviews

## 1.1 Star Wars Galaxies — Pre-CU, versions 1.0 to Publish 14.1

Source: `swg_pre_cu_reference_manual.json`.

### Section 1 — `section_1_character_engine`

**Documented mechanics**
- Three primary HAM pools: Health, Action and Mind.
- Each pool has a secondary regeneration modifier and tertiary ability-cost reducer.
- Regeneration is modified by posture and Battle Fatigue; listed postures range from running through sitting.
- Ability cost is reduced by the associated tertiary modifier.
- Health or Action reaching zero causes incapacitation; Mind reaching zero causes unconsciousness.
- Wounds reduce maximum pool ceilings and are treated by different service roles/facilities.
- Battle Fatigue reduces incoming healing and buff effectiveness.
- Bleeding, poison and disease are distinct damage-over-time types with stated pool targets and a two-second tick.

**Likely player motivation (interpretation)**
- Create resource-management decisions beyond a single health bar.
- Make posture and recovery choices matter.
- Create complementary noncombat support roles and reasons to visit social/medical spaces.
- Give combat damage types distinct tactical identities.

**Strengths and trade-offs (interpretation)**
- Strength: separate pools and recovery vectors can make character builds and support roles more expressive.
- Strength: recovery services can create social and economic interdependence.
- Trade-off: multiple pools, wounds, fatigue and posture modifiers increase cognitive load and can make routine recovery feel punitive.
- Trade-off: exact formulas can become opaque unless the UI clearly explains them.

**RS Galaxies disposition**
- **Adapt later:** consider a small number of clearly differentiated combat/action resources only if the core combat loop demonstrates a real need.
- **Adapt later:** make recovery/support professions valuable through optional services and consumables, with clear explanations and non-punitive access.
- **Reject as a direct port:** do not reproduce HAM names, formulas, posture multipliers, wound/fatigue penalties, or the exact DOT implementation.
- **Defer:** do not add multi-pool complexity before the present health, combat, movement and consumable loops are polished.

### Section 2 — `section_2_skill_point_matrix`

**Documented mechanics**
- A 250-point total skill-point budget.
- No character level cap; horizontal skill-box progression.
- Experience is decentralized and awarded for specific task actions.
- Profession progression is described as four branches with four tiers each and a Master Box.
- Starting and advanced combat/commercial professions are enumerated.

**Likely player motivation (interpretation)**
- Encourage deliberate profession identity, opportunity cost and player interdependence.
- Make advancement visibly connected to the activity performed.

**Strengths and trade-offs (interpretation)**
- Strength: task-attributed XP makes the relationship between activity and advancement legible.
- Strength: profession branches can communicate meaningful mastery milestones.
- Trade-off: a hard skill-point budget forces players to abandon other paths and conflicts with the explicit RS Galaxies preference for broad progression.
- Trade-off: extensive profession grids create content and balance maintenance costs.

**RS Galaxies disposition**
- **Accept/adapt:** activity-attributed skill XP as a readable rule; the current server already awards skill XP from gameplay actions.
- **Adapt:** optional mastery tracks, special equipment and advanced recipes may express specialization without limiting which skills can be trained.
- **Reject:** the 250-point allocation ceiling and any rule that prevents a character from learning broad skill sets.
- **Defer:** deep profession branch grids until the current 18-skill foundation and core loops are stable.

### Section 3 — `section_3_sandbox_economy`

**Documented mechanics**
- Resources have randomized integer attributes on a 1–1000 scale, including overall quality, toughness, conductivity, malleability, heat/cold resistance, decay resistance, potential energy, flavor and other properties.
- Resource spawns have randomized 13-character names and a stated 7–21 real-world-day lifetime.
- Industrial harvesters consume credits according to upkeep and energy cost; output depends on efficiency and node density.
- Experimental crafting success is described as a function of artisan skill, resource quality, station grade and a random roll.
- Outcomes range from great success to critical failure; critical failure can damage core materials and Mind.
- Slicing can improve weapon/armor performance and carries component/durability risk.

**Likely player motivation (interpretation)**
- Make resource discovery, sourcing and quality comparison meaningful.
- Give crafters room to make distinct products from different inputs.
- Create a supply chain involving gatherers, industrial infrastructure, crafters and traders.
- Add risk/reward choices to optional item enhancement.

**Strengths and trade-offs (interpretation)**
- Strength: resource attributes can make otherwise identical raw materials economically distinct.
- Strength: material/station/skill interactions create multiple meaningful crafting decisions.
- Trade-off: many attributes are hard to understand unless recipes reveal which properties matter.
- Trade-off: short-lived resource availability and destructive critical failures can punish casual players or destabilize supply.
- Trade-off: upkeep systems add recurring administration and should not be added without a clear player benefit.

**RS Galaxies disposition**
- **Adapt:** a compact, item-specific material-property model once resource definitions and crafting chains need more depth.
- **Adapt:** crafting quality/experimentation as bounded, explainable choices with clear risk disclosure and recoverable failure.
- **Defer:** regional resource batches, depletion cycles and industrial harvesting until the gathering/economy loops and persistence model can support them.
- **Reject as a direct port:** the exact 1–1000 matrix, randomized naming syntax, 7–21-day lifetime, formula, slicing buffs and destructive failure outcomes.

### Section 4 — `section_4_planetary_colonization`

**Documented mechanics**
- Five settlement tiers from Outpost through Metropolis, with stated population thresholds and facility unlocks.
- Municipal tax vectors include property, citizen income and shuttleport travel surcharges.
- Democratic elections are stated to occur every three real-world weeks.

**Likely player motivation (interpretation)**
- Give players long-term shared construction goals, a reason to settle together and civic identity.
- Make settlement growth unlock useful services.

**Strengths and trade-offs (interpretation)**
- Strength: shared milestones can turn individual resource production into community projects.
- Strength: services unlocked by settlement growth can create geographic hubs.
- Trade-off: population gates and taxes can become a griefing/maintenance burden or create empty towns if population density is low.
- Trade-off: governance systems require permissions, moderation and abuse-resistant persistence.

**RS Galaxies disposition**
- **Adapt later:** player-founded settlements, staged construction and service unlocks funded by community effort.
- **Defer:** elections, taxation and population thresholds until housing, permissions, guild/social tools and persistent building placement exist.
- **Reject as a direct port:** exact tier names, thresholds, facilities, tax model and election schedule.

### Section 5 — `section_5_galactic_civil_war`

**Documented mechanics**
- Neutral, on-leave and overt faction states have different PvP and access rules.
- A temporary 300-second enemy flag can be triggered by assisting overt combatants or attacking faction assets.
- Flagged players face starport/shuttleport restrictions.

**Likely player motivation (interpretation)**
- Let players choose levels of faction involvement and make support actions consequential.
- Connect PvP status to world travel and faction-controlled infrastructure.

**Strengths and trade-offs (interpretation)**
- Strength: explicit states can make PvP eligibility more predictable.
- Trade-off: hidden/unclear flag rules can surprise players; travel restrictions can feel punitive.

**RS Galaxies disposition**
- **Defer:** faction conflict and opt-in PvP until the core multiplayer, social and world rules are robust.
- **Adapt:** any future hostile status must be explicit, visible, time-bounded where appropriate and explained before commitment.
- **Reject:** do not copy the exact three-state labels, 300-second duration or travel denial rules.

### Section 6 — `section_6_jedi_unlock`

**Documented mechanics**
- A hidden unlock loop uses rare boss loot, a randomized profession requirement, mastery, dropping skill boxes and repeating several professions.
- A visibility score can flag players on hunter terminals.
- The described early implementation has a finite Force-life counter and permanent character-slot reset after all lives are lost.

**Likely player motivation (interpretation)**
- Make rare mastery feel secret, prestigious and socially consequential.
- Create a high-risk specialist archetype and a player-hunter counter-system.

**Strengths and trade-offs (interpretation)**
- Strength: hidden mastery can create community discovery and long-term goals.
- Trade-off: opaque random requirements, repeated skill-box dropping and permanent character loss can feel arbitrary or hostile.
- Trade-off: tying the unlock to boss drops creates a content dependency that should not control the core progression.

**RS Galaxies disposition**
- **Accept only the broad direction already recorded by the user:** the six combat specializations are potential future Jedi mastery candidates.
- **Reject:** random holocron profession grinding, skill-box dropping loops, visibility formula and permadeath slot reset.
- **Defer:** any Jedi/Force-like unlock until the user-approved original lore and mastery criteria are designed. It is not implemented.

---

## 1.2 Old School RuneScape

Source: `osrs_system_reference_manual.json`.

### `core_mechanics.game_tick_system`

**Documented mechanics**
- A 600 ms atomic game tick coordinates player input, pathfinding, combat animations and world updates.
- The summary describes 3-tick harvesting, 1-tick prayer flicking and tick-eating as timing techniques.

**Likely player motivation (interpretation)**
- Predictable timing can make actions learnable and create a high skill ceiling.

**Strengths and trade-offs (interpretation)**
- Strength: a consistent simulation cadence simplifies ordering and timing.
- Trade-off: hidden timing windows and input manipulation can reward exploiting implementation details rather than intended play.

**RS Galaxies disposition**
- **Adapt:** deterministic server action ordering and clearly communicated cooldown/timing rules.
- **Reject:** reproduce neither a 600 ms global tick nor tick manipulation exploits.
- **Defer:** select an authoritative simulation cadence based on server performance and gameplay requirements, not nostalgia.

### `core_mechanics.grid_and_movement`

**Documented mechanics**
- A discrete 2D tile grid.
- Walk is one tile per tick; run is two tiles per tick with interpolation.
- Pathfinding is described as modified breadth-first search constrained by line of sight and collision vectors.

**Likely player motivation (interpretation)**
- Discrete spatial rules make movement and interaction positions predictable.

**Strengths and trade-offs (interpretation)**
- Strength: tile rules are simple to reason about and can support deterministic pathfinding.
- Trade-off: grid movement can feel restrictive in a 3D third-person world.

**RS Galaxies disposition**
- **Reject for the current client:** the existing game is a continuous third-person 3D prototype, not a tile-grid game.
- **Adapt later:** bounded pathfinding and collision-aware navigation principles for NPCs, if needed.
- **Preserve:** movement authority must remain server-side for consequential position and action checks.

### `stat_and_experience_matrices`

**Documented mechanics**
- Standard skill level cap 99, XP cap 200,000,000 and a sample of XP milestones.
- A combat-level formula combines defensive and offensive skills.

**Likely player motivation (interpretation)**
- Clear milestones provide long-term goals and a shared vocabulary for progress.

**Strengths and trade-offs (interpretation)**
- Strength: explicit XP thresholds make progress measurable.
- Trade-off: steep curves and a hard cap can create grind or reduce the value of further advancement if item/profession utility is not designed alongside it.

**RS Galaxies disposition**
- **Accept/adapt:** readable level/XP progress and meaningful milestone unlocks. RS Galaxies already has 1–99 skill progression and a 200M XP cap.
- **Reject as a direct port:** exact OSRS XP formula/table and combat-level formula.
- **Review later:** ensure level 99 and post-99 XP have useful, transparent purpose rather than only a number.

### `combat_formulas_and_stat_engine`

**Documented mechanics**
- Melee attack and defense rolls use effective level and equipment bonus terms.
- Hit chance is computed from the relationship between attack and defense rolls.
- Maximum melee hit depends on effective Strength and equipment bonus.
- Weapon speeds are listed in tick counts for selected weapon examples.

**Likely player motivation (interpretation)**
- Give skill level and equipment measurable impact while preserving predictable attack cadence.

**Strengths and trade-offs (interpretation)**
- Strength: formula-driven outcomes are testable and explainable if surfaced clearly.
- Trade-off: opaque hidden rolls or copied coefficients can make balance difficult and would not fit RS Galaxies without calibration.

**RS Galaxies disposition**
- **Adapt:** server-side, deterministic/testable attack and defense resolution with readable feedback.
- **Reject:** copy neither formulas, constants, named weapons nor tick speeds.
- **Priority gap:** server-derived hit/miss outcomes and broader combat accuracy are still roadmap work; combat remains server authoritative.

### `skills_deep_dive`

**Documented mechanics**
- Gathering, artisan, combat and utility skill groupings are enumerated.
- A Runecraft output multiplier increases at listed level milestones.

**Likely player motivation (interpretation)**
- Distinct skill families make the breadth of the game legible; level-based yield multipliers reward long-term training.

**Strengths and trade-offs (interpretation)**
- Strength: skills become useful when they change resource yields or unlock new outputs.
- Trade-off: a skill catalog alone does not create meaningful progression unless the world has item sinks, recipes and unlocks.

**RS Galaxies disposition**
- **Accept/adapt:** skills should have useful outputs and cross-system dependencies, not merely XP bars.
- **Reject:** exact OSRS skill list and multiplier values.
- **Map to current project:** the existing 18-skill roster already spans combat specializations, gathering and production; Farming and other skill loops need continued world integration.

### `economy_and_item_systems`

**Documented mechanics**
- Coins are the currency; the summary describes an automated Grand Exchange market, 1% tax and tax cap.
- Death protection retains the three highest-value items and puts other items in a 15-minute gravestone; a wilderness rule drops unprotected items.

**Likely player motivation (interpretation)**
- A market matching system reduces search friction; a tax acts as a currency sink; death creates risk and item recovery decisions.

**Strengths and trade-offs (interpretation)**
- Strength: market taxes can remove currency from circulation.
- Trade-off: automated markets can centralize trade and undercut player-run services; death loss can be harsh and can create item-recovery chores.

**RS Galaxies disposition**
- **Defer:** player market discovery, regional supply and currency sinks until price/ledger observability and economy safeguards exist.
- **Reject as a direct port:** Grand Exchange matching, tax rate/cap, item-value protection and wilderness drop rules.
- **Adapt later:** a clear, low-friction recovery model and explicit risk zones, after user-facing death rules are decided.

---

## 1.3 RuneScape 3

Source: `rs3_system_reference_manual.json`.

### `game_engine_architecture.tick_and_grid_system`

**Documented mechanics**
- A 600 ms engine heartbeat synchronizes action queues, cooldown counters and animation cycles.
- The summary describes a 2D spatial grid with height levels and instant tile-translation movement abilities that still verify clipping.

**Likely player motivation (interpretation)**
- Centralized action timing and spatial checks keep combat actions coordinated.

**Strengths and trade-offs (interpretation)**
- Strength: a shared action scheduler can align cooldowns and presentation.
- Trade-off: grid-based burst movement may not fit a continuous 3D controller.

**RS Galaxies disposition**
- **Adapt:** a server-owned action/cooldown scheduler if and when a general player ability framework is built.
- **Reject:** exact 600 ms cadence, grid coordinates and named movement mechanics.
- **Preserve:** movement/ability requests must validate range, collision and costs on the server.

### `game_engine_architecture.buff_debuff_and_tick_systems`

**Documented mechanics**
- Dynamic arrays track buffs, debuffs, potion timelines and temporary passive states on the stated tick.
- An adrenaline resource gains from basic abilities, drains outside combat and has a cap that can be increased.

**Likely player motivation (interpretation)**
- A shared combat resource encourages sequencing abilities and creates visible temporary state.

**Strengths and trade-offs (interpretation)**
- Strength: resource bars and buff timelines can communicate combat state.
- Trade-off: another resource and many status effects can overwhelm the player or make balance opaque.

**RS Galaxies disposition**
- **Adapt:** a compact combat resource only if the future ability design needs sequencing; show costs, gain, decay and cap clearly.
- **Reject:** exact adrenaline gain/loss/caps and formulas.
- **Defer:** buff/debuff complexity until the ordinary combat and UI feedback loops are complete.

### `combat_evolution_engine_eoc`

**Documented mechanics**
- Abilities are categorized as basic, threshold and ultimate.
- The summary lists adrenaline gain/cost and a 3-tick global cooldown for basic abilities.
- Styles include melee, ranged, magic and necromancy, with an affinity matrix.
- Damage and hit chance use level, weapon tier, style and affinity terms.

**Likely player motivation (interpretation)**
- Ability tiers create a readable decision structure: build resources, spend some for stronger actions, or commit a full resource bar.
- Cooldowns and target affinity create tactical sequencing.

**Strengths and trade-offs (interpretation)**
- Strength: a small number of ability categories can help players understand their action choices.
- Strength: cooldown and resource feedback make action timing legible.
- Trade-off: complex ability bars can increase cognitive load, particularly on touch devices.
- Trade-off: copied formulas or excessive cooldown layers can make combat feel like UI management.

**RS Galaxies disposition**
- **Adapt:** optional basic/advanced/finisher-style action categories as an original, simplified framework.
- **Accept as architecture principle:** server-owned cost, cooldown, range, target and effect validation; client displays server-confirmed results.
- **Reject:** exact labels, resource percentages, 3-tick global cooldown, affinity formula, damage equation and named style roster.
- **Defer:** ability framework implementation until movement, player animation, interaction and ordinary combat presentation are robust.

### `progression_and_economy_metrics`

**Documented mechanics**
- Standard skills are described at level 99; selected skills can reach level 120 with stated XP milestone and cap.
- Invention dismantles items into components and uses probabilistic perk generation based on component combinations and level.
- The summary describes a 2% Grand Exchange tax and a tiered death reclaim cost with item protection.

**Likely player motivation (interpretation)**
- Extended skill ceilings create longer progression arcs.
- Disassembly makes obsolete items useful as crafting inputs.
- Currency taxes and reclaim costs are sinks.

**Strengths and trade-offs (interpretation)**
- Strength: salvage can connect item progression to crafting and resource circulation.
- Strength: optional higher-level mastery can extend the life of selected skills.
- Trade-off: multiple level caps can confuse progression expectations.
- Trade-off: probabilistic perks can create opaque optimization or material sink pressure.
- Trade-off: death reclaim taxes can disproportionately burden new players.

**RS Galaxies disposition**
- **Adapt candidate:** item salvage/recycling into categorized components, with deterministic and explainable yields first.
- **Defer:** probabilistic item perks and level-120 elite skills until the current skill roster has a coherent progression purpose.
- **Reject as a direct port:** Invention gizmos/perk generation, exact level caps, Grand Exchange tax and Death Office formula.

---

## 2. Cross-source comparison and design implications

| Design domain | SWG evidence | OSRS evidence | RS3 evidence | RS Galaxies direction |
|---|---|---|---|---|
| Character resources | Health/Action/Mind, posture regen, wounds/fatigue | Hitpoints in skill/combat model | Adrenaline and buff/debuff timelines | Keep current health model simple; evaluate extra action resource only with a clear gameplay need |
| Skill progression | 250-point budget, profession boxes, task XP | 1–99, 200M XP, skill families | 99/120 split, elite skills | Broad skill progression; no arbitrary skill-learning cap; clear XP and useful unlocks |
| Specialization | Profession mastery boxes | Skill and equipment progression | Combat style/ability families | Gear, tools, stations, recipes, material access and optional mastery |
| Gathering/resources | Multi-attribute resources, timed spawn batches, harvesters | Gathering skill families and yield multipliers | Not detailed | Build useful resource identity and cross-skill loops; defer quality attributes until needed |
| Crafting | Resource/station/skill experimentation, risk outcomes | Artisan skill families | Item disassembly and components | Deep original crafting chains, explainable quality and later salvage |
| Combat | HAM pools, DOTs, posture and faction status | Attack/defense rolls, max hit, weapon speed | Ability tiers, cooldowns, adrenaline, affinity | Server-authoritative combat, readable outcomes, later action framework; no copied formulas |
| Economy | Player production, upkeep, slicing, settlement taxes | Coins and automated market tax | Market tax and death reclaim sink | Player interdependence first; ledger and economy monitoring before market/tax complexity |
| Housing/settlements | Five settlement tiers, unlocks, elections | Not covered | Not covered | Adapt later after social, permissions and persistent construction foundations |
| Factions/PvP | Explicit faction states and temporary flag | Wilderness risk/death rule | Not detailed | Defer; any PvP must be opt-in/clear and abuse-resistant |
| Loot/death | Boss loot used in hidden mastery; no general death model | Gravestone/protected items/wilderness drops | Reclaim office/protection | Existing multi-item reward claims retained; death policy needs separate user-centered design |
| UI/accessibility/mobile | Not covered | Not covered | Not covered beyond buff bar | Use user's modern-usability goal and current client audits; do not attribute details to JSON |
| World/quests/exploration | Settlement/faction context, but no terrain/quest detail | Not covered | Not covered | Need separate original world/quest design research; current terrain is a prototype foundation |
| Movement/game feel | Posture states; no controller detail | 600 ms tile movement and BFS | 600 ms grid and burst movement | Preserve continuous 3D controller and improve camera, collision, reconciliation and feedback |

### Shared lessons (interpretation)

1. Progression is strongest when actions, skills, resources, items and unlocks reinforce one another.
2. Specialization can create interdependence, but hard restrictions on learning skills conflict with the user's chosen broad-progression model.
3. Resource and item depth is useful only when players can understand which properties matter and what they enable.
4. Combat complexity should be introduced through readable feedback and a manageable action vocabulary, not by copying a reference game's tick or formula.
5. Social/economic systems require security, ownership, persistence and anti-abuse foundations; they should not be rushed ahead of the core world loop.

---

## 3. Existing implementation mapping

| Accepted/adapted direction | Existing implementation | Gap / next dependency | Test expectation |
|---|---|---|---|
| Broad skill progression | 18 active skills, level 1–99, 200M cap; all new characters receive full catalog | Ensure each skill has real item/world utility; Farming and several production loops need expansion | Domain XP tests, API catalog tests, client progression display |
| Activity-attributed XP | Server awards gathering/crafting/combat-related skill XP | Audit all XP paths for correct attribution and idempotence | Unit + PostgreSQL transaction tests |
| Gear-based specialization | Item requirements already include Blades/Defense; equip checks are server-side | Expand tool, station and recipe requirements generically | Domain + integration tests for failure atomicity |
| Resource identity | Server-owned gathering nodes and item catalog exist | Material properties, regional supply and resource lifecycle not yet designed | Deterministic spawn/property tests and economy simulations |
| Crafting depth | Transactional recipes, stations, ingredients, output-capacity checks and XP | Tool requirements, components, quality, experimentation and crafted-item progression | Atomic inventory/XP tests, recipe validation, client explanation |
| Combat authority | Server-calculated damage, cooldowns, Defense/Strength/specialization XP and enemy AI | Broader accuracy/hit outcomes, weapon families, player ability system | Domain + PostgreSQL concurrency tests; client presentation only after server success |
| Multi-item rewards | Durable settlement and idempotent claim system | Rarity/unique item instance design later | Claim concurrency/restart/inventory-full tests |
| Player trade | Escrow, two-party confirmation and atomic settlement | Abuse/load/restart/manual QA gates; currency ledger expansion | PostgreSQL concurrency, restart recovery, load and device QA |
| Settlement building | No complete player housing/city system yet | Requires ownership, permissions, placement, construction, storage and social groups | Permission, persistence, griefing and rollback tests |
| Modern usability | Responsive action toolbar and panels exist | Consistent layout, tooltips, mobile touch layer and physical device QA | Godot smoke tests plus manual desktop/mobile testing |
| World immersion | Deterministic terrain mesh/collision, biome palette, trails | Vegetation, rocks, landmarks, biome-specific placement and performance budgets | Determinism, collision, bounds, performance and scene tests |
| Movement | Client acceleration, deceleration, jump buffering/coyote time, camera zoom; server validates X/Z speed/time/bounds | Latency-aware reconciliation; server vertical authority is unresolved | Movement domain tests plus simulated delayed-response tests |
| Animation/game feel | Mesh-only locomotion feedback and server-confirmed placeholder swing | Explicit presentation state machine and interaction animations | State-transition, cancellation, neutral-pose and collision-separation tests |

### Movement audit snapshot (2026-11-09)

Inspected:
- `client/scripts/player_controller.gd`
- `client/scripts/main.gd`
- `client/tests/player_controller_smoke_test.gd`
- `client/scripts/world_terrain.gd`
- `server/internal/game/movement.go`
- `server/internal/httpapi/server.go`

Observed behavior:
- Godot uses a continuous third-person `CharacterBody3D` controller with camera-relative movement, ground/air acceleration and deceleration, smoothed facing, jump buffering, coyote time, gravity and a fall-speed cap.
- Camera pitch is clamped; spring-arm zoom is bounded; mouse and screen-drag input rotate the camera.
- Locomotion bob/lean/pitch and the initial melee swing affect only the visual mesh. The swing is started after a successful server combat response.
- The client posts X/Z only every 0.2 seconds. It allows only one movement request at a time; while the request is pending, later timer ticks are skipped. A server correction is applied to the local character.
- The server validates action eligibility, finite coordinates, zone bounds and planar speed. Maximum configured speed is 7.0 units/second, movement tolerance is 0.15, and elapsed time is capped to a 250 ms calculation window. Y, jump arc, terrain height and collision are not part of the server contract.
- Terrain is a deterministic 240-unit mesh with 96 subdivisions per side, broad/detail/biome noise, vertex-color blending and four integrated trail paths. The same generated mesh creates the trimesh collision; local client presentation samples the same height function for props/characters.

Risk to investigate (not yet proven as a defect):
- Because client requests are serialized and the server caps the movement calculation window at 250 ms, a slow response can make the actual interval between accepted requests exceed the server's calculation window. The client may have moved farther than the server's capped allowance and receive a correction. Add a latency/delayed-response regression before changing the protocol or constants. Do not weaken speed/bounds validation to mask this.
- The client simulates vertical movement locally, but only planar coordinates are sent. The client must not treat local terrain height or jump state as authoritative.
- Terrain environmental dressing and physical desktop/mobile QA remain incomplete.

No movement protocol change is approved by this synthesis. First reproduce the delayed-request behavior with tests, then design a server-time model that preserves speed and zone-bound checks.

---

## 4. Prioritized original RS Galaxies design

### Core identity

RS Galaxies should be a continuous 3D science-fiction sandbox with broad character progression, strong professions, useful skills, player interdependence, original crafting depth and modern readable interaction.

### Near-term (core-first)

1. Preserve and improve the current movement/controller and server contract; test delayed responses and correction behavior.
2. Improve deterministic terrain with biome-specific environmental dressing, landmarks and coherent resource placement.
3. Build a small, explicit player presentation state machine, starting with idle/move/jump/land/attack/interact; keep it separate from collision and server state.
4. Improve interaction feedback for gathering, crafting, item use and combat.
5. Audit and improve existing panels for responsive sizing, focus, input, tooltips, clear request/error states and mobile ergonomics.
6. Continue connecting the 18 skills to real outputs, items, tools, stations and unlocks.

### Mid-term

- Material properties that matter to selected recipes, not a huge universal stat sheet.
- Crafting components and experimentation with explainable, bounded outcomes.
- Optional specialization/mastery paths that do not restrict broad skill training.
- Salvage/recycling that turns obsolete gear into useful components.
- Player services, social groups, housing and settlement construction after permissions and persistence foundations.
- Regional supply and player markets after a currency ledger and economy telemetry exist.

### Long-term

- Original faction conflict with clear consent/status rules.
- Original mastery unlocks, including any Jedi-like archetype only after separate lore/design approval.
- More advanced abilities and combat resources only after the ordinary action framework is readable and accessible.
- Larger world streaming, events, quests and deeper social institutions.

---

## 5. Explicit exclusions

Do not copy:
- SWG's HAM terminology/formulas, skill-point budget, profession box grid, resource names/attribute matrix, spawn naming/lifetime, slicing, settlement tiers/tax/elections, faction labels/flag rules, holocron grind or Force-life permadeath.
- OSRS's 600 ms tick, tile movement, tick manipulation, exact XP formula, combat formulas, weapon tick speeds, named skill/item/market/death rules.
- RS3's 600 ms tick, grid movement, ability names/categories as distinctive expression, adrenaline constants, hit-chance/damage formulas, Invention gizmos/perks, tax/death formulas.
- Any proprietary code, assets, maps, models, animations, music, dialogue, UI artwork or distinctive protected expression.

The goal is to extract player needs and design patterns, then implement original mechanics, tuning, lore, naming, art and code.

## 6. Open questions that must remain explicit

- What original fiction and lore should govern any Jedi/Force-like mastery? Not decided here.
- Should the core game eventually use a secondary combat resource? Not decided.
- What should the authoritative vertical movement model be? Not decided.
- What death/recovery risk level best serves the intended sandbox? Not decided.
- How should material quality be represented and surfaced in the UI? Candidate for later design.
- Which social/settlement systems should be introduced first after the core loop? Not decided.

Do not silently resolve these questions by importing a reference game's answer.
