# eva documentation

**English** · [Español](es/README.md)

`eva` installs a Spec-Driven Development (SDD) workflow into web and mobile projects and enforces its approval gate: no feature code is written until a developer has approved its spec, plan and tasks.

## For users

| Page | Read it when you want to… |
|------|---------------------------|
| [Getting started](getting-started.md) | install eva, add it to a project, update it and run it in CI |
| [Tutorial](tutorial.md) | take one feature from constitution to verified, step by step |
| [Workflow and documents](workflow.md) | know who does what in each phase and the exact format of every document |
| [`eva check` reference](check-reference.md) | understand an `EVA0xx` finding and how to fix it |

## For contributors

| Page | Read it when you want to… |
|------|---------------------------|
| [Contributing and architecture](contributing.md) | change templates, add a check rule, a stack or an agent |
| [Roadmap and decisions](roadmap.md) | know what is done, what is next and why things are the way they are |

## The flow at a glance

```
/eva-constitution → /eva-specify ⇄ /eva-clarify → /eva-plan → /eva-tasks → /eva-implement → /eva-verify
constitution.md     spec.md        clarify.md     plan.md     tasks.md     code + tests     verification.md
▲ dev approves      ▲ dev approves once clarify.md has no open findings
                                                  ▲ dev approves plan and tasks
```

Agents write documents and move them to `in-review`; only a human sets `approved` or `done`. `eva check` validates the documents and blocks code on `feat/NNN-slug` branches until the gate is passed.
