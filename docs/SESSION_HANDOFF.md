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
- Defensive copying of skill and inventory maps, plus stable player listing.
- One bounded world zone, spawn point and server-time-based movement speed/bounds validation.
- Starter item definitions and stack-count inventory.
- Two gathering nodes (tree and stone) with interaction range, three charges, timed respawn and atomic inventory/XP rewards.
- One frontier workbench and a basic plank recipe: consume two softwood logs, create one wooden plank and award Crafting XP. Crafting checks station range, skill level, ingredient counts and output stack capacity atomically.
- HTTP routes expose skills, world zones, gathering nodes, items, players and player skills. Development-only XP, movement, gathering and crafting routes are disabled when RS_GALAXIES_ENV=production.
- Domain, API, movement, gathering and repository concurrency tests.
- Go CI runs tests with the race detector and go vet.

### Client
- Godot 4.3 provisional client prototype.
- Procedurally generated test floor, placeholder character, third-person orbit camera, keyboard movement, touchscreen D-pad, USE and jump buttons.
- Startup character creation, periodic movement synchronization, and tree/stone gathering and workbench crafting interactions.
- Inventory HUD shows collected starter resources.
- Headless Godot validation and Web export workflow.
- This is a movement/gathering proof of concept, not a fully server-authoritative or production client.

### Development environment
- .devcontainer/ configures a GitHub Codespaces Go/Godot/PostgreSQL workspace using Docker Compose.
- Docker Compose runs PostgreSQL 16 with separate development and integration-test databases. A named volume preserves the development database across Codespaces stops/restarts.
- setup.sh installs OS dependencies, Godot 4.3 and export templates, then runs race-enabled Go tests (including PostgreSQL integration), vet, a headless Godot import and a Web export.
- start.sh launches the Go API and exported browser prototype automatically on Codespaces startup.
- Ports 8000 and 8080 are forwarded; the Web client derives the API origin from the Codespaces hostname, and the development API enables CORS.
- Repository work can be done in GitHub/Codespaces through a browser; no local PC installation is required.

## Limits and decisions

- Go is the server language. Godot is provisional pending real-phone validation of camera, movement and Web export.
- PostgreSQL persistence is implemented and used in Codespaces. Standalone local runs fall back to in-memory storage unless RS_GALAXIES_DATABASE_URL is configured.
- XP award, movement, gathering and crafting shortcuts are development-only. They must not be exposed publicly.
- Authentication, secure sessions, realtime replication, production operations, NPCs, combat, additional recipes, player trading and production hardening remain incomplete.
- Code and tests are committed. GitHub Actions execution results have not yet been inspected in this session; do not claim CI passes until verified.

## Next engineering tasks

1. Inspect GitHub Actions results and fix any Go/Godot/PostgreSQL integration failures.
2. Validate the auto-launched browser build, gathering and crafting loops on a real phone; refine touch controls and camera feel.
3. Add account authentication, secure sessions and character ownership checks; retire development-only state-changing routes before public deployment.
4. Add item instances, equipment, bank storage, currency and atomic player trading.
5. Expand the playable loop with more recipes, tools, resource tiers, NPCs and combat.

## Continuation instruction

Always inspect this file and current repository state first. Work only in ViViD420/RSGalaxies-. Update this handoff after each meaningful milestone. Never assume GalaxyMMO code or decisions apply here.
