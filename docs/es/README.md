# Documentación de eva

[English](../README.md) · **Español**

`eva` instala un flujo de Spec-Driven Development (SDD) en proyectos web y mobile y hace cumplir su gate de aprobación: no se escribe código de una feature hasta que un desarrollador aprueba su spec, su plan y sus tasks.

## Para usuarios

| Página | Léela cuando quieras… |
|--------|------------------------|
| [Primeros pasos](getting-started.md) | instalar eva, agregarlo a un proyecto, actualizarlo y correrlo en CI |
| [Tutorial](tutorial.md) | llevar una feature de la constitución a verificada, paso a paso |
| [Flujo y documentos](workflow.md) | saber quién hace qué en cada fase y el formato exacto de cada documento |
| [Referencia de `eva check`](check-reference.md) | entender un hallazgo `EVA0xx` y cómo arreglarlo |

## Para contribuidores

| Página | Léela cuando quieras… |
|--------|------------------------|
| [Contribuir y arquitectura](contributing.md) | cambiar templates, agregar una regla de chequeo, un stack o un agente |
| [Roadmap y decisiones](roadmap.md) | saber qué está hecho, qué sigue y por qué las cosas son como son |

## El flujo de un vistazo

```
/eva-constitution → /eva-specify ⇄ /eva-clarify → /eva-plan → /eva-tasks → /eva-implement → /eva-verify
constitution.md     spec.md        clarify.md     plan.md     tasks.md     code + tests     verification.md
▲ dev aprueba       ▲ dev aprueba cuando clarify.md no tiene hallazgos abiertos
                                                  ▲ dev aprueba plan y tasks
```

Los agentes escriben los documentos y los pasan a `in-review`; solo una persona pone `approved` o `done`. `eva check` valida los documentos y bloquea el código en ramas `feat/NNN-slug` hasta que se pasa el gate.
