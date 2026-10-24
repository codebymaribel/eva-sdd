# eva-tasks — break the plan into tasks

**Goal:** turn an approved plan into `specs/NNN-slug/tasks.md`: small, ordered, typed, test-first steps. Then stop for review.

**Input:** a feature id (`NNN` or `NNN-slug`), or inferred from the branch `feat/NNN-slug`.

## Gate check — do this first

1. Open `specs/NNN-slug/plan.md`. If its `status` is not `approved`, **stop** and tell the developer the plan must be approved first.

## Write the tasks

1. Copy `.eva/templates/tasks.md` to `specs/NNN-slug/tasks.md`. Front-matter: same `id`/`title`, `type: tasks`, `status: draft`, `author`, `created` (today, `YYYY-MM-DD`). Leave `approved_by`, `approved_at` and `closed_at` empty.
2. Derive tasks from the plan:
   - Number them `T001`, `T002`, … in execution order. Group them in phases: Setup → one phase per user story in priority order → Polish.
   - **Type every task:** `[feat]` adds behaviour required by an FR; `[fix]` makes existing behaviour match its FR; `[improve]` changes no behaviour (refactor, performance, accessibility) and only touches files in the plan's file map.
   - **Cite requirements** at the end of the title: `T002 [feat] [US-1] Validate the form — FR-1, FR-2`. `[feat]` and `[fix]` cite at least one existing `FR-n`; never cite an id the spec does not define.
   - **Size by scope:** one task = one commit, at most 3 code files (5 is the hard limit; tests and docs do not count). If it needs more, split it.
   - Each task lists **Files** (backticked paths, all from the plan's file map), **Tests** (backticked test file(s) + the `FR-n` each verifies; for `[improve]`, "existing suites green" is enough) and **Done when** (a condition someone else can check, e.g. a command and its expected result).
   - **Tests first:** the tests of a task live in the same task as the code they test and are written before it. No "write all tests" task at the end.
   - Mark `[P]` only when the task shares no files and no unfinished dependency with the other `[P]` tasks in its phase.
   - Tag tasks with their user story `[US-n]`.
3. Coverage checks before you finish:
   - Every `FR-n` of the spec is cited by at least one task.
   - Every file in the plan's file map appears in at least one task.
   - The last phase includes an accessibility/performance pass against the plan's §10.

## Finish — then STOP

1. Set `status: in-review`. Never set `approved`, never fill `approved_by`, `approved_at` or `closed_at`.
2. Run `eva check --no-git` if it is installed and fix every finding about `tasks.md` (in review, problems show as warnings).
3. Reply with: the path, number of tasks per phase and per type, the FR → task map, and anything you could not map.
4. End with: *"Please review `tasks.md`. When you are happy, set `status: approved`, `approved_by` and `approved_at`, then run `/eva-implement NNN`."*
5. Do **not** write code in this phase.
