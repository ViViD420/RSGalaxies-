# RS Galaxies — Encounters, Loot and Crafting Direction

Status: architecture direction for future implementation. This document deliberately defines boundaries and invariants, not final boss statistics, drop rates, item tiers or recipe balance.

## Product goal

Build a sandbox MMORPG where combat, exploration, gathering, crafting and trade depend on one another. The intended blend is the freedom and player economy of a profession-driven sandbox with the readable skill advancement, equipment goals, dangerous encounters and boss progression of a classic skill MMORPG. All setting, story, assets, enemies, items and implementation remain original to RS Galaxies.

A boss should not be valuable only because it drops a finished weapon. Bosses should also introduce rare materials, components, schematics, crafting catalysts and resources that give gatherers, crafters and merchants reasons to participate in the wider economy.

## Current prototype status (November 2026)

The ordinary enemy foundation now includes persistent per-spawn idle/chase/telegraph/return state, fresh-presence targeting, sticky single-target selection, bounded movement, server-side attack timing, telegraph interruption, leash return, shared cooldowns, respawn reset, transactional player damage/Defense XP and authenticated Godot snapshots/interpolation. Player attacks use the persisted live enemy position. This is a first-zone prototype, not the final encounter framework.

The server assigns a unique encounter-cycle ID to each enemy spawn life and records server-authoritative damage dealt, damage received and action counts for engaged characters. Target selection preserves a valid sticky target, then scores current-cycle threat (2× damage dealt + damage received + action count), with deterministic distance/character-ID tie-breakers; this score is separate from reward eligibility. Server Actions #471/#472 passed the threat-selection coverage. Ordinary enemy defeats now create durable, idempotent settlement records and account-owned pending claims. Eligibility requires at least 10% of enemy maximum health in verified damage and a recent damage action within two minutes; incoming damage and passive presence do not qualify. XP is split by eligible damage, while the one ordinary loot result is assigned to the highest-damage eligible participant. Claim redemption atomically applies XP and inventory, remains pending on a full stack, and is idempotent. The authenticated claim API and Godot Rewards panel are implemented; server Actions #497 and #498–#503 and client Actions #129–#132 validated these slices. Server-owned weighted ordinary-enemy tables now include no-drop outcomes and cryptographically secure production rolls; Actions #518 passed. Remaining before boss reward policy: party membership, support/healing contribution, group eligibility and boss-specific settlement rules.

## Implementation order

1. Finish the ordinary character and world foundations: stable skill progression, authoritative combat actions, equipment requirements, inventory capacity, player health/death/recovery, movement and persistence.
2. Complete ordinary enemy behavior and persistence: AI/telegraphs, durable spawn-cycle IDs, threat selection, damage-based anti-leech, atomic reward settlement/claims and initial weighted loot tables are implemented; add party and support contribution rules.
3. Introduce the reusable server-owned boss encounter framework and implement one modest first boss as a vertical slice, reusing durable settlement/claim infrastructure.
4. Expand loot tables with multiple/guaranteed/rare drops, quantity bands, versioned tuning and boss-exclusive materials.
5. Expand crafting into a profession-driven production loop using tools, stations, skill requirements, material families and quality.
6. Add more bosses, regions, resources, recipe chains, markets and player services only after the first end-to-end loop is reliable.

## Boss encounter framework

### Data-driven definitions

A boss definition should be server-owned and versioned. It should describe:

- Stable boss and encounter IDs, display name, zone/spawn anchor and recommended progression band.
- Base health, defensive properties, movement/attack profile, interaction range and reset/leash limits.
- Phase thresholds and transitions, including entry conditions and any one-time phase actions.
- Ability definitions: telegraph duration, target rules, range/area, damage/effects, cooldown and recovery windows.
- Adds/minions and environmental hazards, if any, as separately defined encounter actors.
- Respawn/cooldown policy and reward-table reference.

Keep content data separate from encounter execution. The same state machine should support different bosses without hardcoding a separate combat implementation for every creature.

### Encounter state machine

Minimum states: Dormant, Engaged, PhaseTransition, Defeated, Resetting. Transitions are server-authoritative and persisted where a restart or multiple world-server instances could otherwise duplicate rewards or reset a live encounter incorrectly.

Each encounter needs:
- A single authoritative owner/lock for state mutation.
- A unique encounter instance or spawn-cycle ID.
- Server-side phase and cooldown clocks.
- A defined reset policy if players leave, die or fail the encounter.
- Idempotent defeat/reward settlement so retries, concurrent attacks and process restarts cannot pay twice.

### Group contribution and fairness

Before implementing group boss rewards, define party/group membership, damage/healing/support contribution, eligibility windows, encounter participation and anti-leech rules. Never trust client-reported damage, healing, contribution or party membership. Prevent kill-stealing and reward duplication without making casual participation impossible.

Do not commit to personal loot versus shared loot until party/group and trade rules are mature. The reward service should support either policy through explicit server configuration, not duplicate reward logic.

### Telegraphs and counterplay

Boss attacks should communicate danger clearly in the client: wind-up, affected area/target, impact and recovery. The server determines whether a player was in range/area at resolution time. Client animations are presentation only. Boss mechanics should reward positioning and timing rather than only higher statistics.

## Loot and reward system

### Drop tables

Create reusable, server-owned loot tables with:
- Guaranteed rewards and optional weighted rolls.
- Item ID, quantity bounds, eligibility conditions and table references.
- Per-encounter and per-player reward settlement policies.
- Optional unique/rare categories and explicitly versioned drop-rate tuning.
- Deterministic testable random sources; production rolls generated only on the server.
- Audit events for valuable rewards and duplicate-settlement prevention.

A drop table is not a client-supplied list. Clients request an action; the server resolves eligibility, rolls the table, validates inventory/reward capacity and commits the reward exactly once.

### Reward categories

Use a mixture of:
- Ordinary creature materials (hides, meat, bones, glands, fibers, salvage).
- Regional resources and refined components.
- Rare boss materials used in high-value recipes.
- Schematics/recipe discoveries and crafting catalysts.
- Currency or trade goods where the economy supports them.
- Direct equipment drops sparingly, so crafting remains a meaningful route to desirable gear.

Boss-exclusive components should create demand across professions, not simply become vendor trash. A rare boss component may be one input in a crafted item alongside ordinary gathered resources and a crafter's work.

### Inventory and economy invariants

- Validate stack limits and available capacity before committing any reward.
- Define what happens when a player has no room: ground loot with an expiry/ownership window, reward claim storage, or another explicit policy. Never silently destroy valuable drops.
- Valuable rewards need durable settlement records and idempotency keys.
- Keep vendor prices, player market prices and NPC reward values separate; avoid creating unlimited currency from repeatable boss kills.
- Add sinks and regional demand before tuning drop rates around an assumed mature economy.

## Crafting progression

The existing station/recipe model is a useful starting point, not the final system. Expand it in layers:

1. **Recipe requirements:** required skill, minimum level, station type, tool family and known/discovered recipe.
2. **Material inputs:** multiple ingredients, quantities, optional substitutes and boss/region-specific components.
3. **Material properties:** resource family, grade/quality, origin and relevant properties. Avoid arbitrary hidden item bonuses; properties must be inspectable and consistently applied.
4. **Craft outcome:** deterministic baseline result first; later quality bands, experimentation and bounded variation.
5. **Output definition:** stackable goods or unique item instances, equipment requirements, modifiers and crafting provenance.
6. **Progression:** XP awarded for valid completed work, with server-side anti-spam and economic checks.
7. **Production chain:** gather/refine/intermediate component/final assembly loops that naturally connect multiple professions.

Every craft must be atomic: validate station, range, tool, recipe, skill, ingredients and output capacity; consume inputs, award XP and create outputs in one transaction. A failed craft must not partially consume materials or award XP.

### Profession identity without forced skill caps

Players may broadly train skills. Specialization emerges from the equipment, tools, station access, recipe knowledge, scarce materials and market choices needed to perform each job efficiently. Advanced branches for the eight non-combat skills remain deferred; do not add them merely to increase the skill count.

## Technical prerequisites before boss loot at scale

- Unique item instances before per-item durability, binding, modifiers, sockets/crystals or multiple individually tracked weapons.
- Transactional inventory/reward writes and durable idempotency for settlement.
- Encounter persistence and concurrency control across world-server instances.
- Party/group identity and contribution tracking before shared boss reward rules.
- Economy ledger and sinks before high-value repeatable rewards.
- Load tests for high-frequency combat, encounter contention and reward settlement.

## First boss acceptance criteria

The first boss is ready only when:
- Its full encounter can be started, fought, reset and defeated through server-authoritative actions.
- Phase transitions and telegraphs are consistent across clients.
- Player death, disconnect, zone departure and server restart do not corrupt encounter state.
- Concurrent attacks cannot duplicate the kill or rewards.
- Reward eligibility and inventory-full behavior are explicit and tested.
- Loot and crafting outputs persist after reconnect/restart.
- At least one boss material has a useful, tested recipe path that feeds the player economy.


## Delivered ordinary-enemy reward foundation

The current implementation is a first ordinary-enemy policy, not the final party/boss loot system:

- Each encounter has one durable settlement key `(enemy_id, encounter_id)` and one durable claim per eligible character. Both are written in the same PostgreSQL transaction that commits the killing attack.
- Eligibility currently requires verified damage dealt equal to at least 10% of enemy maximum health, at least one recorded action, valid account/character/encounter identity and a damage action within the preceding two minutes. Incoming enemy damage does not refresh the damage-recency clock. Passive presence is never sufficient.
- XP is divided proportionally to eligible damage; integer remainder goes to the highest-damage eligible participant. The single ordinary enemy drop goes to that same participant. The specialization used for damage is persisted in the contribution record and determines XP attribution.
- Claims are account/character scoped. XP, parent XP, loot inventory mutation and the claimed timestamp are committed together. A full loot stack aborts the claim transaction and leaves the whole claim pending; retry after freeing stack capacity is safe. A repeated claim after success is a no-op.
- Authenticated API endpoints expose a character's claims and allow an individual claim to be redeemed. Godot claim UI and client feedback remain to be integrated.
- This policy is intentionally limited: no party membership, support/healing contribution types, multi-drop weighted loot tables, unique item instances, loot rolls or boss-specific distribution rules. Do not treat it as the final group-boss policy.


## Multi-drop loot resolver foundation (implementation committed; persistence integration pending)

The domain catalog now supports versioned loot tables with guaranteed drops and multiple independent weighted bonus rolls. Each bonus roll has its own 10,000-point range; a total weight below 10,000 creates an explicit no-drop interval. Quantity ranges are independently resolved for each guaranteed or selected bonus result, and duplicate item results are aggregated with stack-cap validation. Existing one-roll tables remain supported through a compatibility adapter.

Two version-1 prototype tables demonstrate the intended shape: Frontier Rat guarantees one Rat Pelt and independently rolls a 20% chance of 1–2 Rat Teeth; Wild Boar guarantees one Boar Hide and independently rolls a 15% chance of 1–2 Boar Tusks. These are domain/catalog examples only and are not yet assigned to live enemy definitions. The live settlement and claim schema still stores a single loot item/quantity per character claim, so do not switch live enemies to these tables until the persistence migration, claim DTO/API, atomic redemption and Godot Rewards panel support multiple item lines. The next persistence slice must preserve the existing one-settlement idempotency boundary and full-inventory retry semantics across every item in a claim.

## Initial weighted loot-table implementation (server Actions #518 passed)

The first data-driven tables are server-owned and use a 10,000-point roll range:

- Frontier Rat: 70% Rat Pelt, 20% Rat Tooth, 10% no drop.
- Wild Boar: 75% Boar Hide, 15% Boar Tusk, 10% no drop.
- The world server generates independent cryptographically secure drop and quantity rolls. The pure game resolver is deterministic and boundary-tested.
- The single result (or no-drop result) is resolved once during the unique encounter settlement and attached only to the highest-damage eligible participant's pending claim. XP distribution remains contribution-weighted.
- This is the live settlement slice: one selected item per encounter, with the multi-drop domain resolver implemented separately but not yet wired to persistence. Rarity modifiers, boss tables, unique item instances and per-player roll policy remain future work.


## First rare-material crafting sink

The Rat Tooth now feeds a forge recipe: two Rat Teeth, two Copper Bars and one Wooden Plank produce a Rat-Tooth Blade at Smithing level 2. The blade itself requires Blades level 2 to equip and provides a modest attack bonus. This deliberately separates the crafting profession gate (Smithing) from the combat equipment gate (Blades), reinforcing profession-specific gear progression. The recipe and its skill-gated equipment test are implemented. The latest Rat-Tooth Blade server Actions #520–#523 remain unverified in this handoff because current check-run status is not exposed by the available repository integration; do not infer a pass.


## Second rare-material crafting sink: Boar Tusk

The Boar Tusk now feeds a separate polearm progression path. Two Boar Tusks, two Copper Bars and one Wooden Plank produce a Boar-Tusk Spear at the Frontier Forge, requiring Smithing level 3 and awarding 120 Smithing XP. The spear itself requires Polearms level 2 to equip and grants +4 attack. This preserves the distinction between the production profession gate and the combat specialization equipment gate. Boar Tusks remain loot-only; the starter merchant cannot sell the material or the finished spear. The crafting test covers ingredient consumption, output creation, the Polearms gate and the final equipment bonus. GitHub Actions verification is pending.