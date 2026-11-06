# RS Galaxies — Session Handoff

Last updated: 2026-11-06

## Project identity

- Name: **RS Galaxies**
- Repository: https://github.com/ViViD420/RSGalaxies-
- Default branch: main
- Standalone project. Never reuse or modify the separate GalaxyMMO repository.

## Product direction

An original sandbox MMORPG combining long-form skill progression and item gameplay with a social, open-ended world built around professions, crafting, trade, housing and settlements. Use original code and assets only.

## Current implementation

- Go development server with player creation, retrieval and listing.
- Ten stable skills: combat, cooking, crafting, farming, foraging, mining, ranged, scavenging, smithing and woodcutting.
- Deterministic cumulative XP curve, levels 1–99 and 200,000,000 XP cap.
- Player creation initializes all skills at level one.
- Domain XP validation, level-up detection and atomic in-memory awards.
- Defensive copying of player skill maps and stable player listing.
- HTTP routes for skill catalog, player skills and development-only XP awards.
- Domain, API and concurrency tests.
- GitHub Actions Go test/vet workflow.
- Codespaces development-container configuration in .devcontainer/devcontainer.json.

## Limits and decisions

- Go is the server-side language.
- Godot 4 is the intended client prototype direction; validate movement, camera and world-scale before locking it as final.
- Storage is in memory and data is lost when the server stops.
- The XP-award endpoint is development-only and must never be exposed publicly. Production XP must come from validated game actions.
- PostgreSQL, authentication, secure sessions, realtime networking, production persistence and actual world simulation remain unimplemented.
- Tests are committed, but their pass status must be verified in GitHub Actions before claiming success.

## Next engineering tasks

1. Inspect latest GitHub Actions result and fix any build/test failures.
2. Build a Godot 4 client prototype: 3D test scene, orbit camera, movement and Go API connection.
3. Implement world zones, server-owned coordinates, spawn points and movement validation.
4. Add PostgreSQL schema/migrations and transactional player persistence.
5. Replace development XP awards with validated gathering, crafting and combat actions.

## Phone-first workflow

The repository includes .devcontainer/devcontainer.json for GitHub Codespaces. From GitHub in a browser, use Code > Codespaces to create a hosted workspace. Setup runs Go tests and vet. No local PC installation is needed for repository edits or CI.

## Continuation instruction

Always inspect this file and current repository state first. Work only in ViViD420/RSGalaxies-. Update this handoff after each meaningful milestone. Never assume GalaxyMMO code or decisions apply here.
