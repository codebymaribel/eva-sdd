# Feature specs

Each feature lives in its own folder, created by `/eva-specify`:

```
specs/
└── 004-user-login/
    ├── spec.md          # what & why          (phase 1)
    ├── clarify.md       # open questions      (phase 1b, never edits the spec)
    ├── plan.md          # how                 (phase 2)
    ├── tasks.md         # ordered steps       (phase 3)
    └── verification.md  # evidence it works   (phase 5)
```

- Folder name: three-digit number + kebab-case slug. The number is unique across all branches.
- Branch name: `feat/NNN-slug`, so `eva check` can link code changes to the spec.

## Status lifecycle

Every document has a `type` (`spec`, `clarify`, `plan`, `tasks`, `verification`) and a `created` date in its front-matter; `spec.md`, `plan.md` and `tasks.md` also have a `status`:

```
draft ──► in-review ──► approved ──► done (spec only, after verification)
              ▲             │
              └─────────────┘  any edit to an approved document
```

| Who | Can set |
|-----|---------|
| Agent | `draft`, `in-review` |
| Developer | `approved` (+ `approved_by`, `approved_at`), `done` (+ `closed_at`) |

Dates are `YYYY-MM-DD` and happen in order: `created` ≤ `approved_at` ≤ `closed_at`.

Rules enforced by `eva check`:

- `plan.md` exists only after `spec.md` is approved; `tasks.md` only after `plan.md` is approved.
- Tasks are ticked only after `tasks.md` is approved.
- Every document has the `type` of its file; `created` is set once it leaves `draft`.
- An approved document has `approved_by` and `approved_at`; a `done` document has `closed_at`. Neither date is set before its status.
- An approved spec defines at least one `FR-n` in EARS form under *Functional requirements*, ids are unique, it has *Completion criteria* and no `[NEEDS CLARIFICATION]` left.
- An approved spec has a `clarify.md` with no open findings (`- [ ] **C-n**`); every finding has a unique id, a category (`[ambiguity]`, `[contradiction]`, `[edge-case]`, `[constitution]`) and, once ticked, an answer.
- From `in-review` (warnings) and once approved (errors): the plan cites only `FR-n` ids the spec defines and traces every one in *Requirements traceability*; every task has a type (`[feat]`, `[fix]`, `[improve]`), `Files`, `Done when`, at most 5 code files (more than 3 is a warning), only existing `FR-n` ids, and `[feat]`/`[fix]` tasks cite an FR and list their test files with the FR they verify; every `FR-n` has at least one task.
- A spec is `done` only when every task is ticked and `verification.md` lists every `FR-n` as passing with `**Verdict:** PASS`.
- Code changes on `feat/NNN-*` require spec, plan and tasks of feature `NNN` to be approved.
