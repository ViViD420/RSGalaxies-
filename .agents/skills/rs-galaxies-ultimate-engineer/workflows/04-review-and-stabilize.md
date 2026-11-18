# Workflow 04 — Review and Stabilize

1. **Halt unrelated feature work on any error.** Under the permanent no-hidden-errors rule, any compile failure, failing unit or PostgreSQL integration test, Godot `SCRIPT ERROR:` / `ERROR:`, leak diagnostic, or red GitHub Actions job takes immediate priority.
2. **Reproduce and inspect complete logs.** Run the failing test or Godot script directly and capture the full stack trace or line number.
3. **Identify the root cause.** Distinguish between:
   - Domain bugs (e.g., consumable buff restoration double-counting `Buff*Amount`).
   - Stale test assumptions after catalog expansions (e.g., hardcoded loot table count `5` vs `21`, or zone enemy snapshot count `3` vs `len(game.EnemyDefinitions())`, or XP threshold ordering across levels).
   - Test harness mismatches (e.g., instantiating a bare `CharacterBody3D.new()` instead of `res://scenes/player.tscn` when `main.gd` calls controller methods like `is_walk_mode`).
4. **Fix and add regression coverage.** Fix the defect without weakening assertions or filtering error output, and add a regression test locking in the fix.
5. **Re-run full verification and record in `docs/SESSION_HANDOFF.md`.** Verify all server and client checks pass and document the failure, root cause, fix, and passing run in `docs/SESSION_HANDOFF.md`.
