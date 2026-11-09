# RS Galaxies — Session Handoff

Last updated: 2026-11-09 (boss snapshot and Godot HUD groundwork; CI pending)

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
- The active skill catalog now contains 18 skills with private parent/co-skill metadata. New characters receive all 18 at level 1; PostgreSQL loading backfills missing entries for older characters. Legacy `combat` and `tactics` XP keys remain preserved in existing records but are inactive, omitted from the client skill panel and cannot receive new XP. The starter sword selects Blades; unarmed attacks select Unarmed. Enemy defeats create durable pending XP/loot claims for eligible encounter participants; successful hits award Strength XP and taking damage awards Defense XP. Claims award specialization XP and one-quarter parent XP only when redeemed. Strength contributes to melee damage, Defense to mitigation and item-defined armor requirements. The six specializations alone are marked as future Jedi mastery candidates. Accuracy remains a derived outcome. See docs/SKILL_ROSTER.md and docs/COMBAT_PROGRESSION.md.
- Hidden Jedi progression remains a documented future feature in docs/JEDI_PROGRESSION_DESIGN.md. The six combat specializations—Blades, Polearms, Unarmed, Pistols, Rifles and Magic—are the only future mastery candidates; parent skills, Defense and Strength do not count. The hidden challenge, entitlement and Jedi character archetype are still not implemented or exposed. See docs/JEDI_PROGRESSION_DESIGN.md.
- Defensive in-memory repository copies and concurrency-safe experience awards.
- Starter zone, spawn point, server-time movement validation and world bounds.
- Tree, stone, copper and wild-herb gathering nodes with range checks, charges, respawn and atomic inventory/XP rewards. The Foraging skill feeds a new herbal-bundle workbench recipe. Godot chooses the nearest target and builds resource props from the server node catalog rather than hardcoded node IDs.
- Starter workbench and forge; recipes cover basic planks, herbal bundles, crafted healing herbal tea, copper smelting and a training sword. Herbal tea is a 15-health consumable. Crafting aggregates duplicate ingredients/outputs, simulates net stack changes, validates the exact station and commits inventory/XP only after all checks pass.
- Player inventory, 100-slot bank storage, coin wallet, persistent equipment slots, merchant healing potions and crafted herbal tea. New characters start with 100 coins.
- Starter merchant catalog with server-defined buy/sell prices. Buy/sell operations update coins and inventory atomically.
- PostgreSQL migrations 0001–0015 cover player state, gathering, accounts/sessions, character ownership, bank storage, currency, equipment, health, shared enemy spawn state, persistent zone chat, leased zone presence, player-trade escrow, persistent per-spawn enemy AI state, sticky enemy target identity, durable encounter-cycle/contribution records, durable enemy reward settlements/claims, per-participant combat specialization attribution and child item rows for multi-item claims. Migration 0015 backfills legacy one-item claim rows. Migration 0009 creates trade/participant/offer/escrow tables plus a unique active-character reservation table. Authenticated invite/accept, escrow offer replacement, two-sided confirmation, atomic settlement, cancellation, expiry release and active-trade recovery are now implemented. Invite, acceptance, offer edits and confirmations require fresh same-zone proximity/availability checks; Godot trade UI is now integrated; production routes are disabled unless `RS_GALAXIES_ENABLE_TRADING=true`, while production abuse/load testing remain.
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
## Sticky enemy target selection (new milestone)

- Migration `a11441f` adds nullable `target_character_id` to `enemy_spawn_state`; `1457cee` registers migration 11.
- `d66ce97` adds `TargetCharacterID` to `game.EnemyAIState`. `511d2fa` loads and persists target identity, prioritizes the current target in fresh-presence candidate ordering, and clears target identity when the enemy has no valid target. `7955d4e`, `08c5872` and `46ece55` keep player-engaged target identity consistent, including clearing defeated players; `ef0136b` clears the target when an autonomous attack defeats the player.
- `be61b2e` adds a PostgreSQL integration test: the nearest active player is initially acquired, then a different player moves closer; the enemy must continue moving toward the original target while that target remains active.
- `ae94942` verifies target assignment after player engagement and clearing after leash return. `f37ae26`, `877f35e`, and `c81358e` update combat, architecture and roadmap documentation. `3b44096` isolates the multi-player sticky-target integration test in a synthetic zone after shared starter-zone presence contaminated earlier runs. `9a75850` updates the boss/loot/crafting design to reflect the delivered ordinary AI prototype and remaining encounter reward prerequisites.
- CI runs #443 (migration), #444 (migration registration), #445 (state struct), #447 (sticky-state repository code) and #448 (player-engaged target persistence) passed. #446 failed on intermediate code before target scan variables/helper were completed; #447 corrected it. Sticky integration runs #449–#453 failed because unrelated starter-zone presence from other integration tests could be selected; #457 attempted cleanup and #458 moves the test to an isolated synthetic zone. Runs #457 and #458 passed. #457 cleared unrelated starter-zone presence; #458 further isolates the multi-player test in a synthetic zone. Sticky-target persistence/selection and target-loss switching are now CI-validated.
- Design boundary: this is single-target stickiness, not a threat table. It maintains one target while that character has fresh same-zone presence and remains in the leash selection area. If lost, the nearest eligible active character can be selected. No party aggro or contribution weighting exists yet.

## Encounter cycle and contribution ledger (new milestone)

- Migration `d256729` adds `encounter_id` to each enemy spawn state and creates `enemy_encounter_participants` with per-character damage dealt, damage received, action count and first/last contribution timestamps. `5df0675` registers migration 12.
- `933cc94` adds the internal `game.EncounterParticipant` model. `f408aee` adds server-generated encounter IDs, upserted contribution records and incoming-damage updates. `0a7cfd` assigns/rotates encounter IDs in the world tick, persists them per spawn life and records AI damage received for existing participants. `2259f73` records player attack contributions and stores the encounter ID in all player combat update paths. `139b122` adds the missing encounter-ID read to player combat. `04eddd9` exposes a repository query for a cycle's participants.
- `24c99c0` adds PostgreSQL coverage for unique encounter IDs across respawns and per-cycle contribution records. `d495d6f` asserts that autonomous AI damage updates an already-engaged participant's incoming-damage/action counters.
- Contribution records are created by player attacks; subsequent AI damage updates existing participants but does not make an unengaged bystander eligible. Party membership and support-action contribution types do not exist yet.
- Server CI runs #459–#463, #465 and #466 passed (migration, runner, model, helper, AI encounter-cycle assignment, account persistence and participant listing). #464 failed on an intermediate compile error (`encounterIDValue` referenced before SELECT/Scan update); `139b122` corrects it. #467/#468 failed because the contribution test omitted one return value from `AttackEnemyForAccount`; `4067119` fixes both assignments. #469 passed the corrected encounter contribution integration test and the full server suite. #470 passed: run `42554765267`, commit `7847f47`, including race-enabled Go/PostgreSQL integration tests and `go vet`. The subsequent threat-targeting commits are awaiting their own Actions run.
- Documentation: `07c50f5` updates the encounter/boss/loot design prerequisites; `5bd4638` updates combat documentation; `c9ba4ed` marks the ledger slice on the roadmap; `eeb2afb` documents encounter-cycle and participant persistence in the architecture.

## Latest verified repository state

- Server CI runs #413–#421 all passed, including migration 0010, restart recovery, enemy damage, transactional tick coordinator and snapshot model/repository.
- Server runs #422–#424 failed because the HTTP test double had not yet implemented the newly extended enemy-snapshot interface. This was a test-adapter compile failure; commit `97c5d41` added the fake implementation and #425 passed. #426 passed stale-presence leash-return coverage; #427 passed dynamic live-position combat. #428 failed because its integration assertion still expected the pre-player-attack enemy health; #429 corrected the expected value and passed. Runs #430 and #431 both passed, validating monotonic simulation time across replica clock skew. Runs #432 and #433 passed, covering defeated-target rejection and cross-zone target isolation. #434 passed the respawn-state reset implementation. #435 failed to compile because a local error channel shadowed the imported errors package; commit `fcc331d` renames the channel. Runs #436–#439 failed on the test-only channel shadowing, corrected by `fcc331d`. #440 compiled and reached the integration test but failed because it still expected a cleared last-update timestamp after respawn; the implementation correctly starts a new cooldown and timestamp, and #441 updates that assertion. #442 verifies defeated-player error precedence. Run #442 passed with the corrected test-channel name, respawn cooldown timestamp assertion and defeated-player error precedence. Run #441 was still in progress at the latest check; #442 covers the same code plus the final assertion. Verify Actions before claiming every intermediate commit green.
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
- Monotonic simulation time: `0a0cbe2` prevents a delayed tick from moving persisted simulation time backwards across server replicas; `9a64f8d` tests clock regression. Runs #430/#431 passed. `ebfa1a5` adds a PostgreSQL integration assertion that a defeated character with a fresh presence lease is not attacked and causes the enemy to return; #432 is pending. `49cfd8c` adds a cross-zone presence test proving an enemy does not pursue or damage a character after the character and presence move to another zone; #433 is pending. `d3eb26f` resets enemy position, behavior, telegraph and cooldown when player combat triggers an expired respawn; run #434 passed. `6b3d375` tests the respawn gate and verifies the next life starts at the catalog spawn with cleared transient timers; #435 exposed a test-only shadowing compile error, corrected by `fcc331d`. `cb76074` unifies click retaliation and autonomous AI cooldowns, cancelling a pending telegraph after immediate retaliation; `f981afa` tests the shared cooldown and resumed telegraph; `8024495` clears completed telegraph deadlines; `7179e36` asserts the domain transition clears them. `8049a9b` updates shared cooldown documentation; `8ec40a1` corrects the combat doc to state that `ReceiveEnemyAttack` is now called transactionally by the live world tick. `fcc331d` fixes the test-channel shadowing compile issue. `c4577c1` verifies respawn starts a new 3-second enemy cooldown with the server timestamp. `117b68e` preserves the rule that defeated characters receive the defeated error before enemy respawn/range checks. #440 failed on the stale timestamp assertion; #441/#442 were pending.
- Remaining AI limitations: nearest-active-target selection is simple and not sticky; no threat table, pathfinding, party aggro, enemy abilities or boss phases. No client-authoritative AI ticks or damage are permitted.



## Latest skill-system implementation

- Commits: `588bcc9` (18-skill hierarchy), `be004cb` (hierarchy lookup), `7238d7a` (roster/mastery tests), `21d1c76`/`dd7a39a` (combat progression), `2400283`/`489767b` (combat tests), `8a43b66`/`32e639c`/`2e2fe95` (combat result attribution and XP correctness), `30fd8cf`/`a856a48` (Godot skill display and parent XP feedback), `1bab5b9` (skill roster documentation), `957336d` (combat progression docs), `f9e4a5f` (Jedi mastery design), `4fd6da1` (architecture), `c88039a` (world catalog test), `04d5ce2` (legacy backfill test), `116d32f`/`c03c93b` (PostgreSQL combat persistence test), `ca174c4` (Strength/Defense combat tests), `5b7912f` (README), `7e299ce` (roadmap), `c546fed` (boss/loot/crafting design).
- New PvE behavior: current training sword uses Blades; no supported weapon uses Unarmed. Successful hits train Strength; taking damage trains Defense; Strength contributes to melee damage and Defense mitigates incoming damage. Defeating an enemy awards the selected specialization XP and 25% parent XP. Combat responses report the actual specialization and parent XP awarded; overflow is not falsely reported.
- Existing legacy `combat` and `tactics` progress keys are preserved, not deleted or remapped. The Godot panel displays only the active 18-skill catalog. Verify old-character behavior in integration tests before any future data cleanup.
- Bosses, loot tables and deeper crafting are not being implemented ahead of the foundation. Their architecture and acceptance criteria are captured in `docs/ENCOUNTERS_LOOT_CRAFTING_DESIGN.md`; the roadmap places ordinary combat/enemy AI before boss encounters, and boss encounters before the dedicated loot/crafting expansion.
- Equipment requirements remain server-owned: the starter sword requires Blades 1 and leather armor requires Defense 1.


## Encounter threat-based target selection (implemented; CI passed)

- #470 is confirmed successful: GitHub Actions run `42554765267` on commit `7847f47` completed with race-enabled Go/PostgreSQL integration tests and `go vet` all green.
- `9351fca` updates `AccountRepository.activeEnemyCandidates`: preserve the valid sticky target first; otherwise rank candidates by current encounter contribution score (2× damage dealt + damage received + action count), then squared distance and character ID for deterministic tie-breaking. The score is server-derived from the current encounter ID and participant ledger; no client values are trusted.
- `5cdc3be` adds `TestPostgresEnemyTargetSelectionPrefersEncounterThreat`, using an isolated synthetic zone and two characters to prove a farther high-threat participant outranks a nearer bystander when the enemy has no sticky target.
- `8920372` documents the target-selection score and explicitly distinguishes AI aggro scoring from reward eligibility. `b801932` updates the roadmap.
- These commits were pushed directly to `main`. Server Actions #471 (`42556821479`, commit `9351fca`) and #472 (`42556832552`, commit `5cdc3be`) both completed successfully. The threat-scoring query and PostgreSQL integration test are CI-validated.
- Important boundary: this is target-selection threat scoring only. Party identity, reward eligibility windows, anti-leech thresholds, threat decay/taunt rules and reward settlement execution are not implemented.

## Durable encounter reward settlement and claim execution (CI pending)

- `814ad3a` adds migration 0013 with `enemy_encounter_settlements`, uniquely keyed by `(enemy_id, encounter_id)`, and `enemy_encounter_reward_claims`, uniquely keyed by `(enemy_id, encounter_id, character_id)`. Claims snapshot character/account ownership, specialization and parent XP, optional item/quantity, pending/claimed state and timestamps. Foreign keys, nonnegative checks, item/quantity consistency and claimed-at consistency are database-enforced. Pending claims survive full inventories by design; claim execution is separate and not implemented yet.
- `bc8c74f` registers migration 0013 in the transactional migrator. `902eedb` adds the domain claim DTO. `c34476a` adds PostgreSQL integration coverage for pending-state defaults, settlement replay idempotency, duplicate-claim rejection and invalid claimed-state rejection.
- CI #473–#476 completed successfully: migration, migration registration, claim DTO and PostgreSQL schema/idempotency integration test are green. #477/#478 also passed the initial eligibility helper and unit tests.
- Migration 0014 (`06e15be`) persists the combat specialization associated with each participant; `3280aea` registers it, `210c096` extends the participant model, `e57c08c`/ `3438018` update persistence, and #479–#481 passed. `e5ac5d7`/ `23395ef` add the tested-in-code 10%-of-max-health and two-minute anti-leech policy. `baf5749`/ `48a58b2` add atomic settlement creation, contribution-weighted XP, deterministic single-drop assignment and the defeat-path integration test. `02c490c`/ `1b3cee8` remove direct killer-only XP/loot awards; `ea47749` wires settlement into the combat transaction. `e330aaf`/ `161b896` implement account-scoped claim listing and atomic idempotent redemption; `2d38bd8` adds authenticated API routes; `819b5e3` adds full-inventory retry and exactly-once claim tests; `315bcbf` fixes the test import. `f9a2733` ensures incoming enemy damage does not refresh the damage-based eligibility clock. #482 failed on a transient call-signature mismatch before #483 passed; #486/#488 failed on old direct-reward integration expectations, now updated by `4ca631d`; #491/#494 failed on missing test imports before subsequent fixes. Actions #495/#496 failed only on the old direct-reward integration assertion; #497 (run 42557738699, commit 4ca631d) passed the corrected full server suite, validating defeat-path settlement, claim creation and the full-inventory/idempotent claim flow. #498–#503 are validating HTTP reward list/claim coverage, stricter account/character ownership and test cleanup. They were in progress at last inspection; wait for latest commit's run #503 before declaring the API suite green. #489–#494 also include intermediate compile/import and stale-assertion failures corrected by later commits. Verify #497 and the newest server run before claiming settlement/claim execution green.



## Encounter reward eligibility policy (new; implementation CI pending)

- Commit `e5ac5d7` adds `game.IsEligibleForEncounterReward` and an explicit policy: verified damage dealt must reach at least 10% of the enemy's maximum health, and the participant's last contribution must be no older than two minutes. The policy rejects missing/mismatched encounter identity, zero actions, negative damage, future timestamps and stale contribution. Damage received or passive proximity alone never qualifies.
- Commit `23395ef` adds unit tests for the threshold boundary, stale/future timestamps, malformed identity, missing actions and passive damage-received anti-leech behavior. Actions for these commits have not yet been confirmed.
- This helper is not yet wired into enemy-defeat settlement. Existing `Player.AttackEnemy` still awards kill XP and immediate inventory loot to the killer; the next implementation must remove/replace that direct reward path in the same transaction that creates durable claims, or rewards could be duplicated. Support-only eligibility remains deferred until support actions are represented in the ledger.


 
## Encounter reward policy, settlement and claim API (implementation in CI)

- `game.IsEligibleForEncounterReward` requires verified damage of at least 10% of enemy max health and a player damage action within two minutes; passive presence, damage received alone, missing IDs, stale/future timestamps and zero-action records are rejected. The last-damage recency clock is not refreshed by enemy damage.
- Migration 0014 adds `experience_skill_id` to the participant ledger. This preserves which combat specialization generated damage so XP claims use the actual combat style.- On enemy defeat, the combat transaction inserts the unique settlement row, reads and locks participant rows, filters eligible contributors, splits base XP proportional to damage (integer remainder to highest damage), and assigns the single enemy drop to the highest-damage eligible participant. A settlement replay is a no-op.
- Direct killer-only kill XP and loot were removed from the domain combat path. The reward is represented by pending claims, avoiding duplicate direct grants plus claims.
- Authenticated account APIs: `GET /api/v1/account/characters/{id}/rewards` and `POST /api/v1/account/characters/{id}/rewards/{enemyID}/{encounterID}/claim`. Claiming locks the owned character and claim row, applies XP/parent XP and item inventory, and marks claimed in one transaction. Full item stacks return `reward_inventory_full` and leave the entire claim pending. Repeated successful claims are no-ops.
- The settlement integration test verifies defeat creates one pending claim; claim integration coverage verifies full-stack retry, XP not being consumed on failure, inventory/XP applied once and idempotent retries. Godot now has a Rewards panel with authenticated claim listing, per-claim redemption, inventory refresh and a full-stack retry message; commits d245318, 054de04 and 54749fc. Commit 37193c6 adds HTTP-level tests for authenticated claim listing and missing-claim error mapping; 2824e93 adds successful claim and foreign-character HTTP checks. Commits 0c900aa/4b2ef3a enforce and test account/character ownership on the database claim-list query; 654b7c2 moves cleanup earlier in the integration test to prevent leaked rows if assertions fail. Client runs #129–#132 all passed, including Godot 4.3 project import and script validation for the Rewards panel, panel lifecycle and cancellation of stale requests on character switch/sign-out. Party identity, support actions, loot tables and boss-specific policies remain deferred.


 
## Initial weighted enemy loot tables (server Actions #518 passed)

- The original single-result weighted tables were introduced in Actions #518 (run 42558564921, commit 2dd6867). The initial Frontier Rat and Wild Boar tables included 10% no-drop intervals; cryptographically secure server rolls are generated during the unique encounter settlement.
- The active live enemy definitions now use the version-1 guaranteed-plus-bonus tables: Frontier Rat guarantees 1 Rat Pelt and independently rolls a 20% chance of 1–2 Rat Teeth; Wild Boar guarantees 1 Boar Hide and independently rolls a 15% chance of 1–2 Boar Tusks. The no-drop interval applies to each bonus roll, not the guaranteed base item.
- The first rare-material crafting loops are complete. Rat-Tooth Blade and Boar-Tusk Spear recipes preserve the distinction between Smithing recipe requirements and Blades/Polearms equipment gates. Early intermediate commits had expected failures while definitions and catalog counts were being assembled; final server Actions #528 (run 42560199607, commit 8b373c7) passed the full server workflow for the combined item/recipe/catalog/merchant restriction checkpoint.

## Multi-item encounter reward pipeline (implementation integrated; CI-verified)

- server/internal/game/loot.go supports versioned legacy weighted tables and multi-drop definitions with guaranteed drops, independent weighted bonus rolls, implicit no-drop intervals, quantity ranges, duplicate-item aggregation and stack-limit validation. The legacy one-roll adapter remains available.
- Migration 0015 (f2e8695) creates enemy_encounter_reward_claim_items, keyed by (enemy_id, encounter_id, character_id, item_id), with a foreign key to the durable claim and cascade cleanup. It backfills all existing legacy loot_item_id/loot_quantity pairs before the new child-row reader is used. 5aebec9 registers migration 15.
- game.EncounterRewardClaim now exposes lootItems and retains lootItemId/lootQuantity as compatibility fields for older clients (59d19a2). PostgreSQL settlement uses cryptographically secure roll pairs for every guaranteed and bonus result and attaches all item lines to the highest-damage eligible contributor (0390ccc). Claim listing aggregates all item rows; redemption locks the claim and every child row, validates every stack before any inventory or XP mutation, and commits all item lines, XP and claimed state atomically. Any full stack leaves the entire claim pending for retry. The API and Godot Rewards panel expose/render every item line while retaining legacy payload fallback (01cbb07, df6bb4b, 937124f).
- Live Frontier Rat and Wild Boar definitions now use frontier-rat-harvest and frontier-boar-harvest (45e016e). PostgreSQL settlement coverage now asserts the guaranteed pelt and optional rare tooth (21919cb); multi-item full-inventory retry/idempotency and legacy migration backfill/idempotency tests are included (7ab30e8, 9af2b09).
- CI correction: Actions #529–#532 failed because TestLootTableCatalogIsValidAndDefensivelyCopied assumed legacy tables were at index 0 even though new definitions were inserted first. 0a1e3cf changed the test to locate definitions by stable ID; Actions #533 (run 42586370088) passed. Migration file/registration Actions #534/#535, DTO #536, persistence #537, integration #538, HTTP #539, live enemy switch #540, updated settlement assertions #541 and migration backfill/idempotency #542 all passed. Key final server runs: #541 (run 42586790858, commit 21919cb) and #542 (run 42586828458, commit 9af2b09). Client Rewards UI Actions #133/#134 also passed. The multi-item reward pipeline is CI-validated end to end.
- The full pipeline is integrated and CI-validated by server runs #541/#542 and client runs #133/#134. Do not switch the live enemy tables back to legacy single-drop definitions; they are already on the multi-drop tables.

## Boss encounter phase engine foundation (domain-only; CI verification pending)

- Added `server/internal/game/boss.go` with versioned server-owned boss definitions, validation, defensive catalog copies and a deterministic encounter state machine. The Stoneback Colossus prototype has Guarded (Stone Fist) and Rampage (at or below 50% health; Stone Fist + Ground Shatter) phases.
- The engine resets on a new encounter-cycle ID or definition version, preserves monotonic time within a cycle, keeps phase progression monotonic, starts abilities only when range/cooldown permit, interrupts telegraphs when the target leaves range while retaining cooldown, emits server-calculated damage intents after telegraph completion, and treats defeat as terminal for the cycle.
- Added `server/internal/game/boss_test.go` coverage for catalog defensive copying, invalid definitions, threshold transitions, telegraph timing/resolution, range interruption/cooldowns, new-cycle reset, clock monotonicity, terminal defeat, malformed persisted telegraphs, negative ability cursors and cooldown-map immutability.
- This is deliberately domain-only: Stoneback Colossus is not yet a live enemy; boss state is not yet persisted, damage intents are not yet integrated into the PostgreSQL AI transaction, and Godot does not yet display boss phase/ability state. Do not describe it as a playable boss.
- Relevant commits: `8b584ac` (engine), `4427d6c` (new-cycle clock reset), `fb06bbd` (initialize phase and ability together), `ed6ecfc` (persisted-state validation/cursor safety), `7e90850` (initial tests), `69b5d96` and `0a6d6eb` (test corrections/edge coverage), `5360c24` (cooldown-map immutability test). Server Actions #543 (run 42587425789, commit 8b584ac), #544 (run 42587461425, commit 4427d6c), #548 (run 42587565619, commit ed6ecfc) and #550 (run 42587608297, commit 5360c24) passed. Intermediate runs #545/#546 failed because the first test expectations did not yet match the subsequent initial-phase/reset and interruption-state corrections; those assertions were corrected in later commits #547/#549. The complete domain engine and current test suite are CI-verified by #550; #547/#549 are redundant intermediate runs still completing.

## Next engineering tasks

1. Persist the boss phase/ability state per encounter cycle (migration 0016), integrate boss ability damage with the existing transactional enemy AI/player-health update, expose phase/ability telegraph fields in enemy snapshots and Godot, then ship the first live boss vertical slice using the multi-item reward pipeline.
2. Expand party contribution rules and explicit support/healing contribution types before enabling group-boss reward eligibility. Current anti-leech policy remains damage-only (10% max-health threshold, two-minute active-damage window).
3. Add bounded pathfinding and party aggro; sticky targeting and current-cycle threat ordering are already implemented.
4. Expand crafting into tools, station and skill requirements, material properties, quality and production chains that consume rare and future boss materials.
5. Keep server/client workflows green; run trade abuse/load tests before enabling public production trade.
6. Defer hidden Jedi challenge implementation until progression persistence, concurrency and anti-tampering tests are ready; only the six active combat specializations are mastery candidates.
7. Complete production account safeguards: email verification, recovery, MFA, distributed rate limiting, admin account tools and security/load review.

## Continuation instruction

Always inspect this file, current repository HEAD and recent CI first. Work only in ViViD420/RSGalaxies-. Update this handoff after meaningful milestones. Never assume GalaxyMMO code or decisions apply here.


## Copy/paste continuation prompt

```text
RS GALAXIES — RESUME DEVELOPMENT. Continue only in ViViD420/RSGalaxies- on main. Read docs/SESSION_HANDOFF.md completely, inspect latest HEAD and GitHub Actions before changing anything. Work autonomously, improve existing code, commit meaningful milestones and update this handoff after each milestone. Never claim tests passed unless GitHub Actions confirms it. Preserve the broad 18-skill progression and profession-specific equipment. Threat scoring Actions #471/#472 passed; durable settlement/atomic claims #497 passed; HTTP/API ownership checks #498–#503 passed; Godot Rewards UI #129–#132 passed; initial weighted loot Actions #518 passed; final rare-material crafting/catalog checkpoint #528 passed. Multi-item reward claims are integrated: migration 0015 creates/backfills child item rows; settlement generates cryptographically secure guaranteed + independent bonus rolls; API DTO exposes lootItems with legacy fields retained; listing aggregates item rows; redemption locks and validates all stacks before any XP/inventory mutation; full-stack retry and idempotency are preserved; Godot Rewards panel renders all item lines. Live Frontier Rat and Wild Boar now use the guaranteed-plus-rare-bonus tables. Loot test correction Actions #533 passed; migration file/registration #534/#535 passed; DTO #536 passed; persistence #537 passed; integration #538 passed; HTTP #539 passed; client #133/#134 passed; final live multi-drop server checkpoint #541 passed. Action #542 (run 42586828458, commit 9af2b09) passed migration backfill/repeat-migration idempotency. Next persist and integrate the domain-only boss phase engine, expose boss telegraphs in the client and ship the first live boss vertical slice; then add party/support contribution rules before group-boss rewards. The domain engine commits are 8b584ac, 4427d6c, fb06bbd, ed6ecfc, 7e90850, 69b5d96, 0a6d6eb and 5360c24; Actions #543/#544/#548/#550 passed; #545/#546 were intermediate test failures subsequently corrected; #547/#549 are redundant intermediate runs. Never modify GalaxyMMO or reuse its proprietary code.
```

Before the chat ends, refresh this handoff with the newest commit and CI statuses and provide the continuation prompt again. Give the user an early warning when a conversation handoff is approaching; do not wait until the final message.

---

## Latest verified repository snapshot (2026-11-09; authoritative over older status notes above)

### Repository and commit state
- Repository: `ViViD420/RSGalaxies-`; branch: `main`.
- Latest inspected HEAD: `e9764017c281897cc2a8c9d6d446919dfed0788e` — “Mark boss phase engine domain milestone CI-verified” (documentation-only status update). The immediately preceding documentation commits are `4d6e9955ba3ff119f84a9ec614c1c5f285f16a91` (“Record CI-verified boss domain engine”) and `d4a58e7320b6e99d244aa0d63400ef4db49d9e51` (“Record boss engine intermediate CI failures and corrections”). The boss implementation commits are listed in the boss-engine section above.
- No boss persistence, transactional boss damage integration, boss snapshot fields, Godot boss UI, or live Stoneback Colossus implementation is present in this verified snapshot. The engine remains a domain-only foundation.
- The existing migration inventory in this handoff records migrations 0001–0015, with 0015 supporting multi-item reward claims. Migration 0016 is the anticipated next migration, but before creating it, inspect the actual migration directory and migration registry on current HEAD to confirm there is no later/unregistered migration or naming conflict.

### Verified GitHub Actions
- Server Action #550, run `42587608297`, commit `5360c24`: completed successfully. The job and both substantive steps were explicitly checked: “Test (race detector and PostgreSQL integration)” = success; “Vet” = success.
- Server Action #549, run `42587568842`, commit `0a6d6eb`: completed successfully; race-enabled PostgreSQL test step and vet both succeeded.
- Server Action #548, run `42587565619`, commit `ed6ecfc`: completed successfully; race-enabled PostgreSQL test step and vet both succeeded.
- Server Action #547, run `42587531770`, commit `69b5d96`: completed successfully; race-enabled PostgreSQL test step and vet both succeeded.
- Earlier intermediate Actions #545/#546 failed on test expectation mismatches; those failures were corrected in subsequent commits. Keep that history visible; the later corrected runs above passed.
- Server Action #470, run `42554765267`, commit `7847f47`: rechecked and completed successfully; race-enabled PostgreSQL tests and vet both succeeded.
- Multi-item reward pipeline remains CI-validated by server Actions #541/#542 and client Actions #133/#134, as detailed above. Do not regress the live guaranteed-plus-bonus loot definitions.
- The latest HEAD is a documentation commit after the tested boss implementation commit. Do not describe an absent workflow result for the documentation-only HEAD as a new code-test pass; the strongest verified boss-engine code checkpoint is Action #550.

### Exact next task
Implement the first boss integration vertical slice without restarting or replacing existing systems:
1. Inspect migration files and registration on current HEAD; confirm the next migration number before authoring it.
2. Persist boss state under the existing enemy encounter-cycle identity, including definition ID/version, phase, active ability/telegraph start and resolution timing, cooldowns/cursor, and terminal defeat state. Design for restart recovery, concurrent world workers, stale-cycle isolation and idempotent transitions.
3. Integrate `AdvanceBossEncounter` into the existing server-owned enemy AI transaction. Apply resolved ability damage, player health and Defense XP atomically with boss/AI state; preserve range, telegraph, interruption and cooldown semantics; prevent duplicate damage on retries and prevent defeated enemies attacking.
4. Extend the authenticated enemy snapshot with server-owned boss identity, phase, active ability/telegraph and timing. Preserve ordinary enemy payload compatibility.
5. Add Godot boss health/name, phase and telegraph/ability warning feedback driven only by server snapshots. Preserve the existing Rewards UI.
6. Only after persistence, transaction, snapshots, client rendering and end-to-end CI are verified should Stoneback Colossus become a live encounter. Route rewards through the existing multi-item claim pipeline.
7. Keep the current anti-leech policy explicitly damage-only (10% of enemy max health and recent damage within two minutes); do not silently add support eligibility. Party/support contribution work is a later milestone.

### Fresh-chat start instruction
Use this exact prompt in the new chat:

```text
RS GALAXIES — RESUME DEVELOPMENT FROM THE REPOSITORY HANDOFF.

Work only in ViViD420/RSGalaxies- on branch main. First read docs/SESSION_HANDOFF.md completely, then inspect the live main HEAD, recent commits, migration files/registry and GitHub Actions results. The latest handoff snapshot records HEAD e9764017c281897cc2a8c9d6d446919dfed0788e. Do not assume that SHA is still HEAD; fetch it again.

Continue the existing project; do not restart, recreate or replace it. Do not touch GalaxyMMO or reuse its proprietary code. Work autonomously with minimal questions, improve existing code, commit meaningful milestones, and update docs/SESSION_HANDOFF.md after every meaningful progression. Never report tests as passing without checking the completed GitHub Actions run and job steps.

Current milestone: the boss phase engine in server/internal/game/boss.go is implemented and CI-verified as domain logic only. Server Actions #547 (42587531770), #548 (42587565619), #549 (42587568842) and #550 (42587608297) all completed successfully; #550's race-enabled PostgreSQL integration tests and go vet passed. Earlier #545/#546 failed on intermediate test expectations and were corrected by later commits. Server Action #470 (42554765267) was also rechecked and passed.

Next implement durable cycle-scoped boss-state persistence (verify the next migration number; migration 0015 is the last recorded migration), integrate telegraphed boss damage atomically with existing enemy AI/player-health/Defense-XP transactions, expose authoritative boss phase/telegraph data in enemy snapshots, and build the Godot boss HUD/warnings. Only then register Stoneback Colossus as a live encounter and route its rewards through the existing CI-verified multi-item claim system. Do not call the boss playable until the full vertical slice is complete and verified.

Preserve the broad 18-skill progression, profession-specific equipment, live Frontier Rat/Wild Boar multi-drop tables, durable reward claims, and existing Rewards UI. Current anti-leech eligibility is damage-only: >=10% enemy max health and damage within two minutes. Party/support contribution rules remain future work.
```


---

## Latest development checkpoint — durable boss persistence groundwork (2026-11-09)

### Repository state
- Repository: `ViViD420/RSGalaxies-`; branch: `main`.
- Latest inspected pre-handoff HEAD: `ddb32de766354e992ec9dfdddf932db2c8c9f630` (architecture documentation update). This handoff update will advance main again; always fetch live HEAD before continuing.
- Migration inventory and registry were inspected directly. Migration 0015 is the latest pre-existing migration and migration 0016 was free; no later migration was found.
- Recent implementation commits:
  - `59c3992e0fa2f962912002a6969076a49e040bcf` — add `0016_boss_encounter_state.sql`.
  - `ab2160cb6010ef81b33503d4d68650a36e1f77ba` — register migration 0016 in `Migrate`.
  - `76dd39e2a9cba3e4af8322ede6fd8a3cf281b7b2` — add transactional boss-state persistence helpers.
  - `32455a4146c6730eff80bf3a6672c9bc52326105` — reject stale boss-state writes when the conditional upsert affects zero rows.
  - `dabc0bebe8da7978305c6f28bfff7cf270e07441` — add PostgreSQL integration coverage for state round-trip, telegraph/cooldown recovery, and stale-cycle rejection.
  - `8af3bad2e30c5e798f9f305b0ffaab3a8f7adb0d` — update roadmap.
  - `ddb32de766354e992ec9dfdddf932db2c8c9f630` — update architecture.
- Migration 0016 creates `boss_encounter_states`, keyed by `(enemy_id, encounter_id)`, with definition ID/version, phase, phase start, active ability, telegraph start/deadline, JSONB cooldown map, next-ability cursor, last-updated timestamp and terminal defeated flag. The row references `enemy_spawn_state` with cascade cleanup.
- `boss_persistence.go` provides transaction-scoped load/save helpers. Both lock the parent `enemy_spawn_state` row first, verify the supplied cycle matches the currently active cycle, and lock the boss-state row on load. Cooldown JSON is decoded and validated. Save uses a cycle-scoped upsert with a monotonic timestamp condition and returns an error if a stale write is rejected.
- `boss_persistence_integration_test.go` verifies full state round-trip (including active telegraph and multiple cooldowns), no state inheritance by a new cycle, and rejection of stale-cycle loads/writes.
- `docs/ROADMAP.md` and `docs/ARCHITECTURE.md` now describe the persistence groundwork and its incomplete integration status.

### CI status — do not overstate
- The combined status lookup for implementation/test commit `dabc0bebe8da7978305c6f28bfff7cf270e07441` returned no status entries.
- No new Actions run/job result has yet been verified for migration 0016, persistence helpers or their integration test. The previous verified boss-domain checkpoint remains Server Action #550 (run `42587608297`), whose race-enabled PostgreSQL test step and `go vet` passed. This does not verify the new code.
- Trigger/inspect the server workflow after the next implementation commit and check the actual run conclusion plus job-level steps before claiming tests pass.

### Remaining boss integration work
1. Wire `loadBossEncounterStateForUpdate` and `persistBossEncounterState` into `AccountRepository.TickEnemyAI` inside its existing player-then-enemy transaction, but only for a registered boss enemy.
2. Resolve boss ability damage with a dedicated server-authoritative player damage method using the existing Defense/armor calculation. Do not fake a boss as a normal enemy or change ordinary enemy behavior.
3. Ensure player health, Defense XP, boss state and AI state commit atomically. The parent enemy row lock and encounter-cycle ID are the serialization boundary; test concurrent ticks and retry behavior so a resolved telegraph damages exactly once.
4. Reset boss state only when a new enemy encounter cycle is created; do not reset it on ordinary ticks or stale worker retries.
5. Extend `game.EnemySnapshot` and its PostgreSQL query with optional boss identity, phase, active ability and telegraph timing; ordinary enemy JSON must remain backward-compatible.
6. Extend the existing Godot snapshot rendering/polling with boss name/health/phase/telegraph warnings. Preserve Rewards UI and all 18 skills.
7. Only after those pieces are integrated should Stoneback Colossus be added to live enemy catalog/spawn lifecycle and reward settlement. Reuse the existing multi-item reward pipeline unchanged.

### Next-chat continuation prompt
```text
RS GALAXIES — RESUME DEVELOPMENT FROM THE REPOSITORY HANDOFF.

Continue exclusively in ViViD420/RSGalaxies- on main. First read docs/SESSION_HANDOFF.md completely, fetch live HEAD/recent commits, inspect GitHub Actions run/job results and verify the migration registry. Do not assume the HEAD recorded in this file is still current.

Migration 0016 and cycle-scoped boss persistence groundwork are committed. Inspect:
- server/internal/storage/postgres/migrations/0016_boss_encounter_state.sql
- server/internal/storage/postgres/migrate.go
- server/internal/storage/postgres/boss_persistence.go
- server/internal/storage/postgres/boss_persistence_integration_test.go

The state table stores definition/version, phase/timing, active ability/telegraph, cooldown JSON, ability cursor, last update and defeat state under (enemy_id, encounter_id). Transactional helpers lock the parent enemy spawn row, reject stale cycles and reject stale timestamp writes. Integration coverage was added, but no GitHub Actions result has yet been verified for these new changes. Do not claim CI passed.

Next integrate the helpers into the existing AccountRepository.TickEnemyAI transaction. Keep the established player-then-enemy lock order. Add boss-only server-calculated damage with Defense/armor handling, and commit player health, Defense XP, AI state and boss state atomically. Verify exactly-once damage under concurrent workers/retries and reset only on a new encounter cycle. Then add optional authoritative boss fields to EnemySnapshot and render the boss HUD in the existing Godot snapshot path. Finally register Stoneback Colossus as a live encounter and route defeat rewards through the existing multi-item reward claim pipeline.

Preserve all existing systems, 18-skill progression, crafting, live multi-item enemy loot and Rewards UI. Never modify GalaxyMMO or reuse proprietary code. Improve existing code, commit meaningful milestones, update this handoff after every meaningful progression, and never claim tests passed without checking GitHub Actions results.
```

## Boss persistence validation hardening (2026-11-09)

- Live main advanced from b89fcd58aeaef5ceeeaa71a74a6de4db4ba3a2b7 with two focused commits:
  - d6c12208a3b165da15206d71b4585e2f46b98954 — validBossTelegraphState now requires an inactive ability to have both timestamps zero, and an active ability to have both timestamps populated with a non-inverted interval. This closes a malformed-state acceptance gap in both load and save validation.
  - aac8771ca17c1d52fa7cc247d584546e50771af8 — adds table-driven regression coverage for inactive state, valid active telegraph, missing start/deadline, inverted timestamps and timestamps without an active ability.
- These changes are code/test additions only; no test command has been executed in this environment and no GitHub Actions run for these commits has been verified. The last verified server run remains Action #550 (run 42587608297, commit 5360c24), with race-enabled PostgreSQL tests and go vet successful. It predates migration 0016 and the subsequent persistence changes.
- The durable persistence helpers remain groundwork and are not yet wired into TickEnemyAI. Boss state is not yet included in snapshots or rendered by Godot; Stoneback Colossus is not a live spawn.
- Review note: next integration must preserve player-then-enemy lock ordering, use the boss ability's own maximum range (Ground Shatter reaches farther than the ordinary enemy interaction radius), apply Defense/armor mitigation, and commit player health/Defense XP plus AI/boss state in one transaction. Do not activate the encounter until end-to-end server/client CI validates the full path.

### Immediate continuation
1. Fetch live HEAD and all recent commits; check any available Actions runs/jobs. Re-read this handoff completely.
2. Run/verify the server suite against a PostgreSQL test database after migration 0016 and the new regression test; do not treat Action #550 as validation for these commits.
3. Inspect existing enemy target transaction, EnemySnapshot definition/JSON tests, combat damage/armor helpers and Godot snapshot polling/rendering. Implement a dedicated boss damage calculation and boss-only tick branch while retaining normal enemy behavior.
4. Add concurrent-tick/exactly-once and cycle-reset integration tests before enabling a live boss.
5. Update this handoff after each milestone and provide the latest copy/paste continuation prompt.

### Current continuation prompt
~~~text
RS GALAXIES — RESUME DEVELOPMENT. Work only in ViViD420/RSGalaxies- on main. Read docs/SESSION_HANDOFF.md completely, inspect live HEAD/recent commits and GitHub Actions jobs before changes. Latest known code commits: d6c12208a3b165da15206d71b4585e2f46b98954 (strict complete boss telegraph validation) and aac8771ca17c1d52fa7cc247d584546e50771af8 (regression tests). These commits have not yet been verified by CI. Server Action #550 (42587608297) is the last verified server checkpoint, and predates migration 0016 and the persistence code.

Continue the existing boss vertical slice: integrate cycle-scoped boss persistence into TickEnemyAI, keep player-then-enemy lock order, apply boss ability damage with ability-specific range and existing Defense/armor mitigation, and commit player health/Defense XP plus AI/boss state atomically. Test concurrent ticks, exactly-once damage, stale-cycle rejection and respawn reset. Then add authoritative boss fields to EnemySnapshot, render the boss HUD/telegraph warnings in Godot, and only then activate Stoneback Colossus using the existing multi-item reward pipeline. Preserve all existing systems and the 18-skill progression. Never modify GalaxyMMO or reuse proprietary code. Improve existing code, commit milestones and update this handoff every time. Never claim tests pass without actual Actions job evidence.
~~~

## Boss-specific damage groundwork (2026-11-09)

- Additional code commits:
  - b45319ee4a9fd7fc669f06a1dd2516d8c83013d6 — adds Player.ReceiveBossAttack, which accepts the ability's server-owned minimum/maximum range and raw damage, validates the ability parameters and same-zone distance, then applies existing equipment Defense bonuses and Defense-level mitigation. It awards Defense XP only when positive damage is received and updates LastCombatAt.
  - 58977533977a82cb34ebec3b7a243aa4880cd379 — adds domain tests for Ground Shatter reaching a player beyond ordinary melee range, Defense mitigation/XP, out-of-range rejection and invalid ability parameters without player mutation.
- These commits complement the strict telegraph validation commits d6c1220 and aac8771. No local tests were run and no Actions run/job result has been verified for these code changes. Latest verified server Action remains #550 (42587608297); it predates this work.
- Important boundary: ReceiveBossAttack is a domain primitive only. TickEnemyAI does not yet call it, boss state persistence is not yet wired into the tick, and the Stoneback Colossus is not live. Do not report the boss as playable.
- Current latest known implementation commit: 58977533977a82cb34ebec3b7a243aa4880cd379. Documentation updates will follow this checkpoint.

### Next steps
1. Inspect latest main and any available Actions status/run data. Re-read the entire handoff before the next milestone.
2. Integrate the boss state machine and ReceiveBossAttack into the existing player-then-enemy transaction, using ability-specific range and atomic player/boss/AI persistence.
3. Extend player attack/defeat transaction paths so a boss killed by a player immediately receives terminal boss state and cannot attack during the same cycle.
4. Add PostgreSQL concurrency/retry, defeat and respawn-cycle tests; verify actual server CI.
5. Extend EnemySnapshot and Godot rendering, then register Stoneback Colossus only after full CI verification.

## Authoritative boss snapshot groundwork (2026-11-09)

- Added optional boss fields to game.EnemySnapshot: boss identity/version, phase ID/start time, active ability and telegraph start/deadline.
- PostgreSQL enemy snapshot retrieval now LEFT JOINs boss_encounter_states using both enemy_id and the currently active encounter_id. Old-cycle boss data cannot leak into the current snapshot.
- Boss timestamp fields are pointers so ordinary enemy snapshots omit all optional boss timestamp keys rather than emitting zero-time placeholders. Regression tests verify ordinary snapshots omit boss keys and boss snapshots serialize the supplied phase/ability/timing.
- Commits:
  - 2c7abb53e3a01848df77ab6faab725eb80fe661c — add boss snapshot fields.
  - a44fc761bf983f78d7a3c1407cea067281ef647d — make optional boss timestamps pointer-backed.
  - 1dc7623261eb06238cd0b8e076ea27806253be89 — populate optional timestamps only from non-NULL database values.
  - 3c98572b1392c8a27f72926145df979faeb0c665 — add JSON contract regression tests.
- This snapshot groundwork is not CI-verified and is not yet connected to a live boss. Since Stoneback Colossus is not in the enemy catalog and TickEnemyAI does not persist/advance boss state, current snapshots will continue to contain ordinary enemy data only.
- Latest verified server Actions checkpoint is still #550 (42587608297); no current status checks are exposed for the new code commits. Do not claim tests passed.

### Next integration steps
1. Fetch live HEAD and check current Actions results before edits.
2. Wire boss state into TickEnemyAI and player-initiated defeat paths. Keep the player-then-enemy lock order and persist player health, Defense XP, enemy AI and boss state in one transaction.
3. Add PostgreSQL tests for one boss ability resolving once under concurrent ticks, stale worker rejection, player defeat, boss death, and a fresh state after respawn.
4. Render the optional fields in Godot only after the authoritative transactional server path is integrated; preserve Rewards and all existing HUD panels.
5. Only then register Stoneback Colossus and reuse the existing multi-item reward claims.

## Godot boss HUD groundwork (2026-11-09)

- Godot main.gd now recognizes the optional server boss snapshot fields and renders boss identity, phase and active ability in the world-space enemy label.
- Added a centered boss encounter panel with server-sourced health progress, phase and ability warning. It appears only while an active boss snapshot is present and hides when no live boss is included.
- Client change commit: 8894a3399892b4337410cc08b47ae9f7b34a790c.
- Snapshot/domain groundwork preceding it: 2c7abb5, a44fc76, 1dc7623 and 3c98572. Boss-specific damage groundwork: b45319e and 5897753. Telegraph validation: d6c1220 and aac8771.
- The HUD is intentionally passive: it does not calculate boss health, phase, ability choice or telegraph timing locally. It only renders authoritative snapshot values.
- This does not make Stoneback Colossus playable. The enemy catalog does not yet include the boss, TickEnemyAI does not yet advance/persist boss state, and player-initiated boss defeat is not yet connected.
- No CI job results have been verified for these changes. The last verified server checkpoint remains Action #550 (run 42587608297); no new server/client tests are claimed.

### Next work
1. Integrate the boss engine into the world tick and all boss defeat paths in PostgreSQL transactions.
2. Add PostgreSQL integration tests for exactly-once ability damage under concurrent workers, range interruption, stale-cycle rejection, death and respawn.
3. Verify the server and Godot client workflows and inspect their individual test/import steps.
4. Register the live Stoneback Colossus only after the complete server/client path passes.
