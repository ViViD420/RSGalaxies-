# RS Galaxies — Engineering Rules

These rules apply to every development session, contributor and automated continuation. They are mandatory.

## 1. No hidden errors — absolute rule

**Never hide, omit, minimize, silently bypass, or leave an observed error unresolved while moving on to unrelated work.**

When any error is found:

1. Stop the current feature task immediately.
2. Preserve and report the exact failure, including the command, workflow, job, log excerpt, reproduction or affected behavior when available.
3. Reproduce it where practical and identify the underlying cause. Do not guess, suppress output, weaken assertions, skip tests, disable checks, or add broad fallbacks merely to make the failure disappear.
4. Fix the root cause before starting unrelated work.
5. Add a focused regression test or validation when practical.
6. Run the narrowest relevant check, then the broader affected test suite.
7. Inspect the actual output and GitHub Actions job steps/logs. A queued, running, skipped, cancelled, or unobserved check is not a pass.
8. If the fix reveals another failure, repeat this process until all discovered failures are resolved or explicitly blocked by an external dependency. Never describe a blocked failure as fixed.
9. Record the original error, root cause, fix, regression coverage and exact verification result in `docs/SESSION_HANDOFF.md`.
10. Commit the fix and its verification record before resuming feature work.

A failed CI run remains part of the project's history. It must be documented honestly even after a later correction passes. Never rewrite history or documentation to make the failure disappear.

This rule covers compilation, static analysis, unit/integration tests, race detectors, migrations, API contracts, persistence, security, gameplay logic, movement, rendering, animation, UI, platform exports and runtime behavior.

## 2. Server authority and transaction safety

- The server owns consequential game state: position validation, health, combat results, skills/XP, inventory, currency, crafting, loot, cooldowns and persistence.
- The client sends intent and renders authoritative state; it never becomes the source of truth for rewards, movement limits or combat outcomes.
- Persist related gameplay mutations atomically. Preserve account ownership checks, concurrency controls and idempotency.
- Do not log secrets, credentials or private account data.

## 3. Preserve the existing project

- Work only in `ViViD420/RSGalaxies-` on `main`.
- Never restart, recreate or replace the project.
- Preserve the existing 18 active skills, professions, equipment requirements, gathering, crafting, economy, ordinary enemy progression and multi-item rewards.
- Do not modify GalaxyMMO or reuse proprietary source code, assets or other project artifacts.
- Prefer incremental improvements to existing architecture over parallel prototypes or unnecessary rewrites.

## 4. Quality and documentation

- Every meaningful feature has domain rules, validation, tests and a usable client path.
- Keep `docs/SESSION_HANDOFF.md`, `docs/ROADMAP.md` and `docs/ARCHITECTURE.md` synchronized with actual code and verified CI.
- Do not claim a feature is playable based only on domain logic, a placeholder, or headless import.
- Distinguish automated validation from visual/device playtesting.
- Commit meaningful, reviewable changes with clear messages.
