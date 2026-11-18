# Workflow 01 — Resume Local Project

1. **Confirm active workspace root.** Verify you are operating inside the local RS Galaxies project root containing `server/`, `client/`, `docs/`, `.github/`, and `AGENTS.md`.
2. **Read `docs/SESSION_HANDOFF.md` completely.** Pay special attention to the latest entries at the bottom of the file, any milestones marked *"written locally; CI verification pending"*, and the permanent no-hidden-errors rule.
3. **Read core documentation and manifests.** Inspect `docs/ROADMAP.md`, `docs/ARCHITECTURE.md`, `AGENTS.md`, `server/go.mod`, `client/project.godot`, `client/export_presets.cfg`, and `.github/workflows/{server,client}.yml`.
4. **Inspect Git state without mutating it.** Check `git status`, current branch, `HEAD`, recent commits, staged/unstaged diffs, untracked files, and compare against `origin/main` without pulling, resetting, or cleaning.
5. **Audit handoff claims against actual code.** Verify catalog counts (items, gathering nodes, recipes, enemies, loot tables), server domain logic, database integration tests, and Godot client scripts/smoke tests match what `docs/SESSION_HANDOFF.md` claims.
6. **Run real verification.** Execute `go test -race -count=1 ./...` (including PostgreSQL 16 integration tests), `go vet ./...`, and the Godot 4.3 headless smoke tests and editor import check.
7. **Report verified state and exact unfinished task.** Before moving on to new features, report any discrepancy or failing test uncovered during step 6 and fix it under `workflows/04-review-and-stabilize.md`.
