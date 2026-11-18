# Workflow 02 — Implement Feature

1. **Align with design pillars.** Confirm the feature fits the active core-game priorities in `docs/SESSION_HANDOFF.md` and `docs/ROADMAP.md` (boss work remains paused unless explicitly requested), preserves broad 18-skill progression, and keeps MMO-critical outcomes server-authoritative.
2. **Inspect existing code first.** Extend existing packages (`server/internal/game`, `server/internal/storage/postgres`, `server/internal/httpapi`) and Godot scripts (`client/scripts/main.gd`, `player_controller.gd`, `world_terrain.gd`) instead of duplicating systems.
3. **Implement server domain & persistence.** Add deterministic validation, transactional database updates if state is persisted, and unit + integration tests.
4. **Implement client presentation & UI.** Wire server responses to `main.gd` and `player_controller.gd`, keeping visual meshes decoupled from `CollisionShape3D`.
5. **Verify end-to-end.** Run `go test -race -count=1 ./...`, `go vet ./...`, and Godot 4.3 headless smoke tests + `--editor --quit` import check, confirming zero `SCRIPT ERROR:` or `ERROR:` lines.
6. **Update documentation.** Append a checkpoint to `docs/SESSION_HANDOFF.md` (and update `docs/ARCHITECTURE.md` / `docs/ROADMAP.md` if catalog counts or architectural boundaries changed).
