# RS Galaxies — Godot client

A playable Godot 4.3 third-person prototype for the original RS Galaxies sandbox MMORPG. This is a technical vertical slice, not final art direction or a finished multiplayer client.

## Run

Open `client/` in Godot 4.3 or newer and run the main scene. Start the Go world server on port 8080 with PostgreSQL configured for account-backed play.

### Controls

- **WASD:** move; **Space:** jump.
- **Mouse:** orbit camera; **Escape:** release cursor.
- **Touchscreen:** on-screen movement pad, USE and jump controls; drag to orbit.
- **E / USE:** interact with nearby gathering nodes, stations, bank, merchant and enemies.

### Playable systems

- Account registration/login, account-owned character list, character creation/selection, switching and sign-out.
- Authenticated, server-validated movement and safe-spawn recovery.
- Gathering targets and their resource visuals load from the server's public node catalog; woodcutting, stone gathering, copper mining and foraging are included in the starter zone. The workbench and forge are also positioned from the server's crafting-station catalog.
- Workbench and forge recipes with server-validated ingredients, skill levels, range and output capacity, including a foraging-to-herbal-tea cooking loop. Herbal tea is a crafted healing consumable.
- Persistent inventory, 100-slot bank, coin wallet and merchant purchases/sales. The bag/bank panel supports item search, resource/equipment/consumable filters and sorting by name or carried/banked quantity.
- Equipment slots, gear bonuses and simple equipped-item visuals.
- Server-authoritative PvE attacks, enemy retaliation/respawn, Combat XP, loot and healing consumables.
- Persistent local-zone chat with a bounded 100-line client display.
- Leased zone-presence heartbeats and polling remote-player avatars/nameplates; inactive characters disappear after their presence lease expires.
- Live skills panel with all ten skills, current level, XP totals, per-level progress bars and XP remaining to the next level.

The client uses the public item, skill, crafting and enemy catalogs instead of duplicating those definitions. Gameplay outcomes are committed by the Go server and PostgreSQL repository. API requests use bounded timeouts, and automatic redirects are disabled so bearer credentials are never forwarded to a redirected destination.

## Server configuration

Account login requires PostgreSQL-backed account services. Use the provided GitHub Codespaces development container or run PostgreSQL and set `RS_GALAXIES_DATABASE_URL` before starting the Go server. The in-memory-only server mode is for isolated development and does not register account routes.

When running as a Codespaces Web build, the client derives the forwarded API origin automatically. For a native local run, start the Go API on port 8080 first; change `api_base_url` in `scripts/main.gd` if required.

A Web export preset is included in `export_presets.cfg`. The Codespaces setup installs Godot 4.3 and its export templates, and CI imports the project and validates a Web export.

## Current limitations

- Movement is client-predicted and periodically server-validated; there is no remote-player replication or authoritative continuous world simulation yet.
- The world is a generated starter test zone with placeholder geometry and character/enemy visuals.
- Enemies are stationary and have no aggro, pathfinding, abilities or animated combat.
- Local chat is polling-based. Direct messages, friends, groups, guilds, moderation and retention automation are not implemented.
- Equipment uses catalog items and slot maps; unique item instances, durability, modifiers and binding are future work.
- Access tokens remain in memory only and are not persisted between application launches. Password recovery, email verification and MFA are not implemented.
- Housing, settlements, player-to-player trading and large-scale world streaming remain future milestones.
