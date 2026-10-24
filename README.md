# eva — Spec-Driven Development for web & mobile projects

`eva` installs an agent-agnostic **Spec-Driven Development (SDD)** workflow into a project and enforces its approval gate:

```
/eva-constitution → /eva-specify ⇄ /eva-clarify → /eva-plan → /eva-tasks → /eva-implement → /eva-verify
constitution.md     spec.md        clarify.md     plan.md     tasks.md     code + tests     verification.md
▲ dev approves      ▲ dev approves once clarify.md has no open findings
                                                  ▲ dev approves plan and tasks
```

The constitution is approved once per project. No feature code is written until a developer has approved its spec, plan and tasks. Agents may only move documents to `in-review`; `eva check` fails when the gate is skipped.

Specs are written after a short interview (one question at a time, at most 6) and state their functional requirements as numbered `FR-n` items in [EARS](https://alistairmavin.com/ears/) form, plus the completion criteria that define when the feature is finished:

```md
## 4. Functional requirements

- **FR-1** (US-1) When the user taps Sign in with an empty password, the app shall show "Password is required".
- **FR-2** (US-1) If the server rejects the credentials, then the app shall keep the email and clear the password.

## 7. Completion criteria

- Every FR in §4 has at least one passing test named after it.
```

Before a spec can be approved, `/eva-clarify` reviews it and writes `clarify.md`: ambiguities, contradictions, missing edge cases and conflicts with the constitution, one checkbox per finding. It never edits the spec. The developer answers each finding, `/eva-specify` applies the answer and ticks it, and `eva check` refuses an approved spec while a finding is open:

```md
- [x] **C-1** [ambiguity] (FR-2) "Recent orders": last 30 days or the last 10 orders?
  - Answer: Last 30 days.
  - Applied: FR-2 reworded.
- [ ] **C-2** [edge-case] (§5) What happens when the user submits while offline?
  - Answer:
```

The plan traces every requirement (*Requirements traceability*). Tasks are typed `[feat]`, `[fix]` or `[improve]`, sized to one commit (at most 3 code files, 5 as a hard limit), cite the `FR-n` they implement, list their tests and a *Done when*, and are implemented test-first. Verification goes requirement by requirement and ends with `**Verdict:** PASS` or `FAIL`:

```md
- [ ] T002 [feat] [US-1] Validate the sign-in form — FR-1, FR-2
  - Files: `src/features/auth/SignInForm.tsx`, `src/features/auth/SignInForm.test.tsx`
  - Tests: `src/features/auth/SignInForm.test.tsx` — FR-1, FR-2
  - Done when: `npm test -- SignInForm` passes
```

Every document carries `type`, `created`, `approved_at` and `closed_at` in its front-matter (`YYYY-MM-DD`), so the history of a feature can be read from its files.

- **Agent-agnostic:** `AGENTS.md` is the single source of truth. Claude Code and OpenCode get `/eva-*` shortcuts; any other agent can follow `.eva/commands/*.md`.
- **Stack-specific:** React Native (Expo) and Next.js (App Router) overlays with folder structure, commands, testing conventions and a performance checklist based on Vercel's best practices.
- **Collaborative:** numbered features (`specs/NNN-slug/`), `feat/NNN-slug` branches, front-matter status, ADRs and a CI check.

## Documentation

Full documentation lives in [`docs/`](docs/README.md) — also [in Spanish](docs/es/README.md):

- [Getting started](docs/getting-started.md) · [Tutorial](docs/tutorial.md) · [Workflow and documents](docs/workflow.md) · [`eva check` reference](docs/check-reference.md)
- [Contributing and architecture](docs/contributing.md) · [Roadmap and decisions](docs/roadmap.md)

## Install

Requires Go 1.24+.

```sh
go install github.com/codebymaribel/eva-sdd/cmd/eva@latest
```

Templates are embedded in the binary, so installing a new version also updates the templates. Make sure `$(go env GOPATH)/bin` is on your `PATH`.

## Usage

### `eva init`

Run it in the project root. In a terminal, a wizard asks for stacks, app roots and agents:

```sh
eva init
```

Or non-interactively (CI, scripts):

```sh
eva init --stack rn --agent claude                                   # single app
eva init --stack rn,next --agent claude,opencode \
         --path rn=apps/mobile,next=apps/web                         # monorepo
eva init --dry-run                                                   # preview only
```

| Flag | Description |
|------|-------------|
| `--stack` | `react-native` (`rn`), `nextjs` (`next`), comma-separated |
| `--agent` | `claude`, `opencode`, comma-separated |
| `--path` | app root per stack; default `.` for one stack, `apps/mobile` / `apps/web` for several |
| `--name` | project name (default: directory name) |
| `--dir` | project directory (default `.`) |
| `--dry-run` | print the plan without writing |
| `--force` | overwrite files that differ from the templates, even if you edited them |
| `--prune` | delete files eva installed earlier, no longer ships and nobody edited |
| `--diff` | show how each edited or updated file differs from the new template |
| `--yes` | skip the confirmation prompt |

Re-running `eva init` is safe, and it is how you take new templates after upgrading eva:

- Installed stacks and agents are kept; new ones are added.
- `AGENTS.md`, `CLAUDE.md` and `.gitattributes` only get an eva-managed block; your own content outside it is never touched.
- eva records a hash of every file it writes in `.eva/eva-lock.json`. Files you never edited are updated when their template changes; files you edited (e.g. `.eva/constitution.md`) are kept — `eva init --dry-run --diff` shows what changed, `--force` overwrites them.
- Files eva no longer ships are reported; `--prune` deletes them if you never edited them.

### `eva check`

Validates the SDD documents and the approval gate. Exits `1` on errors.

```sh
eva check                 # documents + branch gate
eva check --no-git        # documents only
eva check --base develop  # custom base branch
eva check --strict        # also fail on code changes in branches without a feature id
```

Document rules: every document has the `type` of its file and valid dates in order (`created` ≤ `approved_at` ≤ `closed_at`); `approved_at` goes with `approved`/`done` and `closed_at` with `done`. An approved spec must define unique `FR-n` requirements in EARS form and non-empty completion criteria. Plan and task rules (EVA080–088) are warnings while the document is `in-review` and errors once it is approved; a `done` spec needs a passing verification (EVA090–093).

On a branch named `feat/NNN-slug` with code changes (anything outside `specs/`, `.eva/`, `docs/`, `.claude/`, `.opencode/` and root `*.md` files), the constitution and feature `NNN`'s `spec.md`, `plan.md` and `tasks.md` must be approved.

| Code | Severity | Meaning |
|------|----------|---------|
| EVA001 | error | `.eva/config.json` missing — run `eva init` |
| EVA010 | warning | folder in `specs/` is not named `NNN-slug` |
| EVA011 | error | two features share the same number |
| EVA020 | error | feature folder without `spec.md` |
| EVA021 | error | invalid or missing front-matter |
| EVA022 | error | front-matter `id` differs from the folder name |
| EVA023 | error | invalid `status` |
| EVA024 | error | `approved`/`done` without `approved_by` |
| EVA025 | error | approved spec still has `[NEEDS CLARIFICATION]` |
| EVA026 | error | approved spec defines no `FR-n` under *Functional requirements* |
| EVA027 | error | an `FR-n` id is defined more than once |
| EVA028 | error | a requirement is not in EARS form |
| EVA029 | error | approved spec has no (or an empty) *Completion criteria* section |
| EVA030 | error | `plan.md` exists before the spec is approved |
| EVA031 | error | `tasks.md` exists before the plan is approved |
| EVA032 | error | tasks ticked before `tasks.md` is approved |
| EVA033 | error | spec `done` with unticked tasks |
| EVA040 | warning | not a git repository (gate skipped) |
| EVA041 | error/warning | branch refers to a feature that does not exist |
| EVA042 | error | code changes but spec/plan/tasks not approved |
| EVA043 | warning (error with `--strict`) | code changes on a branch without a feature id |
| EVA044 | warning | base branch not found (gate skipped) |
| EVA045 | warning | detached HEAD without `--branch` (gate skipped) |
| EVA046 | warning | git diff failed (gate skipped) |
| EVA050 | error | `.eva/constitution.md` missing |
| EVA051 | error | constitution has invalid front-matter or status |
| EVA052 | error | constitution approved without `approved_by` |
| EVA053 | warning (no specs) / error (specs exist) | constitution not approved yet |
| EVA054 | error | approved constitution still has `TODO` markers |
| EVA070 | error | spec approved without `clarify.md` |
| EVA071 | error | spec approved while `clarify.md` has open findings |
| EVA072 | error | a finding id (`C-n`) is used more than once |
| EVA073 | error | a finding has no category or an unknown one |
| EVA074 | error | a finding is ticked without an answer |
| EVA080 | warning / error | task without a type (`[feat]`, `[fix]`, `[improve]`) |
| EVA081 | warning / error | plan or task cites an `FR-n` the spec does not define |
| EVA082 | warning / error | an `FR-n` has no task |
| EVA083 | warning / error | `[feat]`/`[fix]` task cites no `FR-n` |
| EVA084 | warning / error | task without `Files:` or `Done when:` |
| EVA085 | warning / error | `[feat]`/`[fix]` task without test files or without the `FR-n` they verify |
| EVA086 | warning (>3) / error (>5) | too many code files in one task (tests and docs do not count) |
| EVA087 | warning / error | task id used more than once |
| EVA088 | warning / error | plan has no *Requirements traceability* or misses an `FR-n` |
| EVA090 | error | spec `done` without `verification.md` |
| EVA091 | error | an `FR-n` is missing from the verification table |
| EVA092 | error | spec `done` but the verdict is not `PASS` |
| EVA093 | error | an `FR-n` is marked failed (❌/FAIL) in the verification |
| EVA060 | error | `type` missing or not matching the file (`constitution`, `spec`, `clarify`, `plan`, `tasks`, `verification`) |
| EVA061 | error | a date is not a real `YYYY-MM-DD` date |
| EVA062 | error | `created` empty (allowed only while `draft`) |
| EVA063 | error | `approved`/`done` without `approved_at`, or `approved_at` set before approval |
| EVA064 | error | `done` without `closed_at`, or `closed_at` set before `done` (never on the constitution) |
| EVA065 | error | dates out of order (`created` ≤ `approved_at` ≤ `closed_at`) |

#### In CI (GitHub Actions)

```yaml
name: eva
on: pull_request
jobs:
  eva-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0 # the gate diffs against the base branch
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - run: go install github.com/codebymaribel/eva-sdd/cmd/eva@latest
      - run: eva check
```

`GITHUB_HEAD_REF` and `GITHUB_BASE_REF` are picked up automatically. To make sure only humans approve, add a `CODEOWNERS` rule for `specs/**` and require code-owner review.

## What gets installed

```
AGENTS.md                         # eva block: workflow, approval gate, implementation rules
CLAUDE.md                         # (claude) eva block importing @AGENTS.md
.claude/commands/eva-*.md         # (claude) /eva-constitution … /eva-verify
.opencode/commands/eva-*.md       # (opencode) /eva-constitution … /eva-verify
.eva/
  config.json                     # what eva installed (stacks, paths, agents)
  eva-lock.json                   # hash of every installed file (generated — commit it, never edit it)
  constitution.md                 # project principles — /eva-constitution proposes, a dev approves
  stacks/react-native.md          # (rn) baseline, structure, commands, perf checklist
  stacks/nextjs.md                # (next) baseline, structure, commands, perf checklist
  commands/*.md                   # agent-agnostic phase instructions
  templates/{spec,clarify,plan,tasks,verification,adr}.md
specs/README.md                   # feature folder conventions and status lifecycle
docs/adr/README.md                # ADR conventions
.gitattributes                    # eva block: marks eva-lock.json as generated
```

## Development

```sh
make test     # go test ./...
make golden   # regenerate installer golden files after editing templates/
make lint     # go vet + gofmt
make build    # bin/eva
```

Template conventions (see `templates/embed.go`):

- `templates/core/` is always installed; `templates/stacks/<id>/` and `templates/agents/<id>/` per selection.
- `*.tmpl` files are rendered with Go `text/template` using `[[ ]]` delimiters (`[[ .Project ]]`, `[[ range .Stacks ]]`, `[[ stackPath "nextjs" ]]`).
- `*.block` files are merged as a managed block instead of overwriting.

Adding a stack or agent: add it to `internal/catalog/catalog.go`, create its folder under `templates/`, add a golden case in `internal/installer/installer_test.go` and run `make golden`.

The CLI only uses the Go standard library, so `go install` needs no third-party modules.
