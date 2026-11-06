# PvE Combat Prototype

## Server-authoritative rules

- Characters start with 100 health. Health and the last successful combat action are persisted in PostgreSQL.
- Attacks require the same world zone and a distance within the enemy's interaction range (currently 3 world units).
- The server enforces a one-second per-character combat cooldown.
- Player damage is derived from a base hit, the Combat skill level, equipped attack bonuses and enemy defense. Damage has a minimum of one.
- A surviving enemy retaliates. Incoming damage is reduced by equipped defense bonuses and the Combat skill's current defensive contribution.
- Defeating an enemy awards Combat XP and its configured loot if the inventory stack has room. An XP-capped character can still defeat enemies and receive loot.
- Enemy health, initialization and respawn timestamps are shared world state stored in the enemy_spawn_state table. The enemy row and character row are locked and updated in one transaction.
- A defeated character cannot move, gather, craft, trade, bank, equip or earn XP. Resting while alive is restricted to the starter safe spawn; a defeated character is restored at that spawn.

## Starter enemies

| Spawn ID | Enemy | Health | Attack | Defense | Combat XP | Loot | Respawn |
|---|---|---:|---:|---:|---:|---|---:|
| frontier-rat-001 | Frontier Rat | 12 | 2 | 0 | 25 | Rat Pelt ×1 | 30s |
| frontier-boar-001 | Wild Boar | 35 | 6 | 2 | 60 | Boar Hide ×1 | 60s |

## Current limitations

This is the first PvE slice, not a complete combat system. Enemies are stationary, have no aggro/pathfinding/AI, and do not use abilities. There are no animations, hitboxes, combat styles, status effects, group encounters, death penalties, or player-versus-player combat yet.
