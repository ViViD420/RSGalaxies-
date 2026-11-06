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
- Tree, stone, copper and wild-herb gathering nodes with range checks, charges, respawn and atomic inventory/XP rewards. The Foraging skill feeds a new herbal-bundle workbench recipe. Godot chooses the nearest target and builds resource props from the server node catalog rather than hardcoded node IDs.
- Starter workbench and forge; recipes cover basic planks, herbal bundles, copper smelting and a crafted training sword. Crafting validates range, skill level, ingredients and output stack capacity.
- Player inventory, 100-slot bank storage, coin wallet, persistent equipment slots and consumable healing potions. New characters start with 100 coins.
- Starter merchant catalog with server-defined buy/sell prices. Buy/sell operations update coins and inventory atomically.
- PostgreSQL migrations 0001–0008 cover player state, gathering, accounts/sessions, character ownership, bank storage, currency, equipment, health, shared enemy spawn state, persistent zone chat and leased zone presence.
- PostgreSQL transactions and row locks protect XP, movement, gathering, crafting, bank transfers, market trades, equipment changes, PvE combat and presence heartbeat updates.
- Account authentication: PBKDF2-HMAC-SHA-256 with 600,000 iterations and random salts; password length policy is enforced on both hashing and verification; opaque 256-bit bearer tokens stored only as SHA-256 digests; fixed 30-day expiry; logout revocation; last-used timestamps are write-coalesced to at most once per minute to reduce database write amplification; maximum 10 active sessions per account; old-session pruning; dummy PBKDF2 work for unknown/disabled login accounts.
- Account-owned character CRUD and gameplay operations use account ownership in SQL predicates and row-locked transactions.
- Authenticated routes cover login/register/logout/me, character list/create/read, skills/XP, movement, gathering, crafting, bank deposits/withdrawals, merchant trades, equipment equip/unequip, combat attacks, safe-spawn rest, local zone chat and account-scoped zone presence.
- Legacy unscoped player-ID and development mutation routes remain disabled in production. Production registration is disabled unless explicitly enabled.
- Auth routes have per-process IP/path throttling with bounded key memory. Production CORS uses an exact-origin allowlist. Security response headers and no-store responses are applied.
- Public catalog routes include skills, world zones, gathering nodes, crafting stations and recipes, items, enemy spawns and merchant prices.
- `/healthz` is a dependency-independent liveness probe; `/readyz` performs a bounded two-second PostgreSQL ping and returns a generic 503 on dependency failure.
- Zone presence uses a 20-second lease, account-owned heartbeat upserts, stale-entry cleanup, a 100-character roster cap and privacy-limited snapshots that never expose account IDs.

### Godot client

- Godot 4.3 third-person 3D prototype with keyboard, mouse, touch movement and jump.
- Account login/registration, multi-character selection and additional character creation.
- Server-driven enemy visuals, combat attacks, damage/loot/XP feedback, health HUD, Rest recovery/respawn and a polling local-zone chat panel.
- Gathering node target selection and resource-node visuals are driven by the public server node catalog rather than duplicated node coordinates in the client.
- Workbench and forge positions, visuals and interaction selection are driven by the server crafting-station catalog.
- Client sends a presence heartbeat every five seconds, polls the current zone every three seconds and renders remote character avatars/nameplates with smoothed position updates; stale characters are removed.
- All Godot API HTTPRequests use bounded timeouts and disable automatic redirects to prevent bearer-token forwarding to redirected hosts.
- Live authenticated skills panel with all ten skills, per-skill XP progress bars and next-level XP remaining (including the 200M XP cap display at level 99); requests are cancelled when characters change or sign out.
- Bearer-authenticated movement, gathering, crafting, bank transfers, merchant buy/sell panel, equipment panel and full inventory/bank contents panel with search, category filters, quantity sorting and selectable stack transfers.
- Starter bank chest, merchant stall/NPC, copper vein, forge and enemy spawns in the test world; coin balance, health, attack/defense bonuses, bag and bank counts appear in the HUD. Equipped sword/armor pieces have simple client-side visuals.
- Market panel loads item names and prices from the server catalog.
- Sign-out and session-expiry recovery. Access tokens are held in memory only; the user signs in again after restarting the client.
- Combat rules are documented in docs/COMBAT.md; local chat behavior is documented in docs/CHAT.md. This remains a small gameplay prototype, not a full multiplayer client or final art direction.

### Development environment and CI

- .devcontainer/ provides a GitHub Codespaces Go/Godot/PostgreSQL workspace with separate development and integration-test databases.
- Codespaces setup installs Godot 4.3 and export templates, runs race-enabled Go tests, go vet, headless project import and Web export.
- .env.example contains local-only sample settings; docs/API.md, docs/ACCOUNT_AUTH_DESIGN.md, docs/DEPLOYMENT.md, docs/COMBAT.md, docs/CHAT.md, docs/PRESENCE.md and refreshed docs/ARCHITECTURE.md describe routes, security boundaries, deployment and gameplay rules.
- Go server CI runs race-enabled unit/integration tests against PostgreSQL and go vet. Server run #263 passed after the production-route comment cleanup; run #262 passed presence persistence, privacy, heartbeat write-coalescing and account-scoped roster tests. Run #268 exposed a stale public market-catalog test expectation after the herbal items were added; commit 8c0a7fb updates that assertion. Runs #269–#274 are validating the catalog correction, herbal foraging/crafting additions, password-policy enforcement and coalesced session activity writes. Runs #260–#263 passed. Server runs #245–#253 passed chat-history, readiness and rate-limiter hardening.
- Godot client CI uses the official Godot 4.3 release and export templates. Client run #90 passed data-driven crafting-station visuals and target selection; runs #89–#90 passed the client catalog changes. Runs #83–#88 also passed presence, inventory search/sort, gathering visuals and bounded HTTP timeouts/redirect-safe requests.

## Production safeguards and remaining launch gates

- Production startup requires RS_GALAXIES_DATABASE_URL; startup pings PostgreSQL and applies migrations before serving.
- Production CORS is configured through exact comma-separated RS_GALAXIES_ALLOWED_ORIGINS.
- Keep RS_GALAXIES_ENABLE_REGISTRATION unset until email verification, password recovery, distributed edge rate limiting, monitoring and admin account-disable/recovery operations exist.
- Email verification, password reset/recovery, MFA, distributed rate limiting, admin account tools and security/load review remain incomplete.
- The current in-process auth throttle is only a local safety net; production still needs trusted edge controls. When its 10,000-key map is full, new keys fail closed until stale windows expire. `/readyz` provides a bounded PostgreSQL dependency check, while `/healthz` intentionally remains live during database outages.
- The Godot client does not persist access tokens across launches. Secure platform-specific persistence is a future decision.
- Remote character presence and position snapshots are implemented, but there is no authoritative continuous movement replication, interpolation protocol, reconnect recovery or zone ownership yet. Enemy AI/movement, combat animations/abilities, unique equipment instances/durability, housing, settlements, player-to-player trading and large-scale world streaming are also outstanding.
- The account APIs require PostgreSQL. The server's in-memory-only development mode does not register account routes.

## Next engineering tasks

1. Finish validating server runs #269–#274, including the new account-auth and session-write optimizations; then build reconnect recovery and authoritative remote-player movement snapshots.
2. Add player-to-player trade with escrow and transaction safety, then add a real-phone pass on the Godot Web client.
3. Add email verification, password recovery, MFA, distributed abuse controls and administrative account operations before enabling public registration.
4. Add unique item instances/durability, combat animations/AI, currency sinks and player-to-player trade with escrow/transaction safety.
5. Expand professions, gathering tiers, recipes, NPCs, combat, housing and settlements.
6. Plan remote-player replication and scalable world/zone ownership.

## Continuation instruction

Always inspect this file, current repository HEAD and recent CI first. Work only in ViViD420/RSGalaxies-. Update this handoff after meaningful milestones. Never assume GalaxyMMO code or decisions apply here.
