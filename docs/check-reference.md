# `eva check` reference

**English** · [Español](es/check-reference.md) · [Docs index](README.md)

`eva check` validates the SDD documents of a project and enforces the approval gate on feature branches. It is read-only.

```sh
eva check                 # documents + branch gate
eva check --no-git        # documents only
eva check --base develop  # base ref to diff against
eva check --branch feat/004-login   # branch name (e.g. detached HEAD)
eva check --strict        # EVA043 becomes an error
eva check --dir ../app    # project directory (default .)
```

Output: one line per finding (`ERROR` or `WARN`, code, path, message), sorted by path, then a summary or `OK`. Exit codes: `0` no errors (warnings allowed), `1` at least one error, `2` invalid flags.

## How the gate works

1. **Branch:** `--branch`, else `GITHUB_HEAD_REF`, else the current branch.
2. **Base:** `--base`, else `origin/$GITHUB_BASE_REF` / `$GITHUB_BASE_REF`, else the first that exists of `origin/main`, `main`, `origin/master`, `master`. On the base branch itself nothing is gated.
3. **Changed files:** everything that differs from the merge-base with the base, including uncommitted and untracked files.
4. **Code:** any changed file outside `specs/`, `.eva/`, `docs/`, `.claude/`, `.opencode/` and Markdown files in the root.
5. **Feature:** the number in the branch name (`feat/004-login` → `004`) selects `specs/004-*`.
6. With code changes, the constitution and that feature's `spec.md`, `plan.md` and `tasks.md` must be approved.

## Severity

Most rules are always errors. Plan and task rules (EVA080–088) depend on the document's status: **warning** in `in-review` (so the agent can fix them before asking for review), **error** once `approved` or `done`, and not checked in `draft`. Spec content rules (EVA025–029) and the clarify gate (EVA070–071) apply once the spec is approved. Verification rules (EVA090–093) apply once the spec is `done`.

---

## Project

### EVA001 — not an eva project
**Error.** `.eva/config.json` is missing. **Fix:** run `eva init`.

### EVA010 — folder name
**Warning.** A folder in `specs/` is not `NNN-kebab-slug` (e.g. `Login Feature`). It is ignored. **Fix:** rename it, e.g. `specs/004-user-login`.

### EVA011 — duplicate feature number
**Error.** Two folders share a number (`004-login`, `004-signup`). **Fix:** renumber one; check other branches too (`git branch -a`).

## Documents

### EVA020 — missing spec
**Error.** A feature folder has no `spec.md`. **Fix:** run `/eva-specify` or remove the folder.

### EVA021 — invalid front-matter
**Error.** The file does not start with `---`, the block is not closed, or a line is not `key: value`. Applies to spec, clarify, plan, tasks and verification. **Fix:** copy the front-matter from `.eva/templates/`.

### EVA022 — id does not match the folder
**Error.** `id: 001-b` inside `specs/001-a/`. **Fix:** set `id` to the folder name.

### EVA023 — invalid status
**Error.** Status is not `draft`, `in-review`, `approved` or `done`.

### EVA024 — approved without approver
**Error.** `approved`/`done` with an empty `approved_by`. **Fix:** the developer who approves fills it — agents never do.

### EVA025 — open questions in an approved spec
**Error.** An approved spec still contains `[NEEDS CLARIFICATION …]` (text in backticks or code blocks is ignored). **Fix:** resolve it, or set the spec back to `in-review`.

## Spec content (approved specs)

### EVA026 — no requirements
**Error.** No list item starting with `FR-n` under a *Functional requirements* heading. **Fix:** add requirements in §4 — FRs elsewhere are only references.

### EVA027 — duplicate requirement id
**Error.** The same `FR-n` is defined twice. **Fix:** give the new one the next free number; never renumber.

### EVA028 — not EARS
**Error.** A requirement does not start with The/When/While/If/Where, has no `shall <response>`, or an `If` without `, then`. **Fix:** rewrite it, e.g. `If the password is empty, then the app shall show "Required".`

### EVA029 — completion criteria
**Error.** No *Completion criteria* section, or it has no list items (comments do not count). **Fix:** list verifiable criteria in §7.

## Phase order

### EVA030 — plan before spec approval
**Error.** `plan.md` exists but the spec is not approved. **Fix:** approve the spec first, or delete the plan.

### EVA031 — tasks before plan approval
**Error.** `tasks.md` exists but the plan is not approved.

### EVA032 — ticked before approval
**Error.** Tasks are ticked but `tasks.md` is not approved.

### EVA033 — done with open tasks
**Error.** The spec is `done` but `tasks.md` is missing, empty or has unticked items.

## Branch gate

### EVA040 — not a git repository
**Warning.** The gate is skipped. **Fix:** `git init`, or use `--no-git`.

### EVA041 — feature does not exist
**Error** when the branch has code changes, **warning** otherwise. The branch refers to `NNN` but there is no `specs/NNN-*`. **Fix:** create the spec with `/eva-specify`, or rename the branch.

### EVA042 — code before approval
**Error.** The branch has code changes and the constitution or the feature's spec, plan or tasks are not approved; the message lists what is pending (e.g. `tasks.md (in-review)`, `plan.md (missing)`). **Fix:** finish the phases, or move the code to another branch.

### EVA043 — code on a branch without feature id
**Warning** (error with `--strict`). Code changes on a branch that is not `…/NNN-slug`. **Fix:** use `feat/NNN-slug`.

### EVA044 — base branch not found
**Warning.** None of the base candidates exists; the gate is skipped. **Fix:** `--base <ref>`, or `fetch-depth: 0` in CI.

### EVA045 — detached HEAD
**Warning.** **Fix:** `--branch <name>`.

### EVA046 — git diff failed
**Warning.** The message includes git's error; the gate is skipped.

## Constitution

### EVA050 — missing constitution
**Error.** **Fix:** `eva init` restores it.

### EVA051 — invalid constitution
**Error.** Invalid front-matter, or a status other than `draft`, `in-review`, `approved`.

### EVA052 — approved without approver
**Error.** `approved_by` is empty.

### EVA053 — constitution not approved
**Warning** while there are no specs (a fresh project passes); **error** once a spec exists. **Fix:** `/eva-constitution`, then approve it.

### EVA054 — TODO left
**Error.** An approved constitution contains `TODO` (outside code). **Fix:** replace or remove every placeholder.

## Type and dates (every document)

### EVA060 — type
**Error.** `type` is missing or does not match the file: `constitution`, `spec`, `clarify`, `plan`, `tasks`, `verification`.

### EVA061 — date format
**Error.** A date is not a real `YYYY-MM-DD` date (`23/10/2026`, `2026-02-30`).

### EVA062 — created missing
**Error.** `created` is empty outside `draft` (always required in clarify and verification).

### EVA063 — approved_at
**Error.** `approved`/`done` without `approved_at`, or `approved_at` set while not approved. **Fix:** when you set a document back to `in-review`, clear `approved_by` and `approved_at`.

### EVA064 — closed_at
**Error.** `done` without `closed_at`, `closed_at` set before `done`, or any `closed_at` on the constitution.

### EVA065 — date order
**Error.** Within one document, `created ≤ approved_at ≤ closed_at` does not hold.

## Clarify

### EVA070 — clarify missing
**Error.** The spec is approved but `clarify.md` does not exist. **Fix:** set the spec back to `in-review` and run `/eva-clarify`.

### EVA071 — open findings
**Error.** The spec is approved while `clarify.md` has unticked findings. **Fix:** answer them, run `/eva-specify NNN`.

### EVA072 — duplicate finding id
**Error.** Two findings use the same `C-n`.

### EVA073 — category
**Error.** A finding has no category or one other than `[ambiguity]`, `[contradiction]`, `[edge-case]`, `[constitution]`.

### EVA074 — ticked without answer
**Error.** A finding is `- [x]` but its `Answer:` is empty. Only the developer answers.

## Plan and tasks (warning in review, error when approved)

### EVA080 — task type
A task has no `[feat]`, `[fix]` or `[improve]` tag.

### EVA081 — unknown requirement
The plan or a task cites an `FR-n` that the spec does not define.

### EVA082 — requirement without task
An `FR-n` of the spec is not cited by any task.

### EVA083 — feat/fix without requirement
A `[feat]` or `[fix]` task cites no `FR-n`. If it changes no behaviour, it is an `[improve]`.

### EVA084 — Files or Done when missing
A task has no `Files:` line with backticked paths, or no `Done when:`.

### EVA085 — tests
A `[feat]`/`[fix]` task has no `Tests:` line with a test file, or the line does not say which `FR-n` the tests verify.

### EVA086 — task too big
More than 3 code files is always a **warning**; more than 5 follows the status (error once approved). Tests, `.md` and `.snap` files do not count. **Fix:** split the task.

### EVA087 — duplicate task id
Two tasks share a `T` number.

### EVA088 — plan traceability
The plan has no *Requirements traceability* section, or an `FR-n` of the spec is not in it.

## Verification (spec done)

### EVA090 — verification missing
**Error.** **Fix:** run `/eva-verify`.

### EVA091 — requirement not verified
**Error.** An `FR-n` has no row in the *Requirements traceability* table.

### EVA092 — verdict
**Error.** No `Verdict:` line, or it is not `PASS` (a placeholder like `PASS | FAIL` counts as missing).

### EVA093 — failed requirement
**Error.** An FR row is marked ❌ or `FAIL` while the spec is `done`.
