# RS Galaxies — Roadmap

## Milestone 1: Independent foundation
- [x] Establish standalone repository and project identity.
- [x] Document design pillars and system boundaries.
- [x] Add a small server API and deterministic tests.
- [ ] Select client engine after a short movement / camera / world-size prototype.
- [ ] Add PostgreSQL migrations and local development environment.

**Exit criteria:** a clean checkout can run tests and start the development service; project boundaries and client/server contracts are documented.

## Milestone 2: Character and world slice
- [ ] Character creation, identity and persistence.
- [ ] World zones, coordinates, movement validation and spawn points.
- [ ] Interactable objects and NPC definitions.
- [ ] Session lifecycle and basic presence.
- [ ] Client connects to a local server and displays authoritative character state.

## Milestone 3: Skills and progression
- [ ] Data-driven skill definitions and XP curves.
- [ ] Gathering nodes and depletion / respawn.
- [ ] Crafting recipes, requirements and output quality.
- [ ] Combat attributes, abilities, damage and death / recovery loop.
- [ ] Server-side action validation and progression tests.

## Milestone 4: Inventory and economy
- [ ] Item definitions, item instances, stack rules and equipment.
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

Each milestone is delivered as a tested, reviewable slice. Do not mark a feature complete because a placeholder or interface exists; completion requires server rules, persistence where applicable, tests, and a usable client path.
