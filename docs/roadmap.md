# Roadmap and decisions

**English** · [Español](es/roadmap.md) · [Docs index](README.md)

## v1

| # | Feature | Status |
|---|---------|--------|
| 1 | Constitution with status and version, `/eva-constitution`, gate (EVA050–054) | ✅ done |
| 2 | Spec: `FR-n` in EARS, interview (max 6 questions, one at a time), completion criteria, `type`/`created`/`approved_at`/`closed_at` in every document, date validation (EVA026–029, EVA060–065) | ✅ done |
| 3 | Clarify: `/eva-clarify` writes `clarify.md`, detection only; no approval with open findings (EVA070–074) | ✅ done |
| 4 | Plan traceability to FRs; typed, one-commit tasks with *Done when*; tests first; FR-by-FR verification with a verdict (EVA080–093) | ✅ done |
| 5 | Change: `/eva-change` reopens an unmerged spec (back to `in-review`) or, once on main, creates a new feature with `amends: NNN-slug`; `eva check` validates `amends` | planned |
| 6 | Fixes, improvements and maintenance: while the feature is open, a new `[fix]`/`[improve]` task in its `tasks.md` (back to `in-review`); once on main, `specs/NNN/fixes/NN-slug.md` with `restores: FR-n` and `specs/NNN/improvements/NN-slug.md`; `specs/maintenance/` for work that belongs to no feature; branches `fix/NNN-NN-slug`, `improve/NNN-NN-slug` and `fix/maintenance-NN-slug` with their own gate; commit type validation | planned |
| 7 | `AGENTS.md` rules: a mandatory "When you finish any task" verification section, and "never modify `specs/` without an explicit request" | planned |
| 8 | **Template hashes:** `eva init` records a SHA-256 of every file it writes in `.eva/eva-lock.json` (generated — it says so in a `_generated` field — committed, and marked `linguist-generated` in a managed block of `.gitattributes`). Untouched files whose template changed are updated; edited files are skipped and `eva init --dry-run --diff` shows what changed in the template; `eva init --prune` deletes obsolete files only if untouched; projects without a lock record hashes on their first run; line endings are normalised before hashing. `AGENTS.md`/`CLAUDE.md` blocks and `config.json` work as today | ✅ done |
| 9 | **Template placeholders:** text left from the templates (taken automatically from the embedded templates — every `<…>` — plus "Feature title", "Short imperative description" and `path/to/…`) is a warning in review and an error once approved, in every feature document and the constitution. Guidance tables in the templates (e.g. the EARS patterns) move to HTML comments | planned |
| 10 | **File map:** every `Files:` path of a task must be in the plan's file map (warning in review, error once approved); entries are exact paths or folders ending in `/`; file map, `Files:` and `covers` paths are relative to the repo root. Gate EVA047: code changed on the branch outside the file map is one grouped warning (error with `--strict`), only once tasks are approved; lockfiles and `ignore` patterns in `.eva/config.json` do not count; when it appears, `/eva-implement` stops and asks whether to add the file to the plan or revert it | planned |

Every feature updates the README, these docs (both languages) and the goldens.

## Design decisions

- **One namespace: eva.** Commands are `/eva-<phase>`, the project folder is `.eva/` and the manifest `.eva/config.json`, matching `eva init` / `eva check`. Generic `sdd` names could collide with other SDD tools, and eva never overwrites a file someone else wrote or edited.
- **Agent-agnostic.** `AGENTS.md` is the single source of truth; per-agent shortcuts only point to `.eva/commands/`.
- **Only humans approve.** Agents move documents to `in-review`; `approved`, `done`, `approved_by`, `approved_at` and `closed_at` are human-only. Approval is read from the file, never from chat.
- **EARS only** for requirements (no Gherkin): one testable statement per `FR-n`, ids never reused.
- **Clarify detects, specify applies.** Clarify never edits the spec; answers come from the developer; `/eva-specify` applies them and ticks the finding.
- **Severity follows status.** Plan and task rules are warnings in review (the agent can self-correct) and errors once approved; drafts are free.
- **Dates per document.** `created ≤ approved_at ≤ closed_at` is checked inside each document, not across documents, so reopening and re-approving a spec does not invalidate older plans. Dates feed `eva status` and `eva log NNN` in v2.
- **Fix vs change.** An approved spec is never edited to record a fix. If *what* the system does changes, it is a change (feature 5); otherwise a fix or improvement (feature 6).
- **Task size by scope**, not by time: one commit, ≤ 3 code files (hard limit 5), tests and docs excluded.
- **Tests first** is enforced structurally (test files and FRs declared per task) and by instruction (red step recorded); eva does not inspect git history for it.
- **Standard library only** in v1 so `go install` needs no third-party modules (the v2 wizard is the planned exception).
- **Templates in English**, embedded in the binary.

## Known limitations

- **Files from before the lock.** Projects installed before v1 feature 8 have no record of what eva wrote: files that differ from the templates are treated as edited, and obsolete files from that time (e.g. the old `sdd-*.md` shortcuts) are not detected by `--prune`.
- **File map.** `eva check` does not verify that task files belong to the plan's file map (the `[improve]` rule is instruction-only). Fixed by v1 feature 10.
- **Template placeholders.** An approved spec that still contains `<…>` placeholders from the template passes. Fixed by v1 feature 9.
- **Branch number.** The gate takes the first `NNN-` in the branch name, so future `fix/NNN-NN-slug` branches already map to feature `NNN`; dedicated rules come with feature 6.

## v2 (planned)

| # | Feature | Status |
|---|---------|--------|
| 1 | **`eva status`** — one table with every feature of the project: number and title, status of spec / plan / tasks, open clarify findings, tasks done/total, verification verdict, and `created` / `closed_at`. Works in CI (plain text) and in a terminal | planned |
| 2 | **`eva log NNN`** — chronological history of one feature from the front-matter dates: spec, fixes, improvements and the features that `amends` it, with event (`created`, `approved`, `closed`) and status. Needs v1 features 5 and 6 | planned |
| 3 | **Interactive wizard** — `eva init` with arrow-key menus, multi-select and confirmation (`charmbracelet/huh`) behind `prompt.Prompter`; the current line prompts stay as the fallback outside a terminal and in tests | planned |
| 4 | **Existing projects (brownfield), "spec on touch" mode** — `eva adopt` (Go, deterministic) prints the inventory: stack from `package.json`, candidate features from Expo Router / Next.js App Router routes, existing tests per feature and which features already have a spec (`covers`). `/eva-adopt <feature>` writes a baseline spec (`type: baseline`, `origin: reverse`) from the code; FRs it is unsure about are marked `[review]` with `Evidence:`, and each one is confirmed, turned into a fix (`restores: FR-n`) or discarded. A baseline spec cannot be approved while `[review]` items remain; missing tests are warnings | planned |
| 5 | **Ownership** — `covers` (globs from the repo root) says which code belongs to a feature: in the plan front-matter of regular features (`/eva-verify` updates it when closing) and in baseline specs. `eva owner <path>` tells which feature owns a file, and `/eva-plan` uses it to fill a new **Owner** column in the file map. Warnings: a file-map file owned by another feature with an empty or wrong Owner; two features claiming the same file; a branch touching another feature's code without declaring it | planned |
| 6 | **Claude Code hook** (optional, per agent) — before each edit, a hook installed by eva blocks files that are not in the current task's `Files:` and asks for approval first, turning detection into prevention. Other agents only if they offer an equivalent | planned |

Decisions for v2:

- Baseline specs use `docs/NNN-slug` branches and `docs(NNN):` commits (a standard Conventional Commits type); a `docs/` branch cannot contain code changes.
- The inventory is not committed: `eva adopt` prints it (`--out <file>` writes it for reading). Manual decisions (`ignore`, `group`) live in the committed `.eva/adopt.json`.
- `covers` is required in baseline specs and optional in regular features, where it lives in the plan because the spec never names files. It only produces warnings.
- The wizard brings the first third-party module (`huh`); accepted, with the line prompter as the fallback.

## v3 (ideas)

- Brownfield "full inventory" mode: batches and progress over the whole inventory, and a constitution inferred from the existing code.

## Ideas for later

- Releases with GoReleaser and a Homebrew tap, when v1 is published, so eva installs without Go.
- More agents (Cursor, GitHub Copilot, Gemini CLI) after v1, as they are requested.
