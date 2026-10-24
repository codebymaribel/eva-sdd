# eva-clarify — find what the spec leaves unclear

**Goal:** review `specs/NNN-slug/spec.md` and record every ambiguity, contradiction, missing edge case and conflict with the constitution in `specs/NNN-slug/clarify.md`. This phase only **detects**: you never edit `spec.md`. Then stop for the developer's answers.

**Input:** a feature id (`NNN` or `NNN-slug`). If none is given, infer it from the current branch name (`feat/NNN-slug`); if that fails, ask.

## Gate check — do this first

1. `.eva/constitution.md` must be `approved` and `specs/NNN-slug/spec.md` must exist. Otherwise stop and say which is missing.
2. If the spec is already `approved` or `done`, stop: clarifying it would reopen it. Ask the developer whether to set it back to `in-review` first.

## Review

1. Read the whole spec, `.eva/constitution.md` and the stack file(s) for the spec's `stack`. Read the code only to understand existing behaviour the spec relies on.
2. Look for findings in four categories:
   - **[ambiguity]** a word or requirement two people could read differently ("fast", "recent", "large", "the user" when there are several roles), or a requirement a test could not verify.
   - **[contradiction]** two requirements, or a requirement and a story, non-functional requirement or out-of-scope item, that cannot both be true.
   - **[edge-case]** a situation from §5 (empty, loading, offline, errors, permissions, long text, large lists, locales) or an obvious failure path that no `FR-n` covers.
   - **[constitution]** something that breaks or is not covered by a principle of the constitution (e.g. personal data without a privacy requirement).
3. Also collect every `[NEEDS CLARIFICATION: …]` marker in the spec as a finding (usually `[ambiguity]`).
4. Keep findings that change behaviour, scope or data. Skip wording and style. If there are more than 10, keep the 10 with the biggest impact and say how many you dropped.

## Write `clarify.md`

1. If `specs/NNN-slug/clarify.md` does not exist, copy `.eva/templates/clarify.md`. Front-matter: same `id`/`title` as the spec, `type: clarify`, `created` (today, `YYYY-MM-DD`). Leave `closed_at` empty.
2. Add one list item per finding under *Findings*:
   `- [ ] **C-n** [category] (FR-n or §section) <one specific question>` with an empty `  - Answer:` line below it. Offer concrete options in the question when you can ("30 days or the last 10 orders?").
3. On a re-run, keep existing findings and their ids untouched, never reopen or edit a resolved one, and give new findings the next free number.
4. If you found nothing, write `No findings.` under *Findings*.
5. Never tick a finding, never write an answer and never change `spec.md` in this phase.

## Finish — then STOP

1. Run `eva check --no-git` if it is installed and fix any finding about `clarify.md`.
2. Reply with: the path and the open findings (id, category, question), one line each.
3. End with: *"Please answer each finding under `Answer:` in `specs/NNN-slug/clarify.md`, then run `/eva-specify NNN` to apply the answers. The spec can be approved once no finding is open."* If there are no findings: *"No findings. When you are happy with the spec, set `status: approved`, `approved_by` and `approved_at`, then run `/eva-plan NNN`."*
