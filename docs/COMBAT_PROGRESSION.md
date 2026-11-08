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
- A weapon family maps to its appropriate skill: blades, polearms, unarmed, bows/thrown weapons, pistols or rifles.
- Magic abilities use Magic, with ability-specific requirements and effects.
- Melee, Strength, Defense and Tactics can contribute to defined combat actions, but their contribution must be explicit per ability/action rather than silently stacking every skill into every roll.
- The legacy Combat skill remains a broad baseline during the initial core phase. As specialized combat styles arrive, migrate combat formulas deliberately and preserve existing character progression.
- Equipment modifiers are server-owned item data. Client visuals and local stats never determine hit outcomes.
- NPC and player defenses are resolved server-side. Random rolls, if used, are generated server-side and are reproducible in tests through injected deterministic randomness.

## Implementation sequence

1. Keep the current simple PvE loop stable while adding skill metadata and the expanded roster.
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
