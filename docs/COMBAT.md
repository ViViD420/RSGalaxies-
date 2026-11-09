# PvE Combat Prototype

## Server-authoritative rules

- Characters start with 100 health. Health and the last successful combat action are persisted in PostgreSQL.
- Attacks require the same world zone and a distance within the enemy's interaction range (currently 3 world units).
- The server enforces a one-second per-character combat cooldown.
- The current starter sword selects Blades; unarmed attacks select Unarmed. The governing specialization is resolved from server-owned equipped item data, never from a client-selected skill ID.
- Damage currently uses a deterministic prototype formula based on the selected specialization, Melee parent level, Strength, equipment attack bonus and enemy Defense. Accuracy is not a skill and is not yet rolled as a separate hit/miss outcome.
- Successful hits award Strength XP. Defeating an enemy awards the selected specialization XP and one-quarter parent XP (minimum one where applicable). Taking positive damage awards Defense XP.
- Strength contributes to melee damage; Defense reduces incoming damage. Defense also governs armor eligibility through server-owned item requirements.
- In the current click-to-attack loop, a surviving enemy retaliates in the same player/enemy transaction. The domain layer now also exposes `Player.ReceiveEnemyAttack`, which resolves a completed enemy telegraph against the persisted live enemy position, applies equipment/Defense mitigation, awards Defense XP for positive damage and reports defeat. This method is unit-tested but is not yet wired into the autonomous AI transaction.
- Defeating an enemy awards its configured starter loot when the corresponding inventory stack has capacity. Dedicated loot tables, durable overflow claims, contribution eligibility and reward audit records are not implemented yet; these are explicit prerequisites for boss rewards.
- Enemy health, initialization and respawn timestamps are shared world state stored in `enemy_spawn_state`. Migration 0010 adds persisted AI behavior, position, telegraph deadline, cooldown and last-update fields. `PlayerRepository.AdvanceEnemyAIForPlayer` locks and persists a spawn tick; its PostgreSQL restart test covers recovery. The current click-to-attack operation still locks and updates the player and enemy rows in one transaction.
- A defeated character cannot perform gameplay actions. Resting while alive is restricted to the starter safe spawn; a defeated character is restored at that spawn.
- Combat is currently a single-target action loop. The server-side AI domain state machine models idle, aggro/chase, attack telegraph/interruption and leash/return behavior with bounded timestamp-based movement, and its state now persists per spawn. A PostgreSQL-backed server-owned world tick coordinator now selects fresh same-zone character presence, locks the target character before the enemy spawn, advances the AI and commits completed telegraph damage, player health, Defense XP and enemy state in one transaction. A background worker advances catalogued spawns every 250 ms; multiple server instances serialize through row locks and the persisted attack cooldown. An authenticated zone-scoped enemy snapshot endpoint exposes authoritative positions and behavior for clients. The Godot client has not yet been updated to poll/render these snapshots, so live enemy visuals remain at their static catalog positions until that client integration lands. Enemies still have no pathfinding, abilities or multi-phase behavior.

## Starter enemies

| Spawn ID | Enemy | Health | Attack | Defense | Defeat XP | Loot | Respawn |
|---|---|---:|---:|---:|---:|---|---:|
| frontier-rat-001 | Frontier Rat | 12 | 2 | 0 | 25 | Rat Pelt ×1 | 30s |
| frontier-boar-001 | Wild Boar | 35 | 6 | 2 | 60 | Boar Hide ×1 | 60s |

## Next combat foundation

1. Validate the world tick coordinator in CI and add any missing stale-presence, zone-change, death/respawn and multi-instance concurrency coverage.
2. Replicate authenticated enemy snapshots and interpolate authoritative enemy positions/telegraphs in Godot.
3. Introduce server-owned weapon-family and ability definitions, beginning with the existing melee families; resolve hit outcomes from validated ability, weapon, skill, equipment and target state with injectable randomness.
4. Add durable combat telemetry and encounter participation records that can later support fair group rewards.
5. Add a reusable boss encounter state machine only after ordinary enemy behavior and reward settlement are reliable.

## Current limitations

There are no ranged weapon action loops, magic abilities, player combat animations, hitboxes, status effects, party/group encounters, death penalties or player-versus-player combat. The current starter loot path is not a substitute for the dedicated loot framework described in `ENCOUNTERS_LOOT_CRAFTING_DESIGN.md`.
