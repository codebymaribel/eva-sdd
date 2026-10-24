---
type: constitution
status: draft # draft | in-review | approved — only a human sets approved
version: 1
created:
approved_by:
approved_at:
closed_at: # always empty: a constitution is amended, never closed
---

# Constitution — demo

The non-negotiable rules of this project. No spec is written until this document is approved. Every plan includes a *Constitution check* against it, and every verification re-checks it.

> Agents: read this file before any SDD phase. Draft or revise it only through `/eva-constitution` (`.eva/commands/constitution.md`). Never set `status: approved`.
>
> Developers: run `/eva-constitution` to get a proposal, or replace every `TODO` below yourself. An approved constitution cannot contain `TODO` markers. Defaults are opinionated starting points — change them.

## 1. Product

- **What we are building:** TODO — one paragraph.
- **Primary users:** TODO
- **Platforms:**
  - Next.js (App Router) (`.`) — see `.eva/stacks/nextjs.md`
- **Out of scope for the product:** TODO

## 2. Core principles

Short and verifiable: each principle says how a reviewer checks it.

1. **The spec rules.** No behaviour is implemented unless an approved spec requires it; if a decision is missing, work stops and the developer is asked. *Check:* every task cites the requirements it implements.
2. **Tests are the gate.** Each task ends with its tests green, written before the code; no task advances with red tests. *Check:* every requirement maps to at least one passing test.
3. **Performance is a feature.** Every plan applies the stack checklist; regressions against the stack budgets block the merge. *Check:* `verification.md` reports each applicable rule with evidence.
4. **Scalable by structure.** Code is organised by feature; features do not import each other's internals. *Check:* cross-feature imports appear in the plan's file map.
5. **Type safety.** TypeScript `strict: true`; no `any`, `@ts-ignore` or non-null assertion without a comment explaining why. *Check:* typecheck passes and a search finds no unexplained escapes.
6. **Small, reviewable changes.** One task = one commit touching at most 3–5 source files. *Check:* the diff of each commit.
7. **Simplicity.** No speculative abstractions; no new dependency without a reason in the plan (or an ADR when cross-cutting). *Check:* the plan's dependency section.
8. **Accessible by default.** Interactive elements have accessible names and roles; text scales with system settings; contrast meets WCAG AA. *Check:* the accessibility items in the stack file.

## 3. Engineering standards

- **Language:** TypeScript (strict). JavaScript files are not added.
- **Formatting & linting:** Prettier + ESLint; the CI fails on lint errors.
- **Naming:** components `PascalCase.tsx`, hooks `useThing.ts`, utilities `camelCase.ts`, tests next to the code as `*.test.ts(x)`.
- **Imports:** import from the file that defines the symbol; no barrel (`index.ts`) re-export files.
- **Dependencies:** adding, removing or upgrading a major version is declared in the plan. Cross-cutting choices (state, navigation, data layer, styling) are recorded as ADRs in `docs/adr/`.
- **Errors:** user-facing errors are handled and shown; unexpected errors are reported (TODO: error reporting tool).
- **Secrets:** never committed; read from environment configuration.

## 4. Testing policy

- **Unit / component tests** for logic, hooks and components (see the stack file for tools).
- **Integration tests** for flows that cross modules (data fetching + UI state).
- **End-to-end tests** for the critical user journeys: TODO list them.
- **Traceability:** test names start with the requirement id, e.g. `it('FR-2: shows an error when the password is empty', …)`.
- **Coverage:** new or changed code ≥ 80% lines (TODO: adjust). Coverage is a signal, not the goal.
- Tests must be deterministic: no real network, no real time, no order dependence.

## 5. Performance & scalability

The budgets and checklist for each platform live in its stack file. Plans reference checklist items by id (e.g. `list-performance-virtualize`, `async-parallel`) and the verification reports each applicable item.

## 6. Security & privacy

- Validate all external input (forms, URL params, API responses) with schemas.
- Authenticate and authorise on the server for every mutation.
- Store tokens only in secure storage (mobile) or httpOnly cookies (web).
- Personal data collected by a feature is listed in its spec.

## 7. Definition of Done

- [ ] Spec, plan and tasks approved by a developer.
- [ ] All tasks ticked; every requirement (`FR-n`) has a passing test; the spec's completion criteria are met.
- [ ] Typecheck, lint and test suites pass in CI.
- [ ] Stack performance checklist reviewed in `verification.md`, whose verdict is PASS.
- [ ] `eva check` passes.
- [ ] Pull request reviewed and approved by a developer.

## 8. Amendments

Any change to an approved constitution sets it back to `status: in-review`, increments `version` and adds a row below. A developer approves it again in a dedicated pull request. Architectural changes also add an ADR.

| Version | Date | Summary |
|---------|------|---------|
| 1 | | Initial constitution |
