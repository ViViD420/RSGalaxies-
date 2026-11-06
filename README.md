# RS Galaxies

An original persistent-world sandbox MMORPG combining long-term skill progression and item-driven gameplay with a highly social, player-shaped open world.

> **Project boundary:** RS Galaxies is completely separate. It does not share source code, assets, build outputs, databases or runtime dependencies with any other game project.

## Project status

**Foundation and progression prototype — in progress.** The Go server supports character creation, ten data-driven skills, deterministic XP-to-level progression, a starter world zone, server-validated movement, two resource nodes, stack-count inventory, a gathering loop and a first workbench crafting recipe. The Godot 4.3 client has a 3D movement prototype, touch controls and browser export. Player and world data are in memory and reset when the process stops. PostgreSQL, authentication, secure sessions, combat, broader crafting, realtime multiplayer and production deployment are not yet implemented.

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

A development container is configured for GitHub Codespaces. On the repository page, open **Code**, choose **Codespaces**, and create a codespace on main. Setup installs Go and Godot, runs server checks and exports the Web client. When the workspace starts, it launches the Go API and browser prototype on forwarded ports 8080 and 8000. Codespaces availability and usage limits depend on the GitHub account.

## Run the development server

Requires Go 1.23 or newer.

    cd server
    go test ./...
    go vet ./...
    go run ./cmd/worldd

The server listens on http://localhost:8080 by default. Set RS_GALAXIES_ADDR to change the listen address.

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

Create a character with POST /api/v1/players and JSON body {"name":"Wayfarer"}. The XP curve is cumulative and deterministic, with levels 1–99 and a 200,000,000 XP cap. Development-only XP, movement, gathering and crafting shortcuts are disabled when RS_GALAXIES_ENV=production; production use still requires authentication and durable persistence.

## Engineering rules

1. The server is authoritative; clients submit intent, never trusted state changes.
2. Game rules belong in domain packages, not HTTP handlers or UI code.
3. Persistent mutations will be transactionally committed before success is acknowledged.
4. Every progression system must have deterministic tests.
5. No secrets, proprietary assets or copied game code are committed.
6. Keep this repository and all project artifacts independent.

See docs/ARCHITECTURE.md, docs/ROADMAP.md and docs/SESSION_HANDOFF.md.
