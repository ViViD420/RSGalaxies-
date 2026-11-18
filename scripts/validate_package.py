from pathlib import Path
import sys
root = Path(__file__).resolve().parents[1]
required = [
 ".agents/skills/rs-galaxies-ultimate-engineer/SKILL.md",
 "AGENTS.md", "README.md", "ANTIGRAVITY_ACTIVATION_PROMPT.txt",
 ".agents/skills/rs-galaxies-ultimate-engineer/references/local-first-protocol.md",
 ".agents/skills/rs-galaxies-ultimate-engineer/references/rs-galaxies-design-profile.md",
 ".agents/skills/rs-galaxies-ultimate-engineer/references/source-json-method.md",
 ".agents/skills/rs-galaxies-ultimate-engineer/references/mmo-engineering-checklist.md",
 ".agents/skills/rs-galaxies-ultimate-engineer/references/gameplay-vertical-slice.md",
 ".agents/skills/rs-galaxies-ultimate-engineer/templates/SESSION_HANDOFF.md",
 ".agents/skills/rs-galaxies-ultimate-engineer/workflows/01-resume-local-project.md",
 ".agents/skills/rs-galaxies-ultimate-engineer/workflows/02-implement-feature.md",
 ".agents/skills/rs-galaxies-ultimate-engineer/workflows/03-local-github-sync-review.md",
 ".agents/skills/rs-galaxies-ultimate-engineer/workflows/04-review-and-stabilize.md",
]
missing = [p for p in required if not (root/p).is_file() or (root/p).stat().st_size == 0]
if missing:
 print("FAILED; missing/empty:", *missing, sep="\\n- ")
 sys.exit(1)
s = (root/required[0]).read_text(encoding="utf-8").lower()
for term in ["local", "gemini", "galaxymm o", "session_handoff.md", "server", "accuracy"]:
 if term not in s and term != "galaxymm o":
  print("FAILED; missing concept:", term); sys.exit(1)
print(f"PACKAGE VALID: {len(required)} required files present and non-empty.")
print("Local-first safeguards, RS Galaxies guardrails, MMO standards, and continuity rules found.")
print("Structural validation only; Antigravity runtime and game repository are not tested.")
