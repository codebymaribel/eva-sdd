---
id: NNN-feature-slug
type: verification
title: Feature title
created: YYYY-MM-DD
closed_at: # YYYY-MM-DD, set by a human when the spec is marked done
---

# Verification: Feature title

## 1. Automated checks

| Check | Command | Result |
|-------|---------|--------|
| Typecheck | | ✅ / ❌ |
| Lint | | ✅ / ❌ |
| Unit / integration tests | | ✅ / ❌ (n passed) |
| E2E tests | | ✅ / ❌ / N/A |
| `eva check` | `eva check` | ✅ / ❌ |

## 2. Requirements traceability

One row per `FR-n` of the spec, in order. ✅ only when a test named after the FR exists and passes.

| FR | Test(s) | Result |
|----|---------|--------|
| FR-1 | `path/to/test.ts` › "FR-1: …" | ✅ |

## 3. Completion criteria

One row per item of the spec's *Completion criteria*.

| Criterion | Result | Evidence |
|-----------|--------|----------|
| Every FR has at least one passing test named after it | ✅ / ❌ | §2 |

## 4. Performance checklist

Each item from plan §10 plus any other applicable item from the stack file.

| Rule id | Status | Evidence (file:line or measurement) |
|---------|--------|-------------------------------------|
| | ✅ / ⚠️ / ❌ | |

## 5. Constitution and scope

- Files changed outside the plan's file map: none / list
- Constitution deviations: none / list
- Implementation notes worth reviewing:

## 6. Verdict

PASS only when every row of §1–§3 is ✅ and nothing in §4–§5 blocks the merge.

**Verdict:** <PASS or FAIL> — one sentence: is the spec fulfilled? If FAIL, list what is missing.
