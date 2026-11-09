# RS Galaxies — Core Skill Roster

Status: approved foundational roster; advanced non-combat professions are intentionally deferred.

The initial ordinary-character catalog contains 18 active skills. Skill IDs are stable API and persistence keys. Display names and descriptions may evolve, but IDs must not be renamed without a deliberate migration.

## Combat foundations

| Skill | ID | Role |
|---|---|---|
| Melee | `melee` | Broad close-quarters parent skill. |
| Blades | `blades` | Melee specialization for edged weapons. |
| Polearms | `polearms` | Melee specialization for staffs, spears and long-reach weapons. |
| Unarmed | `unarmed` | Melee specialization for hand-to-hand combat. |
| Ranged | `ranged` | Broad ranged-combat parent skill. |
| Pistols | `pistols` | Ranged specialization for one-handed firearms. |
| Rifles | `rifles` | Ranged specialization for long guns and precision firearms. |
| Magic | `magic` | Ranged specialization for magical abilities. |
| Defense | `defense` | Independent co-skill governing defensive progression and armor eligibility. |
| Strength | `strength` | Independent co-skill contributing to maximum melee damage. |

Melee and Ranged are broad parent skills. Their child proficiencies are distinct progression tracks. The server awards specialization XP for the selected combat style and a smaller share of parent XP. The current starter weapon maps to Blades; attacking without a supported weapon maps to Unarmed. Ranged weapon and Magic action loops will be introduced separately.

Defense and Strength are independent co-skills; neither branches into specializations. Defense levels contribute to incoming damage mitigation and determine equipment eligibility through item-defined requirements. Strength contributes to maximum melee damage. They are not Jedi mastery skills.

## Gathering skills

| Skill | ID | Purpose |
|---|---|---|
| Farming | `farming` | Cultivating crops and productive land. |
| Foraging | `foraging` | Finding plants, fungi and natural materials. |
| Mining | `mining` | Extracting ores, stone and minerals. |
| Scavenging | `scavenging` | Recovering components and materials from ruins. |

## Production skills

| Skill | ID | Purpose |
|---|---|---|
| Cooking | `cooking` | Preparing food and provisions. |
| Crafting | `crafting` | Making tools, equipment and goods. |
| Smithing | `smithing` | Refining ores and working metals. |
| Woodcutting | `woodcutting` | Harvesting timber and renewable plant resources. |

The eight non-combat skills will be expandable into advanced professions later. Their specialization names and mechanics are deliberately not locked in during the core phase. Players may broadly progress across skills; profession identity will come from equipment, tools, recipes, materials, stations and access requirements rather than arbitrary skill-count limits.

## Progression and compatibility

- Every active skill uses the level 1–99 progression curve and 200,000,000 XP cap.
- New characters initialize all 18 active skills. Existing character records are backfilled without resetting existing XP.
- Retired `combat` and `tactics` progress keys are preserved in old character data for compatibility, but are not active skills, do not appear in the Godot skill panel, and cannot receive new XP through the active catalog. They do not count toward Jedi mastery.
- The six combat specialization skills (Blades, Polearms, Unarmed, Pistols, Rifles and Magic) are the only ordinary skills marked eligible for the future Jedi mastery challenge. Parent skills and co-skills do not count.
- Accuracy is not a skill. The authoritative server derives hit outcomes from the governing skill, ability, weapon/equipment, target defenses and situational modifiers.
- Jedi remains a separate hidden character progression system. See `JEDI_PROGRESSION_DESIGN.md`.

A skill definition does not automatically imply a complete gameplay loop. XP sources, tools, abilities, recipes and balancing are added in focused vertical slices.
