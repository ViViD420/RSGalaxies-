# Source JSON Reference Method

Follow `docs/NEW_PROJECT_SOURCE_REFERENCE_GUIDE.md` and `docs/DESIGN_SYNTHESIS.md` whenever working with the Star Wars Galaxies (SWG), Old School RuneScape (OSRS), and RuneScape 3 (RS3) reference summaries.

## Rules

1. **Discover and validate real files.** Enumerate actual paths, validate JSON syntax, and read every embedded Markdown section in full. Never invent filenames or stop at metadata keys.
2. **Keep sources separate before synthesizing.** Analyze SWG, OSRS, and RS3 mechanics independently first, then compare them across domain matrices (world, progression, gathering, crafting, combat, HAM/attributes, loot, economy, UI/mobile).
3. **Attribute and originalize.** For every adopted or adapted pattern, record the source game, the player motivation it serves, trade-offs, and the original RS Galaxies adaptation. Never copy copyrighted names, lore, quest text, UI artwork, models, animations, or code.
4. **Reconcile against live code.** Map every design decision in `docs/DESIGN_SYNTHESIS.md` and `docs/DESIGN_DECISIONS.md` to concrete server (`server/internal/game/*`) and client (`client/scripts/*`) implementations and tests.
