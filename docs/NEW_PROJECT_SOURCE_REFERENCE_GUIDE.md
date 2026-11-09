# RS Galaxies — New Project Source Reference Guide

Last updated: 2026-11-09

## Purpose

This guide tells a future RS Galaxies development chat how to use the three user-supplied JSON reference summaries for Star Wars Galaxies (SWG), Old School RuneScape (OSRS), and RuneScape 3 (RS3), while preserving the existing implementation and the user's design intent.

The JSON files are research inputs. They are not source code, asset packs, implementation specifications, or permission to reproduce another game's protected expression.

## 1. First-contact checklist

Before making changes in a new ChatGPT project or code workspace:

1. Identify whether the user has created a ChatGPT Project, a new GitHub repository, a new local source tree, or some combination. Do not assume these are the same thing.
2. Enumerate all available source folders and actual files. Record exact paths.
3. Locate the three JSON reference summaries by content, not by assuming a filename.
4. Read each JSON completely. Validate its syntax and structure. If malformed, report the path, parser error and exact section/key; do not silently omit the file.
5. Discover how Markdown is embedded. It may be under keys such as markdown, content, summary, notes, description, body, or nested arrays. Do not assume one schema.
6. Extract and read every Markdown section, including nested sub-sections, appendices, tables and lists. Track completion per file/section.
7. Preserve the source-game identity and source attribution for every extracted design observation.
8. Only after all three are read, create the comparison matrix and RS Galaxies synthesis.
9. Read the current implementation repository's SESSION_HANDOFF, ARCHITECTURE and ROADMAP before changing code.
10. Inspect the current branch/HEAD and latest CI. Treat live repository state as authoritative over old chat notes.

## 2. Suggested JSON organization

The user may choose any layout. If they have not supplied one, this structure is recommended:

- source/references/star_wars_galaxies.json
- source/references/old_school_runescape.json
- source/references/runescape_3.json

Do not create these files or move user content unless asked. If a different layout already exists, use it.

A possible JSON shape is:

{
  "source_game": "Star Wars Galaxies",
  "short_name": "SWG",
  "summary_version": 1,
  "last_reviewed": "YYYY-MM-DD",
  "sections": [
    {
      "id": "professions_and_skills",
      "title": "Professions and Skills",
      "markdown": "# Professions and Skills\n\nFull research notes...",
      "topics": ["professions", "skills", "specialization"],
      "design_signals": ["player identity", "economic interdependence"]
    }
  ],
  "references": [
    {
      "title": "Reference title",
      "url": "https://example.org/reference",
      "note": "What this source supports"
    }
  ]
}

This schema is illustrative, not mandatory. Adapt to the real files. JSON strings containing Markdown must be parsed as JSON before rendering/reading the Markdown. Preserve headings, tables, links, code examples, citations and section order. If the JSON stores one giant Markdown field, read it in complete sequential chunks and maintain a checklist of offsets/sections so no content is lost.

## 3. Research handling and evidence quality

For every source claim:

- Keep the source-game label (SWG, OSRS or RS3).
- Keep the section ID/title and any included citation/reference metadata.
- Distinguish documented fact from interpretation, opinion or inferred design intent.
- Note version/era differences where the summary mentions them. Do not flatten different historical versions into one timeless feature.
- If the summary contains conflicting descriptions, record the conflict and investigate it rather than selecting whichever is convenient.
- Do not invent details to fill gaps. Mark them unknown or seek reliable supporting sources if the user wants further research.
- Treat community opinions as evidence of player experience, not universal fact.
- Keep historical mechanics distinct from modern/private-server/community implementations.

A reference matrix entry should be traceable back to the exact source file and section. Example columns:

| Domain | Source | Observed pattern | Player need | Strength | Trade-off | RS Galaxies proposal | Decision |
|---|---|---|---|---|---|---|---|

## 4. Required design synthesis domains

Review each source for all applicable domains:

1. World structure, zones, exploration, travel, world scale, terrain, landmarks and environmental storytelling.
2. Character identity, creation, broad progression, skill leveling, professions, class restrictions and respecialization.
3. Gathering, resource nodes, quality, depletion/respawn, tools and resource circulation.
4. Crafting, recipes, stations, experimentation, quality, schematics, component chains and professions.
5. Combat, targeting, weapon families, accuracy, damage, defense, abilities, cooldowns, statuses, animations, telegraphs, enemy AI and group encounters.
6. Loot, drop tables, rarity, reward fairness, contribution, economic sinks/sources and reward UI.
7. Equipment, attributes, armor, durability, item identity, modifications, requirements and profession compatibility.
8. Economy, vendors, player trading, markets, currency, scarcity, taxes and anti-exploit protections.
9. Social play, chat, groups, guilds, parties, mentoring, events and community.
10. Housing, player cities, settlements, structures, ownership, permissions and decoration.
11. Quests, missions, narrative, exploration incentives and repeatable activities.
12. UI/UX, inventory, banks, skills, maps, action bars, tooltips, accessibility and mobile ergonomics.
13. Progression pacing, XP curves, level caps, prestige/mastery, endgame and retention.
14. Game feel and technical observations: movement, camera, animation timing, hit feedback, sound, readability, network authority and performance.

For each idea, record:

- Reference game and source section.
- Neutral description of the observed mechanic.
- The player motivation or need it serves.
- What works and why.
- Costs, weaknesses, exploits or design trade-offs.
- What could be adapted for RS Galaxies.
- What should be rejected or changed.
- Originalization requirements (names, tuning, lore, visuals, implementation).
- Dependencies on existing systems.
- Security, economy and multiplayer implications.
- Acceptance criteria and regression tests.
- Decision status: candidate, accepted, rejected, deferred, implemented, or needs user decision.

## 5. Initial lenses for the three references

These are starting hypotheses, not binding instructions. The actual supplied summaries and user decisions take precedence.

### Star Wars Galaxies

Study especially:
- Open-ended sandbox identity and player-defined roles.
- Profession structures, skill advancement and meaningful specialization.
- Crafting depth, experimentation, resource properties and component chains.
- Player-driven economy, vendors, trade and interdependence.
- Social play, player cities, housing and settlement systems.
- Exploration, world scale and emergent community activity.

Do not assume the game should reproduce SWG's exact profession tree, UI, item naming, setting, lore, maps, models, animations or numerical balance. Extract the underlying player motivations and rebuild them with original systems.

### Old School RuneScape

Study especially:
- Clear, readable skill progression and long-term goals.
- Gathering-to-crafting loops and the utility of resources/items.
- Strong connections between skills, equipment, quests, areas and unlocks.
- Understandable combat feedback and meaningful gear progression.
- Economy loops, resource scarcity and player trading.
- Consistent interaction patterns and recognizable progression milestones.

Avoid blindly importing OSRS's exact XP curve, level gates, item names, monsters, quest text, map layouts or distinctive visual style.

### RuneScape 3

Study especially:
- Modern UI affordances, information density and customization.
- Action/ability feedback, cooldown readability and combat presentation.
- Accessibility and quality-of-life systems.
- Inventory, bank, skill and equipment information design.
- Modern onboarding and player guidance.
- Performance/usability lessons from a mature live-service game.

Do not equate complexity with quality. Prefer clarity, optional depth, responsive controls and accessible information over visual clutter or needless menus.

## 6. RS Galaxies design pillars and known user decisions

The intended game is an original science-fiction sandbox MMORPG with:

- The open-ended sandbox, profession interdependence, crafting depth, social world, player economy, housing and settlements associated with the SWG reference.
- The broad, readable skill progression, gathering/crafting loops, useful items and long-term progression goals associated with OSRS.
- Modern usability, interface affordances, ability feedback and quality-of-life thinking informed by RS3.
- Original lore, world, factions, characters, names, quests, models, animations, UI artwork, music, sounds, assets and code.

Explicit player preference:
- Broad progression: characters may train many or all skills.
- Do not impose an arbitrary cap on the number of skills a character can train.
- Meaningful specialization should emerge from equipment/tool requirements, recipes, materials, stations, access, player choices and economic interdependence.
- Equipment may be profession/class-specific where appropriate; this is a meaningful gearing/progression gate, not a ban on learning other skills.

When a proposed mechanic conflicts with these decisions, flag the conflict and explain the trade-off. Do not quietly override user decisions based on a reference game.

## 7. Existing implementation baseline

The current code baseline is the repository https://github.com/ViViD420/RSGalaxies- on main. The project already has a Go/PostgreSQL server and Godot 4.3 client. It is not an empty project.

Major existing systems include:
- Account authentication, character ownership and persistence.
- 18-skill catalog and broad progression.
- Server-validated movement and starter zone.
- Gathering nodes and atomic gathering rewards.
- Crafting stations, recipes and transactional inventory changes.
- Inventory, bank, coins, equipment, merchant and consumables.
- Zone chat and multiplayer presence.
- Escrow-based player trade.
- Persistent enemy AI, target stickiness, encounter cycles and contribution tracking.
- Multi-item reward settlement and claims.
- A Stoneback boss vertical slice.

The client already has:
- Third-person movement, jump, camera look and bounded camera zoom.
- Responsive action toolbar and multiple gameplay panels.
- Procedural deterministic terrain with collision, elevation/detail/moisture noise, biome palette and integrated dirt trails.
- Initial server-confirmed melee swing animation, isolated to the visual mesh.
- Existing UI for skills, inventory/bank, equipment, merchant, crafting, chat, presence and player trade.

Read the complete session handoff and architecture docs for details. Do not replace working systems just because a new project has source summaries. Improve and extend the implementation in place unless the user explicitly chooses a separate repository and migration plan.

Important authority boundary: current server movement primarily validates X/Z position, speed, bounds and elapsed time. The Godot client generates terrain height and presentation. Do not treat client-only Y position, terrain collision or animation as server-authoritative. Any server terrain-height/jump/collision model requires a deliberate design and tests for slopes, gravity, latency, anti-cheat and respawn.

## 8. Source-to-implementation workflow

After completing the comparison matrix:

1. Create or update a design synthesis document, preferably docs/DESIGN_SYNTHESIS.md, with source-attributed observations and the selected original RS Galaxies design.
2. Create a decision log. Each decision should include rationale, alternatives considered, impacts, dependencies and status.
3. Map each accepted concept to existing code/docs/tests. Identify reuse, missing functionality and risks.
4. Separate immediate core-loop work from later features. The user currently wants movement, terrain, mechanics, animations, ability feedback, UI and world systems prioritized. Boss development is paused.
5. Convert accepted ideas into small milestones with measurable acceptance criteria.
6. Implement one coherent vertical slice at a time. Preserve server authority and transactional persistence.
7. Update architecture, roadmap and session handoff after each meaningful milestone.
8. Re-read the source summary only when a design question needs evidence; do not repeatedly dump all files into chat when a precise section is enough.

## 9. Core-first implementation priorities

Suggested order, to be revised after reading the JSONs:

### Movement and world foundation
- Audit input-to-server movement path, speed checks, timestamps, zone bounds, presence, respawn/rest and latency behavior.
- Review acceleration/deceleration, jump buffering, coyote time, camera orbit/zoom, grounding and input focus.
- Decide whether to retain the current X/Z server contract or design server-authoritative terrain height/jumping. Do not add a partial Y coordinate check.
- Improve deterministic biome transitions, terrain landmarks, resource placement, vegetation, rocks, paths and environmental identity.
- Add performance budgets, deterministic-seed tests, collision consistency checks and zone bounds tests.

### Game feel and animation
- Introduce an explicit presentation state machine: idle, walk, run, start/stop, turn, jump, fall, land, attack, gather, craft, use item, hit reaction, defeat and respawn.
- Separate presentation from authoritative state and collision.
- Trigger combat animation after server-confirmed action; handle cancellation, repeated attacks, target loss, death and scene changes.
- Build weapon-family and ability presentation with original animation/VFX/audio.
- Add interaction prompts, target outlines/markers, gather progress, hit feedback, damage numbers, cooldowns and resource depletion/respawn feedback.
- Keep tests for transitions, cleanup and renderer resource lifetime.

### UI and mobile
- Audit current panels, action callbacks, request cancellation, focus/input capture, modal stacking and stale state.
- Improve inventory, bank, equipment, skills, crafting, merchant, chat and trade consistency.
- Add clear validation/error/success states and useful item/recipe tooltips.
- Design actual mobile controls: movement stick, camera-look area, action/interact controls, safe areas and scalable touch targets. Do not simply shrink desktop controls.
- Headless CI does not replace physical device QA.

### World activities and progression
- Improve resource quality, tool requirements, resource loops, farming and skill unlocks.
- Deepen crafting chains and meaningful outputs while preserving server-side recipe validation and transactional inventory changes.
- Make equipment identity and profession requirements meaningful without preventing broad skill progression.
- Build exploration, quests, social systems and economy loops from the approved synthesis.

### Bosses and advanced content
Boss expansion is explicitly paused by the user. Do not spend the next development cycles adding bosses, threat systems or boss abilities. Revisit only after the core gameplay loop has been materially improved or the user requests it.

## 10. Absolute error-handling requirement

This rule overrides any preference for speed or autonomous progress:

- Never hide, suppress, filter away, ignore or misreport an error.
- Stop unrelated feature work as soon as a genuine error is found.
- Reproduce and inspect the full logs and relevant code.
- Identify the actual cause; label uncertain diagnoses as hypotheses.
- Fix the cause, add a regression test, rerun the affected workflow and inspect the full logs.
- If the first fix fails, document that failure and keep investigating.
- Never weaken tests or remove CI checks to get a green badge.
- Preserve failing run URLs and correction commits in SESSION_HANDOFF.
- Clearly separate verified behavior, assumptions, pending tests and manual QA.

## 11. Required handoff routine

After each meaningful development change:
- Commit the code and tests.
- Update docs/SESSION_HANDOFF.md with exact commits, CI run IDs, results, failure/fix history and next actions.
- Update docs/ARCHITECTURE.md when system behavior/contracts change.
- Update docs/ROADMAP.md with completed and pending milestones.
- Verify all three docs match the actual current branch.
- Before handing over, fetch current main HEAD and latest Actions again. Never rely on this guide's checkpoint as current state.

## 12. Suggested new-chat opening prompt

RS GALAXIES — CONTINUE FROM THE SOURCE-DRIVEN HANDOFF.

Read the entire docs/SESSION_HANDOFF.md and docs/NEW_PROJECT_SOURCE_REFERENCE_GUIDE.md first. Inspect the new RS Galaxies project's actual source folders, locate and fully read the Star Wars Galaxies, OSRS and RS3 JSON summaries, validate their JSON syntax, and extract every embedded Markdown section. Report the exact discovered paths and schemas. Do not invent paths or claim unseen files were read.

Use the three files as attributed design research. Create a comparison matrix and an original RS Galaxies design synthesis/decision log. Do not copy proprietary code, assets, maps, dialogue, UI artwork, animations, music, sounds or distinctive expression.

The existing implementation baseline is ViViD420/RSGalaxies- on main. Inspect its latest HEAD, commits and CI before coding. Preserve working systems and history. Do not touch the separate GalaxyMMO project. If this new project uses a different code repository, inspect it and confirm the migration target before pushing changes.

The user wants broad skill progression without arbitrary skill caps. Specialization comes from equipment, tools, recipes, materials, stations, access and economic interdependence. The game blends SWG-style sandbox depth, OSRS-style readable progression and RS3-informed modern usability, all in an original science-fiction world.

Boss work is paused. Prioritize core movement, terrain, world generation, game mechanics, animations, ability feedback, interaction, UI/UX, mobile controls, gathering, crafting and progression. Improve existing code rather than restarting or duplicating it. Work autonomously and keep handoffs current.

ABSOLUTE RULE: never hide an error. If any error, failing test, parser issue, runtime problem, CI failure or correctness-affecting warning is found, stop unrelated work, inspect the complete logs, fix the cause, add regression coverage and rerun the checks before moving on. Never weaken tests, filter errors or claim success without evidence. Record every failure and fix in the handoff.
