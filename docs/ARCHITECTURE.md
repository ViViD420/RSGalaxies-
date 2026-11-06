# RS Galaxies — Architecture

## Goals

Build an original, server-authoritative sandbox MMORPG with independently testable game rules, a replaceable client, and persistence boundaries that can evolve without rewriting gameplay systems.

## Initial technology direction

| Area | Initial direction | Rationale |
|---|---|---|
| Game rules / world service | Go | Fast iteration, strong concurrency and simple deployment |
| Client | Godot 4.3 prototype (provisional) | Open-source 3D engine; validating third-person movement, touch controls and Web export before final commitment |
| Durable data | PostgreSQL (planned) | Transactions and relational integrity for characters, inventories, ownership and economy |
| Cache / ephemeral coordination | Redis (planned, not required yet) | Later use for presence, rate limits and ephemeral coordination |
| Realtime transport | WebSocket or UDP-based protocol (planned) | Choose after latency and replication prototypes |
| CI | GitHub Actions | Repeatable tests on every change |

The Go service in this milestone is a development API, not yet a production MMO server. Its in-memory repository is deliberately replaceable.

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
- Development-only `POST` routes for XP awards, movement and gathering.

The first world slice contains one bounded zone, a spawn point, server-time-validated horizontal movement, two starter resource nodes, stack-count inventory and XP rewards. Resource gathering updates player inventory, skill XP and node charges under one repository lock; resource nodes have limited charges and timed respawn. A first workbench recipe consumes two logs to create a wooden plank and award Crafting XP. Crafting checks station range, skill level, ingredient counts and output stack capacity before committing the inventory/XP change. These systems are in-memory development implementations, not durable multiplayer services. Development state-changing shortcuts are not registered when RS_GALAXIES_ENV=production.

## Data ownership and transaction strategy

Planned PostgreSQL aggregates:
- Account and character identity
- Character skills and progression
- Inventory and item instances
- Bank storage
- Currency ledger
- Trade offers and completed trades
- Housing / settlement ownership

A trade must lock or version-check both parties' affected inventories, validate capacity and ownership, write both sides and the currency ledger in one transaction, then publish an event after commit. Never acknowledge a trade based only on a client message.

## Security baseline

- Never trust client-supplied position, XP, inventory, item ownership, currency or combat results.
- Validate request size, JSON shape, names and identifiers.
- Use authentication before exposing private character data.
- Rate-limit state-changing actions.
- Avoid logging credentials, session tokens or private account data.
- Use parameterized SQL and explicit transaction boundaries.
