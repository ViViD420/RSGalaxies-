---
name: rs-galaxies-ultimate-engineer
description: Principal local-first engineering skill for RS Galaxies — an original server-authoritative sandbox MMORPG built with a Go + PostgreSQL world server and a Godot 4.3 client. Protects local Gemini/Antigravity work, enforces the permanent no-hidden-errors rule, preserves the GalaxyMMO boundary, and guides feature implementation, verification, and GitHub synchronization.
---

# RS Galaxies — Ultimate Engineer Skill

Use this skill for every development, verification, documentation, or synchronization task in the **RS Galaxies** project.

## 1. Non-negotiable guardrails

1. **Local-first source of truth.** The active local workspace is the primary source of truth. Local files may contain newer work written on mobile (Gemini) or in Antigravity that has not yet been synced to GitHub. Never assume `origin/main` is newer. Never run destructive Git commands (`reset --hard`, `clean -fd`, `checkout -- .`, force-overwriting pulls) to make local match remote. Compare local and remote explicitly before any push or pull (see `references/local-first-protocol.md` and `workflows/03-local-github-sync-review.md`).
2. **Project boundary — never touch GalaxyMMO.** RS Galaxies is completely separate from GalaxyMMO and any other project. Never modify GalaxyMMO or reuse its proprietary code, assets, schemas, or configurations.
3. **Permanent no-hidden-errors rule (accuracy over convenience).** Every error, test failure, parser/compile error, Godot `SCRIPT ERROR:` / `ERROR:`, race detector report, or CI failure must be reproduced, diagnosed to its actual root cause, fixed, covered by a regression test, and re-verified. Never filter logs, weaken assertions, skip failing tests, or claim tests/builds/commits/pushes/CI passed without real execution evidence.
4. **Original game from attributed design research.** Use Star Wars Galaxies (SWG), Old School RuneScape (OSRS), and RuneScape 3 (RS3) source JSON summaries strictly as design research (see `references/source-json-method.md` and `references/rs-galaxies-design-profile.md`). Never copy proprietary code, assets, lore, dialogue, maps, UI art, or distinctive expression.
5. **Server authority for MMO-critical outcomes.** The Go + PostgreSQL server is authoritative for account sessions, character ownership, movement bounds/speed/time, skill XP and level gates, gathering node charges/respawns, crafting ingredient consumption, inventory/bank/escrow trade mutations, combat damage/accuracy/defense, HAM + secondary attribute pools (`Health`/`Action`/`Mind` and `Constitution`/`Stamina`/`Willpower`), enemy AI telegraphs, encounter contribution ledgers, and reward claims (see `references/mmo-engineering-checklist.md`). Client terrain height and presentation state machines are visual only unless an authoritative server model is explicitly designed and tested.
6. **Continuity via `docs/SESSION_HANDOFF.md`.** Read `docs/SESSION_HANDOFF.md` completely before editing code, and append an honest, evidence-backed checkpoint after every meaningful milestone using the structure in `templates/SESSION_HANDOFF.md`. Update `docs/ROADMAP.md` and `docs/ARCHITECTURE.md` whenever their contents genuinely change.

## 2. Architecture & stack quick reference

| Layer | Stack & key paths | Verification command |
| --- | --- | --- |
| World server | Go 1.23 (`server/cmd/worldd`, `server/internal/{auth,game,httpapi,storage/postgres}`), PostgreSQL 16 (`migrations/0001..0016`) | `cd server && go test -race -count=1 ./... && go vet ./...` (with `RS_GALAXIES_TEST_DATABASE_URL` set so integration tests run rather than skip) |
| Game client | Godot 4.3 (`client/project.godot`, `client/scenes/{main,player}.tscn`, `client/scripts/{main,player_controller,world_terrain}.gd`) | Headless smoke tests (`tests/interaction_hint_smoke_test.gd`, `tests/player_controller_smoke_test.gd`), `--editor --quit` import check, Web and Android debug APK exports; fail on any `SCRIPT ERROR:` or `ERROR:` line |
| CI workflows | `.github/workflows/server.yml`, `.github/workflows/client.yml` | `gh run list` / `gh run view --log-failed` against `ViViD420/RSGalaxies-` |

## 3. Standard workflows

- **Resuming work:** Follow `workflows/01-resume-local-project.md` to confirm workspace state, diff local vs. remote, verify claims in `docs/SESSION_HANDOFF.md` against actual code, and identify the exact unfinished task.
- **Implementing features:** Follow `workflows/02-implement-feature.md` and `references/gameplay-vertical-slice.md` to build end-to-end across server domain, PostgreSQL persistence, HTTP API, Godot client presentation/UI, and regression tests.
- **Synchronizing with GitHub:** Follow `workflows/03-local-github-sync-review.md` to inspect divergence, verify all local changes pass server and client gates, commit cleanly, push safely, and confirm remote GitHub Actions.
- **Stabilizing & reviewing:** Follow `workflows/04-review-and-stabilize.md` whenever tests, strict Godot logs, or CI runs surface failures.
