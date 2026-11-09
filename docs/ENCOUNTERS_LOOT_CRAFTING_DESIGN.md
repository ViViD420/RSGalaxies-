# RS Galaxies — Encounters, Loot and Crafting Direction

Status: architecture direction for future implementation. This document deliberately defines boundaries and invariants, not final boss statistics, drop rates, item tiers or recipe balance.

## Product goal

Build a sandbox MMORPG where combat, exploration, gathering, crafting and trade depend on one another. The intended blend is the freedom and player economy of a profession-driven sandbox with the readable skill advancement, equipment goals, dangerous encounters and boss progression of a classic skill MMORPG. All setting, story, assets, enemies, items and implementation remain original to RS Galaxies.

A boss should not be valuable only because it drops a finished weapon. Bosses should also introduce rare materials, components, schematics, crafting catalysts and resources that give gatherers, crafters and merchants reasons to participate in the wider economy.

## Current prototype status (November 2026)

The ordinary enemy foundation now includes persistent per-spawn idle/chase/telegraph/return state, fresh-presence targeting, sticky single-target selection, bounded movement, server-side attack timing, telegraph interruption, leash return, shared cooldowns, respawn reset, transactional player damage/Defense XP and authenticated Godot snapshots/interpolation. Player attacks use the persisted live enemy position. This is a first-zone prototype, not the final encounter framework.

The server now assigns a unique encounter-cycle ID to each enemy spawn life and records server-authoritative damage dealt, damage received and action counts for characters who engage. Autonomous incoming damage updates an existing participant record without making an unengaged bystander eligible. Enemy target candidate ordering now preserves the active sticky target first, then scores current-cycle threat (2× damage dealt + damage received + action count), then uses distance and character ID as deterministic tie-breakers. This threat score only chooses an AI target; it is not yet a reward-eligibility rule. PostgreSQL integration coverage verifies that a higher-threat participant outranks a nearer bystander; server Actions #471 and #472 passed. Migration 0013 now establishes a durable settlement key per enemy/encounter cycle and one character-owned pending/claimed reward record per eligible participant, with database-enforced idempotency, ownership, state and loot-quantity constraints. This is schema foundation only: the kill path does not yet create settlements, no claim API exists, and eligibility/anti-leech policy is not yet active. Still outstanding before boss rewards: party rules, eligibility windows and anti-leech policy, transactional settlement creation, idempotent claim execution, inventory-full claim behavior, and a dedicated loot-table service. Current ordinary enemies still use fixed starter drops. Do not jump directly to high-value boss loot.

## Implementation order

1. Finish the ordinary character and world foundations: stable skill progression, authoritative combat actions, equipment requirements, inventory capacity, player health/death/recovery, movement and persistence.
2. Complete ordinary enemy behavior and persistence: the single-target AI/telegraph/movement/return vertical slice, durable spawn-cycle IDs and per-character contribution ledger are implemented; add threat/party contribution rules and reliable reward settlement.
3. Introduce the reusable boss encounter framework and implement one modest first boss as a vertical slice.
4. Add a dedicated loot-table and reward service, then migrate ordinary enemies and the first boss from direct fixed drops.
5. Expand crafting from simple recipes into a profession-driven production loop using tools, stations, skill requirements, material families and quality.
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
