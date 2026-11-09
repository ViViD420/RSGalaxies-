# RS Galaxies — Roadmap

## Milestone 1: Independent foundation

- [x] Establish standalone repository and original project identity.
- [x] Document architecture, design pillars and system boundaries.
- [x] Add Go world API, deterministic unit tests, race checks and vet.
- [x] Add Godot 4.3 third-person client, headless import and Web export validation.
- [x] Add PostgreSQL migrations, integration tests and Codespaces development database.
- [x] Add hosted Codespaces development-container configuration.

## Milestone 2: Account and world slice

- [x] Character identity, account registration/login/logout and bearer-session lifecycle.
- [x] PostgreSQL-backed account and character ownership; account-scoped character APIs.
- [x] First world zone, coordinates, spawn point and server-validated movement.
- [x] Godot account UI, multi-character selection and authenticated gameplay requests.
- [x] Interactable gathering nodes, workbench, forge, bank, merchant and enemy spawn points.
- [x] Safe-spawn recovery and session-expiry recovery.
- [x] PostgreSQL-backed leased character presence and account-scoped zone roster API.
- [x] Godot polling heartbeat and remote-character placeholder visuals with stale-lease cleanup.
- [ ] Reconnect recovery, server-driven movement snapshots and fully authoritative client replication.
- [ ] Email verification, password recovery, MFA and distributed abuse controls before public registration.

## Milestone 3: Skills and progression

- [x] Eighteen active core skills: Melee/Ranged parents, six combat specializations, independent Defense/Strength co-skills, four gathering and four production skills; level 1–99 progression. Legacy Combat/Tactics XP is preserved but inactive.
- [x] Authenticated live skills panel with per-skill XP progress and next-level thresholds.
- [x] Tree, stone, copper and wild-herb gathering with range, skill, charge and respawn rules.
- [x] Godot selects gathering targets and builds resource props from the server node catalog.
- [x] Consolidated public world catalog reduces startup round trips and keeps client definitions server-owned.
- [x] Godot builds and selects workbench/forge stations from the server crafting-station catalog.
- [x] Workbench plank, herbal-bundle and herbal-tea recipes plus forge copper smelting/training-sword recipes.
- [x] Cooking profession progression through crafting a server-defined healing tea consumable.
- [x] First rare enemy material sink: Rat Tooth + Copper Bars + Plank → Smithing level 2 Rat-Tooth Blade, equippable only at Blades level 2 (final combined crafting/catalog server Actions #528 passed).
- [x] Second rare enemy material sink: Boar Tusks + Copper Bars + Plank → Smithing level 3 Boar-Tusk Spear, equippable only at Polearms level 2 (final combined crafting/catalog server Actions #528 passed).
- [x] Transactional crafting with ingredient consumption, output capacity and skill XP.
- [x] Basic PvE enemy definitions, server-calculated damage, cooldowns, retaliation, specialization/parent/co-skill XP, starter loot and respawns.
- [x] Player health, defeated-state action restrictions and safe-spawn recovery.
- [x] Persistent server-owned enemy AI: aggro, chase, attack telegraphs/interruption, leash and return-to-spawn behavior; 250 ms PostgreSQL world worker, fresh-presence target selection, transactional enemy attacks and row-lock concurrency.
- [x] Authenticated zone-scoped enemy snapshots; Godot polls, interpolates enemy movement and displays health, behavior, telegraphs and respawn state. Player attacks use persisted live enemy positions.
- [x] Persisted sticky enemy target identity: maintain a valid active target between ticks and select the nearest eligible character when that target is lost.
- [x] Unique encounter-cycle IDs per enemy spawn life and durable per-character damage/action contribution ledger; autonomous incoming damage updates existing participants.
- [x] Threat scoring for enemy target selection (active sticky target first; otherwise current-cycle contribution score, then distance); PostgreSQL integration test passed in server Actions #472.
- [x] Initial server-side anti-leech policy and atomic ordinary-enemy settlement: 10% enemy-max-health damage threshold, two-minute active-damage window, contribution-weighted XP and one deterministic loot recipient; server Actions #497 passed.
- [ ] Party membership, healing/support contribution types and group/boss reward distribution.
- [x] Migration 0013: durable per-encounter settlement keys and per-character pending/claimed reward-claim records with uniqueness, ownership, quantity and state constraints; migration/registration/model/integration Actions #473–#476 passed.
- [x] Migration 0014 persists each participant's combat specialization for XP attribution; Actions #479–#481 passed.
- [x] Authenticated reward-claim API, account-scoped listing, idempotent redemption and full-stack pending behavior; settlement/claim integration suite passed in #497.
- [x] Authenticated HTTP API ownership/success-path tests (#498–#503) and Godot Rewards panel follow-up CI (#129–#132).
- [ ] Bounded pathfinding, party aggro, combat animations, enemy abilities, weapon-family expansion and server-derived hit/miss outcomes (see docs/COMBAT_PROGRESSION.md).
- [x] Reusable versioned boss phase/ability domain engine with validated definitions, threshold transitions, telegraphs, cooldowns, range interruption, encounter-cycle reset and deterministic damage intents (domain-only; server Actions verification pending).
- [ ] Durable boss encounter persistence groundwork: migration 0016 adds encounter-cycle-scoped state, versioned definition/phase, telegraph timestamps, JSONB cooldowns, ability cursor and terminal defeat state; PostgreSQL load/save helpers lock the parent enemy cycle and reject stale-cycle or stale-timestamp writes. Integration test added; GitHub Actions verification pending, and the world tick has not yet been wired to these helpers.
- [x] Initial server-owned weighted ordinary-enemy loot tables with cryptographically secure rolls and no-drop outcomes (server Actions #518 passed; later superseded in live enemy definitions by the multi-drop pipeline).
- [x] Versioned multi-drop loot-table domain model with guaranteed drops, independent weighted bonus rolls, no-drop intervals, quantity bands and aggregated stack validation.
- [x] Durable multi-item reward claims: migration/backfill, account-scoped listing, atomic all-item redemption, full-inventory retry, idempotency, API DTO and Godot Rewards panel.
- [x] Activate guaranteed-plus-rare-bonus loot tables for Frontier Rat and Wild Boar.
- [ ] Expand loot tables with rarity tiers, boss-exclusive materials, unique item instances, per-player policies, versioned tuning and economy safeguards (multi-item foundation now in place).
- [ ] Deeper crafting framework: richer recipes, tool/station/skill requirements, material properties, quality, experimentation and crafted-item progression.
- [ ] Recipe quality, tool requirements, advanced stations and broader profession progression. Advanced non-combat specialization trees remain deferred until core loops are stable.

## Milestone 4: Inventory, equipment and economy

- [x] Starter item catalog, stack-count inventory, server-side stack validation and consumable healing potions.
- [x] Persistent 100-slot bank with atomic deposit/withdraw operations.
- [x] Client inventory/bank search, category filters and quantity-based sorting.
- [x] Godot bag/bank panel with item selection and quantity-based stack transfers.
- [x] Persistent coin wallet, starter merchant catalog and atomic buy/sell transactions.
- [x] Equipment slots, equip/unequip transactions and derived attack/defense bonuses.
- [ ] Unique item instances, durability, binding, item modifiers and richer stack management.
- [ ] Currency ledger, player-to-player trading with escrow and anti-duplication guarantees (backend and Godot multi-item UI vertical slice implemented; concurrency, restart-recovery, abuse/load and real-device gates remain; see `docs/PLAYER_TRADE_DESIGN.md`).
- [ ] Regional supply, market discovery, currency sinks and player-run shops.

## Milestone 5: Social sandbox

- [x] Persistent account-owned local zone chat with polling, message cursors, newest-50 initial history, bounded client history and per-character cooldown.
- [ ] Direct messages, friends, groups and guilds.
- [ ] Player housing, permissions, storage and decoration.
- [ ] Settlement plots, construction, shared storage and governance.
- [ ] Player services, profession specialization and player-run institutions.
- [ ] Factions, reputation, world events and persistent world consequences.

## Milestone 6: Persistent multiplayer alpha

- [ ] Realtime replication, zone ownership and reconnect recovery.
- [ ] Durable world-state ownership, backup/restore drills and migration operations.
- [ ] Load testing, observability, capacity planning and operational runbooks.
- [ ] Distributed abuse prevention, economy monitoring, moderation and admin tools.
- [ ] Closed alpha deployment and feedback loop.

## Core game experience focus — active priority

Boss roster expansion and boss-specific polish are paused. Development now prioritizes the core player experience across the whole game.

- [x] Implement third-person movement acceleration/deceleration, air control, coyote time, jump buffering, facing/turn smoothing, fall-speed cap and mesh-only locomotion feedback; headless regression and strict Godot import/Web export passed in client Action #153.
- [ ] Interactive desktop/mobile playtesting, touch-control feel, camera collision/zoom and movement reconciliation tuning.
- [ ] Complete movement authority and reconciliation: server snapshots, bounded correction, reconnect recovery and smooth remote-player interpolation without client trust.
- [ ] Replace the flat placeholder floor with server-aligned, collision-safe terrain that supports elevation, slopes, traversable paths, resource placement and future biome expansion.
- [ ] Establish a real character presentation/animation foundation: idle, locomotion, start/stop, jump, fall, land, interact and combat states, with equipment-aware attachment points.
- [ ] Improve the core HUD and responsive UI: health/resources, skills, inventory, contextual interaction, feedback, scalable panels and mobile-safe controls.
- [ ] Build a reusable player ability/action framework with server-owned costs, cooldowns, range/target validation and client animation/feedback; keep it independent of boss mechanics.
- [ ] Improve world interaction feedback, navigation, environmental readability and performance before adding more content systems.

## Working rule

Each milestone is a tested, reviewable slice. A feature is not complete because a placeholder or interface exists: completion requires server rules, persistence where applicable, tests, and a usable client path. Keep production registration disabled until the documented security launch gates are met. **Never hide an error:** stop, diagnose, fix, regression-test and verify every discovered failure before moving to unrelated work; preserve the failure and correction in the handoff. See docs/DEVELOPMENT_RULES.md.

## Boss encounter vertical slice

- [x] Migration 0016 persists versioned boss phase, telegraph, cooldown, ability cursor and terminal defeat state under the active spawn encounter-cycle ID.
- [x] Reject stale-cycle and stale-timestamp boss state writes; validate complete telegraph timestamp invariants.
- [x] Advance boss phases and telegraphs inside the PostgreSQL enemy world tick; resolve ability-specific range checks and Defense/equipment mitigation transactionally with player health, contribution, boss state and AI state.
- [x] Interrupt active telegraphs when the target moves out of ability range or no eligible target remains.
- [x] Cover concurrent exactly-once damage, lethal player defeat, target/telegraph cleanup, respawn-cycle reset, public authenticated snapshots and concurrent reward-claim retries with PostgreSQL integration tests (server Actions #602 passed race-enabled integration tests and Vet).
- [x] Register Stoneback Colossus in the public enemy catalog at Verdant Frontier (32,32): 100 HP, 2 Defense, no ordinary attack, 500 encounter XP and 300-second respawn.
- [x] Assign the versioned Stoneback multi-item harvest table and verify defeat, durable settlement, claim listing, claim redemption, XP and idempotent retry.
- [x] Render the boss world model and consume authoritative boss health/phase/ability snapshots in the Godot HUD.
- [x] Keep the encounter HUD visible on authoritative defeat snapshots, show explicit reward-claim feedback, make active telegraphs communicate the ability warning and move-out-of-range instruction, surface short phase-transition cues, and size the HUD with viewport-relative margins for narrow screens.
- [x] Verify server Actions #599/#600, including race-enabled PostgreSQL integration tests and Vet; verify client Actions #137 for Godot import/script validation.
- [ ] Continue player-facing playtesting and polish: telegraph timing feedback, phase transition cues, target lifecycle, defeat/respawn presentation and accessibility at desktop/mobile aspect ratios.
- [ ] Add group encounters, party threat and support contribution before implementing group reward distribution.
- [ ] Expand boss roster only after the Stoneback encounter remains robust under repeated combat, respawn and reward-claim testing.

