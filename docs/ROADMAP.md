# RS Galaxies — Roadmap

## Milestone 1: Independent foundation
- [x] Establish standalone repository and project identity.
- [x] Document design pillars and system boundaries.
- [x] Add server API and deterministic tests.
- [ ] Validate client engine through movement, camera and world-size prototype.
- [ ] Add PostgreSQL migrations and local development environment.
- [x] Add hosted Codespaces development-container configuration.

## Milestone 2: Character and world slice
- [x] Basic character identity and in-memory development storage.
- [ ] Durable character persistence and migrations.
- [x] First world zone, coordinates, spawn point and server-validated movement prototype.
- [ ] Interactable objects and NPC definitions.
- [ ] Session lifecycle and basic presence.
- [ ] Client displays fully authoritative character state (current movement sync is development-only).

## Milestone 3: Skills and progression
- [x] Data-driven starting skill definitions and XP curve prototype.
- [x] Starter gathering nodes with charges, depletion and timed respawn.
- [ ] Crafting recipes, requirements and output quality.
- [ ] Combat attributes, abilities, damage and death / recovery loop.
- [x] Award XP through validated starter gathering actions (development prototype).\n- [ ] Full production gameplay-action validation for gathering, crafting and combat.
- [x] Deterministic progression boundary and repository concurrency tests.

## Milestone 4: Inventory and economy
- [x] Starter item definitions and stack-count inventory.\n- [ ] Item instances, equipment, ownership rules and full stack management.
- [ ] Bank storage and item ownership.
- [ ] Atomic player-to-player trades.
- [ ] Currency ledger and NPC/player shops.
- [ ] Resource properties, regional supply and market discovery.

## Milestone 5: Social sandbox
- [ ] Friends, groups, guilds and communication.
- [ ] Player housing and permissions.
- [ ] Settlement plots, construction and shared storage.
- [ ] Player services and profession specialization.
- [ ] Factions, reputation and world events.

## Milestone 6: Persistent multiplayer alpha
- [ ] Realtime replication and reconnect recovery.
- [ ] Durable world state, backups and migrations.
- [ ] Load testing, observability and operational runbooks.
- [ ] Abuse prevention, economy monitoring and moderation tools.
- [ ] Closed alpha deployment and feedback loop.

## Working rule

Each milestone is a tested, reviewable slice. Do not mark a feature complete because a placeholder or interface exists; completion requires server rules, persistence where applicable, tests, and a usable client path.
