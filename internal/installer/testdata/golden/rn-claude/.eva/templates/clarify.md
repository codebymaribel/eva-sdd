---
id: NNN-feature-slug
type: clarify
title: Feature title
spec: ./spec.md
created: YYYY-MM-DD
closed_at: # YYYY-MM-DD, set by a human when the spec is approved
---

# Clarify: Feature title

> Findings about `spec.md`, written by `/eva-clarify`. This file only **detects**: it never changes the spec. The developer answers each finding under *Answer*; `/eva-specify` then applies the answer to the spec and ticks the finding. A spec cannot be approved while a finding is open (`- [ ]`).

## Findings

<!--
One list item per finding: checkbox, id, category, the part of the spec it is about, and one specific question.
Categories: [ambiguity] [contradiction] [edge-case] [constitution]

- [ ] **C-1** [ambiguity] (FR-2) "Recent orders" is not defined: last 30 days, last 10 orders, or something else?
  - Answer:
- [x] **C-2** [edge-case] (§5) What happens when the user submits while offline?
  - Answer: Queue the request and retry when back online.
  - Applied: FR-6 added.

Write "No findings." below if the review found nothing.
-->
