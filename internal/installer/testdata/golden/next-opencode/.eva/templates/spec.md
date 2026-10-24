---
id: NNN-feature-slug
type: spec
title: Feature title
status: draft # draft | in-review | approved | done — only a human sets approved/done
stack: react-native # react-native | nextjs | multiple (comma-separated)
author:
created: YYYY-MM-DD
approved_by:
approved_at: # YYYY-MM-DD, set by the approver
closed_at: # YYYY-MM-DD, set by a human together with status: done
issue: # optional link to the issue/ticket
---

# Spec: Feature title

> **What** the feature does and **why**. No implementation details: no libraries, file names, endpoints or component names — those belong in `plan.md`.

## 1. Summary

One or two sentences a non-developer can understand.

## 2. Problem and motivation

What problem do users have today? How do we know (data, feedback, support tickets)? What happens if we do nothing?

## 3. User stories

| ID   | Priority | Story |
|------|----------|-------|
| US-1 | P1 | As a <role>, I want <capability> so that <benefit>. |
| US-2 | P2 | As a <role>, I want <capability> so that <benefit>. |

P1 = must ship, P2 = should ship, P3 = nice to have.

## 4. Functional requirements

Every user story has at least one requirement. Each requirement is one list item that starts with its id, the user story in brackets and a single [EARS](https://alistairmavin.com/ears/) statement. Ids are never reused or renumbered; tests and tasks cite them (`FR-1: …`).

| Pattern | Template |
|---------|----------|
| Ubiquitous | The <system> shall <response>. |
| Event-driven | When <trigger>, the <system> shall <response>. |
| State-driven | While <state>, the <system> shall <response>. |
| Unwanted behaviour | If <condition>, then the <system> shall <response>. |
| Optional feature | Where <feature is included>, the <system> shall <response>. |
| Complex | While <state>, when <trigger>, the <system> shall <response>. |

- **FR-1** (US-1) When <trigger>, the <system> shall <observable response>.
- **FR-2** (US-1) If <unwanted condition>, then the <system> shall <observable response>.
- **FR-3** (US-2) While <state>, the <system> shall <observable response>.

## 5. Edge cases and error states

Every case that changes behaviour becomes a requirement in §4 (usually `If …, then …`). Write "N/A" when a case does not apply.

- Empty state:
- Loading state:
- Offline / slow network:
- Server or validation error:
- Permissions denied / unauthenticated:
- Very long text, large lists, unusual locales:

## 6. Non-functional requirements

| Area | Requirement |
|------|-------------|
| Performance | e.g. list scrolls at 60 fps with 1,000 items; page LCP ≤ 2.5 s |
| Accessibility | e.g. screen reader labels, dynamic type |
| Security & privacy | personal data collected, who can access it |
| Offline | expected behaviour without network |
| i18n | languages, RTL |
| Analytics | events to track |

## 7. Completion criteria

The feature is finished when every item below is true. Each item is verifiable by someone who did not build it; `/eva-verify` checks them one by one.

- Every FR in §4 has at least one passing test named after it.
- <observable outcome a reviewer can check, e.g. "works on iOS 16+ and Android 10+">

## 8. Out of scope

What this feature explicitly does **not** do.

## 9. Dependencies and assumptions

Other features, services, designs or decisions this relies on.

## 10. Open questions

Use `[NEEDS CLARIFICATION: question]` inline wherever something is still ambiguous after the interview, and list them here. `/eva-clarify` collects them, with everything else it finds, in `clarify.md`. A spec with open questions or open findings cannot be approved.

## 11. Review checklist

For the reviewer before approving:

- No implementation details (libraries, files, APIs).
- Every user story has at least one `FR-n`, and every `FR-n` is a single EARS statement a test can verify.
- Edge cases and error states are covered by requirements.
- Completion criteria are concrete and verifiable.
- Out of scope is explicit.
- No `[NEEDS CLARIFICATION]` markers remain and `clarify.md` has no open findings.
