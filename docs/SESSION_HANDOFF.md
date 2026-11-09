# RS Galaxies — Session Handoff

Last updated: 2026-11-09

## Project identity

- Repository: https://github.com/ViViD420/RSGalaxies-
- Owner: ViViD420
- Branch: main
- This is a standalone original MMORPG project. Never reuse or modify the separate GalaxyMMO repository.

## Product direction

An original sandbox MMORPG combining long-form skill progression and item gameplay with a social, open-ended world built around professions, crafting, trade, housing and settlements. Use original code and assets only.

## Current implementation

### Latest approved design direction

- Broad progression is the intended character model: players may train across many skills. Equipment, tools, recipes and access requirements—not arbitrary caps on the number of skills—will create profession identity and economic interdependence.
- The active core roster is now implemented as 18 skills: Melee and Ranged parent skills; Blades, Polearms, Unarmed, Pistols, Rifles and Magic combat specializations; independent Defense and Strength co-skills; and Farming, Foraging, Mining, Scavenging, Cooking, Crafting, Smithing and Woodcutting. Combat is a gameplay system, not a skill; Tactics is deferred. Advanced non-combat professions are deferred. Broad progression is allowed; profession identity comes from equipment, tools, recipes, materials, stations and access requirements.

### Go world server

- Go authoritative world API, domain model and PostgreSQL persistence.
- The active skill catalog now contains 18 skills with private parent/co-skill metadata. New characters receive all 18 at level 1; PostgreSQL loading backfills missing entries for older characters. Legacy `combat` and `tactics` XP keys remain preserved in existing records but are inactive, omitted from the client skill panel and cannot receive new XP. The starter sword selects Blades; unarmed attacks select Unarmed. Enemy defeat awards specialization XP and one-quarter parent XP; successful hits award Strength XP; taking damage awards Defense XP. Strength contributes to melee damage, Defense to mitigation and item-defined armor requirements. The six specializations alone are marked as future Jedi mastery candidates. Accuracy remains a derived outcome. See docs/SKILL_ROSTER.md and docs/COMBAT_PROGRESSION.md.
- Hidden Jedi progression remains a documented future feature in docs/JEDI_PROGRESSION_DESIGN.md. The six combat specializations—Blades, Polearms, Unarmed, Pistols, Rifles and Magic—are the only future mastery candidates; parent skills, Defense and Strength do not count. The hidden challenge, entitlement and Jedi character archetype are still not implemented or exposed. See docs/JEDI_PROGRESSION_DESIGN.md.
- Defensive in-memory repository copies and concurrency-safe experience awards.
- Starter zone, spawn point, server-time movement validation and world bounds.
- Tree, stone, copper and wild-herb gathering nodes with range checks, charges, respawn and atomic inventory/XP rewards. The Foraging skill feeds a new herbal-bundle workbench recipe. Godot chooses the nearest target and builds resource props from the server node catalog rather than hardcoded node IDs.
- Starter workbench and forge; recipes cover basic planks, herbal bundles, crafted healing herbal tea, copper smelting and a training sword. Herbal tea is a 15-health consumable. Crafting aggregates duplicate ingredients/outputs, simulates net stack changes, validates the exact station and commits inventory/XP only after all checks pass.
- Player inventory, 100-slot bank storage, coin wallet, persistent equipment slots, merchant healing potions and crafted herbal tea. New characters start with 100 coins.
- Starter merchant catalog with server-defined buy/sell prices. Buy/sell operations update coins and inventory atomically.
- PostgreSQL migrations 0001–0010 cover player state, gathering, accounts/sessions, character ownership, bank storage, currency, equipment, health, shared enemy spawn state, persistent zone chat, leased zone presence, player-trade escrow and persistent per-spawn enemy AI state. Migration 0009 creates trade/participant/offer/escrow tables plus a unique active-character reservation table. Authenticated invite/accept, escrow offer replacement, two-sided confirmation, atomic settlement, cancellation, expiry release and active-trade recovery are now implemented. Invite, acceptance, offer edits and confirmations require fresh same-zone proximity/availability checks; Godot trade UI is now integrated; production routes are disabled unless `RS_GALAXIES_ENABLE_TRADING=true`, while production abuse/load testing remain.
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
- Live authenticated skills panel with all 18 active ordinary skills, per-skill XP progress bars and next-level XP remaining (including the 200M XP cap display at level 99); requests are cancelled when characters change or sign out.
- Bearer-authenticated movement, gathering, crafting, bank transfers, merchant buy/sell panel, equipment panel and full inventory/bank contents panel with search, category filters, quantity sorting and selectable stack transfers.
- Starter bank chest, merchant stall/NPC, copper vein, forge and enemy spawns in the test world; coin balance, health, attack/defense bonuses, bag and bank counts appear in the HUD. Equipped sword/armor pieces have simple client-side visuals.
- Market panel loads item names and prices from the server catalog.
- Sign-out and session-expiry recovery. Access tokens are held in memory only; the user signs in again after restarting the client.
- Combat rules are documented in docs/COMBAT.md; local chat behavior is documented in docs/CHAT.md. This remains a small gameplay prototype, not a full multiplayer client or final art direction.

### Development environment and CI

- .devcontainer/ provides a GitHub Codespaces Go/Godot/PostgreSQL workspace with separate development and integration-test databases.
- Codespaces setup installs Godot 4.3 and export templates, runs race-enabled Go tests, go vet, headless project import and Web export.
- .env.example contains local-only sample settings; docs/API.md, docs/ACCOUNT_AUTH_DESIGN.md, docs/DEPLOYMENT.md, docs/COMBAT.md, docs/CHAT.md, docs/PRESENCE.md, docs/PLAYER_TRADE_DESIGN.md, docs/JEDI_PROGRESSION_DESIGN.md, docs/SKILL_ROSTER.md, docs/COMBAT_PROGRESSION.md and refreshed docs/ARCHITECTURE.md describe routes, security boundaries, deployment, gameplay rules, skill progression, combat accuracy design, escrow-safe trade requirements and the future Jedi unlock.
- Go server CI runs race-enabled unit/integration tests against PostgreSQL and go vet. Runs #278–#284 passed the password-policy correction, migration-009 schema registration/idempotency, escrow constraints and consolidated catalog checks. Server runs #364 and #365 passed, including the production trading feature-gate correction and trade HTTP integration suite. Run #363's PostgreSQL confirm/cancel race test passed; its overall failure was the earlier HTTP test setup omitting the explicit production trading flag, corrected and verified in #364/#365. Server run #372 passed with corrected concurrent offer-edit escrow accounting and movement-driven presence integration coverage. Server run #373 passed the two-party concurrent-confirmation test, including exactly-once settlement and duplicate retry protection. Server run #374 failed because its restart-recovery test created a pending invitation but asserted that it was an accepted/open trade; the failure was in test setup, not a demonstrated production recovery defect. The test has been corrected to heartbeat both characters, accept the invitation, then instantiate a fresh repository and recover the open trade before expiry. The correction is committed as 53d981a and server CI run #375 passed: https://github.com/ViViD420/RSGalaxies-/actions/runs/42526911104. The expanded 18-skill catalog, internal category/mastery metadata, legacy-character skill backfill, XP-map normalization and regression tests are committed; server run #389 passed the expanded roster, backfill and updated catalog tests: https://github.com/ViViD420/RSGalaxies-/actions/runs/42529387627. Run #390 is validating the final small XP-map mutation-order refinement: https://github.com/ViViD420/RSGalaxies-/actions/runs/42529401720. Earlier runs #376–#384 caught the initial stale 10-skill catalog assertion, malformed helper insertion and outdated mastery-count assertion; later commits correct these issues. Runs #366–#371 exposed intermediate test issues and compile mistakes that were corrected in later commits. Earlier failures exposed duplicate definitions, stale test helpers and Go ServeMux route-pattern conflicts; current code removes those conflicts and corrects affected tests. Runs #260–#263 passed presence persistence, privacy, heartbeat write-coalescing and account-scoped roster tests. Runs #245–#253 passed chat-history, readiness and rate-limiter hardening.
- Godot client CI uses the official Godot 4.3 release and export templates. Client runs #99–#121 passed the trade HUD, presence polling, multi-item offer editor, inventory refresh, responsive panel layout, incoming/outgoing invitation notices and per-participant confirmation display. Client run #122 passed framerate-independent remote-avatar interpolation and movement-facing orientation. Runs #89–#98 passed data-driven gathering/station visuals, foraging, presence, inventory search/sort, world-catalog startup and bounded HTTP timeouts/redirect-safe requests.

## Production safeguards and remaining launch gates

- Production startup requires RS_GALAXIES_DATABASE_URL; startup pings PostgreSQL and applies migrations before serving.
- Production CORS is configured through exact comma-separated RS_GALAXIES_ALLOWED_ORIGINS.
- Keep RS_GALAXIES_ENABLE_REGISTRATION unset until email verification, password recovery, distributed edge rate limiting, monitoring and admin account-disable/recovery operations exist. Keep RS_GALAXIES_ENABLE_TRADING unset until trade concurrency, restart recovery and abuse/load gates pass.
- Email verification, password reset/recovery, MFA, distributed rate limiting, admin account tools and security/load review remain incomplete.
- The current in-process auth throttle is only a local safety net; production still needs trusted edge controls. When its 10,000-key map is full, new keys fail closed until stale windows expire. `/readyz` provides a bounded PostgreSQL dependency check, while `/healthz` intentionally remains live during database outages.
- The Godot client does not persist access tokens across launches. Secure platform-specific persistence is a future decision.
- Remote character presence and position snapshots are implemented; successful authoritative movement refreshes shared presence in the same transaction, coalesced to one database write per character per second. Enemy AI now has persisted per-spawn state, a server-owned 250 ms world tick worker, fresh-presence target selection, transactional telegraph attacks with player health/Defense XP, and an authenticated zone-scoped enemy snapshot endpoint. The Godot client still needs to poll/render/interpolate those snapshots. The current world tick uses a nearest eligible active character within the enemy leash; target stickiness/party threat, pathfinding and more advanced behavior remain future work. Combat animations/abilities, boss encounters, dedicated loot tables, deeper crafting, unique equipment instances/durability, housing, settlements and large-scale world streaming are also outstanding.
- The account APIs require PostgreSQL. The server's in-memory-only development mode does not register account routes.

## Latest verified repository state

- Server CI runs #413–#421 all passed, including migration 0010, restart recovery, enemy damage, transactional tick coordinator and snapshot model/repository.
- Server runs #422–#424 failed because the HTTP test double had not yet implemented the newly extended enemy-snapshot interface. This was a test-adapter compile failure; commit `97c5d41` added the fake implementation and #425 passed. #426 passed stale-presence leash-return coverage; #427 passed dynamic live-position combat. #428 failed because its integration assertion still expected the pre-player-attack enemy health; #429 corrected the expected value and passed. Runs #430 and #431 both passed, validating monotonic simulation time across replica clock skew. Runs #432 and #433 passed, covering defeated-target rejection and cross-zone target isolation. #434 passed the respawn-state reset implementation. #435 failed to compile because a local error channel shadowed the imported errors package; commit `fcc331d` renames the channel. Runs #436–#439 failed on the test-only channel shadowing, corrected by `fcc331d`. #440 validates that correction and the current integrated changes; #441 adds the final respawn cooldown timestamp assertion; #442 verifies defeated-player error precedence. Runs #440–#442 were pending at the latest check. Verify Actions before claiming these changes green.
- Client runs #125–#127 all passed. They validate snapshot polling/interpolation code and the respawn-label/type-cast refinements.
- Latest documentation updates `28c86c8` (combat), `77b8195` (architecture), `662df6b` (roadmap) and `1a035ca` (README) record the transactional enemy simulation and Godot integration. API docs update `ad6b6e5` documents the authenticated snapshot endpoint and server tick.

## Enemy AI foundation, transactional world and Godot integration (current milestone)

- Domain commits: `b805b26` added `server/internal/game/enemy_ai.go`; `b8bdd94` added behavior tests; `cf2540c` and `1621e3f` refined bounded return-path tests. Runs #411/#412 passed; intermediate #410 failed and was corrected.
- Persistence commits: `a5432cb` adds migration 0010; `9db87ff` registers it; `5c52656` adds row-locked persistence; `1a3e8b3` adds restart recovery. `df989e2` adds `Player.ReceiveEnemyAttack`; `cd54f78` tests range, mitigation, Defense XP and lethal hits. Runs #413–#418 passed.
- World integration commits: `8f633bd` adds `AccountRepository.TickEnemyAI`, selecting fresh same-zone presence and locking player then enemy before atomically committing AI state and telegraphed damage; `f9eee97` adds the snapshot model; `ba02b5f` adds account-scoped snapshot retrieval; `a01f2d7` adds the authenticated endpoint; `71ec8ec` starts the PostgreSQL-backed 250 ms world worker; `632c6d1` adds concurrency integration coverage; `97c5d41` tests authenticated snapshot authorization; `3308ed9` adds stale-presence leash-return coverage. `ad6b6e5` documents the API.
- CI corrections: #422–#424 failed because the HTTP test double did not yet implement the new snapshot method; `97c5d41` fixes the test adapter and #425 passes. #426 passes stale-presence/leash-return coverage. #427 passes the live-position attack repository change. #428/#429 are validating the extended integration test, including a player attacking a moving enemy and checking the snapshot health. #430/#431 validate monotonic timestamps across server clock skew; pending at the latest check.
- The world tick queries fresh presence within the enemy leash, revalidates presence after locking the authoritative character row, then locks the enemy spawn. Completed telegraphs call `ReceiveEnemyAttack` and persist player health/Defense XP and enemy state in the same transaction. This preserves player-then-enemy lock order. Multiple workers serialize on PostgreSQL rows and the persisted cooldown prevents duplicate hits at one telegraph deadline.
- The world worker runs only with PostgreSQL-backed account storage and advances catalogued spawns every 250 ms. No-player ticks cause pursuing enemies to return to spawn. Enemy snapshots are authenticated and limited to the requesting character's current zone.
- Godot commits: `401e1e5` adds 0.5-second authenticated snapshot polling, health/behavior labels, telegraph warnings and interpolation; `424415e` keeps the respawn label visible when the creature mesh is hidden; `816147a` uses safe typed mesh casts; `b3e2b0f` prevents selecting defeated/respawning enemies for interaction. Client runs #125–#127 passed; #128 validates the latest defeated-target interaction filter.
- PostgreSQL player-initiated attacks now read the persisted enemy position (`a13b0b9`), so the client can attack a moving enemy rather than the static catalog spawn. `6f4369c` adds integration coverage; `4e403d9` updates expected snapshot health.
- Monotonic simulation time: `0a0cbe2` prevents a delayed tick from moving persisted simulation time backwards across server replicas; `9a64f8d` tests clock regression. Runs #430/#431 passed. `ebfa1a5` adds a PostgreSQL integration assertion that a defeated character with a fresh presence lease is not attacked and causes the enemy to return; #432 is pending. `49cfd8c` adds a cross-zone presence test proving an enemy does not pursue or damage a character after the character and presence move to another zone; #433 is pending. `d3eb26f` resets enemy position, behavior, telegraph and cooldown when player combat triggers an expired respawn; run #434 passed. `6b3d375` tests the respawn gate and verifies the next life starts at the catalog spawn with cleared transient timers; #435 exposed a test-only shadowing compile error, corrected by `fcc331d`. `cb76074` unifies click retaliation and autonomous AI cooldowns, cancelling a pending telegraph after immediate retaliation; `f981afa` tests the shared cooldown and resumed telegraph; `8024495` clears completed telegraph deadlines; `7179e36` asserts the domain transition clears them. `8049a9b` updates combat docs. `fcc331d` fixes the test-channel shadowing compile issue. `c4577c1` verifies respawn starts a new 3-second enemy cooldown with the server timestamp. `117b68e` preserves the rule that defeated characters receive the defeated error before enemy respawn/range checks. Runs #440–#442 were pending.
- Remaining AI limitations: nearest-active-target selection is simple and not sticky; no threat table, pathfinding, party aggro, enemy abilities or boss phases. No client-authoritative AI ticks or damage are permitted.



## Latest skill-system implementation

- Commits: `588bcc9` (18-skill hierarchy), `be004cb` (hierarchy lookup), `7238d7a` (roster/mastery tests), `21d1c76`/`dd7a39a` (combat progression), `2400283`/`489767b` (combat tests), `8a43b66`/`32e639c`/`2e2fe95` (combat result attribution and XP correctness), `30fd8cf`/`a856a48` (Godot skill display and parent XP feedback), `1bab5b9` (skill roster documentation), `957336d` (combat progression docs), `f9e4a5f` (Jedi mastery design), `4fd6da1` (architecture), `c88039a` (world catalog test), `04d5ce2` (legacy backfill test), `116d32f`/`c03c93b` (PostgreSQL combat persistence test), `ca174c4` (Strength/Defense combat tests), `5b7912f` (README), `7e299ce` (roadmap), `c546fed` (boss/loot/crafting design).
- New PvE behavior: current training sword uses Blades; no supported weapon uses Unarmed. Successful hits train Strength; taking damage trains Defense; Strength contributes to melee damage and Defense mitigates incoming damage. Defeating an enemy awards the selected specialization XP and 25% parent XP. Combat responses report the actual specialization and parent XP awarded; overflow is not falsely reported.
- Existing legacy `combat` and `tactics` progress keys are preserved, not deleted or remapped. The Godot panel displays only the active 18-skill catalog. Verify old-character behavior in integration tests before any future data cleanup.
- Bosses, loot tables and deeper crafting are not being implemented ahead of the foundation. Their architecture and acceptance criteria are captured in `docs/ENCOUNTERS_LOOT_CRAFTING_DESIGN.md`; the roadmap places ordinary combat/enemy AI before boss encounters, and boss encounters before the dedicated loot/crafting expansion.
- Equipment requirements remain server-owned: the starter sword requires Blades 1 and leather armor requires Defense 1.

## Next engineering tasks

1. Finish CI verification for runs #440–#442 and correct any failures. Defeated-target, cross-zone, stale-presence, leash-return, respawn-reset and shared-cooldown coverage is in the integration suite.
2. Improve enemy target stickiness/threat selection and add bounded pathfinding; do this before party aggro or multi-enemy encounters.
3. Add server-owned enemy abilities and combat telegraph types, then combat animations/feedback in Godot.
4. Add durable encounter participation, contribution tracking and reward settlement before boss phases or group loot.
5. Implement reusable boss encounter phases after ordinary enemy behavior and reward settlement are reliable.
6. Add dedicated loot tables, idempotent reward claims and explicit inventory-overflow handling before valuable boss drops.
7. Expand crafting into tools, station and skill requirements, material properties, quality and production chains that consume boss materials.
8. Keep server/client workflows green; run trade abuse/load tests before enabling public production trade.
9. Defer hidden Jedi challenge implementation until progression persistence, concurrency and anti-tampering tests are ready; only the six active combat specializations are mastery candidates.
10. Complete production account safeguards: email verification, recovery, MFA, distributed rate limiting, admin account tools and security/load review.

## Continuation instruction

Always inspect this file, current repository HEAD and recent CI first. Work only in ViViD420/RSGalaxies-. Update this handoff after meaningful milestones. Never assume GalaxyMMO code or decisions apply here.


## Copy/paste continuation prompt

```text
RS GALAXIES — RESUME DEVELOPMENT. Continue work in ViViD420/RSGalaxies- on main. Read docs/SESSION_HANDOFF.md first, inspect latest HEAD and GitHub Actions statuses, and resume the exact unfinished task. Work autonomously with minimal questions. Improve existing code, commit each meaningful change, and update docs/SESSION_HANDOFF.md after every milestone. Never claim CI passed unless Actions confirms it. Current HEAD at handoff: 5245407d323509f079b3db2f67934543a83314bb. Current focus: finish server CI runs #440–#442 for unified enemy retaliation cooldowns, respawn reset, integration coverage and defeated-player error precedence; finish client run #128 for skipping defeated enemy interaction targets. Then improve enemy target threat/pathfinding, add encounter contribution/reward settlement, and progress toward bosses, loot and crafting. Godot polls authenticated enemy snapshots and interpolates positions. Preserve the 18-skill broad-progression design and profession-specific equipment. Never modify GalaxyMMO or reuse its proprietary code.
```

Before the chat ends, refresh this handoff with the newest commit and CI statuses and provide the continuation prompt again. Give the user an early warning when a conversation handoff is approaching; do not wait until the final message.
