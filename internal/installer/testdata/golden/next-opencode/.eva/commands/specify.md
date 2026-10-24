# eva-specify — write or revise a feature spec

**Goal:** turn the developer's feature description into `specs/NNN-feature-slug/spec.md` (the *what* and *why*), then stop for review.

**Input:** a feature description, or an existing feature id (`NNN` or `NNN-slug`) plus feedback to revise it (or the answers in its `clarify.md`).

## Before you start

1. **Gate check:** open `.eva/constitution.md`. If its `status` is not `approved`, **stop** and tell the developer to run `/eva-constitution` and approve it first.
2. Read `AGENTS.md`, `.eva/constitution.md` and `.eva/config.json` (active stacks and their paths).
3. If the input is an existing feature id, open its `spec.md` and jump to **Revising**.

## Creating a new spec

1. **Number the feature.** List `specs/` and take the highest `NNN` + 1, zero-padded to three digits. Also run `git branch -a --list '*feat/*'` and skip numbers already used by other branches, so two developers do not collide.
2. **Name it.** Slug = 2–5 lowercase words in kebab-case (`user-login`, `offline-cart-sync`). Directory: `specs/NNN-slug/`.
3. **Branch.** If you are on the default branch, ask the developer whether to create `feat/NNN-slug`. Do not create it without asking.
4. **Interview — before writing anything.** Ask the developer questions **one at a time**, waiting for each answer, **at most 6 in total**. Pick the questions with the biggest impact on scope, user-visible behaviour, error handling and data. Do not ask what the description, the constitution or the code already answer. Stop early when nothing important is left. Typical topics:
   - who uses the feature and the main flow, step by step;
   - what must happen on errors, empty states and offline;
   - limits and numbers (sizes, counts, time-outs, supported versions);
   - what is explicitly out of scope;
   - how the developer will decide the feature is finished.
5. **Create the file** by copying `.eva/templates/spec.md` to `specs/NNN-slug/spec.md`. Fill the front-matter: `id: NNN-slug`, `type: spec`, `title`, `status: draft`, `stack` (from `.eva/config.json`; ask if ambiguous), `author` (`git config user.name`), `created` (today, `YYYY-MM-DD`). Leave `approved_by`, `approved_at` and `closed_at` empty.
6. **Write the spec.**
   - Describe behaviour from the user's point of view. No libraries, file names, endpoints, components or database tables.
   - Number user stories `US-n` with a priority.
   - Write **functional requirements** `FR-n` in §4, one list item each, in **EARS** form: `The …`, `When …`, `While …`, `If …, then …` or `Where …`, followed by `the <system> shall <observable response>`. One behaviour per requirement; every user story has at least one. No Gherkin.
   - Turn every edge case that changes behaviour into a requirement (usually `If …, then …`).
   - Fill **completion criteria** (§7): concrete, verifiable statements of when the feature is finished. Keep the first item ("Every FR in §4 has at least one passing test named after it").
   - Fill non-functional requirements and out of scope. Prefer concrete numbers ("under 1 s", "1,000 items") over adjectives ("fast", "many").
   - Do not invent requirements. If something is still ambiguous after the interview, write `[NEEDS CLARIFICATION: <specific question>]` inline and add it to §10.
7. **Self-review** against §11 of the spec and fix what fails.

## Revising

1. Apply the developer's answers and feedback; remove each resolved `[NEEDS CLARIFICATION]` marker.
2. **Answers in `clarify.md`.** For every open finding (`- [ ]`) whose `Answer:` the developer filled: apply the answer to the spec, tick the finding (`- [x]`) and add a line `  - Applied: <what changed, e.g. "FR-2 reworded, FR-7 added">` below the answer. Never tick a finding without an answer, never write or change an answer yourself, and leave unanswered findings open.
3. Never renumber or reuse an `FR-n` id: a new requirement takes the next free number.
4. If the spec was `approved`, set it back to `in-review` and clear `approved_by` and `approved_at` (an edit invalidates the approval), then tell the developer.

## Finish — then STOP

1. Set `status: in-review`. Never set `approved` or `done`, never fill `approved_by`, `approved_at` or `closed_at`.
2. Run `eva check --no-git` if it is installed and fix any finding in the spec.
3. Reply with: the file path, a 3–5 line summary, the list of requirements (`FR-n` one line each), the findings you applied (if revising) and the open questions (if any).
4. End with: *"Please review `specs/NNN-slug/spec.md`, then run `/eva-clarify NNN`. A spec can only be approved once `clarify.md` exists and has no open findings; then set `status: approved`, `approved_by` and `approved_at`, and run `/eva-plan NNN`."*
5. Do **not** create `clarify.md`, `plan.md`, `tasks.md` or any code in this phase.
