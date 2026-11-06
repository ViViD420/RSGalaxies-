# RS Galaxies — Architecture

## Goals

Build an original, server-authoritative sandbox MMORPG with independently testable game rules, a replaceable client, and persistence boundaries that can evolve without rewriting gameplay systems.

## Initial technology direction

| Area | Initial direction | Rationale |
|---|---|---|
| Game rules / world service | Go | Fast iteration, strong concurrency and simple deployment |
| Client | Godot 4.3 prototype (provisional) | Open-source 3D engine; validating third-person movement, touch controls and Web export before final commitment |
| Durable data | PostgreSQL 16 (implemented for development) | Transactions and relational integrity for characters, skills, inventory and gathering-node state |
| Cache / ephemeral coordination | Redis (planned, not required yet) | Later use for presence, rate limits and ephemeral coordination |
| Realtime transport | WebSocket or UDP-based protocol (planned) | Choose after latency and replication prototypes |
| CI | GitHub Actions | Repeatable tests on every change |

The Go service is a development API, not yet a production MMO server. It selects a transactional PostgreSQL repository when RS_GALAXIES_DATABASE_URL is configured and otherwise uses the replaceable in-memory repository. PostgreSQL migrations and integration tests are included; authentication and production operational hardening remain future work.

## Runtime boundaries

1. **Client:** rendering, input, local presentation, prediction where appropriate. The client is untrusted.
2. **Gateway / session layer (future):** authentication, connection lifecycle, rate limits and routing.
3. **Authoritative world simulation:** validates movement and actions, owns live entity state, applies combat and skill rules.
4. **Domain services:** character, skills, inventory, crafting, gathering, trade, housing and social systems.
5. **Persistence:** PostgreSQL transactions for durable state; append-only audit/event records for economically sensitive operations.
6. **Operations:** structured logs, metrics, tracing, backups, migrations and deployment automation.

## Domain design principles

- Use stable IDs rather than display names as identity.
- Represent skill XP as integer values; derive levels from a single tested progression curve.
- Treat inventory and currency as owned resources. Transfers must be atomic and idempotent.
- Model item definitions separately from item instances.
- Model resources with explicit properties so crafting can use quality, purity, rarity and region without hard-coding every resource.
- Keep recipes data-driven and validate all ingredients and requirements on the server.
- Use server timestamps and sequence numbers for ordering consequential actions.
- Add version fields to persisted aggregates to support optimistic concurrency.

## Initial service contract

The current development API exposes:

- `GET /healthz`
- `GET /api/v1/skills`
- `GET /api/v1/world/zones`
- `GET /api/v1/world/nodes`
- `GET /api/v1/items`
- `GET /api/v1/players`, `POST /api/v1/players`, `GET /api/v1/players/{id}`
- `GET /api/v1/players/{id}/skills`
- `GET /api/v1/crafting/recipes` and `GET /api/v1/crafting/stations`
- Development-only `POST` routes for XP awards, movement, gathering and crafting.

The first world slice contains one bounded zone, a spawn point, server-time-validated horizontal movement, two starter resource nodes, stack-count inventory and XP rewards. Resource gathering updates player inventory, skill XP and node charges inside one transaction (with a shared node row lock in PostgreSQL); resource nodes have limited charges and timed respawn. A first workbench recipe consumes two logs to create a wooden plank and award Crafting XP. Crafting checks station range, skill level, ingredient counts and output stack capacity before committing the inventory/XP change. These systems have both in-memory and PostgreSQL repository implementations for development. PostgreSQL persists character skills, inventory, positions, movement timestamps and gathering-node charge/respawn state. All player-specific routes and development state-changing shortcuts are not registered when RS_GALAXIES_ENV=production; public player APIs will return only after authentication and ownership checks exist.

## Data ownership and transaction strategy

Implemented PostgreSQL aggregates:
- Character identity, skills, inventory stacks and world position
- Gathering-node charges and respawn timestamps

Planned PostgreSQL aggregates:
- Account identity and authenticated sessions
- Item instances, equipment and bank storage
- Currency ledger and trade offers/completions
- Housing and settlement ownership

A trade must lock or version-check both parties' affected inventories, validate capacity and ownership, write both sides and the currency ledger in one transaction, then publish an event after commit. Never acknowledge a trade based only on a client message.

## Security baseline

- Never trust client-supplied position, XP, inventory, item ownership, currency or combat results.
- Validate request size, JSON shape, names and identifiers.
- Use authentication before exposing private character data.
- Rate-limit state-changing actions.
- Avoid logging credentials, session tokens or private account data.
- Use parameterized SQL and explicit transaction boundaries.
