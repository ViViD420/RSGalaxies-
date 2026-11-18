# MMO Engineering Checklist

Use this checklist before considering any server or client change complete.

## Server authority & data integrity
- [ ] Authentication and character ownership verified via bearer session digests (`internal/auth`, `internal/storage/postgres/accounts.go`).
- [ ] All inventory, bank, equipment, currency, gathering, crafting, consumable, and trade mutations run inside PostgreSQL transactions with deterministic row locking where concurrent access is possible.
- [ ] Skill requirements (`RequiredLevel`), interaction range (`InteractionRange`), node charges/respawns, and net stack capacities (`MaxStack`, inventory slot caps) are enforced server-side before mutating state.
- [ ] Combat attribution awards XP to the active specialization, its parent skill (`melee` or `ranged`), `defense`, and `strength` (only when parent is `melee`).
- [ ] HAM pools (`Health`, `Action`, `Mind`) and secondary attributes (`Constitution`, `Stamina`, `Willpower`) clamp strictly to `EffectiveMaxHealth()`, `EffectiveMaxAction()`, and `EffectiveMaxMind()`, and buff consumables never double-count `Buff*Amount` on top of `HealAmount` / `ActionRestoreAmount` / `MindRestoreAmount`.
- [ ] Catalog accessors (`ItemDefinitions`, `GatheringNodeDefinitions`, `RecipeDefinitions`, `EnemyDefinitions`, `LootTableDefinitions`) return defensive copies including nested slices.

## Client stability & headless CI gates
- [ ] `player_controller.gd` presentation states (`IDLE`, `WALK`, `RUN`, `JUMP`, `FALL`, `LAND`, `ATTACK`, `GATHER`, `CRAFT`, `USE_ITEM`, `HIT`, `DEFEAT`, `RESPAWN`) affect visual meshes only and never mutate `CollisionShape3D`.
- [ ] Touch input isolates camera drag from bottom-left D-pad and bottom-right action button exclusion zones.
- [ ] All dynamic calls on `player_node` in `main.gd` match methods defined on `player_controller.gd`, and smoke tests instantiate `res://scenes/player.tscn` so the controller script is attached.
- [ ] Every headless test frees its nodes cleanly so Godot exits with zero `SCRIPT ERROR:`, `ERROR:`, RID leak, or ObjectDB leak lines.
