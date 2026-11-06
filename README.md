# RS Galaxies

An original persistent-world sandbox MMORPG combining the readable, long-term skill progression and item-driven gameplay of classic MMORPGs with a highly social, player-shaped open world.

> **Project boundary:** RS Galaxies is a completely separate project. It does not share source code, assets, build outputs, databases, or runtime dependencies with any other game project.

## Project status

**Milestone 1 — independent foundation (in progress)**

The initial repository establishes the project boundaries, technical direction, a small authoritative world-service slice, tests, and a continuation document. The current server uses in-memory storage for development only; durable persistence and multiplayer transport are later milestones.

## Design pillars

- **Skill-first progression:** actions improve relevant skills; progression is broad rather than locked to a single class.
- **Open-world sandbox:** exploration, resource discovery, travel, social play, and player projects matter.
- **Meaningful crafting:** resources have properties, recipes have requirements, and crafted items participate in the economy.
- **Player-shaped economy:** gathering, production, trade, shops, and scarcity create interdependent roles.
- **Persistent community:** housing, guilds, settlements, and shared projects give players reasons to cooperate.
- **Server authority:** the server owns game state and validates all consequential actions.

## Repository layout

- `server/` — Go authoritative game-service foundation.
- `client/` — reserved for the game client; engine selection is intentionally not locked yet.
- `docs/` — architecture, roadmap, and session handoff.
- `.github/workflows/` — automated server checks.

## Run the development server

Requires Go 1.23 or newer.

```sh
cd server
go test ./...
go run ./cmd/worldd
```

The server listens on `http://localhost:8080` by default. Set `RS_GALAXIES_ADDR` to change the listen address.

### Initial API

- `GET /healthz` — process health.
- `GET /api/v1/players` — list development players.
- `POST /api/v1/players` — create a player.
- `GET /api/v1/players/{id}` — retrieve a player.

Example:

```sh
curl -X POST http://localhost:8080/api/v1/players \
  -H 'Content-Type: application/json' \
  -d '{"name":"Wayfarer"}'
```

## Engineering rules

1. The server is authoritative; clients submit intent, never trusted state changes.
2. Game rules belong in domain packages, not HTTP handlers or UI code.
3. Persistent mutations will be transactionally committed before success is acknowledged.
4. Every progression system must have deterministic tests.
5. No secrets, proprietary assets, or copied game code are committed.
6. Keep this repository and all project artifacts independent.

See [Architecture](docs/ARCHITECTURE.md), [Roadmap](docs/ROADMAP.md), and [Session handoff](docs/SESSION_HANDOFF.md).