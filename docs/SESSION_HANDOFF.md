# RS Galaxies — Session Handoff

Last updated: 2026-11-06

## Project identity

- Repository: https://github.com/ViViD420/RSGalaxies-
- Owner: ViViD420
- Branch: main
- This is a standalone original MMORPG project. Never reuse or modify the separate GalaxyMMO repository.

## Product direction

An original sandbox MMORPG combining long-form skill progression and item gameplay with a social, open-ended world built around professions, crafting, trade, housing and settlements. Use original code and assets only.

## Current implementation

### Go world server

- Go authoritative world API, domain model and PostgreSQL persistence.
- Ten skills, level 1–99, cumulative XP curve and 200,000,000 XP cap.
- Defensive in-memory repository copies and concurrency-safe experience awards.
- Starter zone, spawn point, server-time movement validation and world bounds.
- Tree, stone and copper gathering nodes with range checks, charges, respawn and atomic inventory/XP rewards.
- Starter workbench and forge; recipes cover basic planks, copper smelting and a crafted training sword. Crafting validates range, skill level, ingredients and output stack capacity.
- Player inventory, 100-slot bank storage, coin wallet and persistent equipment slots. New characters start with 100 coins.
- Starter merchant catalog with server-defined buy/sell prices. Buy/sell operations update coins and inventory atomically.
- PostgreSQL migrations 0001–0006 cover player state, gathering, accounts/sessions, character ownership, bank storage, currency, equipment, health and shared enemy spawn state.
- PostgreSQL transactions and row locks protect XP, movement, gathering, crafting, bank transfers, market trades, equipment changes and PvE combat.
- Account authentication: PBKDF2-HMAC-SHA-256 with 600,000 iterations and random salts; opaque 256-bit bearer tokens stored only as SHA-256 digests; fixed 30-day expiry; logout revocation; last-used updates; maximum 10 active sessions per account; old-session pruning; dummy PBKDF2 work for unknown/disabled login accounts.
- Account-owned character CRUD and gameplay operations use account ownership in SQL predicates and row-locked transactions.
- Authenticated routes cover login/register/logout/me, character list/create/read, skills/XP, movement, gathering, crafting, bank deposits/withdrawals, merchant trades, equipment equip/unequip, combat attacks and safe-spawn rest.
- Legacy unscoped player-ID and development mutation routes remain disabled in production. Production registration is disabled unless explicitly enabled.
- Auth routes have per-process IP/path throttling. Production CORS uses an exact-origin allowlist. Security response headers and no-store responses are applied.
- Public catalog routes include skills, world zones, gathering nodes, items, crafting definitions and merchant prices.

### Godot client

- Godot 4.3 third-person 3D prototype with keyboard, mouse, touch movement and jump.
- Account login/registration, multi-character selection and additional character creation.
- Server-driven enemy visuals, combat attacks, damage/loot/XP feedback, health HUD and a Rest control for safe-zone recovery/respawn.
- Bearer-authenticated movement, gathering, crafting, bank transfers, merchant buy/sell panel and equipment panel.
- Starter bank chest and merchant stall in the test world; coin balance, attack/defense bonuses, bag and bank counts appear in the HUD.
- Market panel loads item names and prices from the server catalog.
- Sign-out and session-expiry recovery. Access tokens are held in memory only; the user signs in again after restarting the client.
- This remains a small gameplay prototype, not a full multiplayer client or final art direction.

### Development environment and CI

- .devcontainer/ provides a GitHub Codespaces Go/Godot/PostgreSQL workspace with separate development and integration-test databases.
- Codespaces setup installs Godot 4.3 and export templates, runs race-enabled Go tests, go vet, headless project import and Web export.
- .env.example contains local-only sample settings; docs/API.md, docs/ACCOUNT_AUTH_DESIGN.md and docs/DEPLOYMENT.md describe routes, security boundaries and deployment.
- Go server CI runs race-enabled unit/integration tests against PostgreSQL and go vet. Server run #187 passed the latest market/equipment/copper data changes with race-enabled Go tests and PostgreSQL integration. Client runs #45–#47 passed, including copper mining and forge interactions.
- Godot client CI uses the official Godot 4.3 release and export templates. Client runs #42–#44 passed, including market catalog loading, character selection, merchant panel, equipment panel and gear request flows.

## Production safeguards and remaining launch gates

- Production startup requires RS_GALAXIES_DATABASE_URL; startup pings PostgreSQL and applies migrations before serving.
- Production CORS is configured through exact comma-separated RS_GALAXIES_ALLOWED_ORIGINS.
- Keep RS_GALAXIES_ENABLE_REGISTRATION unset until email verification, password recovery, distributed edge rate limiting, monitoring and admin account-disable/recovery operations exist.
- Email verification, password reset/recovery, MFA, distributed rate limiting, admin account tools and security/load review remain incomplete.
- The current in-process auth throttle is only a local safety net; production still needs trusted edge controls.
- The Godot client does not persist access tokens across launches. Secure platform-specific persistence is a future decision.
- No remote-player replication, NPC behaviors, combat calculations, equipment visuals/item instances/durability, housing, settlements, player-to-player trading or large-scale world streaming yet.
- The account APIs require PostgreSQL. The server's in-memory-only development mode does not register account routes.

## Next engineering tasks

1. Validate the first server-authoritative PvE combat loop, including enemy spawn persistence, cooldowns, loot and safe-spawn recovery.
2. Add merchant/bank usability improvements and a real-phone pass on the Godot Web client.
3. Add email verification, password recovery, MFA, distributed abuse controls and administrative account operations before enabling public registration.
4. Add equipment visuals, item instances/durability, combat calculations, currency sinks and player-to-player trade with escrow/transaction safety.
5. Expand professions, gathering tiers, recipes, NPCs, combat, housing and settlements.
6. Plan remote-player replication and scalable world/zone ownership.

## Continuation instruction

Always inspect this file, current repository HEAD and recent CI first. Work only in ViViD420/RSGalaxies-. Update this handoff after meaningful milestones. Never assume GalaxyMMO code or decisions apply here.
