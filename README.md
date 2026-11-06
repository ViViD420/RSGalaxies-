# RS Galaxies

An original persistent-world sandbox MMORPG combining long-term skill progression and item-driven gameplay with a highly social, player-shaped open world.

> **Project boundary:** RS Galaxies is completely separate. It does not share source code, assets, build outputs, databases or runtime dependencies with any other game project.

## Project status

**Foundation and progression prototype — in progress.** The Go server supports character creation, ten data-driven skills, deterministic XP-to-level progression and a development-only XP award route. Player data is in memory and resets when the process stops. PostgreSQL, authentication, secure sessions, validated gameplay actions, realtime multiplayer and the world client are not yet implemented.

## Design pillars

- **Skill-first progression:** actions improve relevant skills; progression is broad rather than class-locked.
- **Open-world sandbox:** exploration, resource discovery, travel, social play and player projects matter.
- **Meaningful crafting:** resources have properties, recipes have requirements and crafted items participate in the economy.
- **Player-shaped economy:** gathering, production, trade, shops and scarcity create interdependent roles.
- **Persistent community:** housing, guilds, settlements and shared projects give players reasons to cooperate.
- **Server authority:** the server owns game state and validates consequential actions.

## Repository layout

- server/ — Go authoritative game-service foundation.
- client/ — reserved for the Godot client prototype.
- .devcontainer/ — Codespaces / VS Code cloud workspace configuration.
- docs/ — architecture, roadmap and session handoff.
- .github/workflows/ — automated server checks.

## Work from a phone or browser

A development container is configured for GitHub Codespaces. On the repository page, open **Code**, choose **Codespaces**, and create a codespace on main. Setup installs Go dependencies and runs tests and vet. Codespaces availability and usage limits depend on the GitHub account.

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
- POST /api/v1/players/{id}/skills/{skill}/experience — development-only XP award.

Create a character with POST /api/v1/players and JSON body {"name":"Wayfarer"}. The XP curve is cumulative and deterministic, with levels 1–99 and a 200,000,000 XP cap. The development award route is for testing only; production progression must be granted through validated gathering, crafting, combat and other gameplay actions.

## Engineering rules

1. The server is authoritative; clients submit intent, never trusted state changes.
2. Game rules belong in domain packages, not HTTP handlers or UI code.
3. Persistent mutations will be transactionally committed before success is acknowledged.
4. Every progression system must have deterministic tests.
5. No secrets, proprietary assets or copied game code are committed.
6. Keep this repository and all project artifacts independent.

See docs/ARCHITECTURE.md, docs/ROADMAP.md and docs/SESSION_HANDOFF.md.
