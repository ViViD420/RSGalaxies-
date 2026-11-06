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
- Player inventory, 100-slot bank storage, coin wallet, persistent equipment slots and consumable healing potions. New characters start with 100 coins.
- Starter merchant catalog with server-defined buy/sell prices. Buy/sell operations update coins and inventory atomically.
- PostgreSQL migrations 0001–0007 cover player state, gathering, accounts/sessions, character ownership, bank storage, currency, equipment, health, shared enemy spawn state and persistent zone chat.
- PostgreSQL transactions and row locks protect XP, movement, gathering, crafting, bank transfers, market trades, equipment changes and PvE combat.
- Account authentication: PBKDF2-HMAC-SHA-256 with 600,000 iterations and random salts; opaque 256-bit bearer tokens stored only as SHA-256 digests; fixed 30-day expiry; logout revocation; last-used updates; maximum 10 active sessions per account; old-session pruning; dummy PBKDF2 work for unknown/disabled login accounts.
- Account-owned character CRUD and gameplay operations use account ownership in SQL predicates and row-locked transactions.
- Authenticated routes cover login/register/logout/me, character list/create/read, skills/XP, movement, gathering, crafting, bank deposits/withdrawals, merchant trades, equipment equip/unequip, combat attacks, safe-spawn rest and local zone chat.
- Legacy unscoped player-ID and development mutation routes remain disabled in production. Production registration is disabled unless explicitly enabled.
- Auth routes have per-process IP/path throttling with bounded key memory. Production CORS uses an exact-origin allowlist. Security response headers and no-store responses are applied.
- Public catalog routes include skills, world zones, gathering nodes, items, crafting definitions and merchant prices. `/healthz` is a liveness probe; `/readyz` performs a bounded PostgreSQL readiness check without exposing dependency errors.
- `/healthz` is a dependency-independent liveness probe; `/readyz` performs a bounded two-second PostgreSQL ping and returns a generic 503 on dependency failure.

### Godot client

- Godot 4.3 third-person 3D prototype with keyboard, mouse, touch movement and jump.
- Account login/registration, multi-character selection and additional character creation.
- Server-driven enemy visuals, combat attacks, damage/loot/XP feedback, health HUD, Rest recovery/respawn and a polling local-zone chat panel.
- Live authenticated skills panel with all ten skills, per-skill XP progress bars and next-level XP remaining (including the 200M XP cap display at level 99); requests are cancelled when characters change or sign out.
- Bearer-authenticated movement, gathering, crafting, bank transfers, merchant buy/sell panel, equipment panel and full inventory/bank contents panel with selectable stack quantities for deposits and withdrawals.
- Starter bank chest, merchant stall/NPC, copper vein, forge and enemy spawns in the test world; coin balance, health, attack/defense bonuses, bag and bank counts appear in the HUD. Equipped sword/armor pieces have simple client-side visuals.
- Market panel loads item names and prices from the server catalog.
- Sign-out and session-expiry recovery. Access tokens are held in memory only; the user signs in again after restarting the client.
- Combat rules are documented in docs/COMBAT.md; local chat behavior is documented in docs/CHAT.md. This remains a small gameplay prototype, not a full multiplayer client or final art direction.

### Development environment and CI

- .devcontainer/ provides a GitHub Codespaces Go/Godot/PostgreSQL workspace with separate development and integration-test databases.
- Codespaces setup installs Godot 4.3 and export templates, runs race-enabled Go tests, go vet, headless project import and Web export.
- .env.example contains local-only sample settings; docs/API.md, docs/ACCOUNT_AUTH_DESIGN.md, docs/DEPLOYMENT.md, docs/COMBAT.md and docs/CHAT.md describe routes, security boundaries, deployment and gameplay rules.
- Go server CI runs race-enabled unit/integration tests against PostgreSQL and go vet. Server run #245 passed; run #251 is validating the latest bounded chat-history and operational hardening changes. Earlier run #244 passed the full suite, including persistent zone chat, combat, economy, equipment, copper progression and transactional consumable use.
- Godot client CI uses the official Godot 4.3 release and export templates. Client run #73 passed, including the level-99 XP-cap progress bar. Client runs #69–#72 also passed the trader NPC, bank item selection, consumable use, character panel lifecycle and skill progression UI.

## Production safeguards and remaining launch gates

- Production startup requires RS_GALAXIES_DATABASE_URL; startup pings PostgreSQL and applies migrations before serving.
- Production CORS is configured through exact comma-separated RS_GALAXIES_ALLOWED_ORIGINS.
- Keep RS_GALAXIES_ENABLE_REGISTRATION unset until email verification, password recovery, distributed edge rate limiting, monitoring and admin account-disable/recovery operations exist.
- Email verification, password reset/recovery, MFA, distributed rate limiting, admin account tools and security/load review remain incomplete.
- The current in-process auth throttle is only a local safety net; production still needs trusted edge controls. When its 10,000-key map is full, new keys fail closed until stale windows expire. `/readyz` provides a bounded PostgreSQL dependency check, while `/healthz` intentionally remains live during database outages.
- The Godot client does not persist access tokens across launches. Secure platform-specific persistence is a future decision.
- No remote-player replication, enemy AI/movement, combat animations/abilities, unique equipment instances/durability, housing, settlements, player-to-player trading or large-scale world streaming yet.
- The account APIs require PostgreSQL. The server's in-memory-only development mode does not register account routes.

## Next engineering tasks

1. Finish validating the latest PostgreSQL chat-history tests and server hardening runs, then add player-to-player trade with escrow and transaction safety.
2. Complete a real-phone pass on the Godot Web client and add better inventory/bank sorting and item filters.
3. Add email verification, password recovery, MFA, distributed abuse controls and administrative account operations before enabling public registration.
4. Add unique item instances/durability, combat animations/AI, currency sinks and player-to-player trade with escrow/transaction safety.
5. Expand professions, gathering tiers, recipes, NPCs, combat, housing and settlements.
6. Plan remote-player replication and scalable world/zone ownership.

## Continuation instruction

Always inspect this file, current repository HEAD and recent CI first. Work only in ViViD420/RSGalaxies-. Update this handoff after meaningful milestones. Never assume GalaxyMMO code or decisions apply here.
