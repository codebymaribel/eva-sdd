---
id: NNN-feature-slug
type: tasks
title: Feature title
status: draft # draft | in-review | approved — only a human sets approved
plan: ./plan.md
author:
created: YYYY-MM-DD
approved_by:
approved_at: # YYYY-MM-DD, set by the approver
closed_at: # YYYY-MM-DD, set by a human together with status: done
---

# Tasks: Feature title

> Ordered, small steps derived from the approved `plan.md`. Every task:
>
> - has a **type**: `[feat]` new behaviour, `[fix]` behaviour that does not match its FR yet, `[improve]` no behaviour change (refactor, performance, accessibility) inside the plan's file map;
> - **cites the requirements** it implements at the end of its title (`— FR-1, FR-2`); `[feat]` and `[fix]` cite at least one, and every `FR-n` of the spec has at least one task;
> - is **one commit**: at most 3 code files (5 is the hard limit; tests and docs do not count);
> - lists **Files**, **Tests** and a verifiable **Done when**;
> - is **test-first**: the tests in *Tests* are written and seen failing before the code.
>
> Legend: `[P]` can run in parallel with other `[P]` tasks of the same phase (no shared files) · `[US-n]` user story.
> Only tick a box after the task's tests, typecheck and lint pass.

## Phase 1 — Setup

- [ ] T001 [improve] Short imperative description
  - Files: `path/to/config.ts`
  - Tests: existing suites green
  - Done when: typecheck and tests pass

## Phase 2 — US-1 (P1)

- [ ] T002 [feat] [US-1] Short imperative description — FR-1, FR-2
  - Files: `path/to/file.ts`, `path/to/file.test.ts`
  - Tests: `path/to/file.test.ts` — FR-1, FR-2 (unit)
  - Done when: the FR-1 and FR-2 tests pass

- [ ] T003 [feat] [P] [US-1] Short imperative description — FR-3
  - Files:
  - Tests:
  - Done when:

## Phase 3 — Polish

- [ ] T004 [improve] Accessibility and performance pass against the stack checklist
  - Files: as needed within the file map
  - Tests: existing suites green
  - Done when: checklist items from plan §10 verified

## Implementation notes

The red step of every `[feat]`/`[fix]` task, plus deviations, discoveries and decisions made while implementing (with the task id):

- T002 red: `path/to/file.test.ts` › "FR-1: …" failed with "…" before the code existed.

Anything that changes the plan must be raised with the developer first.
