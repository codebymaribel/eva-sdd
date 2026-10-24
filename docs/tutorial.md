# Tutorial: one feature, end to end

**English** · [Español](es/tutorial.md) · [Docs index](README.md)

This tutorial takes a small feature — *sign in with email and password* in a React Native app — through every phase. You will see what the agent writes, what **you** edit to approve, and what `eva check` says at each step. All outputs below are real.

You need eva installed ([Getting started](getting-started.md)) and Claude Code (or OpenCode).

## 0. A playground project

Use a throwaway project so you can break things on purpose:

```sh
mkdir -p ~/Code/eva-playground && cd ~/Code/eva-playground
npx create-expo-app@latest .          # optional: real code for the gate
git init -b main                      # create-expo-app may already have created the repo
git add -A && git commit -m "init"
eva init --stack rn --agent claude
git add -A && git commit -m "eva init"
eva check
```

```
eva check — 0 feature(s)
  WARN  EVA053 .eva/constitution.md: constitution is not approved yet (status: draft): run /eva-constitution and approve it before writing specs
0 error(s), 1 warning(s)
```

> If `create-expo-app` already created the repo on `master`, run `git branch -m master main`. The gate diffs feature branches against `main` (or `master`, or `--base`).

## 1. Constitution

In Claude Code:

```
/eva-constitution
```

The agent reads the repository, asks you **up to 6 questions, one at a time**, and writes `.eva/constitution.md`: 6–10 short principles, each with a *Check:* line saying how a reviewer verifies it. It sets `status: in-review` and stops.

Review it, replace anything you disagree with, and approve it by editing the front-matter yourself:

```yaml
---
type: constitution
status: approved        # you set this — agents never do
version: 1
created: 2026-10-24
approved_by: Maribel
approved_at: 2026-10-24
closed_at:              # always empty for the constitution
---
```

An approved constitution cannot contain `TODO` (EVA054). Commit it.

## 2. Spec

```
/eva-specify Users sign in with email and password
```

The agent:

1. numbers the feature (`001`) and proposes a branch `feat/001-sign-in` — accept it;
2. **interviews you, one question at a time (max 6)**, before writing anything;
3. writes `specs/001-sign-in/spec.md` with user stories, **functional requirements in EARS form** and **completion criteria**, then sets `status: in-review`.

```md
## 4. Functional requirements

- **FR-1** (US-1) When the user taps "Sign in" with a registered email and the right password, the app shall open the home screen.
- **FR-2** (US-1) If the email or the password is empty, then the app shall show "Required" below each empty field and stay on the sign-in screen.
- **FR-3** (US-1) If the server rejects the credentials, then the app shall show "Wrong email or password" and clear the password field.

## 7. Completion criteria

- Every FR in §4 has at least one passing test named after it.
- Works on iOS 16+ and Android 10+.
```

Don't approve yet. If you try, `eva check` stops you:

```
  ERROR EVA070 specs/001-sign-in/spec.md: spec is approved but clarify.md is missing: run /eva-clarify before approving
```

## 3. Clarify

```
/eva-clarify 001
```

The agent reviews the spec — it **never edits it** — and writes `specs/001-sign-in/clarify.md` with one checkbox per finding (ambiguities, contradictions, missing edge cases, conflicts with the constitution):

```md
- [ ] **C-1** [edge-case] (§5) What happens when the user taps "Sign in" while offline: show a message, or queue the request?
  - Answer:
- [ ] **C-2** [ambiguity] (FR-2) Is an email without "@" also "Required", or a different message?
  - Answer:
```

While findings are open, the spec cannot be approved:

```
  ERROR EVA071 specs/001-sign-in/clarify.md: 2 open finding(s) but spec.md is approved: answer them and apply them with /eva-specify
```

**You** write each answer under `Answer:`. Then:

```
/eva-specify 001
```

The agent applies every answered finding to the spec, ticks it and records what changed. It never answers for you:

```md
- [x] **C-1** [edge-case] (§5) What happens when the user taps "Sign in" while offline: show a message, or queue the request?
  - Answer: Show "You are offline" and do not send the request.
  - Applied: FR-4 added.
```

```md
- **FR-4** (US-1) While the device is offline, when the user taps "Sign in", the app shall show "You are offline" and not send the request.
```

Run `/eva-clarify 001` again if the changes were big. When no finding is open, **approve the spec** (`status: approved`, `approved_by`, `approved_at`) and commit.

## 4. Plan

```
/eva-plan 001
```

The plan explains **how**. Its §2 *Requirements traceability* maps every FR to the parts of the plan that implement it. The agent sets `status: in-review` and runs `eva check`; in review, problems are warnings:

```
  WARN  EVA088 specs/001-sign-in/plan.md: FR-4 is not in the "Requirements traceability" section
```

After the fix:

```md
| FR | Plan sections | Components / files | Test level |
|----|---------------|--------------------|------------|
| FR-1 | §4, §7 | `SignInScreen`, `useSignIn` | unit + e2e |
| FR-2 | §4 | `SignInForm` | unit |
| FR-3 | §6, §7 | `useSignIn` | unit |
| FR-4 | §7 | `useSignIn`, `useNetworkStatus` | unit |
```

Review the file map (§9) and the performance section (§10) especially. Approve the plan.

## 5. Tasks

```
/eva-tasks 001
```

Each task is **typed** (`[feat]`, `[fix]`, `[improve]`), **cites the FRs** it implements, is one commit (≤ 3 code files, 5 at most), and lists **Files**, **Tests** and **Done when**:

```md
- [ ] T001 [feat] [US-1] Sign-in form with required fields — FR-2
  - Files: `src/features/auth/SignInForm.tsx`, `src/features/auth/SignInForm.test.tsx`
  - Tests: `src/features/auth/SignInForm.test.tsx` — FR-2
  - Done when: `npm test -- SignInForm` passes
```

If the agent forgets something, `eva check` warns while the tasks are in review (and fails once they are approved):

```
  WARN  EVA080 specs/001-sign-in/tasks.md: T003 has no type: tag it [feat], [fix] or [improve]
  WARN  EVA084 specs/001-sign-in/tasks.md: T002 has no "Done when:" condition
```

Approve the tasks and commit.

### Try the gate

Before approving the tasks, create any code file on the feature branch and run `eva check`:

```
eva check — 1 feature(s), branch feat/001-sign-in vs main (1 code change(s))
  ERROR EVA042 specs/001-sign-in: 1 code change(s) on "feat/001-sign-in" but not approved yet: tasks.md (in-review)
```

Code means anything outside `specs/`, `.eva/`, `docs/`, `.claude/`, `.opencode/` and root `*.md` files — uncommitted and untracked files included.

## 6. Implement

```
/eva-implement 001          # next task
/eva-implement 001 T002     # a specific task
/eva-implement 001 all      # every remaining task
```

For each `[feat]`/`[fix]` task the agent writes the tests first, runs them and **sees them fail for the expected reason**, records it under *Implementation notes* (`T001 red: …`), then writes the smallest code that makes them pass. Tests are named after the requirement (`it('FR-2: shows Required under empty fields', …)`). It ticks a task only when tests, typecheck and lint pass, and stops if it needs a file outside the plan's file map.

## 7. Verify and close

```
/eva-verify 001
```

The agent runs the checks, then goes **requirement by requirement** and writes `verification.md` with one row per FR and a verdict:

```md
| FR | Test(s) | Result |
|----|---------|--------|
| FR-1 | `useSignIn.test.ts` › "FR-1: opens home on valid credentials" | ✅ |
| FR-2 | `SignInForm.test.tsx` › "FR-2: shows Required under empty fields" | ✅ |
| FR-3 | `useSignIn.test.ts` › "FR-3: shows error and clears password" | ✅ |
| FR-4 | `useSignIn.test.ts` › "FR-4: does not send while offline" | ✅ |

**Verdict:** PASS — every FR and completion criterion is met.
```

If you agree, close the feature: in `spec.md` set `status: done` and `closed_at`, and set `closed_at` in `verification.md`. `eva check` only accepts it with a PASS verdict and every FR passing:

```
  ERROR EVA092 specs/001-sign-in/verification.md: spec is done but the verdict is not PASS (FAIL — FR-4 has no test.)
  ERROR EVA093 specs/001-sign-in/verification.md: FR-4 failed verification but the spec is done
```

```
eva check — 1 feature(s), branch feat/001-sign-in vs main (1 code change(s))
OK
```

Open the pull request with `verification.md`. With `eva check` in CI ([Getting started](getting-started.md#run-it-in-ci-github-actions)), the gate runs on every PR.

## Things to try

| Do this | Expect |
|---------|--------|
| Approve a document without `approved_at` | EVA063 |
| Write `created: 24/10/2026` | EVA061 |
| Write an FR without "shall" | EVA028 (once the spec is approved) |
| Tick a clarify finding with no answer | EVA074 |
| Cite `FR-9` in a task | EVA081 |
| List 6 code files in one task | EVA086 |

Every code is explained in the [`eva check` reference](check-reference.md).
