# RS Galaxies

An original persistent-world sandbox MMORPG combining long-term skill progression and item-driven gameplay with a highly social, player-shaped open world.

> **Project boundary:** RS Galaxies is completely separate. It does not share source code, assets, build outputs, databases or runtime dependencies with any other game project.

## Project status

**Core gameplay prototype — in progress; not production-ready.** The Go server now includes account registration/login and secure sessions, account-owned characters, 18 active core skills with parent/specialization combat progression and level 1–99 advancement, a starter world, validated movement, gathering, crafting, skill-attributed PvE combat, inventory/bank, merchant economy, equipment, local chat, leased zone presence and transaction-backed player trading behind a production feature gate. PostgreSQL persists accounts and gameplay state. The Godot 4.3 client includes a third-person world, touch/keyboard controls, account and character UI, gathering/crafting, combat, inventory/economy panels, chat, presence and trade UI. Production launch still requires account recovery/verification, distributed abuse controls, operational monitoring, trade abuse/load review and fuller realtime replication. Jedi remains a documented hidden future progression path and is not implemented or exposed.

## Design pillars

- **Skill-first progression:** actions improve relevant skills; progression is broad rather than class-locked.
- **Open-world sandbox:** exploration, resource discovery, travel, social play and player projects matter.
- **Meaningful crafting:** resources have properties, recipes have requirements and crafted items participate in the economy.
- **Player-shaped economy:** gathering, production, trade, shops and scarcity create interdependent roles.
- **Persistent community:** housing, guilds, settlements and shared projects give players reasons to cooperate.
- **Server authority:** the server owns game state and validates consequential actions.

## Repository layout

- server/ — Go authoritative game-service foundation.
- client/ — Godot 4.3 3D prototype with touch controls and Web export.
- .devcontainer/ — Codespaces / VS Code cloud workspace configuration.
- docs/ — architecture, roadmap and session handoff.
- .github/workflows/ — automated server checks.

## Work from a phone or browser

A development container is configured for GitHub Codespaces. On the repository page, open **Code**, choose **Codespaces**, and create a codespace on main. Setup starts a PostgreSQL service, installs Go and Godot, runs race-enabled server tests (including database integration) and exports the Web client. When the workspace starts, it launches the Go API and browser prototype on forwarded ports 8080 and 8000. Codespaces availability and usage limits depend on the GitHub account.

## Run the development server

Requires Go 1.23 or newer.

    cd server
    go test ./...
    go vet ./...
    go run ./cmd/worldd

The server listens on http://localhost:8080 by default. Without RS_GALAXIES_DATABASE_URL it uses in-memory development storage; set that variable to a PostgreSQL DSN to enable durable storage. Set RS_GALAXIES_ADDR to change the listen address. The Codespaces workspace supplies PostgreSQL automatically; its checked-in credentials are development-only and must never be reused in production.

## API

- GET /healthz — process health.
- GET /api/v1/skills — skill definitions and progression caps.
- GET /api/v1/players — list development players.
- POST /api/v1/players — create a player.
- GET /api/v1/players/{id} — retrieve a player.
- GET /api/v1/players/{id}/skills — retrieve skill progress.
- GET /api/v1/world/zones — zone bounds and spawn data.
- GET /api/v1/world/nodes — resource node definitions.
- GET /api/v1/items — starter item definitions.
- GET /api/v1/crafting/recipes and GET /api/v1/crafting/stations — crafting data.
- POST /api/v1/players/{id}/skills/{skill}/experience — development-only XP award.
- POST /api/v1/players/{id}/move — development movement request with server speed/bounds validation.
- POST /api/v1/players/{id}/gather/{nodeID} — development gathering action that updates inventory and XP.
- POST /api/v1/players/{id}/craft/{recipeID} — development crafting action that consumes ingredients and creates outputs.

Create a character with POST /api/v1/players and JSON body {"name":"Wayfarer"}. The XP curve is cumulative and deterministic, with levels 1–99 and a 200,000,000 XP cap. Legacy unscoped player-ID and development mutation routes are disabled in production. Authenticated character and gameplay APIs enforce account ownership. Production registration and player trading remain disabled unless their explicit feature flags are enabled; see docs/DEPLOYMENT.md and docs/ACCOUNT_AUTH_DESIGN.md.

## Engineering rules

1. The server is authoritative; clients submit intent, never trusted state changes.
2. Game rules belong in domain packages, not HTTP handlers or UI code.
3. Persistent mutations will be transactionally committed before success is acknowledged.
4. Every progression system must have deterministic tests.
5. No secrets, proprietary assets or copied game code are committed.
6. Keep this repository and all project artifacts independent.

See docs/ARCHITECTURE.md, docs/ROADMAP.md, docs/SESSION_HANDOFF.md, docs/SKILL_ROSTER.md, docs/COMBAT_PROGRESSION.md and docs/JEDI_PROGRESSION_DESIGN.md.


## API and security design

- [HTTP API reference](docs/API.md)
- [Account authentication and session design](docs/ACCOUNT_AUTH_DESIGN.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Development roadmap](docs/ROADMAP.md)
- [Core skill roster](docs/SKILL_ROSTER.md)
- [Combat progression and derived accuracy](docs/COMBAT_PROGRESSION.md)
- [Future hidden Jedi progression design](docs/JEDI_PROGRESSION_DESIGN.md)

Copy `.env.example` for local configuration and adjust the PostgreSQL connection string for your environment.
