# Workflow 03 — Local & GitHub Sync Review

1. **Fetch remote safely.** Run `git fetch origin` without modifying the local working tree.
2. **Compare local workspace against `origin/main`.** Inspect `git status` and `git diff origin/main --stat` to enumerate every modified and untracked file.
3. **Run full verification before committing.**
   - Run `python3 scripts/validate_package.py` to confirm skill package integrity.
   - Run `go test -race -count=1 ./...` (with PostgreSQL integration enabled) and `go vet ./...` in `server/`.
   - Run Godot 4.3 headless smoke tests (`tests/interaction_hint_smoke_test.gd`, `tests/player_controller_smoke_test.gd`) and `--editor --quit` in `client/`, verifying zero `SCRIPT ERROR:` or `ERROR:` output.
4. **Stage and commit cleanly.** Ensure temporary build directories (`client/.godot/`, `client/build/`, `server/bin/`) are ignored, stage all intentional project and skill files, and create a descriptive commit on `main`.
5. **Push and verify GitHub Actions.** Push to `origin/main`, inspect `.github/workflows/server.yml` and `.github/workflows/client.yml` runs via `gh run list` / `gh run watch`, and record the verified commit SHA and run status in `docs/SESSION_HANDOFF.md`.
