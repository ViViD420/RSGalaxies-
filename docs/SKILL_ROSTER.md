# RS Galaxies — Core Skill Roster

Status: foundational roster; skill progression sources are added as their gameplay systems are built.

The initial ordinary-character roster contains 20 skills. IDs are permanent API and persistence keys. Display names and descriptions may evolve, but existing IDs must not be renamed without a deliberate data migration.

## Combat skills (12)

| Skill | ID | Purpose |
|---|---|---|
| Combat | `combat` | General combat experience and baseline combat fundamentals. |
| Melee | `melee` | Broad close-quarters weapon fundamentals. |
| Defense | `defense` | Defensive technique, timing and mitigation. |
| Strength | `strength` | Physical power and demanding combat techniques. |
| Tactics | `tactics` | Positioning, timing and combat decisions. |
| Unarmed | `unarmed` | Hand-to-hand combat without a weapon. |
| Blades | `blades` | Edged melee weapons and their techniques. |
| Polearms | `polearms` | Staffs, spears and other long-reach melee weapons. |
| Ranged | `ranged` | Bows, thrown weapons and non-firearm ranged weapons. |
| Magic | `magic` | Offensive, defensive and utility magic. |
| Pistols | `pistols` | One-handed firearms. |
| Rifles | `rifles` | Long guns and precision firearms. |

Accuracy is deliberately **not** a skill. It is a derived combat outcome based on the attacker's relevant skill, the selected ability, the equipped weapon and other relevant equipment, plus target defenses and situational modifiers. The authoritative server will calculate hit chance and results.

Combat remains a broad, foundational skill for compatibility with the current PvE slice and counts as one of the combat mastery tracks. Accuracy is not a skill and never counts as a separate mastery.

## Gathering skills (4)

| Skill | ID | Purpose |
|---|---|---|
| Farming | `farming` | Cultivating crops and productive land. |
| Foraging | `foraging` | Finding plants, fungi and natural materials. |
| Mining | `mining` | Extracting ores, stone and minerals. |
| Scavenging | `scavenging` | Recovering components and materials from ruins. |

## Production skills (4)

| Skill | ID | Purpose |
|---|---|---|
| Cooking | `cooking` | Preparing food and provisions. |
| Crafting | `crafting` | Making tools, equipment and goods. |
| Smithing | `smithing` | Refining ores and working metals. |
| Woodcutting | `woodcutting` | Harvesting timber and renewable plant resources. |

## Progression and implementation rules

- Every skill uses the existing level 1–99 progression curve and XP cap.
- All 20 skills are initialized for new characters. Existing character JSON maps remain compatible; missing newly introduced skills are treated as level 1 when loaded or displayed.
- A skill definition does not automatically imply a complete gameplay loop. XP sources, tools, abilities, recipes and balancing are added in focused vertical slices.
- The first implementation keeps skill categories and Jedi mastery eligibility server-internal. Public catalogs and ordinary character skill payloads do not serialize these fields.
- Combat XP currently awards the legacy `combat` skill. Specialized skills are catalog foundations only until their corresponding combat styles and ability rules are implemented.
- Future skill additions should solve a clear gameplay need. Keep the ordinary roster near 20–25 for the initial core; add more only when a distinct progression loop warrants it.
- Jedi remains a separate hidden character progression system, not one of these 20 ordinary skills. See `JEDI_PROGRESSION_DESIGN.md`.

## Planned combat progression model

Combat resolution will select a governing skill from the server-defined ability and weapon family. The server will combine that skill with ability properties, weapon/equipment attributes, target defenses and situational effects to calculate hit chance and damage. No client-provided accuracy score or hit result is trusted. See `COMBAT_PROGRESSION.md`.
