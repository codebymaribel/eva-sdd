# eva-verify — verify a finished feature

**Goal:** prove, requirement by requirement, that the implementation satisfies the spec, the constitution and the stack checklist, and write `specs/NNN-slug/verification.md` with a final verdict for the developer's review.

**Input:** a feature id, or inferred from the branch `feat/NNN-slug`.

## Gate check — do this first

1. Open `specs/NNN-slug/tasks.md`. If any task is unticked, stop and list them.

## Verify

1. **Automated checks.** Run typecheck, lint, the unit/integration tests and, if configured, the e2e tests (commands in the stack file). Run `eva check` if it is installed. Record each command and its result.
2. **Requirement by requirement.** Go through every `FR-n` of the spec, in order. For each one, find the test(s) whose name starts with `FR-n`, confirm they pass and that they really assert the EARS statement (trigger/state/condition → response). Record one row per FR: ✅ with the test, or ❌ with what is missing. An FR without a passing test is ❌.
3. **Completion criteria.** Check every item of the spec's *Completion criteria* with evidence.
4. **Performance.** Review each rule listed in plan §10, plus any other rule from the stack checklist that applies to the changed code. For each one, record ✅/⚠️/❌ with evidence (file:line or a measurement). Do not mark ✅ without evidence.
5. **Scope.** Compare changed files (`git diff --name-only <base>...HEAD`) with the plan's file map and report anything outside it.
6. **Constitution.** Re-run the constitution check against the final code.

## Report

1. Create `specs/NNN-slug/verification.md` from `.eva/templates/verification.md` and fill it. Front-matter: same `id`/`title`, `type: verification`, `created` (today, `YYYY-MM-DD`). Leave `closed_at` empty.
2. **Verdict.** Write `**Verdict:** PASS` only when every FR and completion criterion is ✅ and the automated checks pass; otherwise `**Verdict:** FAIL` and list what is missing. `eva check` refuses a `done` spec without a PASS verdict or with an FR that is missing or ❌.
3. If anything failed, say what and propose the fix (as a new task when it is real work). Do not fix it silently in this phase.

## Finish — then STOP

Reply with the FR table, the verdict and: *"If you agree, set `status: done` and `closed_at` in `spec.md` (and `closed_at` in `verification.md`), then open the pull request including `verification.md`."* If the verdict is FAIL, say what has to happen before the spec can be closed.

Never set `status: done` yourself.
