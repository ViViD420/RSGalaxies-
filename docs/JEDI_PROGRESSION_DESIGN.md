# Jedi Hidden Progression and Character Unlock Design

Status: proposed design; not enabled in gameplay or public catalogs.

## Player-facing concept

Jedi is a hidden endgame progression path. A player first masters a randomly determined number of distinct combat skills to level 99 on an eligible character. Completing that requirement permanently unlocks the ability to create a separate Jedi character on the same account. The Jedi character has its own identity, character progression, skills, equipment and inventory; it is not a reskin or conversion of the character that earned the unlock.

The unlock should feel like a discovery, not a visible checklist. Before it is earned, public skill catalogs, character creation responses and the Godot UI must not reveal the Jedi skill, its requirement count, the eligible combat-skill pool or the unlock state. After completion, the server may reveal a discovery message and the Jedi character option.

## Unlock rule

1. Use explicit server-internal skill metadata, including a stable skill ID, progression category and a private combat-mastery eligibility flag. Only explicitly eligible skills count. Do not infer eligibility from skill names or client-supplied IDs. Never serialize the category or eligibility flag in public catalogs or character skill payloads.
2. At the start of an eligible character's progression, generate and persist one required mastery count from a server-configured inclusive range. The proposed initial tuning range is 2–5, but it must be reviewed against the size of the combat-skill catalog before activation.
3. Count distinct eligible combat skills whose authoritative level is exactly 99. A skill counts once, regardless of XP above its level-99 threshold. Non-combat skills never count.
4. The required count is fixed for that character for the lifetime of the challenge. It must not reroll on login, restart, character transfer, or repeated API calls.
5. The client never chooses the count, the random seed, eligible skills or unlock result. Generation uses a cryptographically secure server-side random source and persists only the resulting count and challenge version. Do not expose the count in public catalogs or ordinary character/skills payloads.
6. Evaluate completion in the same database transaction that awards the XP which may take a combat skill to level 99. Persist the unlock exactly once with a unique constraint/idempotent transition. Concurrent XP awards must not create duplicate unlocks or duplicate entitlements.
7. The unlock is an account entitlement earned by one eligible character. It permanently permits creation of the Jedi character archetype even if the qualifying character is later deleted, subject to the account's normal character-slot policy. The unlock record retains the earning character ID and timestamp for audit.
8. Only one challenge may be active per eligible character. A later design may support multiple Jedi characters, but the first release should use a clear server-enforced account limit to prevent accidental duplicate creation.

### Combat roster dependency

The approved core roster contains 18 active ordinary skills. Only the six combat specializations—Blades, Polearms, Unarmed, Pistols, Rifles and Magic—are eligible for Jedi mastery. Melee and Ranged parent skills, Defense and Strength co-skills, and all non-combat skills do not count. Retired Combat and Tactics progress keys are preserved only for compatibility and are not eligible. Accuracy is not a skill. The proposed random 2–5 requirement remains technically possible but should be rebalanced against a six-skill mastery pool before implementation; a 5-of-6 requirement may be excessively restrictive. The unlock itself remains unimplemented until challenge persistence, transaction integration, concurrency tests and archetype gates are built. The eligibility flag and category are not serialized to public API responses. See SKILL_ROSTER.md and COMBAT_PROGRESSION.md.

## Jedi character and progression

- Character creation gains a server-authorized archetype, not a client-controlled isJedi flag. The Jedi option appears only after the account entitlement exists.
- A Jedi character starts with a separate progression profile and its own skill set. It does not inherit the earning character's levels, inventory, bank, coins, equipment or combat bonuses.
- The Jedi progression is the umbrella mastery skill. Separate trainable disciplines can include Force Control, Force Sensitivity, Telekinesis, Force Defense, Force Focus and Lightsaber Mastery. Final names, XP curves, abilities and unlock levels are content decisions.
- Force powers are server-defined abilities with validated costs, cooldowns, range, target rules, line-of-sight where applicable, and authoritative effects. Client animation or requests never determine damage, healing, control duration, resource consumption or success.
- Jedi character creation and skill progression are separate from ordinary character progression. A character's archetype and progression rules are immutable after creation unless an explicit future respecialization feature is designed.

## Lightsaber weapon families

Represent these as distinct server item definitions / weapon families, not cosmetic labels:

- One-handed lightsaber
- Dual-blade lightsaber
- Two-handed lightsaber
- Dual-wield lightsabers (two separately equipped one-handed weapons)

Weapon validation must enforce compatible equipment slots and hands. Dual-wield requires two compatible one-handed item instances; a dual-blade weapon occupies its intended two-hand slot and must not be counted as two independently equipped weapons. Weapon-family mastery can influence abilities, animations and derived combat behavior, but the server remains authoritative.

The current inventory uses stack-based item quantities and equipment maps. Before introducing individual lightsaber variants with durability, quality, color crystals or per-item modifiers, implement unique item instances and transactional equipment invariants. Do not simulate dual-wield by placing one item ID into two slots if the underlying item model cannot represent two independent instances.

## Persistence and API shape (future implementation)

Proposed durable records:
- jedi_unlock_challenges: character ID, challenge version, required mastery count, created timestamp, completion timestamp; unique active challenge per character.
- account_unlocks: account ID, unlock key, source character ID, earned timestamp; unique (account_id, unlock_key).
- Character archetype field or normalized character-profile table with a server-owned enum (standard, jedi).
- Jedi skill progress stored in the existing skill-progress structure only after the character archetype permits those skill IDs, or in a normalized progression table if archetype-specific schemas become more complex.

Proposed server behavior:
- Generate the challenge during eligible character initialization using crypto/rand; persist it transactionally.
- Evaluate mastery after each authoritative combat-skill XP award, under the existing owned-character row lock and in the same transaction.
- Expose a generic hidden-state-safe character creation capability only after entitlement exists. Do not put hidden challenge fields in the public skills endpoint or world catalog.
- Make Jedi character creation idempotency-safe and enforce account slot limits in the database transaction.
- Ensure all XP paths (combat, rewards, admin grants and future quests) call the same domain progression evaluator. No alternate route may bypass unlock checks.
- Add audit events for challenge creation, completion and Jedi character creation without logging secrets or exposing hidden requirements prematurely.

## Security, testing and rollout gates

Tests must cover:
- Random count bounds, secure generation failure and stable persistence across repository recreation.
- Challenge count does not reroll on restart or repeated initialization.
- Only explicitly categorized combat skills count; non-combat level 99 skills do not.
- A skill at level 98 does not count; level 99 counts exactly once.
- Concurrent XP awards that complete the threshold create one account unlock.
- A character cannot create Jedi archetype without the account entitlement, even with a forged request.
- A qualifying character can unlock the account, and a different character on that account can create the Jedi archetype.
- Duplicate creation attempts, slot-limit races, account ownership violations and character deletion after unlock.
- Jedi skills cannot be awarded to standard characters; standard skills and Jedi-only skills cannot be injected by editing JSON or calling hidden endpoints.
- Force powers and weapon-family rules are validated server-side, including cooldowns, costs, range, incompatible slots and dual-wield item identity.
- Migrations work on both a fresh database and an existing populated database.

Rollout sequence:
1. Maintain the six explicitly eligible combat specializations and complete combat ability/weapon loops in focused slices.
2. Implement and test the hidden challenge and account entitlement without exposing Jedi in the client.
3. Add character archetypes and secure Jedi character creation.
4. Add Jedi skills and Force ability definitions.
5. Add unique lightsaber item instances, equipment rules, combat behavior and client presentation.
6. Enable the discovery only after database migration, concurrency, abuse and client compatibility tests pass.

No Jedi unlock logic, secret challenge values, Jedi skill catalog entries or character-creation options should be exposed in production until all rollout gates pass.
