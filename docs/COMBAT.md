# PvE Combat Prototype

## Server-authoritative rules

- Characters start with 100 health. Health and the last successful combat action are persisted in PostgreSQL.
- Attacks require the same world zone and a distance within the enemy's interaction range (currently 3 world units).
- The server enforces a one-second per-character combat cooldown.
- The current starter sword selects Blades; unarmed attacks select Unarmed. The governing specialization is resolved from server-owned equipped item data, never from a client-selected skill ID.
- Damage currently uses a deterministic prototype formula based on the selected specialization, Melee parent level, Strength, equipment attack bonus and enemy Defense. Accuracy is not a skill and is not yet rolled as a separate hit/miss outcome.
- Successful hits award Strength XP. Defeating an enemy awards the selected specialization XP and one-quarter parent XP (minimum one where applicable). Taking positive damage awards Defense XP.
- Strength contributes to melee damage; Defense reduces incoming damage. Defense also governs armor eligibility through server-owned item requirements.
- A surviving enemy retaliates. Incoming damage is reduced by equipped defense bonuses and Defense level.
- Defeating an enemy awards its configured starter loot when the corresponding inventory stack has capacity. Dedicated loot tables, durable overflow claims, contribution eligibility and reward audit records are not implemented yet; these are explicit prerequisites for boss rewards.
- Enemy health, initialization and respawn timestamps are shared world state stored in the enemy_spawn_state table. The enemy row and character row are locked and updated in one transaction.
- A defeated character cannot perform gameplay actions. Resting while alive is restricted to the starter safe spawn; a defeated character is restored at that spawn.
- Combat is currently a single-target action loop. Enemies are stationary and have no aggro, pathfinding, telegraphs, abilities or multi-phase behavior.

## Starter enemies

| Spawn ID | Enemy | Health | Attack | Defense | Defeat XP | Loot | Respawn |
|---|---|---:|---:|---:|---:|---|---:|
| frontier-rat-001 | Frontier Rat | 12 | 2 | 0 | 25 | Rat Pelt ×1 | 30s |
| frontier-boar-001 | Wild Boar | 35 | 6 | 2 | 60 | Boar Hide ×1 | 60s |

## Next combat foundation

1. Introduce server-owned weapon-family and ability definitions, beginning with the existing melee families.
2. Resolve hit outcomes and damage from validated ability, weapon, skill, equipment and target state; inject deterministic randomness for tests.
3. Add enemy aggro, leash, movement and simple behavior state before complex boss phases.
4. Add combat telemetry and encounter participation records that can later support fair group rewards.
5. Add a reusable boss encounter state machine only after ordinary enemy behavior and reward settlement are reliable.

## Current limitations

There are no ranged weapon action loops, magic abilities, player combat animations, hitboxes, status effects, party/group encounters, death penalties or player-versus-player combat. The current starter loot path is not a substitute for the dedicated loot framework described in `ENCOUNTERS_LOOT_CRAFTING_DESIGN.md`.
