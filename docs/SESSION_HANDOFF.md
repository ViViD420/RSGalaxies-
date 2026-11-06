# RS Galaxies — Session Handoff

Last updated: 2026-11-06

## Project identity

- Name: **RS Galaxies**
- Repository: https://github.com/ViViD420/RSGalaxies-
- Default branch: `main`
- This is a standalone project. Do not reuse or modify the user's separate GalaxyMMO repository or its source tree.

## Product direction

An original sandbox MMORPG combining long-form, skill-driven progression and item gameplay with a social, open-ended world built around player professions, crafting, trade, housing and settlements.

## Current state

Milestone 1 is underway. The repository began empty. The initial commit series adds:
- Project README and repository rules.
- Initial architecture and milestone roadmap.
- A Go development server with a minimal player API.
- Domain-level player creation and progression foundations.
- Unit and HTTP API tests.
- GitHub Actions test workflow.

## Important technical decisions

- Go is the first server-side implementation language.
- The client engine has intentionally not been selected yet.
- In-memory storage is temporary development infrastructure, not production persistence.
- PostgreSQL, authentication, realtime networking and a real world simulation are future work and must not be described as already implemented.

## Immediate next tasks

1. Add player skill definitions and deterministic XP-to-level progression.
2. Introduce repository interfaces and PostgreSQL implementation with migrations.
3. Add character attributes, spawn location and validated movement.
4. Prototype client engine options against the documented server contract.
5. Add request limits, structured logging and API versioning.

## Continuation instruction

When continuing this project, first inspect this file and the current repository state. Continue only in `ViViD420/RSGalaxies-`. Update this handoff after each meaningful milestone. Never assume work from the separate GalaxyMMO project applies here.
