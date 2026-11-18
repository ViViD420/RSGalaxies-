# Local-First Protocol

## Purpose

Gemini on mobile and Antigravity sessions edit files directly in the local RS Galaxies workspace before those changes are committed or pushed to GitHub. GitHub (`ViViD420/RSGalaxies-`) is a synchronization and CI target, **not** the default source of truth.

## Rules

1. **Inspect before touching Git state.** Before any branch switch, pull, merge, rebase, or push:
   - Check `git status`, `git diff`, `git log -n 5`, and untracked files.
   - Fetch remote metadata (`git fetch origin`) without modifying the working tree.
   - Diff the local working tree against `origin/main` (`git diff origin/main --stat`).
2. **Never discard local work.** Do not run `git reset --hard`, `git checkout -- .`, `git restore .`, `git clean -fd`, or `git pull --force`. If a local file differs from `origin/main`, assume the local file contains intentional newer work until proven otherwise by comparing it against `docs/SESSION_HANDOFF.md`.
3. **Verify unverified local milestones first.** When `docs/SESSION_HANDOFF.md` lists milestones marked *"written locally; CI verification pending"*, run the full server and client verification suites against the local files before adding new features or pushing to GitHub. Fix any defects uncovered during verification in place rather than reverting the milestone.
4. **Stage and commit deliberately.** Review `git status` and `git diff --cached` to ensure build artifacts (`client/.godot/`, `client/build/`, `server/bin/`, `.env`, temporary logs) remain excluded by `.gitignore`.
5. **Confirm remote CI after pushing.** Once local changes are committed and pushed to `main` (or an explicit branch requested by the user), monitor `.github/workflows/server.yml` and `.github/workflows/client.yml` runs and record the run IDs and commit SHAs in `docs/SESSION_HANDOFF.md`.
