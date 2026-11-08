# RS Galaxies — Session Handoff

Last updated: 2026-11-08

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
- Hidden Jedi progression is a documented future feature in docs/JEDI_PROGRESSION_DESIGN.md. The proposed design requires a randomly generated, persistent level-99 combat-skill mastery count, then grants an account entitlement to create a separate Jedi character with Force disciplines and lightsaber weapon families. It is deliberately not implemented or exposed yet: the current skill catalog has only Combat and Ranged as combat-relevant entries, so the prerequisite combat-skill pool must be expanded and categorized first.
- Ten skills, level 1–99, cumulative XP curve and 200,000,000 XP cap.
- Defensive in-memory repository copies and concurrency-safe experience awards.
- Starter zone, spawn point, server-time movement validation and world bounds.
- Tree, stone, copper and wild-herb gathering nodes with range checks, charges, respawn and atomic inventory/XP rewards. The Foraging skill feeds a new herbal-bundle workbench recipe. Godot chooses the nearest target and builds resource props from the server node catalog rather than hardcoded node IDs.
- Starter workbench and forge; recipes cover basic planks, herbal bundles, crafted healing herbal tea, copper smelting and a training sword. Herbal tea is a 15-health consumable. Crafting aggregates duplicate ingredients/outputs, simulates net stack changes, validates the exact station and commits inventory/XP only after all checks pass.
- Player inventory, 100-slot bank storage, coin wallet, persistent equipment slots, merchant healing potions and crafted herbal tea. New characters start with 100 coins.
- Starter merchant catalog with server-defined buy/sell prices. Buy/sell operations update coins and inventory atomically.
- PostgreSQL migrations 0001–0009 cover player state, gathering, accounts/sessions, character ownership, bank storage, currency, equipment, health, shared enemy spawn state, persistent zone chat, leased zone presence and the initial player-trade escrow schema. Migration 0009 creates trade/participant/offer/escrow tables plus a unique active-character reservation table. Authenticated invite/accept, escrow offer replacement, two-sided confirmation, atomic settlement, cancellation, expiry release and active-trade recovery are now implemented. Invite, acceptance, offer edits and confirmations require fresh same-zone proximity/availability checks; Godot trade UI is now integrated; production routes are disabled unless `RS_GALAXIES_ENABLE_TRADING=true`, while production abuse/load testing remain.
- PostgreSQL transactions and row locks protect XP, movement, gathering, crafting, bank transfers, market trades, equipment changes, PvE combat and presence heartbeat updates.
- Account authentication: PBKDF2-HMAC-SHA-256 with 600,000 iterations and random salts; password length policy is enforced on both hashing and verification; opaque 256-bit bearer tokens stored only as SHA-256 digests; fixed 30-day expiry; logout revocation; last-used timestamps are write-coalesced to at most once per minute to reduce database write amplification; maximum 10 active sessions per account; old-session pruning; dummy PBKDF2 work for unknown/disabled login accounts.
- Account-owned character CRUD and gameplay operations use account ownership in SQL predicates and row-locked transactions.
- Authenticated routes cover login/register/logout/me, character list/create/read, skills/XP, movement, gathering, crafting, bank deposits/withdrawals, merchant trades, equipment equip/unequip, combat attacks, safe-spawn rest, local zone chat, account-scoped zone presence and escrow-based player trade.
- Legacy unscoped player-ID and development mutation routes remain disabled in production. Production registration is disabled unless explicitly enabled.
- Auth routes have per-process IP/path throttling with bounded key memory. Production CORS uses an exact-origin allowlist. Security response headers and no-store responses are applied.
- Public catalog routes include skills, world zones, gathering nodes, crafting stations and recipes, items, enemy spawns and merchant prices. `GET /api/v1/world/catalog` returns the complete world definition set in one bounded bootstrap response to reduce client startup round trips.
- `/healthz` is a dependency-independent liveness probe; `/readyz` performs a bounded two-second PostgreSQL ping and returns a generic 503 on dependency failure.
- Zone presence uses a 20-second lease, account-owned heartbeat upserts, stale-entry cleanup, a 100-character roster cap and privacy-limited snapshots that never expose account IDs.

### Godot client

- Godot 4.3 third-person 3D prototype with keyboard, mouse, touch movement and jump.
- Account login/registration, multi-character selection and additional character creation.
- Server-driven enemy visuals, combat attacks, damage/loot/XP feedback, health HUD, Rest recovery/respawn and a polling local-zone chat panel.
- Gathering node target selection and resource-node visuals are driven by the public server node catalog rather than duplicated node coordinates in the client. A wild-herb patch has distinct low-poly plant visuals.
- The client loads skills, items, zones, gathering nodes, enemy spawns, crafting recipes and stations through one consolidated world-catalog request.
- Workbench and forge positions, visuals and interaction selection are driven by the server crafting-station catalog.
- Client sends a presence heartbeat every five seconds, polls the current zone every three seconds and renders remote character avatars/nameplates with smoothed position updates; stale characters are removed.
- Player-trade panel supports inviting nearby players, accepting invitations, staging up to 20 item types plus coins, editing/removing staged stacks, updating offers, confirming/cancelling, polling active trades after reconnect and refreshing inventory whenever assets enter or leave escrow.
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
- .env.example contains local-only sample settings; docs/API.md, docs/ACCOUNT_AUTH_DESIGN.md, docs/DEPLOYMENT.md, docs/COMBAT.md, docs/CHAT.md, docs/PRESENCE.md, docs/PLAYER_TRADE_DESIGN.md, docs/JEDI_PROGRESSION_DESIGN.md, docs/SKILL_ROSTER.md, docs/COMBAT_PROGRESSION.md and refreshed docs/ARCHITECTURE.md describe routes, security boundaries, deployment, gameplay rules, skill progression, combat accuracy design, escrow-safe trade requirements and the future Jedi unlock.
- Go server CI runs race-enabled unit/integration tests against PostgreSQL and go vet. Runs #278–#284 passed the password-policy correction, migration-009 schema registration/idempotency, escrow constraints and consolidated catalog checks. Server runs #364 and #365 passed, including the production trading feature-gate correction and trade HTTP integration suite. Run #363's PostgreSQL confirm/cancel race test passed; its overall failure was the earlier HTTP test setup omitting the explicit production trading flag, corrected and verified in #364/#365. Server run #372 passed with corrected concurrent offer-edit escrow accounting and movement-driven presence integration coverage. Server run #373 passed the two-party concurrent-confirmation test, including exactly-once settlement and duplicate retry protection. Server run #374 failed because its restart-recovery test created a pending invitation but asserted that it was an accepted/open trade; the failure was in test setup, not a demonstrated production recovery defect. The test has been corrected to heartbeat both characters, accept the invitation, then instantiate a fresh repository and recover the open trade before expiry. The correction is committed as 53d981a and server CI run #375 is currently validating it: https://github.com/ViViD420/RSGalaxies-/actions/runs/42526911104. Runs #366–#371 exposed intermediate test issues and compile mistakes that were corrected in later commits. Earlier failures exposed duplicate definitions, stale test helpers and Go ServeMux route-pattern conflicts; current code removes those conflicts and corrects affected tests. Runs #260–#263 passed presence persistence, privacy, heartbeat write-coalescing and account-scoped roster tests. Runs #245–#253 passed chat-history, readiness and rate-limiter hardening.
- Godot client CI uses the official Godot 4.3 release and export templates. Client runs #99–#121 passed the trade HUD, presence polling, multi-item offer editor, inventory refresh, responsive panel layout, incoming/outgoing invitation notices and per-participant confirmation display. Client run #122 passed framerate-independent remote-avatar interpolation and movement-facing orientation. Runs #89–#98 passed data-driven gathering/station visuals, foraging, presence, inventory search/sort, world-catalog startup and bounded HTTP timeouts/redirect-safe requests.

## Production safeguards and remaining launch gates

- Production startup requires RS_GALAXIES_DATABASE_URL; startup pings PostgreSQL and applies migrations before serving.
- Production CORS is configured through exact comma-separated RS_GALAXIES_ALLOWED_ORIGINS.
- Keep RS_GALAXIES_ENABLE_REGISTRATION unset until email verification, password recovery, distributed edge rate limiting, monitoring and admin account-disable/recovery operations exist. Keep RS_GALAXIES_ENABLE_TRADING unset until trade concurrency, restart recovery and abuse/load gates pass.
- Email verification, password reset/recovery, MFA, distributed rate limiting, admin account tools and security/load review remain incomplete.
- The current in-process auth throttle is only a local safety net; production still needs trusted edge controls. When its 10,000-key map is full, new keys fail closed until stale windows expire. `/readyz` provides a bounded PostgreSQL dependency check, while `/healthz` intentionally remains live during database outages.
- The Godot client does not persist access tokens across launches. Secure platform-specific persistence is a future decision.
- Remote character presence and position snapshots are implemented; successful authoritative movement now refreshes the shared presence snapshot in the same transaction, coalesced to one database write per character per second. The client still polls snapshots; there is no interpolation protocol or zone ownership yet. Enemy AI/movement, combat animations/abilities, unique equipment instances/durability, housing, settlements and large-scale world streaming are also outstanding.
- The account APIs require PostgreSQL. The server's in-memory-only development mode does not register account routes.

## Next engineering tasks

1. Verify server CI for the restart-recovery test correction committed as 53d981a; run #373 passed and run #374 failed on the test's pending-vs-open setup mismatch.
2. Keep the server/client workflows green, then run trade abuse/load tests before enabling public production trade.
3. Expand the combat skill catalog and assign explicit progression categories before implementing the hidden Jedi unlock. Follow docs/JEDI_PROGRESSION_DESIGN.md; do not expose the unlock until its persistence, concurrency and anti-tampering tests pass.
4. Test the multi-item Godot trade panel, invitation/confirmation feedback and remote-avatar movement on a real phone.
5. Implement client-side interpolation/prediction correction using the now-fresher authoritative presence snapshots, then zone ownership; add enemy AI/movement and combat animations.
6. Add unique item instances/durability, currency sinks, additional professions/recipes/NPCs, housing and settlements. Unique item instances are a prerequisite for correct dual-wield lightsaber equipment.
7. Complete production account safeguards: email verification, recovery, MFA, distributed rate limiting, admin account tools and security/load review.

## Continuation instruction

Always inspect this file, current repository HEAD and recent CI first. Work only in ViViD420/RSGalaxies-. Update this handoff after meaningful milestones. Never assume GalaxyMMO code or decisions apply here.
