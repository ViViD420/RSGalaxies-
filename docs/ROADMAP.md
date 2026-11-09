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
- [x] Transactional crafting with ingredient consumption, output capacity and skill XP.
- [x] Basic PvE enemy definitions, server-calculated damage, cooldowns, retaliation, specialization/parent/co-skill XP, starter loot and respawns.
- [x] Player health, defeated-state action restrictions and safe-spawn recovery.
- [x] Persistent server-owned enemy AI: aggro, chase, attack telegraphs/interruption, leash and return-to-spawn behavior; 250 ms PostgreSQL world worker, fresh-presence target selection, transactional enemy attacks and row-lock concurrency.
- [x] Authenticated zone-scoped enemy snapshots; Godot polls, interpolates enemy movement and displays health, behavior, telegraphs and respawn state. Player attacks use persisted live enemy positions.
- [x] Persisted sticky enemy target identity: maintain a valid active target between ticks and select the nearest eligible character when that target is lost.
- [x] Unique encounter-cycle IDs per enemy spawn life and durable per-character damage/action contribution ledger; autonomous incoming damage updates existing participants.
- [x] Threat scoring for enemy target selection (active sticky target first; otherwise current-cycle contribution score, then distance); PostgreSQL integration test passed in server Actions #472.
- [ ] Party membership and support-action contribution rules; initial damage-based anti-leech policy and atomic ordinary-enemy settlement are implemented, but need CI verification and group-policy expansion.
- [x] Migration 0013: durable per-encounter settlement keys and per-character pending/claimed reward-claim records with uniqueness, ownership, quantity and state constraints; migration/registration/model/integration Actions #473–#476 passed.
- [ ] Verify CI for migration 0014 specialization attribution, damage-based anti-leech policy, atomic enemy-defeat settlement, authenticated claim APIs and full-inventory/idempotent claim integration tests. API/client claim UX still needs integration.
- [ ] Bounded pathfinding, party aggro, combat animations, enemy abilities, weapon-family expansion and server-derived hit/miss outcomes (see docs/COMBAT_PROGRESSION.md).
- [ ] Boss encounter framework: multi-phase bosses, encounter state, group contribution/eligibility, telegraphs, persistence and reward integration. Build after ordinary enemy/combat foundations.
- [ ] Dedicated loot framework: server-owned weighted drop tables, guaranteed/conditional rewards, unique drops, contribution rules, anti-duplication and economy safeguards.
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

## Working rule

Each milestone is a tested, reviewable slice. A feature is not complete because a placeholder or interface exists: completion requires server rules, persistence where applicable, tests, and a usable client path. Keep production registration disabled until the documented security launch gates are met.
