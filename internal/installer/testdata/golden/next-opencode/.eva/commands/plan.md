# eva-plan — write the technical plan

**Goal:** turn an approved spec into `specs/NNN-slug/plan.md` (the *how*), then stop for review.

**Input:** a feature id (`NNN` or `NNN-slug`). If none is given, infer it from the current branch name (`feat/NNN-slug`); if that fails, ask.

## Gate check — do this first

1. Open `specs/NNN-slug/spec.md` and read its front-matter.
2. If `status` is not `approved` (or `done`), **stop** and tell the developer the spec must be approved first. Do not write a plan for an unapproved spec, even partially.
3. Open `specs/NNN-slug/clarify.md`. If it is missing or has open findings (`- [ ]`), **stop** and tell the developer to finish `/eva-clarify` first.

## Research

1. Read the full spec, `.eva/constitution.md` and the stack file(s) in `.eva/stacks/` for the spec's `stack`.
2. Read the existing ADRs in `docs/adr/` so you do not reopen decisions already made.
3. Explore the code that the feature touches: existing features with similar patterns, shared UI, API client, state, navigation. Reuse what exists; match local conventions.
4. If something needed for the plan is unknown (an endpoint shape, a design), ask the developer instead of assuming.

## Write the plan

1. Copy `.eva/templates/plan.md` to `specs/NNN-slug/plan.md`. Front-matter: same `id` and `title` as the spec, `type: plan`, `status: draft`, `author`, `created` (today, `YYYY-MM-DD`). Leave `approved_by`, `approved_at` and `closed_at` empty.
2. Fill every section. In particular:
   - **Requirements traceability (§2):** one row per `FR-n` of the spec, pointing to the plan sections, components or files that implement it and the test level. Every FR appears; no part of the plan exists without a requirement (or a stated reason). Cite only FR ids the spec defines.
   - **Constitution check (§3):** one row per principle. A "No" needs a justification or the plan is not ready.
   - **File map (§9):** every file to create or modify, with its purpose. Implementation may not touch files outside this list.
   - **Performance (§10):** go through the stack checklist; for every rule that applies, cite its id and say how it is met. Note the expected effect on the stack's budgets.
   - **Testing strategy (§11):** every `FR-n` in the spec maps to at least one test with its level and file.
   - **Dependencies (§12):** justify each new package (size, maintenance, alternatives). If the choice is cross-cutting (state, navigation, data layer, styling, auth), draft an ADR in `docs/adr/` from `.eva/templates/adr.md` with `Status: proposed`.
3. Keep it proportional: a small feature gets a short plan, but no section is skipped.

## Finish — then STOP

1. Set `status: in-review`. Never set `approved`, never fill `approved_by`, `approved_at` or `closed_at`.
2. Run `eva check --no-git` if it is installed and fix every finding about `plan.md` (in review, problems show as warnings).
3. Reply with: the path, the approach in 3–5 lines, new dependencies, proposed ADRs, risks and open questions.
4. End with: *"Please review `plan.md`. When you are happy, set `status: approved`, `approved_by` and `approved_at`, then run `/eva-tasks NNN`."*
5. Do **not** create `tasks.md` or write code in this phase.
