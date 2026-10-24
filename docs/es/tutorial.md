# Tutorial: una feature de punta a punta

[English](../tutorial.md) · **Español** · [Índice](README.md)

Este tutorial lleva una feature chica, *iniciar sesión con email y contraseña* en una app React Native, por todas las fases. Vas a ver qué escribe el agente, qué editas **tú** para aprobar y qué dice `eva check` en cada paso. Todas las salidas son reales.

Necesitas eva instalado ([Primeros pasos](getting-started.md)) y Claude Code (u OpenCode).

## 0. Un proyecto de prueba

Usa un proyecto descartable para poder romper cosas a propósito:

```sh
mkdir -p ~/Code/eva-playground && cd ~/Code/eva-playground
npx create-expo-app@latest .          # opcional: código real para el gate
git init -b main                      # create-expo-app puede haber creado el repo ya
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

> Si `create-expo-app` ya creó el repo en `master`, corre `git branch -m master main`. El gate compara las ramas de feature contra `main` (o `master`, o `--base`).

## 1. Constitución

En Claude Code:

```
/eva-constitution
```

El agente lee el repositorio, te hace **hasta 6 preguntas, de a una**, y escribe `.eva/constitution.md`: entre 6 y 10 principios cortos, cada uno con una línea *Check:* que dice cómo lo verifica quien revisa. Pone `status: in-review` y se detiene.

Revísala, cambia lo que no te convenza y apruébala editando tú el front-matter:

```yaml
---
type: constitution
status: approved        # lo pones tú, los agentes nunca
version: 1
created: 2026-10-24
approved_by: Maribel
approved_at: 2026-10-24
closed_at:              # siempre vacío en la constitución
---
```

Una constitución aprobada no puede tener `TODO` (EVA054). Haz commit.

## 2. Spec

```
/eva-specify Users sign in with email and password
```

El agente:

1. numera la feature (`001`) y propone la rama `feat/001-sign-in`; acéptala;
2. **te entrevista, de a una pregunta (máx. 6)**, antes de escribir nada;
3. escribe `specs/001-sign-in/spec.md` con historias de usuario, **requisitos funcionales en EARS** y **criterios de finalización**, y pone `status: in-review`.

```md
## 4. Functional requirements

- **FR-1** (US-1) When the user taps "Sign in" with a registered email and the right password, the app shall open the home screen.
- **FR-2** (US-1) If the email or the password is empty, then the app shall show "Required" below each empty field and stay on the sign-in screen.
- **FR-3** (US-1) If the server rejects the credentials, then the app shall show "Wrong email or password" and clear the password field.

## 7. Completion criteria

- Every FR in §4 has at least one passing test named after it.
- Works on iOS 16+ and Android 10+.
```

Todavía no la apruebes. Si lo intentas, `eva check` te frena:

```
  ERROR EVA070 specs/001-sign-in/spec.md: spec is approved but clarify.md is missing: run /eva-clarify before approving
```

## 3. Clarificación

```
/eva-clarify 001
```

El agente revisa la spec (**nunca la edita**) y escribe `specs/001-sign-in/clarify.md` con un checkbox por hallazgo: ambigüedades, contradicciones, casos límite que faltan y conflictos con la constitución.

```md
- [ ] **C-1** [edge-case] (§5) What happens when the user taps "Sign in" while offline: show a message, or queue the request?
  - Answer:
- [ ] **C-2** [ambiguity] (FR-2) Is an email without "@" also "Required", or a different message?
  - Answer:
```

Mientras haya hallazgos abiertos, la spec no se puede aprobar:

```
  ERROR EVA071 specs/001-sign-in/clarify.md: 2 open finding(s) but spec.md is approved: answer them and apply them with /eva-specify
```

**Tú** escribes cada respuesta debajo de `Answer:`. Después:

```
/eva-specify 001
```

El agente aplica cada hallazgo respondido a la spec, lo tilda y anota qué cambió. Nunca responde por ti:

```md
- [x] **C-1** [edge-case] (§5) What happens when the user taps "Sign in" while offline: show a message, or queue the request?
  - Answer: Show "You are offline" and do not send the request.
  - Applied: FR-4 added.
```

```md
- **FR-4** (US-1) While the device is offline, when the user taps "Sign in", the app shall show "You are offline" and not send the request.
```

Si los cambios fueron grandes, corre `/eva-clarify 001` otra vez. Cuando no quede ningún hallazgo abierto, **aprueba la spec** (`status: approved`, `approved_by`, `approved_at`) y haz commit.

## 4. Plan

```
/eva-plan 001
```

El plan explica el **cómo**. Su §2 *Requirements traceability* conecta cada FR con las partes del plan que lo implementan. El agente pone `status: in-review` y corre `eva check`; en revisión, los problemas son warnings:

```
  WARN  EVA088 specs/001-sign-in/plan.md: FR-4 is not in the "Requirements traceability" section
```

Ya corregido:

```md
| FR | Plan sections | Components / files | Test level |
|----|---------------|--------------------|------------|
| FR-1 | §4, §7 | `SignInScreen`, `useSignIn` | unit + e2e |
| FR-2 | §4 | `SignInForm` | unit |
| FR-3 | §6, §7 | `useSignIn` | unit |
| FR-4 | §7 | `useSignIn`, `useNetworkStatus` | unit |
```

Revisa sobre todo el file map (§9) y la sección de performance (§10). Aprueba el plan.

## 5. Tasks

```
/eva-tasks 001
```

Cada task tiene **tipo** (`[feat]`, `[fix]`, `[improve]`), **cita los FR** que implementa, es un commit (≤ 3 archivos de código, 5 como máximo) y lista **Files**, **Tests** y **Done when**:

```md
- [ ] T001 [feat] [US-1] Sign-in form with required fields — FR-2
  - Files: `src/features/auth/SignInForm.tsx`, `src/features/auth/SignInForm.test.tsx`
  - Tests: `src/features/auth/SignInForm.test.tsx` — FR-2
  - Done when: `npm test -- SignInForm` passes
```

Si al agente se le olvida algo, `eva check` avisa mientras las tasks están en revisión (y falla cuando están aprobadas):

```
  WARN  EVA080 specs/001-sign-in/tasks.md: T003 has no type: tag it [feat], [fix] or [improve]
  WARN  EVA084 specs/001-sign-in/tasks.md: T002 has no "Done when:" condition
```

Aprueba las tasks y haz commit.

### Prueba el gate

Antes de aprobar las tasks, crea cualquier archivo de código en la rama de la feature y corre `eva check`:

```
eva check — 1 feature(s), branch feat/001-sign-in vs main (1 code change(s))
  ERROR EVA042 specs/001-sign-in: 1 code change(s) on "feat/001-sign-in" but not approved yet: tasks.md (in-review)
```

Es código todo lo que esté fuera de `specs/`, `.eva/`, `docs/`, `.claude/`, `.opencode/` y de los `*.md` de la raíz, incluidos archivos sin commitear o sin trackear.

## 6. Implementar

```
/eva-implement 001          # la siguiente task
/eva-implement 001 T002     # una task puntual
/eva-implement 001 all      # todas las que faltan
```

En cada task `[feat]`/`[fix]`, el agente escribe primero los tests, los corre y **los ve fallar por la razón esperada**, lo anota en *Implementation notes* (`T001 red: …`) y después escribe el código mínimo para que pasen. Los tests llevan el nombre del requisito (`it('FR-2: shows Required under empty fields', …)`). Solo tilda una task cuando pasan tests, typecheck y lint, y se detiene si necesita un archivo fuera del file map del plan.

## 7. Verificar y cerrar

```
/eva-verify 001
```

El agente corre los chequeos, recorre **requisito por requisito** y escribe `verification.md` con una fila por FR y un veredicto:

```md
| FR | Test(s) | Result |
|----|---------|--------|
| FR-1 | `useSignIn.test.ts` › "FR-1: opens home on valid credentials" | ✅ |
| FR-2 | `SignInForm.test.tsx` › "FR-2: shows Required under empty fields" | ✅ |
| FR-3 | `useSignIn.test.ts` › "FR-3: shows error and clears password" | ✅ |
| FR-4 | `useSignIn.test.ts` › "FR-4: does not send while offline" | ✅ |

**Verdict:** PASS — every FR and completion criterion is met.
```

Si estás de acuerdo, cierra la feature: en `spec.md` pon `status: done` y `closed_at`, y pon `closed_at` en `verification.md`. `eva check` solo lo acepta con veredicto PASS y todos los FR aprobados:

```
  ERROR EVA092 specs/001-sign-in/verification.md: spec is done but the verdict is not PASS (FAIL — FR-4 has no test.)
  ERROR EVA093 specs/001-sign-in/verification.md: FR-4 failed verification but the spec is done
```

```
eva check — 1 feature(s), branch feat/001-sign-in vs main (1 code change(s))
OK
```

Abre el pull request con `verification.md`. Con `eva check` en CI ([Primeros pasos](getting-started.md#correrlo-en-ci-github-actions)), el gate corre en cada PR.

## Cosas para probar

| Haz esto | Esperas |
|----------|---------|
| Aprobar un documento sin `approved_at` | EVA063 |
| Escribir `created: 24/10/2026` | EVA061 |
| Escribir un FR sin "shall" | EVA028 (con la spec aprobada) |
| Tildar un hallazgo de clarify sin respuesta | EVA074 |
| Citar `FR-9` en una task | EVA081 |
| Listar 6 archivos de código en una task | EVA086 |

Cada código está explicado en la [referencia de `eva check`](check-reference.md).
