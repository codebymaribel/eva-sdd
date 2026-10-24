# Getting started

**English** · [Español](es/getting-started.md) · [Docs index](README.md)

## Requirements

- Homebrew, or Go 1.24 or newer, to install eva (your project does not need Go).
- Git, for the branch gate of `eva check`.
- A coding agent. Claude Code and OpenCode get `/eva-*` shortcuts; any other agent can follow `AGENTS.md` and `.eva/commands/*.md`.

## Install

With Homebrew (macOS and Linux):

```sh
brew install codebymaribel/tap/eva
brew upgrade eva      # later, to update
```

Homebrew builds eva from the source of the release and checks the download against the checksum in the formula. It installs Go by itself as a build dependency.

With Go:

```sh
go install github.com/codebymaribel/eva-sdd/cmd/eva@latest
```

From a local clone (during development):

```sh
cd eva-sdd
make install          # go install ./cmd/eva
```

`go install` and `make install` put the binary in `$(go env GOPATH)/bin`. If your shell says `command not found: eva`, add that folder to your `PATH`:

```sh
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc   # or ~/.bashrc
source ~/.zshrc
eva -h
```

`eva version` prints the installed version (`dev` for local builds). Templates are embedded in the binary: installing a new eva also brings new templates.

## Add eva to a project

Run it in the project root. Commit your work first so the changes eva makes are easy to review.

```sh
eva init
```

In a terminal, a wizard asks for the stacks, the app root of each stack (only when there are several) and the agents, prints the plan of files and asks for confirmation. Without a terminal (CI, scripts) pass flags:

```sh
eva init --stack rn --agent claude                                   # one app
eva init --stack rn,next --agent claude,opencode \
         --path rn=apps/mobile,next=apps/web                         # monorepo
eva init --dry-run                                                   # preview only
```

| Flag | Description |
|------|-------------|
| `--stack` | `react-native` (aliases `rn`, `expo`), `nextjs` (aliases `next`, `next.js`), comma-separated |
| `--agent` | `claude` (alias `claude-code`), `opencode` (alias `open-code`), comma-separated |
| `--path` | app root per stack, e.g. `rn=apps/mobile`; default `.` for one stack, `apps/mobile` / `apps/web` for several |
| `--name` | project name (default: directory name) |
| `--dir` | project directory (default `.`) |
| `--dry-run` | print the plan without writing |
| `--force` | overwrite files that differ from the templates, even if you edited them |
| `--prune` | delete files eva installed earlier, no longer ships and nobody edited |
| `--diff` | show how each edited or updated file differs from the new template |
| `--yes` | do not ask for confirmation |

Exit codes: `0` success, `1` failure or cancelled, `2` invalid flags or arguments.

## What gets installed

```
AGENTS.md                         # eva block: phases, approval gate, implementation rules
CLAUDE.md                         # (claude) eva block importing @AGENTS.md
.claude/commands/eva-*.md         # (claude) /eva-constitution … /eva-verify
.opencode/commands/eva-*.md       # (opencode) /eva-constitution … /eva-verify
.eva/
  config.json                     # what eva installed: stacks, paths, agents, eva version
  eva-lock.json                   # hash of every installed file (generated — commit it, never edit it)
  constitution.md                 # project principles, approved once by a developer
  stacks/react-native.md          # (rn) baseline, structure, commands, performance checklist
  stacks/nextjs.md                # (next) baseline, structure, commands, performance checklist
  commands/*.md                   # agent-agnostic instructions for each phase
  templates/{spec,clarify,plan,tasks,verification,adr}.md
specs/README.md                   # feature folder conventions, status lifecycle, rules
docs/adr/README.md                # architecture decision records
.gitattributes                    # eva block: marks eva-lock.json as generated so PRs fold it
```

After installing:

1. In your agent, run `/eva-constitution`, then review and approve `.eva/constitution.md` (see [the tutorial](tutorial.md#1-constitution)).
2. Review `.eva/stacks/*.md` (commands, structure, budgets) and adjust them to the project.
3. Commit the generated files.
4. Start a feature with `/eva-specify <what you want to build>`.

## Re-running and updating

`eva init` is safe to re-run, and it is how you take the new templates after upgrading eva:

- Installed stacks and agents are kept; new ones are added.
- `AGENTS.md`, `CLAUDE.md` and `.gitattributes` only get an eva-managed block (`<!-- eva:start -->…<!-- eva:end -->`, or `# eva:start`…`# eva:end` in `.gitattributes`). Your content outside the block is never touched; the block is always updated.
- `.eva/config.json` and `.eva/eva-lock.json` are owned by eva and always rewritten.
- For every other file, eva compares it with the hash it recorded in `.eva/eva-lock.json` when it wrote it:

| The file… | `eva init` |
|-----------|------------|
| does not exist | creates it |
| is identical to the new template | leaves it alone |
| matches the recorded hash (you never edited it) | **updates it** to the new template |
| was edited by you | keeps it and says so; if the template changed too, `eva init --dry-run --diff` shows the difference |
| is no longer shipped by eva | reports it; `--prune` deletes it if you never edited it |

`--force` overwrites every file that differs from its template, edited or not — commit first and recover what you want with `git checkout -p`.

**`.eva/eva-lock.json`** is generated: commit it so the whole team shares the same record, and never edit it. `eva init` marks it as generated in `.gitattributes`, so GitHub folds it in pull requests. Hashes ignore line endings, so a checkout that converts `\n` to `\r\n` does not count as an edit.

**Projects installed before the lock existed:** on the first `eva init` with a newer eva, files identical to the templates are recorded, and files that differ are treated as edited (eva cannot know their origin): they are kept, with `--diff` to compare. Files eva installed before the lock and no longer ships (for example the old `sdd-*.md` shortcuts) are not detected; delete them yourself.

## Check the project

```sh
eva check                 # documents + branch gate
eva check --no-git        # documents only
eva check --base develop  # custom base branch
eva check --strict        # also fail on code changes in branches without a feature id
```

`eva check` exits `1` when there are errors; warnings never fail. Every finding has a code — see the [`eva check` reference](check-reference.md).

## Run it in CI (GitHub Actions)

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

`GITHUB_HEAD_REF` and `GITHUB_BASE_REF` are read automatically. To make sure only humans approve, add a `CODEOWNERS` rule for `specs/**` and `.eva/constitution.md` and require code-owner review.

Next: [Tutorial](tutorial.md).
