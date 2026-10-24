# eva-constitution — propose or amend the project constitution

**Goal:** produce `.eva/constitution.md`: short, verifiable, non-negotiable principles for this project. Then stop for review. No spec can be written until the constitution is approved.

**Input:** optional context from the developer (what the project is, constraints), or the change to make to an approved constitution.

## Before you start

1. Read `AGENTS.md`, `.eva/config.json`, the stack file(s) in `.eva/stacks/` and the current `.eva/constitution.md`.
2. Look at the repository for facts you should not ask about: `package.json` (scripts, dependencies, TypeScript config), lint/format config, existing folder structure, CI files, README.
3. If the constitution is `approved`, jump to **Amending**.

## Proposing

1. **Interview.** Ask the developer questions **one at a time**, waiting for each answer, **at most 6 in total**. Only ask what you cannot infer from the repository and what changes the rules. Typical topics:
   - what the product is and who uses it;
   - quality bar: coverage expectation, required e2e journeys, supported OS/browser versions;
   - hard constraints: offline support, compliance, data residency, budgets;
   - what agents must never touch or add without asking;
   - language of code, UI copy and documentation.
2. **Write** `.eva/constitution.md` keeping its structure. Replace every `TODO` with real values or remove the line. Keep the default principles that fit, adjust the ones that do not, and add project-specific ones.
3. **Principles must be short and verifiable.** One sentence each plus a *Check:* line saying how a reviewer verifies it. No adjectives without a threshold ("fast", "clean", "robust"). Aim for 6–10 principles.
4. **Front-matter:** `status: in-review`, `version: 1`, `created` = today (`YYYY-MM-DD`) if empty. Leave `approved_by` and `approved_at` empty.

## Amending

1. Never edit an approved constitution on your own initiative. Only amend it when the developer asks for a specific change.
2. Apply the change, increment `version`, add a row to the *Amendments* table (version, today's date, one-line summary) and set `status: in-review`.
3. Show the diff of the change in your reply.

## Finish — then STOP

1. Never set `status: approved`, never fill `approved_by`/`approved_at`.
2. Reply with: the list of principles (one line each), what you inferred from the repository, and anything still uncertain.
3. End with: *"Please review `.eva/constitution.md`. When you are happy, set `status: approved`, `approved_by` and `approved_at`. Then start a feature with `/eva-specify <description>`."*
4. Do not create specs or code in this phase.
