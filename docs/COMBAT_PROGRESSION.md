# RS Galaxies — Combat Progression Foundation

Status: design baseline; existing PvE remains unchanged until each combat style is implemented and tested.

## Core rule

There is no standalone Accuracy skill or player-assigned accuracy stat. Accuracy is calculated for each action by the authoritative server using:

1. The attacker's level in the skill governing the selected ability and weapon family.
2. The ability's own accuracy modifier, execution properties and requirements.
3. The equipped weapon's quality, condition and relevant attributes.
4. Other relevant equipment bonuses and penalties.
5. Target defense, movement/state and applicable environmental or situational modifiers.

A conceptual model for future tuning is:

`hitChance = clamp(base + skillContribution + abilityModifier + equipmentModifier - targetDefense - situationalPenalty, min, max)`

This is a design relationship, not a final numeric formula. Do not implement arbitrary constants until combat abilities, weapon data and target defense rules are defined and testable.

## Skill and weapon selection

- The server resolves the selected ability ID and equipped weapon; the client cannot nominate a different governing skill.
- A weapon family maps to its appropriate specialization: Blades, Polearms, Unarmed, Pistols or Rifles. Magic abilities use Magic and their own ability-specific requirements and effects.
- Melee and Ranged are broad parent skills. Specialized actions award the governing specialization XP and a smaller share of the matching parent XP.
- Strength and Defense are independent co-skills. Strength contributes to maximum melee damage; Defense contributes to damage mitigation and item-defined armor eligibility. Their effects are explicit and server-owned.
- Combat and Tactics are not active skills. Legacy progress records are preserved but inactive; they are neither displayed as active skills nor eligible for Jedi mastery.
- Equipment modifiers are server-owned item data. Client visuals and local stats never determine hit outcomes.
- NPC and player defenses are resolved server-side. Random rolls, if used, are generated server-side and are reproducible in tests through injected deterministic randomness.

## Implementation sequence

1. Keep the simple PvE loop stable while routing XP through the agreed parent, specialization and co-skill structure.
2. Introduce a server-owned ability definition model with explicit governing skill, range, cooldown, resource cost and modifiers.
3. Add weapon-family definitions and equipment compatibility rules.
4. Add hit resolution with bounded chance, deterministic unit tests and server-side random source.
5. Add one combat style at a time (melee first, then ranged firearms and magic when their item/ability loops exist).
6. Add client targeting, ability selection, feedback and animations only after server rules are authoritative.

## Invariants

- Never accept hit chance, damage, target defense, weapon stats or XP amounts as authoritative client input.
- A missing or incompatible weapon/ability combination fails validation; it does not silently fall back to a more powerful skill.
- Skill level, ability, weapon and equipment effects must be independently testable.
- Accuracy is an outcome of the full action context, not a separately trainable skill.
- Jedi Force powers and lightsaber weapon disciplines remain a separate future archetype/progression system.
