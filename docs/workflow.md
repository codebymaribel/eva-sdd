# Workflow and documents

**English** · [Español](es/workflow.md) · [Docs index](README.md)

This page is the reference for the SDD workflow that eva installs: the phases, who does what, the status lifecycle and the exact format of every document. The installed source of truth in a project is `AGENTS.md`, `.eva/commands/*.md` and `.eva/templates/*.md`.

## Phases

| # | Phase | Command | Output | Gate to continue |
|---|-------|---------|--------|------------------|
| 0 | Constitution | `/eva-constitution` | `.eva/constitution.md` (once per project) | Developer approves |
| 1 | Specify | `/eva-specify` | `specs/NNN-slug/spec.md` — what and why | — |
| 1b | Clarify | `/eva-clarify` | `specs/NNN-slug/clarify.md` — findings only | Developer answers; `/eva-specify` applies; developer approves the spec |
| 2 | Plan | `/eva-plan` | `plan.md` — how | Developer approves |
| 3 | Tasks | `/eva-tasks` | `tasks.md` — ordered, typed steps | Developer approves |
| 4 | Implement | `/eva-implement` | code + tests, one task at a time | Tests pass per task |
| 5 | Verify | `/eva-verify` | `verification.md` — FR by FR, verdict | Developer marks the spec `done` |

Agents without slash commands follow the same steps from `.eva/commands/<phase>.md`.

## Roles

| | Agent | Developer |
|---|-------|-----------|
| Writes documents | yes, and sets `status: in-review` | reviews and edits |
| `type`, `created` | fills them when creating a document | — |
| `status: approved` / `done` | **never** | yes |
| `approved_by`, `approved_at`, `closed_at` | **never** fills them; clears `approved_by`/`approved_at` when it edits an approved document | yes |
| Clarify answers | **never** writes or invents them | writes them |
| Starts the next phase | only when the previous document says `approved` in the file | — |

Approval is read from the file, not from chat: "looks good" is not approval.

## Status lifecycle

`spec.md`, `plan.md` and `tasks.md` have a `status`; the constitution has one too (without `done`):

```
draft ──► in-review ──► approved ──► done (spec only, after verification)
              ▲             │
              └─────────────┘  any edit to an approved document
```

Phase order enforced by `eva check`:

- specs exist only after the constitution is approved;
- `plan.md` exists only after `spec.md` is approved (and `clarify.md` has no open findings);
- `tasks.md` exists only after `plan.md` is approved;
- tasks are ticked only after `tasks.md` is approved;
- a spec is `done` only when every task is ticked and `verification.md` passes;
- code changes on `feat/NNN-slug` need the constitution and that feature's spec, plan and tasks approved.

## Front-matter

Every document starts with a flat `key: value` block between `---` lines. Comments after ` #` are ignored.

| Field | constitution | spec | clarify | plan | tasks | verification |
|-------|:-----------:|:----:|:-------:|:----:|:-----:|:------------:|
| `id` (= folder name) | — | ✔ | ✔ | ✔ | ✔ | ✔ |
| `type` | `constitution` | `spec` | `clarify` | `plan` | `tasks` | `verification` |
| `status` | ✔ | ✔ | — | ✔ | ✔ | — |
| `created` | ✔ | ✔ | ✔ | ✔ | ✔ | ✔ |
| `approved_by`, `approved_at` | ✔ | ✔ | — | ✔ | ✔ | — |
| `closed_at` | always empty | ✔ | ✔ | ✔ | ✔ | ✔ |
| other | `version` | `title`, `stack`, `author`, `issue` | `title`, `spec` | `title`, `spec`, `author` | `title`, `plan`, `author` | `title` |

### Dates

- Format `YYYY-MM-DD`, and it must be a real date (`2026-02-30` fails).
- `created` is required once a document leaves `draft` (always for clarify and verification).
- `approved_at` is required with `approved`/`done` and must be empty before.
- `closed_at` is required with `done` and must be empty before; the constitution never has one. For clarify and verification it is optional (the developer sets it when closing the feature).
- Within a document: `created ≤ approved_at ≤ closed_at`. Dates are not compared across documents, so re-approving a reopened spec does not break older plans.

## Constitution — `.eva/constitution.md`

Product context, 6–10 **core principles** (one sentence each plus a *Check:* line saying how a reviewer verifies it), engineering standards, testing policy, performance and security rules, a Definition of Done and an **Amendments** table. Amending an approved constitution increments `version`, adds a row to the table and sets it back to `in-review`. An approved constitution has no `TODO` markers.

## Spec — `spec.md`

The **what** and **why**, with no implementation details. Written after an interview of at most 6 questions, asked one at a time.

| Section | Content |
|---------|---------|
| 1–2 | Summary, problem and motivation |
| 3 | User stories `US-n` with priority P1/P2/P3 |
| 4 | **Functional requirements** `FR-n` in EARS form |
| 5 | Edge cases and error states (each one that changes behaviour becomes an FR) |
| 6 | Non-functional requirements |
| 7 | **Completion criteria** |
| 8–11 | Out of scope, dependencies, open questions, review checklist |

### Functional requirements (EARS)

One list item per requirement under the *Functional requirements* heading, starting with its id. `FR-n` mentioned anywhere else (criteria, questions) are references, not definitions. Ids are never reused or renumbered.

```md
- **FR-1** (US-1) When <trigger>, the <system> shall <observable response>.
```

| Pattern | Template |
|---------|----------|
| Ubiquitous | The <system> shall <response>. |
| Event-driven | When <trigger>, the <system> shall <response>. |
| State-driven | While <state>, the <system> shall <response>. |
| Unwanted behaviour | If <condition>, then the <system> shall <response>. |
| Optional feature | Where <feature is included>, the <system> shall <response>. |
| Complex | While <state>, when <trigger>, the <system> shall <response>. |

`eva check` accepts a statement that starts with The/When/While/If/Where, contains `shall` followed by a response, and — for `If` — contains `, then`.

### Completion criteria

A list of verifiable statements of when the feature is finished. The first item is always *"Every FR in §4 has at least one passing test named after it."* `/eva-verify` checks each one.

## Clarify — `clarify.md`

Written by `/eva-clarify` with the spec in `draft` or `in-review`. It only **detects**; it never edits the spec. At most 10 findings per run; re-runs keep ids and never reopen resolved findings. Leftover `[NEEDS CLARIFICATION: …]` markers in the spec become findings.

```md
- [ ] **C-1** [edge-case] (§5) What happens when the user submits while offline?
  - Answer:
```

- Categories: `[ambiguity]`, `[contradiction]`, `[edge-case]`, `[constitution]`.
- The developer writes the answer; `/eva-specify NNN` applies it, ticks the finding and adds `  - Applied: <what changed>`.
- If nothing is found, the file says `No findings.`
- A spec cannot be approved without `clarify.md` or with open findings.

## Plan — `plan.md`

The **how**. Sections: 1 Approach · **2 Requirements traceability** · 3 Constitution check · 4 Architecture · 5 Data model · 6 API contracts · 7 State · 8 Navigation · **9 File map** · **10 Performance** · **11 Testing strategy** · 12 Dependencies · 13 Risks · 14 Rollout · 15 Alternatives.

- **Requirements traceability:** a table with one row per FR → plan sections, components or files, test level. Every FR appears; the plan only cites FRs the spec defines.
- **File map:** the only files implementation may create or modify.
- **Performance:** cites rule ids from the stack checklist (`.eva/stacks/<stack>.md`) and how each is met.
- Cross-cutting decisions get an ADR in `docs/adr/` (`Status: proposed`).

## Tasks — `tasks.md`

```md
- [ ] T002 [feat] [P] [US-1] Validate the sign-in form — FR-1, FR-2
  - Files: `src/features/auth/SignInForm.tsx`, `src/features/auth/SignInForm.test.tsx`
  - Tests: `src/features/auth/SignInForm.test.tsx` — FR-1, FR-2
  - Done when: `npm test -- SignInForm` passes
```

| Rule | Detail |
|------|--------|
| Id | `T001`, `T002`, … unique, in execution order; phases Setup → one per user story → Polish |
| Type | `[feat]` new behaviour · `[fix]` behaviour that does not match its FR · `[improve]` no behaviour change, inside the file map |
| Requirements | cited at the end of the title (and in *Tests*); `[feat]`/`[fix]` cite at least one; every FR has at least one task |
| Size | one commit; at most 3 code files (more is a warning, more than 5 an error); tests, `.md` and `.snap` files do not count |
| Files | backticked paths, all from the plan's file map |
| Tests | `[feat]`/`[fix]`: backticked test files (`*.test.*`, `*.spec.*`, `*.e2e.*`, `*_test.*`, `__tests__/`, `e2e/`) and the FR each verifies; `[improve]`: "existing suites green" |
| Done when | a condition someone else can check |
| Optional tags | `[P]` parallel within the phase · `[US-n]` user story |

*Implementation notes* at the end record the **red step** of every `[feat]`/`[fix]` task (`T002 red: <test> failed with "<reason>"`) and any deviation.

## Implement

One task per run by default (`/eva-implement NNN [T00X|all]`):

1. Read the task, plan and stack file; touch only the task's files.
2. **Tests first:** write the tests, run them, confirm they fail for the expected reason, record it. `[improve]`: existing suites green before and after, no test changed to fit.
3. Implement the smallest change that passes; follow the stack checklist.
4. Run typecheck, lint and tests; tick the task; note deviations.
5. Commit only when asked: `<feat|fix|improve>(NNN): T00X <description>`.

It stops and asks when a task needs a file outside the file map, a new dependency, a different approach, or a change to an FR.

## Verification — `verification.md`

| Section | Content |
|---------|---------|
| 1 | Automated checks (typecheck, lint, tests, e2e, `eva check`) |
| 2 | **Requirements traceability:** one row per FR with the test and ✅/❌ |
| 3 | Completion criteria with evidence |
| 4 | Performance checklist (plan §10 + stack rules) with evidence |
| 5 | Constitution and scope (files outside the file map) |
| 6 | **Verdict:** `**Verdict:** PASS` or `**Verdict:** FAIL — what is missing` |

PASS only when every FR and completion criterion is ✅ and the automated checks pass. A `done` spec needs a PASS verdict with every FR listed and none failed.

## Folders, branches and commits

```
specs/
└── 001-sign-in/
    ├── spec.md
    ├── clarify.md
    ├── plan.md
    ├── tasks.md
    └── verification.md
```

- Folder: three-digit number + kebab-case slug; the number is unique across all branches.
- Branch: `feat/NNN-slug`, so `eva check` links code changes to the spec.
- Commits: `feat(NNN): T003 add login form validation` — the type is the task's type.
- The pull request links `specs/NNN-slug/` and includes `verification.md`.

## Stacks and ADRs

`.eva/stacks/<stack>.md` holds the baseline libraries, commands, folder structure, performance budgets, a performance checklist with rule ids (based on Vercel's React Native and React best-practice skills) and testing and accessibility conventions. Change it there, with an ADR, rather than per feature. ADRs live in `docs/adr/NNNN-title.md` from `.eva/templates/adr.md`.
