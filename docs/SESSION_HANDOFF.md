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
- Starter item definitions, stack-count inventory and persistent 100-slot bank storage with validated deposit/withdraw operations.
- Two gathering nodes (tree and stone) with interaction range, three charges, timed respawn and atomic inventory/XP rewards.
- One frontier workbench and a basic plank recipe: consume two softwood logs, create one wooden plank and award Crafting XP. Crafting checks station range, skill level, ingredient counts and output stack capacity atomically.
- HTTP routes expose public skills, world zones, gathering nodes, items and crafting definitions. Player-specific routes and development-only XP, movement, gathering and crafting routes are all disabled when RS_GALAXIES_ENV=production until authentication and character ownership are implemented.
- Domain, API, movement, gathering, crafting, PostgreSQL persistence and repository concurrency tests.
- Go CI runs tests with the race detector and go vet.

### Client
- Godot 4.3 provisional client prototype.
- Procedurally generated test floor, placeholder character, third-person orbit camera, keyboard movement, touchscreen D-pad, USE and jump buttons.
- Account login/registration UI, account character discovery/creation, bearer-authenticated movement, gathering and crafting, sign-out, and session-expiry recovery.
- The client keeps the bearer token in memory only; users sign in again after app restart.
- Inventory HUD shows collected starter resources.
- Headless Godot validation and Web export workflow. The corrected Godot CI installer now downloads the official 4.3 release and export templates.
- This is still a movement/gathering proof of concept, not a fully replicated production MMO client.

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
- Account authentication now includes PBKDF2-HMAC-SHA-256 password hashing (600,000 iterations, random salt), opaque 256-bit bearer sessions stored by SHA-256 digest, PostgreSQL accounts/sessions, 30-day expiry, last-used updates, logout revocation, generic login failures, account-owned character CRUD and account-scoped XP/movement/gathering/crafting/bank transactions. Active sessions are capped at 10 per account with oldest-session pruning; unknown/disabled login attempts receive dummy PBKDF2 work to reduce timing-based account enumeration. Registration is gated off by default in production. A per-process auth throttle and explicit production CORS allowlist are implemented. Email verification, password recovery, MFA, distributed rate limiting, admin disable tooling and public registration readiness remain incomplete.
- Account API reference is in docs/API.md; authentication threat boundaries and enablement decisions are in docs/ACCOUNT_AUTH_DESIGN.md. Realtime replication, production operations, NPCs, combat, additional recipes and player trading also remain incomplete.
- Go CI runs #117–#122 passed, including race-enabled Go tests, PostgreSQL account/session and ownership integration tests, and go vet. The timestamp precision regression was corrected and passed run #78. The client workflow was repaired after discovering its former container tag did not exist; the replacement workflow installs the official Godot 4.3 binary and export templates. Client run #24 passed import and Web export validation for the account UI; run #25 validates the latest small client lifecycle cleanup.

## Next engineering tasks

1. Verify the movement timestamp precision fix in GitHub Actions; investigate any remaining Go/PostgreSQL failures.
2. Validate the account login and world interactions in the auto-launched browser build on a real phone; refine touch controls and camera feel.
3. Add email verification, password recovery, MFA, distributed abuse controls and admin account operations before enabling public registration.
4. Add client bank UI and character selection/deletion; then add equipment, currency and player trading.
4. Add item instances, equipment, bank storage, currency and atomic player trading.
5. Expand the playable loop with more recipes, tools, resource tiers, NPCs and combat.

## Continuation instruction

Always inspect this file and current repository state first. Work only in ViViD420/RSGalaxies-. Update this handoff after each meaningful milestone. Never assume GalaxyMMO code or decisions apply here.
