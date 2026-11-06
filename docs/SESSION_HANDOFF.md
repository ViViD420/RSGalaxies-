# RS Galaxies — Session Handoff

Last updated: 2026-11-06

## Project identity

- Name: **RS Galaxies**
- Repository: https://github.com/ViViD420/RSGalaxies-
- Default branch: main
- Standalone project. Never reuse or modify the separate GalaxyMMO repository.

## Product direction

An original sandbox MMORPG combining long-form skill progression and item gameplay with a social, open-ended world built around professions, crafting, trade, housing and settlements. Use original code and assets only.

## Current implementation

### Server
- Go development server with player creation, retrieval and listing.
- Ten stable skills: combat, cooking, crafting, farming, foraging, mining, ranged, scavenging, smithing and woodcutting.
- Deterministic cumulative XP curve, levels 1–99 and 200,000,000 XP cap.
- Player creation initializes all skills at level one.
- Domain XP validation, level-up detection and atomic in-memory awards.
- Defensive copying of player skill maps and stable player listing.
- HTTP routes for skill catalog, world zones, player skills and development-only XP awards/movement. Development shortcuts are disabled when RS_GALAXIES_ENV=production.
- World zone bounds, default spawn point and server-time-based movement speed/bounds validation. Repository movement updates are atomic.
- Starter item definitions, per-character stack inventory and two gathering nodes (tree/stone). Gathering atomically validates range, awards skill XP and an item, depletes node charges and respawns nodes after a cooldown.
- HTTP endpoints expose item/node definitions and development-only gathering actions; development shortcuts are disabled when RS_GALAXIES_ENV=production.
- Domain, API and repository concurrency tests.
- Go CI runs tests with the race detector and go vet.

### Client
- Godot 4.3 prototype in client/.
- Procedurally generated test floor, placeholder character, third-person orbit camera, keyboard movement and touchscreen D-pad/jump controls; touch drag or mouse controls the camera.
- Startup character creation and HTTP health check against the Go service; client position is periodically sent to the server and checked against movement rules.
- Touchscreen D-pad/jump buttons and touch-drag camera controls.
- Headless Godot validation and Web export workflow.
- This is a movement/API proof of concept, not a final client or a network-authoritative character controller.

### Development environment
- .devcontainer/ configures a GitHub Codespaces Go workspace.
- .devcontainer/setup.sh installs Godot 4.3 and export templates, then runs Go tests/vet, a headless Godot import and a Web export. .devcontainer/start.sh launches the Go API and exported browser prototype automatically on Codespaces startup. Ports 8000 and 8080 are forwarded; the client derives the API port from the Codespaces hostname and the development API enables CORS.
- User can edit files in GitHub/Codespaces through a browser; no local PC installation is required for repository work.

## Limits and decisions

- Go is the server language. Godot is the provisional client engine, pending camera, movement and web/mobile export validation.
- Storage is in memory; data is lost when the server stops.
- The XP-award HTTP endpoint is development-only and must not be exposed publicly. Production XP must come from validated game actions.
- PostgreSQL, authentication, secure sessions, realtime replication, production persistence, NPCs, combat, crafting, player trading and production inventory persistence are not implemented. The current gathering loop is a small in-memory tutorial slice. Movement synchronization is a development prototype, not authenticated production multiplayer.
- Code and tests are committed, but GitHub Actions execution results have not yet been inspected in this session. Do not claim CI passes until verified.

## Next engineering tasks

1. Inspect GitHub Actions results and fix any Go/Godot build failures.
2. Validate the auto-launched browser build on a real phone and refine touch controls and camera feel.
3. Implement world zones, server-owned coordinates, spawn points and validated movement.
4. Add PostgreSQL schema/migrations and transactional player persistence.
5. Replace development XP awards with validated gathering, crafting and combat actions.

## Continuation instruction

Always inspect this file and current repository state first. Work only in ViViD420/RSGalaxies-. Update this handoff after each meaningful milestone. Never assume GalaxyMMO code or decisions apply here.
