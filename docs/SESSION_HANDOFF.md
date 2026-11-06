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
- HTTP routes expose public skills, world zones, gathering nodes, items and crafting definitions. Player-specific routes and development-only XP, movement, gathering and crafting routes are all disabled when RS_GALAXIES_ENV=production until authentication and character ownership are implemented.
- Domain, API, movement, gathering, crafting, PostgreSQL persistence and repository concurrency tests.
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
- All player-specific APIs and XP/movement/gathering/crafting shortcuts are disabled in production until authentication and character ownership are implemented.
- Production startup fails closed if RS_GALAXIES_DATABASE_URL is missing; PostgreSQL connection and migrations must succeed before the HTTP service starts.
- Authentication, secure sessions, realtime replication, production operations, NPCs, combat, additional recipes, player trading and production hardening remain incomplete.
- Code and PostgreSQL integration tests are committed. GitHub Actions run #77 (commit 6e9ca12) failed in the PostgreSQL integration test. The movement timestamp assertion compared a nanosecond-precision Go time with PostgreSQL TIMESTAMPTZ's microsecond precision. The test now normalizes its timestamp to microseconds before asserting persistence; the corrective commit's CI result must be checked before treating this as resolved. Runs #73–#76 passed.

## Next engineering tasks

1. Verify the movement timestamp precision fix in GitHub Actions; investigate any remaining Go/PostgreSQL failures.
2. Validate the auto-launched browser build, gathering and crafting loops on a real phone; refine touch controls and camera feel.
3. Design and implement account authentication, secure sessions and character ownership checks; keep player routes disabled in production until ownership enforcement is tested.
4. Add item instances, equipment, bank storage, currency and atomic player trading.
5. Expand the playable loop with more recipes, tools, resource tiers, NPCs and combat.

## Continuation instruction

Always inspect this file and current repository state first. Work only in ViViD420/RSGalaxies-. Update this handoff after each meaningful milestone. Never assume GalaxyMMO code or decisions apply here.
