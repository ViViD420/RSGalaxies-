# RS Galaxies — Architecture

## Product and engineering goals

Build an original, server-authoritative sandbox MMORPG with long-form skills, gathering, crafting, economy, combat and social systems. Keep game rules independently testable, the client replaceable and durable state behind transactional repositories. All code and assets must be original.

## Technology

| Area | Current implementation | Intended evolution |
|---|---|---|
| World service | Go HTTP API and domain services | Authoritative zone simulation and realtime gateway |
| Client | Godot 4.3 third-person prototype | Production-ready desktop/mobile client |
| Durable data | PostgreSQL 16, versioned SQL migrations | Partitioning, backups, restore drills and capacity planning |
| Ephemeral coordination | PostgreSQL leased presence prototype | Redis or another dedicated presence/coordination service when scale requires it |
| Realtime transport | HTTP polling for movement/presence/chat | WebSocket or UDP-based transport after latency/replication prototypes |
| CI | GitHub Actions | Expanded load, security and deployment pipelines |
| Hosted development | GitHub Codespaces devcontainer | Reproducible team workspaces |

The Go service selects PostgreSQL repositories when `RS_GALAXIES_DATABASE_URL` is configured; otherwise it uses an in-memory development repository. Account-backed gameplay requires PostgreSQL. Production startup fails closed without the database URL, applies migrations before serving and exposes separate liveness/readiness probes.

## Runtime boundaries

1. **Godot client:** input, rendering, local presentation and short-term smoothing. Treat all client state and requests as untrusted.
2. **HTTP/session boundary:** bearer-token authentication, account ownership checks, request limits, throttling and route-level security.
3. **Game/domain layer:** skill progression, movement validation, gathering, crafting, bank, merchant, equipment, consumables and PvE rules. Domain logic is independent of HTTP and SQL.
4. **Repository/persistence layer:** PostgreSQL transactions and row locks protect account-owned character state and shared world state. The in-memory repository supports isolated development tests.
5. **Presence/social layer:** PostgreSQL leased zone presence and persistent local chat currently use bounded polling. This is a prototype, not a realtime fan-out service.
6. **Operations:** migrations, CI, liveness/readiness, deployment configuration, monitoring, backups and incident procedures.

## Implemented vertical slices

### Identity and ownership

- Account registration/login/logout and opaque bearer sessions.
- PBKDF2-HMAC-SHA-256 password hashes with per-account random salts; random session tokens are stored as SHA-256 digests.
- Account-owned character CRUD and ownership predicates in character operations.
- Production registration disabled by default. Email verification, recovery, MFA, distributed edge rate limiting and admin account operations remain launch gates.

### World and progression

- One bounded starter zone with a spawn point and server-time movement validation.
- Eighteen active data-driven ordinary skills: two combat parent skills, six combat specializations, two independent combat co-skills, four gathering skills and four production skills. All use level 1–99 progression and a 200,000,000 XP cap. Skill categories, parent relationships and Jedi mastery eligibility are server-internal and omitted from public catalogs. Accuracy is a derived server-side combat outcome, not a trainable skill. Legacy Combat/Tactics progress keys are retained in old character records but are inactive and omitted from the client skill panel. See docs/SKILL_ROSTER.md and docs/COMBAT_PROGRESSION.md. A future hidden Jedi unlock is specified in docs/JEDI_PROGRESSION_DESIGN.md; it is not implemented or exposed yet.
- Tree/stone/copper gathering, node charges and respawn, workbench/forge recipes and transactional skill XP awards.
- Server-authoritative PvE with enemy health/respawn state, cooldowns, retaliation, skill-attributed XP (combat specialization, parent and co-skills), starter loot and safe-spawn recovery. Per-spawn AI state persists behavior, position, sticky target identity, telegraph deadlines and cooldowns. A PostgreSQL-backed 250 ms world worker selects fresh same-zone presence, advances aggro/chase/telegraph/leash/return behavior and commits telegraphed enemy damage plus player health/Defense XP atomically. Authenticated zone snapshots feed Godot health/behavior labels and interpolated enemy movement; player attacks validate against persisted live enemy positions.

### Inventory and economy

- Stack-based inventory, 100-slot bank, coin wallet and starter merchant.
- Atomic bank transfers, merchant buy/sell and equipment equip/unequip. Item definitions can declare server-enforced skill and level requirements; the starter sword requires Blades and armor requires Defense. Failed equip attempts leave inventory/equipment unchanged, and already-equipped items can always be unequipped if a skill later changes.
- Consumable healing items and derived equipment combat bonuses. Profession/class-specific gear can expand this generic requirement model later without hardcoding restrictions into the client.
- Item definitions are separate from player stack maps. Unique item instances, durability, binding and modifiers are future work.
- Player-to-player trading must use atomic multi-party transactions, escrow and an auditable currency/item ledger before release.

### Social and multiplayer prototype

- Persistent local-zone chat with server-side cooldown, cursor pagination, newest-50 initial history and bounded client history.
- Leased zone presence: account-owned heartbeat, 20-second lease, stale cleanup, capped zone roster and privacy-limited remote-character snapshots.
- Godot polls presence and renders placeholder remote avatars/nameplates. There is no authoritative continuous movement replication, reconnect recovery, zone handoff or realtime transport yet.

## Persistence and transaction invariants

- Never trust client-supplied XP, item counts, currency, position, combat results, account IDs or ownership.
- Character operations resolve the character using the authenticated account and lock affected rows in a transaction.
- Shared gathering and enemy state is stored in PostgreSQL and locked to prevent duplicate resource consumption or combat rewards across server instances. Enemy world ticks and player attacks both lock the player before the enemy; a completed AI telegraph, enemy state, player health and Defense XP commit in one transaction. Multiple world workers serialize through PostgreSQL row locks and persisted enemy cooldowns.
- Inventory and currency changes validate all inputs and capacity before commit.
- Any future trade must lock both parties in a stable order, reserve assets in escrow, validate acceptance and capacity, atomically deliver both sides, and write an append-only ledger in the same transaction.
- Presence is ephemeral and lease-based; it must not be treated as proof of an active socket or as durable player state.
- Use server timestamps and monotonically increasing database IDs/cursors for ordering consequential events.

## Security and production boundaries

- Parameterized SQL, strict JSON decoding, request body limits, account ownership checks and generic external error responses.
- Production player mutation routes are account-scoped; legacy unscoped development routes are not registered in production.
- Production CORS requires exact origins. Authentication throttling is process-local and bounded; trusted edge rate limiting is still required.
- `/healthz` reports process liveness independent of PostgreSQL. `/readyz` performs a bounded database dependency check.
- Never log passwords, bearer tokens, authorization headers or private account data.
- Do not enable public registration until recovery, verification, distributed abuse controls, monitoring and administrative account-disable/recovery workflows are complete.

## Next architectural decisions

1. Improve enemy target stickiness, threat selection and bounded pathfinding before party aggro and multi-phase encounters.
2. Replace polling with a measured realtime transport and authoritative zone snapshots when latency/load measurements justify it.
3. Define online/session ownership, reconnect and multi-device character rules.
4. Add an append-only economy ledger before expanding player-to-player exchange.
5. Introduce item instances only when durability, binding or per-item modifiers require them.
6. Split presence/cache and zone simulation only when load tests demonstrate PostgreSQL/HTTP polling is insufficient.
