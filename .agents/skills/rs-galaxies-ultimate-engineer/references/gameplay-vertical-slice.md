# Gameplay Vertical Slice Guide

Every gameplay addition in RS Galaxies must be delivered as a complete, verified vertical slice rather than an isolated stub.

## Required layers for a vertical slice

1. **Authoritative server domain (`server/internal/game/`):**
   - Item, node, recipe, enemy, or loot definitions with deterministic validation.
   - Unit tests covering happy paths, boundary conditions (level gates, range checks, stack overflow, full pools), and defensive copy invariants.
2. **Persistence & HTTP API (`server/internal/storage/postgres/`, `server/internal/httpapi/`):**
   - PostgreSQL schema/query updates and transactional state mutations where persisted state changes.
   - HTTP catalog and action endpoint wiring plus `server_test.go` / integration test updates.
3. **Godot client presentation & UI (`client/scripts/`):**
   - Terrain-grounded 3D visuals (`_terrain_height_at(x, z)`) with distinct readable materials/silhouettes.
   - Responsive HUD/panel integration, interaction hints, and server-confirmed presentation triggers.
   - Headless smoke test coverage in `client/tests/interaction_hint_smoke_test.gd` or `client/tests/player_controller_smoke_test.gd`.
4. **Documentation & handoff (`docs/`):**
   - Honest verification evidence appended to `docs/SESSION_HANDOFF.md` (and `ARCHITECTURE.md` / `ROADMAP.md` when catalog totals or system boundaries change).
