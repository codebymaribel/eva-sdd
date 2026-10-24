# eva-implement — implement approved tasks

**Goal:** implement the next task(s) of an approved `tasks.md`, each with passing tests.

**Input:** a feature id, optionally followed by a task id (`T004`) or `all`. Without a task id, implement the next unticked task. Default is **one task per run**; continue to the next only if the input says `all` or the developer says to continue.

## Gate check — do this first

1. Open `specs/NNN-slug/tasks.md`. If its `status` is not `approved`, **stop** and tell the developer.
2. Confirm `spec.md` and `plan.md` are also `approved`. If not, stop.

## For each task

1. **Read** the task, the plan sections it relates to, the stack file and the code it touches.
2. **Scope:** only create or modify the files listed in the task (and the plan's file map). If you discover you need another file, a new dependency or a different approach, **stop and ask** — propose the plan change instead of improvising.
3. **Tests first (mandatory).**
   - `[feat]` / `[fix]`: write or update the tests listed in the task **before any production code**. Test names start with the requirement id (`FR-1: …`). Run them and confirm they **fail for the expected reason** (missing behaviour, not a typo or a broken import). Record it under *Implementation notes*: `T00X red: <test> failed with "<reason>"`. If a test passes before the code exists, it does not test the requirement: fix the test first.
   - `[improve]`: run the existing suites and confirm they are green before you change anything; they must stay green with no test changed to fit the new code.
4. **Implement** the smallest change that makes the tests pass, following the constitution and the stack checklist (performance rules apply now, not only at verification). Stay within the task's type: a `[improve]` task changes no behaviour, a `[feat]`/`[fix]` task only the behaviour of its FRs.
5. **Validate:** run typecheck, lint and the relevant tests (commands are in the stack file). Fix failures before moving on. Never disable a test or a lint rule to get green.
6. **Record:** tick the task (`- [x]`). If anything deviated from the plan, add a line under *Implementation notes* with the task id.
7. **Commit** only if the developer has asked you to commit; the commit type is the task's type: `<feat|fix|improve>(NNN): T00X <description>`.

## Stop conditions

Stop and report (do not push through) when:

- a task requires touching files outside the file map;
- the plan or spec turns out to be wrong or incomplete;
- a test cannot pass without changing a requirement (`FR-n`);
- typecheck, lint or tests fail for reasons outside the task.

## Finish

Reply with: tasks completed, files changed, test results, implementation notes, and the next task. When every task is ticked, suggest `/eva-verify NNN`.

Never edit the front-matter status of `spec.md`, `plan.md` or `tasks.md` in this phase, except setting an approved document back to `in-review` if the developer agreed to change it.
