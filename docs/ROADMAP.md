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

- [x] Ten data-driven skills, level 1–99 progression curve and XP cap.
- [x] Authenticated live skills panel with per-skill XP progress and next-level thresholds.
- [x] Tree, stone and copper gathering with range, skill, charge and respawn rules.
- [x] Workbench plank recipe and forge copper smelting/training-sword recipes.
- [x] Transactional crafting with ingredient consumption, output capacity and skill XP.
- [x] Basic PvE enemy definitions, server-calculated damage, cooldowns, retaliation, combat XP, loot and respawns.
- [x] Player health, defeated-state action restrictions and safe-spawn recovery.
- [ ] Enemy movement/AI, combat animations, abilities, target selection and deeper combat progression.
- [ ] Recipe quality, tool requirements, advanced stations and broader profession progression.

## Milestone 4: Inventory, equipment and economy

- [x] Starter item catalog, stack-count inventory, server-side stack validation and consumable healing potions.
- [x] Persistent 100-slot bank with atomic deposit/withdraw operations.
- [x] Client inventory/bank search, category filters and quantity-based sorting.
- [x] Godot bag/bank panel with item selection and quantity-based stack transfers.
- [x] Persistent coin wallet, starter merchant catalog and atomic buy/sell transactions.
- [x] Equipment slots, equip/unequip transactions and derived attack/defense bonuses.
- [ ] Unique item instances, durability, binding, item modifiers and richer stack management.
- [ ] Currency ledger, player-to-player trading with escrow and anti-duplication guarantees.
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
