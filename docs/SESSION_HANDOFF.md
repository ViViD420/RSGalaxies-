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

### Server
- Go development server with player creation, retrieval and listing.
- Ten stable skills: combat, cooking, crafting, farming, foraging, mining, ranged, scavenging, smithing and woodcutting.
- Deterministic cumulative XP curve, levels 1–99 and 200,000,000 XP cap.
- Player creation initializes all skills at level one.
- Domain XP validation, level-up detection and atomic in-memory awards.
- Defensive copying of player skill maps and stable player listing.
- HTTP routes for skill catalog, player skills and development-only XP awards. The XP shortcut is disabled when RS_GALAXIES_ENV=production.
- Domain, API and repository concurrency tests.
- Go CI runs tests with the race detector and go vet.

### Client
- Godot 4.3 prototype in client/.
- Procedurally generated test floor, placeholder character, third-person orbit camera and WASD/jump movement.
- Startup HTTP health check against the Go service.
- Headless Godot validation workflow.
- This is a movement/API proof of concept, not a final client or a network-authoritative character controller.

### Development environment
- .devcontainer/ configures a GitHub Codespaces Go workspace.
- .devcontainer/setup.sh installs Godot 4.3 and export templates, then runs Go tests/vet and a headless Godot project import.
- User can edit files in GitHub/Codespaces through a browser; no local PC installation is required for repository work.

## Limits and decisions

- Go is the server language. Godot is the provisional client engine, pending camera, movement and web/mobile export validation.
- Storage is in memory; data is lost when the server stops.
- The XP-award HTTP endpoint is development-only and must not be exposed publicly. Production XP must come from validated game actions.
- PostgreSQL, authentication, secure sessions, realtime networking, production persistence, world simulation and server-authoritative movement are not implemented.
- Code and tests are committed, but GitHub Actions execution results have not yet been inspected in this session. Do not claim CI passes until verified.

## Next engineering tasks

1. Inspect GitHub Actions results and fix any Go/Godot build failures.
2. Add a browser-playable Godot web export and phone-friendly input controls.
3. Implement world zones, server-owned coordinates, spawn points and validated movement.
4. Add PostgreSQL schema/migrations and transactional player persistence.
5. Replace development XP awards with validated gathering, crafting and combat actions.

## Continuation instruction

Always inspect this file and current repository state first. Work only in ViViD420/RSGalaxies-. Update this handoff after each meaningful milestone. Never assume GalaxyMMO code or decisions apply here.
