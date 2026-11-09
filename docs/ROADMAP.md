# RS Galaxies — Roadmap

## Milestone 1: Independent foundation

- [x] Establish standalone repository and original project identity.
- [x] Document architecture, design pillars and system boundaries.
- [x] Add Go world API, deterministic unit tests, race checks and vet.
- [x] Add Godot 4.3 third-person client, headless import and Web export validation.
- [x] Add PostgreSQL migrations, integration tests and Codespaces development database.
- [x] Add hosted Codespaces development-container configuration.

## Milestone 2: Account and world slice

- [x] Character identity, account registration/login/logout and bearer-session lifecycle.
- [x] PostgreSQL-backed account and character ownership; account-scoped character APIs.
- [x] First world zone, coordinates, spawn point and server-validated movement.
- [x] Godot account UI, multi-character selection and authenticated gameplay requests.
- [x] Interactable gathering nodes, workbench, forge, bank, merchant and enemy spawn points.
- [x] Safe-spawn recovery and session-expiry recovery.
- [x] PostgreSQL-backed leased character presence and account-scoped zone roster API.
- [x] Godot polling heartbeat and remote-character placeholder visuals with stale-lease cleanup.
- [ ] Reconnect recovery, server-driven movement snapshots and fully authoritative client replication.
- [ ] Email verification, password recovery, MFA and distributed abuse controls before public registration.

## Milestone 3: Skills and progression

- [x] Eighteen active core skills: Melee/Ranged parents, six combat specializations, independent Defense/Strength co-skills, four gathering and four production skills; level 1–99 progression. Legacy Combat/Tactics XP is preserved but inactive.
- [x] Authenticated live skills panel with per-skill XP progress and next-level thresholds.
- [x] Tree, stone, copper and wild-herb gathering with range, skill, charge and respawn rules.
- [x] Godot selects gathering targets and builds resource props from the server node catalog.
- [x] Consolidated public world catalog reduces startup round trips and keeps client definitions server-owned.
- [x] Godot builds and selects workbench/forge stations from the server crafting-station catalog.
- [x] Workbench plank, herbal-bundle and herbal-tea recipes plus forge copper smelting/training-sword recipes.
- [x] Cooking profession progression through crafting a server-defined healing tea consumable.
- [x] First rare enemy material sink: Rat Tooth + Copper Bars + Plank → Smithing level 2 Rat-Tooth Blade, equippable only at Blades level 2 (final combined crafting/catalog server Actions #528 passed).
- [x] Second rare enemy material sink: Boar Tusks + Copper Bars + Plank → Smithing level 3 Boar-Tusk Spear, equippable only at Polearms level 2 (final combined crafting/catalog server Actions #528 passed).
- [x] Transactional crafting with ingredient consumption, output capacity and skill XP.
- [x] Basic PvE enemy definitions, server-calculated damage, cooldowns, retaliation, specialization/parent/co-skill XP, starter loot and respawns.
- [x] Player health, defeated-state action restrictions and safe-spawn recovery.
- [x] Persistent server-owned enemy AI: aggro, chase, attack telegraphs/interruption, leash and return-to-spawn behavior; 250 ms PostgreSQL world worker, fresh-presence target selection, transactional enemy attacks and row-lock concurrency.
- [x] Authenticated zone-scoped enemy snapshots; Godot polls, interpolates enemy movement and displays health, behavior, telegraphs and respawn state. Player attacks use persisted live enemy positions.
- [x] Persisted sticky enemy target identity: maintain a valid active target between ticks and select the nearest eligible character when that target is lost.
- [x] Unique encounter-cycle IDs per enemy spawn life and durable per-character damage/action contribution ledger; autonomous incoming damage updates existing participants.
- [x] Threat scoring for enemy target selection (active sticky target first; otherwise current-cycle contribution score, then distance); PostgreSQL integration test passed in server Actions #472.
- [x] Initial server-side anti-leech policy and atomic ordinary-enemy settlement: 10% enemy-max-health damage threshold, two-minute active-damage window, contribution-weighted XP and one deterministic loot recipient; server Actions #497 passed.
- [ ] Party membership, healing/support contribution types and group/boss reward distribution.
- [x] Migration 0013: durable per-encounter settlement keys and per-character pending/claimed reward-claim records with uniqueness, ownership, quantity and state constraints; migration/registration/model/integration Actions #473–#476 passed.
- [x] Migration 0014 persists each participant's combat specialization for XP attribution; Actions #479–#481 passed.
- [x] Authenticated reward-claim API, account-scoped listing, idempotent redemption and full-stack pending behavior; settlement/claim integration suite passed in #497.
- [x] Authenticated HTTP API ownership/success-path tests (#498–#503) and Godot Rewards panel follow-up CI (#129–#132).
- [ ] Bounded pathfinding, party aggro, enemy abilities, weapon-family expansion and server-derived hit/miss outcomes (see docs/COMBAT_PROGRESSION.md).
- [ ] Expand the initial melee swing into weapon-family animations, hit reactions and ability effects. Initial server-confirmed placeholder swing and bounded third-person camera zoom are implemented; headless verification is in progress: Actions #164/#165 passed import/export; #166/#167 exposed an animation completion timing issue; #168 passed the animation assertions but exposed a retained-tween renderer shutdown error. Tween release did not resolve the renderer error. Replaced the tween with an allocation-free physics-driven sine swing and visible-motion/neutral-pose tests in 136f497/f948208; Client Action #172 passed the allocation-free visible-swing and neutral-pose runtime checks without Godot errors.
- [x] Reusable versioned boss phase/ability domain engine with validated definitions, threshold transitions, telegraphs, cooldowns, range interruption, encounter-cycle reset and deterministic damage intents (domain-only; server Actions verification pending).
- [ ] Durable boss encounter persistence groundwork: migration 0016 adds encounter-cycle-scoped state, versioned definition/phase, telegraph timestamps, JSONB cooldowns, ability cursor and terminal defeat state; PostgreSQL load/save helpers lock the parent enemy cycle and reject stale-cycle or stale-timestamp writes. Integration test added; GitHub Actions verification pending, and the world tick has not yet been wired to these helpers.
- [x] Initial server-owned weighted ordinary-enemy loot tables with cryptographically secure rolls and no-drop outcomes (server Actions #518 passed; later superseded in live enemy definitions by the multi-drop pipeline).
- [x] Versioned multi-drop loot-table domain model with guaranteed drops, independent weighted bonus rolls, no-drop intervals, quantity bands and aggregated stack validation.
- [x] Durable multi-item reward claims: migration/backfill, account-scoped listing, atomic all-item redemption, full-inventory retry, idempotency, API DTO and Godot Rewards panel.
- [x] Activate guaranteed-plus-rare-bonus loot tables for Frontier Rat and Wild Boar.
- [ ] Expand loot tables with rarity tiers, boss-exclusive materials, unique item instances, per-player policies, versioned tuning and economy safeguards (multi-item foundation now in place).
- [ ] Deeper crafting framework: richer recipes, tool/station/skill requirements, material properties, quality, experimentation and crafted-item progression.
- [ ] Recipe quality, tool requirements, advanced stations and broader profession progression. Advanced non-combat specialization trees remain deferred until core loops are stable.

## Milestone 4: Inventory, equipment and economy

- [x] Starter item catalog, stack-count inventory, server-side stack validation and consumable healing potions.
- [x] Persistent 100-slot bank with atomic deposit/withdraw operations.
- [x] Client inventory/bank search, category filters and quantity-based sorting.
- [x] Godot bag/bank panel with item selection and quantity-based stack transfers.
- [x] Persistent coin wallet, starter merchant catalog and atomic buy/sell transactions.
- [x] Equipment slots, equip/unequip transactions and derived attack/defense bonuses.
- [ ] Unique item instances, durability, binding, item modifiers and richer stack management.
- [ ] Currency ledger, player-to-player trading with escrow and anti-duplication guarantees (backend and Godot multi-item UI vertical slice implemented; concurrency, restart-recovery, abuse/load and real-device gates remain; see `docs/PLAYER_TRADE_DESIGN.md`).
- [ ] Regional supply, market discovery, currency sinks and player-run shops.

## Milestone 5: Social sandbox

- [x] Persistent account-owned local zone chat with polling, message cursors, newest-50 initial history, bounded client history and per-character cooldown.
- [ ] Direct messages, friends, groups and guilds.
- [ ] Player housing, permissions, storage and decoration.
- [ ] Settlement plots, construction, shared storage and governance.
- [ ] Player services, profession specialization and player-run institutions.
- [ ] Factions, reputation, world events and persistent world consequences.

## Milestone 6: Persistent multiplayer alpha

- [ ] Realtime replication, zone ownership and reconnect recovery.
- [ ] Durable world-state ownership, backup/restore drills and migration operations.
- [ ] Load testing, observability, capacity planning and operational runbooks.
- [ ] Distributed abuse prevention, economy monitoring, moderation and admin tools.
- [ ] Closed alpha deployment and feedback loop.

## Core game experience focus — active priority

Boss roster expansion and boss-specific polish are paused. Development now prioritizes the core player experience across the whole game.

- [x] Implement third-person movement acceleration/deceleration, air control, coyote time, jump buffering, facing/turn smoothing, fall-speed cap and mesh-only locomotion feedback; headless regression and strict Godot import/Web export passed in client Action #153.
- [ ] Interactive desktop/mobile playtesting, touch-control feel, camera collision and movement reconciliation tuning. Bounded wheel/keyboard camera zoom is implemented; automated min/max and attack-animation checks are pending.
- [ ] Complete movement authority and reconciliation: server snapshots, bounded correction, reconnect recovery and smooth remote-player interpolation without client trust.
- [ ] Replace the flat placeholder floor with collision-safe deterministic terrain. A noise-based 240m terrain mesh and matching collision now place local/remote players, enemies, gathering nodes and stations consistently; runtime collision smoke coverage now includes an actual physics ray hit and player grounding. Client Action #161 passed the runtime physics ray/grounding and movement checks without Godot errors. Action #162 exposed a test-cleanup regression, reverted in 71eafa2; validate Action #163 before closing this item. All failures remain documented in the handoff. Server-authoritative vertical movement and richer biomes/paths remain future work.
- [ ] Establish a real character presentation/animation foundation: idle, locomotion, start/stop, jump, fall, land, interact and combat states, with equipment-aware attachment points.
- [ ] Improve the core HUD and responsive UI: health/resources, skills, inventory, contextual interaction, feedback, scalable panels and mobile-safe controls.
- [ ] Build a reusable player ability/action framework with server-owned costs, cooldowns, range/target validation and client animation/feedback; keep it independent of boss mechanics.
- [ ] Improve world interaction feedback, navigation, environmental readability and performance before adding more content systems.

## Working rule

Each milestone is a tested, reviewable slice. A feature is not complete because a placeholder or interface exists: completion requires server rules, persistence where applicable, tests, and a usable client path. Keep production registration disabled until the documented security launch gates are met. **Never hide an error:** stop, diagnose, fix, regression-test and verify every discovered failure before moving to unrelated work; preserve the failure and correction in the handoff. See docs/DEVELOPMENT_RULES.md.

## Boss encounter vertical slice

- [x] Migration 0016 persists versioned boss phase, telegraph, cooldown, ability cursor and terminal defeat state under the active spawn encounter-cycle ID.
- [x] Reject stale-cycle and stale-timestamp boss state writes; validate complete telegraph timestamp invariants.
- [x] Advance boss phases and telegraphs inside the PostgreSQL enemy world tick; resolve ability-specific range checks and Defense/equipment mitigation transactionally with player health, contribution, boss state and AI state.
- [x] Interrupt active telegraphs when the target moves out of ability range or no eligible target remains.
- [x] Cover concurrent exactly-once damage, lethal player defeat, target/telegraph cleanup, respawn-cycle reset, public authenticated snapshots and concurrent reward-claim retries with PostgreSQL integration tests (server Actions #602 passed race-enabled integration tests and Vet).
- [x] Register Stoneback Colossus in the public enemy catalog at Verdant Frontier (32,32): 100 HP, 2 Defense, no ordinary attack, 500 encounter XP and 300-second respawn.
- [x] Assign the versioned Stoneback multi-item harvest table and verify defeat, durable settlement, claim listing, claim redemption, XP and idempotent retry.
- [x] Render the boss world model and consume authoritative boss health/phase/ability snapshots in the Godot HUD.
- [x] Keep the encounter HUD visible on authoritative defeat snapshots, show explicit reward-claim feedback, make active telegraphs communicate the ability warning and move-out-of-range instruction, surface short phase-transition cues, and size the HUD with viewport-relative margins for narrow screens.
- [x] Verify server Actions #599/#600, including race-enabled PostgreSQL integration tests and Vet; verify client Actions #137 for Godot import/script validation.
- [ ] Continue player-facing playtesting and polish: telegraph timing feedback, phase transition cues, target lifecycle, defeat/respawn presentation and accessibility at desktop/mobile aspect ratios.
- [ ] Add group encounters, party threat and support contribution before implementing group reward distribution.
- [ ] Expand boss roster only after the Stoneback encounter remains robust under repeated combat, respawn and reward-claim testing.



## Core client presentation

- [x] Deterministic terrain mesh/collision, player grounding and physics-ray smoke tests (Client #161/#163/#172).
- [x] Bounded third-person camera zoom with min/max regression coverage (Client #172).
- [x] Initial server-confirmed melee swing presentation with visible motion and neutral-pose regression coverage (Client #172).
- [ ] Add richer terrain paths, environmental dressing and biome transitions. Four starter-zone trails are integrated into terrain vertex colors and passed Client Action #177. A deterministic moisture-driven grass/moss/dry-scrub/rock palette is implemented; new palette regression CI is pending.
- [ ] Improve interaction feedback, combat hit reactions, ability VFX and weapon-family animation sets.
- [ ] Responsive HUD layout and consistent desktop/mobile interaction playtesting.

- [ ] Validate responsive HUD on narrow mobile, tablet and desktop aspect ratios. Primary action toolbar now uses a responsive two-column grid and status text wraps; Actions #178/#179 caught an unsupported viewport API; corrected and passed in Client Action #180. Follow-up shrinkable-button refinement passed in #181; interactive device playtesting remains outstanding.


## Core world and client milestone — 2026-11-09

- [x] Deterministic moisture noise and multiple biome surface palettes integrated into the existing terrain mesh.
- [x] Regression coverage for deterministic biome colors and multiple surface palettes; Client Actions #182 and #183 passed headless runtime smoke, Godot import and Web export.
- [x] Integrated starter trail coloration in the same terrain mesh; Client Action #177 passed.
- [x] Responsive HUD action toolbar and viewport sizing; Client Actions #180/#181 passed.
- [x] Movement acceleration/deceleration, jump/coyote/buffer behavior, bounded camera zoom and initial allocation-free attack swing; Client Action #172 and subsequent full client runs passed.
- [ ] Terrain environmental dressing, vegetation/rocks/landmarks and biome-specific resource placement.
- [ ] Full locomotion/action animation state machine, hit reactions, gather/craft animations and weapon-family attacks.
- [ ] Mobile virtual controls and physical-device usability/visual QA.
- [ ] Movement authority review and explicit server-side terrain-height/jump/collision design decision.
- [ ] End-to-end core gameplay loop polish: gathering, crafting, equipment, interaction, feedback and progression.
- Boss expansion is deliberately paused per user direction until the core game is substantially stronger.


## Source-driven design and decision checkpoint — 2026-11-09

- [x] Validate and fully read the supplied SWG Pre-CU JSON manual (`/mnt/data/swg_pre_cu_reference_manual.json`).
- [x] Validate and fully read the supplied OSRS JSON manual (`/mnt/data/osrs_system_reference_manual.json`).
- [x] Validate and fully read the supplied RS3 JSON manual (`/mnt/data/rs3_system_reference_manual.json`).
- [x] Create source-attributed comparative synthesis in `docs/DESIGN_SYNTHESIS.md`, including source section attribution, inferred player motivations, strengths/trade-offs, dispositions, implementation mapping and source coverage gaps.
- [x] Create `docs/DESIGN_DECISIONS.md` with accepted, adapted, rejected, deferred and open decisions.
- [x] Preserve broad skill progression and reject an arbitrary skill-learning cap/250-point budget.
- [x] Audit current client movement, server planar movement validation, terrain/collision and regression smoke test.
- [ ] Add delayed-request/latency movement regression coverage and determine whether the 250 ms elapsed-window cap causes avoidable corrections under realistic request latency. Do not change server limits until reproduced and tested.
- [ ] Complete movement reconciliation and reconnect design without trusting client coordinates.
- [ ] Add terrain environmental dressing, landmarks, vegetation, rocks and biome-aware resource placement while preserving deterministic generation and shared visual/collision sampling.
- [ ] Build the player presentation state machine for locomotion, jump/fall/land, attack and interaction, separate from collision and authoritative state.
- [ ] Improve gather/craft/use-item feedback and responsive gameplay panels.
- [ ] Design mobile virtual movement/look/action controls and perform physical-device QA.

### Source interpretation boundary

The three JSON manuals do not provide sufficient source detail for quest design, mobile controls, accessibility, detailed UI, terrain/world layout, animation systems or housing decoration. Those topics remain original design work and are not marked as source-derived. See `docs/DESIGN_SYNTHESIS.md` for the exact coverage statement.

### Current source/CI notes

The raw JSON manuals were available as current conversation attachments under `/mnt/data`; they were not found as files under a tracked repository source folder and were not committed as raw source data. The source paths above are the actual mounted paths in this workspace, not proposed Git paths.

Client Action #183 (run 42604378304) passed; #182 (42604361140) passed. Client #178 (42603926295) and #179 (42603956717) failed due to unsupported `get_viewport_rect()` in `client/scripts/main.gd` plus a dependent type inference error; correction to `get_viewport().get_visible_rect().size.x` was validated by #180 (42604040649) and #181 (42604113289). Full logs were inspected. Server Action #602 (42598330356) passed race-enabled PostgreSQL integration tests and Vet. Its PostgreSQL service log includes intentional negative-path constraint errors from integration tests; the Go test packages and Vet completed successfully. The latest CI logs also include an upstream Node.js 20 deprecation warning from checkout@v4.

### Source documents committed

- `docs/DESIGN_SYNTHESIS.md` — source-attributed comparison and original design direction.
- `docs/DESIGN_DECISIONS.md` — decision log with explicit status and implementation boundaries.

 
## Movement latency correction — verification pending

- [x] Commit `df58725e8b22246203ebef653255bcacbabbda91` increases the bounded server movement calculation window from 250 ms to one second while retaining the 7 units/second speed limit, 0.15 tolerance, zone bounds and finite-coordinate checks.
- [x] Add a regression for serialized client updates delayed to 400 ms.
- [x] Add a regression proving a five-second idle interval remains capped to one second.
- [ ] Server Action #603 / run `42606055274` must complete race-enabled PostgreSQL integration tests and `go vet ./...`; inspect full logs before marking the movement correction complete.


### Core world presentation — active

- [x] Add deterministic foliage and rock environmental dressing to the existing terrain root using batched MultiMesh rendering.
- [x] Keep decorative placement aligned to the shared terrain height sampler and clear of spawn/trail corridors and steep slopes.
- [x] Add regression coverage for deterministic transforms, minimum generated instance counts, protected-site exclusion and terrain-height alignment.
- [ ] Replace placeholder primitive prop meshes with original art while preserving batching and deterministic transforms.
- [ ] Add collision/navigation policy for substantial world obstacles only after a server-authority and pathing design is agreed; current decorative props are intentionally non-colliding.


### Environmental dressing verification follow-up

- [ ] Inspect client run #186 (42619558244) after the regression-test correction; inspect the complete log for the earlier dummy-renderer cleanup error.
- [ ] Verify actual MultiMesh transforms and visual placement in a non-headless Godot renderer.
- [ ] Keep decorative props non-colliding until a separate collision/navigation and server-authority design is implemented.


- [x] Client Action #186 (run 42619558244) passed deterministic environmental placement generator checks, terrain collision/movement smoke, Godot import and Web export with no Godot errors.
- [ ] Verify MultiMesh appearance and GPU placement in a non-headless rendered client and on desktop/mobile hardware.


### Biome-aware prop presentation — active

- [x] Implement moisture-derived wet/moss, transition and dry foliage tint sampling plus rock tint variation.
- [x] Add deterministic palette diversity and color-data configuration regression coverage.
- [ ] Verify biome-aware MultiMesh tinting through Client Action #187 (42619776488).
- [ ] Review the result in a non-headless rendered client and on physical desktop/mobile devices.
