# Flujo y documentos

[English](../workflow.md) · **Español** · [Índice](README.md)

Esta página es la referencia del flujo SDD que instala eva: las fases, quién hace qué, el ciclo de estados y el formato exacto de cada documento. En un proyecto, la fuente de verdad instalada es `AGENTS.md`, `.eva/commands/*.md` y `.eva/templates/*.md`.

## Fases

| # | Fase | Comando | Resultado | Para seguir |
|---|------|---------|-----------|-------------|
| 0 | Constitución | `/eva-constitution` | `.eva/constitution.md` (una vez por proyecto) | El dev aprueba |
| 1 | Spec | `/eva-specify` | `specs/NNN-slug/spec.md`: qué y por qué | — |
| 1b | Clarificación | `/eva-clarify` | `specs/NNN-slug/clarify.md`: solo hallazgos | El dev responde; `/eva-specify` aplica; el dev aprueba la spec |
| 2 | Plan | `/eva-plan` | `plan.md`: cómo | El dev aprueba |
| 3 | Tasks | `/eva-tasks` | `tasks.md`: pasos ordenados y tipados | El dev aprueba |
| 4 | Implementación | `/eva-implement` | código + tests, de a una task | Los tests de cada task pasan |
| 5 | Verificación | `/eva-verify` | `verification.md`: FR por FR, veredicto | El dev marca la spec como `done` |

Los agentes sin slash commands siguen los mismos pasos desde `.eva/commands/<fase>.md`.

## Roles

| | Agente | Desarrollador |
|---|--------|---------------|
| Escribe documentos | sí, y pone `status: in-review` | revisa y edita |
| `type`, `created` | los completa al crear un documento | — |
| `status: approved` / `done` | **nunca** | sí |
| `approved_by`, `approved_at`, `closed_at` | **nunca** los completa; vacía `approved_by`/`approved_at` cuando edita un documento aprobado | sí |
| Respuestas de clarify | **nunca** las escribe ni las inventa | las escribe |
| Empieza la fase siguiente | solo cuando el documento anterior dice `approved` en el archivo | — |

La aprobación se lee del archivo, no del chat: "se ve bien" no es una aprobación.

## Ciclo de estados

`spec.md`, `plan.md` y `tasks.md` tienen `status`; la constitución también (sin `done`):

```
draft ──► in-review ──► approved ──► done (solo la spec, después de verificar)
              ▲             │
              └─────────────┘  cualquier edición de un documento aprobado
```

Orden de fases que hace cumplir `eva check`:

- las specs existen solo con la constitución aprobada;
- `plan.md` existe solo con `spec.md` aprobada (y `clarify.md` sin hallazgos abiertos);
- `tasks.md` existe solo con `plan.md` aprobado;
- las tasks se tildan solo con `tasks.md` aprobado;
- una spec es `done` solo con todas las tasks tildadas y `verification.md` aprobada;
- el código en `feat/NNN-slug` necesita la constitución y la spec, el plan y las tasks de esa feature aprobados.

## Front-matter

Cada documento empieza con un bloque plano `clave: valor` entre líneas `---`. Lo que sigue a ` #` es un comentario.

| Campo | constitution | spec | clarify | plan | tasks | verification |
|-------|:-----------:|:----:|:-------:|:----:|:-----:|:------------:|
| `id` (= nombre de la carpeta) | — | ✔ | ✔ | ✔ | ✔ | ✔ |
| `type` | `constitution` | `spec` | `clarify` | `plan` | `tasks` | `verification` |
| `status` | ✔ | ✔ | — | ✔ | ✔ | — |
| `created` | ✔ | ✔ | ✔ | ✔ | ✔ | ✔ |
| `approved_by`, `approved_at` | ✔ | ✔ | — | ✔ | ✔ | — |
| `closed_at` | siempre vacío | ✔ | ✔ | ✔ | ✔ | ✔ |
| otros | `version` | `title`, `stack`, `author`, `issue` | `title`, `spec` | `title`, `spec`, `author` | `title`, `plan`, `author` | `title` |

### Fechas

- Formato `YYYY-MM-DD`, y tiene que ser una fecha real (`2026-02-30` falla).
- `created` es obligatorio cuando el documento sale de `draft` (en clarify y verification, siempre).
- `approved_at` es obligatorio con `approved`/`done` y tiene que estar vacío antes.
- `closed_at` es obligatorio con `done` y tiene que estar vacío antes; la constitución nunca lo tiene. En clarify y verification es opcional (lo pone el dev al cerrar la feature).
- Dentro de un documento: `created ≤ approved_at ≤ closed_at`. No se comparan fechas entre documentos, así reaprobar una spec reabierta no rompe planes anteriores.

## Constitución: `.eva/constitution.md`

Contexto del producto, entre 6 y 10 **principios** (una oración cada uno más una línea *Check:* que dice cómo se verifica), estándares de ingeniería, política de testing, reglas de performance y seguridad, una Definition of Done y una tabla de **enmiendas** (*Amendments*). Enmendar una constitución aprobada incrementa `version`, agrega una fila a la tabla y la vuelve a `in-review`. Una constitución aprobada no tiene `TODO`.

## Spec: `spec.md`

El **qué** y el **por qué**, sin detalles de implementación. Se escribe después de una entrevista de hasta 6 preguntas, de a una.

| Sección | Contenido |
|---------|-----------|
| 1–2 | Resumen, problema y motivación |
| 3 | Historias de usuario `US-n` con prioridad P1/P2/P3 |
| 4 | **Requisitos funcionales** `FR-n` en EARS |
| 5 | Casos límite y errores (cada uno que cambia el comportamiento se vuelve un FR) |
| 6 | Requisitos no funcionales |
| 7 | **Criterios de finalización** (*Completion criteria*) |
| 8–11 | Fuera de alcance, dependencias, preguntas abiertas, checklist de revisión |

### Requisitos funcionales (EARS)

Un ítem de lista por requisito bajo el título *Functional requirements*, empezando por su id. Un `FR-n` mencionado en otra parte (criterios, preguntas) es una referencia, no una definición. Los ids nunca se reutilizan ni se renumeran.

```md
- **FR-1** (US-1) When <trigger>, the <system> shall <observable response>.
```

| Patrón | Plantilla |
|--------|-----------|
| Ubicuo | The <system> shall <response>. |
| Por evento | When <trigger>, the <system> shall <response>. |
| Por estado | While <state>, the <system> shall <response>. |
| Comportamiento no deseado | If <condition>, then the <system> shall <response>. |
| Funcionalidad opcional | Where <feature is included>, the <system> shall <response>. |
| Complejo | While <state>, when <trigger>, the <system> shall <response>. |

`eva check` acepta un enunciado que empieza con The/When/While/If/Where, contiene `shall` seguido de una respuesta y, si empieza con `If`, contiene `, then`.

### Criterios de finalización

Una lista de afirmaciones verificables de cuándo la feature está terminada. El primer ítem siempre es *"Every FR in §4 has at least one passing test named after it."* `/eva-verify` revisa cada uno.

## Clarificación: `clarify.md`

La escribe `/eva-clarify` con la spec en `draft` o `in-review`. Solo **detecta**: nunca edita la spec. Máximo 10 hallazgos por corrida; al volver a correrla conserva los ids y nunca reabre hallazgos resueltos. Los marcadores `[NEEDS CLARIFICATION: …]` que quedan en la spec se convierten en hallazgos.

```md
- [ ] **C-1** [edge-case] (§5) What happens when the user submits while offline?
  - Answer:
```

- Categorías: `[ambiguity]`, `[contradiction]`, `[edge-case]`, `[constitution]`.
- El dev escribe la respuesta; `/eva-specify NNN` la aplica, tilda el hallazgo y agrega `  - Applied: <qué cambió>`.
- Si no encuentra nada, el archivo dice `No findings.`
- Una spec no se puede aprobar sin `clarify.md` o con hallazgos abiertos.

## Plan: `plan.md`

El **cómo**. Secciones: 1 Approach · **2 Requirements traceability** · 3 Constitution check · 4 Architecture · 5 Data model · 6 API contracts · 7 State · 8 Navigation · **9 File map** · **10 Performance** · **11 Testing strategy** · 12 Dependencies · 13 Risks · 14 Rollout · 15 Alternatives.

- **Requirements traceability:** una tabla con una fila por FR → secciones del plan, componentes o archivos, nivel de test. Aparecen todos los FR; el plan solo cita FR que la spec define.
- **File map:** los únicos archivos que la implementación puede crear o modificar.
- **Performance:** cita los ids de reglas del checklist del stack (`.eva/stacks/<stack>.md`) y cómo se cumple cada una.
- Las decisiones transversales llevan un ADR en `docs/adr/` (`Status: proposed`).

## Tasks: `tasks.md`

```md
- [ ] T002 [feat] [P] [US-1] Validate the sign-in form — FR-1, FR-2
  - Files: `src/features/auth/SignInForm.tsx`, `src/features/auth/SignInForm.test.tsx`
  - Tests: `src/features/auth/SignInForm.test.tsx` — FR-1, FR-2
  - Done when: `npm test -- SignInForm` passes
```

| Regla | Detalle |
|-------|---------|
| Id | `T001`, `T002`, … únicos, en orden de ejecución; fases Setup → una por historia → Polish |
| Tipo | `[feat]` comportamiento nuevo · `[fix]` comportamiento que no cumple su FR · `[improve]` sin cambio de comportamiento, dentro del file map |
| Requisitos | se citan al final del título (y en *Tests*); `[feat]`/`[fix]` citan al menos uno; todo FR tiene al menos una task |
| Tamaño | un commit; máximo 3 archivos de código (más es warning, más de 5 es error); tests, `.md` y `.snap` no cuentan |
| Files | rutas entre backticks, todas del file map del plan |
| Tests | `[feat]`/`[fix]`: archivos de test entre backticks (`*.test.*`, `*.spec.*`, `*.e2e.*`, `*_test.*`, `__tests__/`, `e2e/`) y el FR que verifica cada uno; `[improve]`: "existing suites green" |
| Done when | una condición que otra persona puede comprobar |
| Tags opcionales | `[P]` en paralelo dentro de la fase · `[US-n]` historia de usuario |

Al final, *Implementation notes* registra el **paso en rojo** de cada task `[feat]`/`[fix]` (`T002 red: <test> failed with "<reason>"`) y cualquier desvío.

## Implementación

Por defecto, una task por corrida (`/eva-implement NNN [T00X|all]`):

1. Lee la task, el plan y el archivo del stack; toca solo los archivos de la task.
2. **Tests primero:** escribe los tests, los corre, confirma que fallan por la razón esperada y lo anota. `[improve]`: los tests existentes en verde antes y después, sin cambiar ninguno para que encaje.
3. Implementa el cambio mínimo que los hace pasar, siguiendo el checklist del stack.
4. Corre typecheck, lint y tests; tilda la task; anota los desvíos.
5. Hace commit solo si se lo piden: `<feat|fix|improve>(NNN): T00X <descripción>`.

Se detiene y pregunta cuando una task necesita un archivo fuera del file map, una dependencia nueva, otro enfoque o un cambio en un FR.

## Verificación: `verification.md`

| Sección | Contenido |
|---------|-----------|
| 1 | Chequeos automáticos (typecheck, lint, tests, e2e, `eva check`) |
| 2 | **Requirements traceability:** una fila por FR con su test y ✅/❌ |
| 3 | Criterios de finalización con evidencia |
| 4 | Checklist de performance (plan §10 + reglas del stack) con evidencia |
| 5 | Constitución y alcance (archivos fuera del file map) |
| 6 | **Veredicto:** `**Verdict:** PASS` o `**Verdict:** FAIL — qué falta` |

PASS solo cuando todos los FR y criterios de finalización están en ✅ y pasan los chequeos automáticos. Una spec `done` necesita veredicto PASS con todos los FR listados y ninguno fallido.

## Carpetas, ramas y commits

```
specs/
└── 001-sign-in/
    ├── spec.md
    ├── clarify.md
    ├── plan.md
    ├── tasks.md
    └── verification.md
```

- Carpeta: número de tres dígitos + slug en kebab-case; el número es único en todas las ramas.
- Rama: `feat/NNN-slug`, así `eva check` vincula los cambios de código con la spec.
- Commits: `feat(NNN): T003 add login form validation`; el tipo es el de la task.
- El pull request enlaza `specs/NNN-slug/` e incluye `verification.md`.

## Stacks y ADRs

`.eva/stacks/<stack>.md` tiene las librerías base, comandos, estructura de carpetas, presupuestos de performance, un checklist de performance con ids de reglas (basado en las skills de buenas prácticas de React Native y React de Vercel) y convenciones de testing y accesibilidad. Se cambia ahí, con un ADR, no feature por feature. Los ADRs viven en `docs/adr/NNNN-title.md` a partir de `.eva/templates/adr.md`.
