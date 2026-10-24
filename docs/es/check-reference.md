# Referencia de `eva check`

[English](../check-reference.md) · **Español** · [Índice](README.md)

`eva check` valida los documentos SDD de un proyecto y hace cumplir el gate de aprobación en las ramas de feature. Solo lee, nunca escribe.

```sh
eva check                 # documentos + gate por rama
eva check --no-git        # solo documentos
eva check --base develop  # ref base contra la que comparar
eva check --branch feat/004-login   # nombre de la rama (p. ej. con HEAD desacoplado)
eva check --strict        # EVA043 pasa a ser error
eva check --dir ../app    # carpeta del proyecto (por defecto .)
```

Salida: una línea por hallazgo (`ERROR` o `WARN`, código, ruta, mensaje), ordenadas por ruta, y al final un resumen u `OK`. Códigos de salida: `0` sin errores (puede haber warnings), `1` al menos un error, `2` flags inválidos.

## Cómo funciona el gate

1. **Rama:** `--branch`, si no `GITHUB_HEAD_REF`, si no la rama actual.
2. **Base:** `--base`, si no `origin/$GITHUB_BASE_REF` / `$GITHUB_BASE_REF`, si no la primera que exista entre `origin/main`, `main`, `origin/master`, `master`. En la propia rama base no se bloquea nada.
3. **Archivos cambiados:** todo lo que difiere del merge-base con la base, incluidos archivos sin commitear y sin trackear.
4. **Código:** cualquier archivo cambiado fuera de `specs/`, `.eva/`, `docs/`, `.claude/`, `.opencode/` y de los Markdown de la raíz.
5. **Feature:** el número del nombre de la rama (`feat/004-login` → `004`) elige `specs/004-*`.
6. Si hay cambios de código, la constitución y la spec, el plan y las tasks de esa feature tienen que estar aprobados.

## Severidad

La mayoría de las reglas son siempre error. Las de plan y tasks (EVA080–088) dependen del estado del documento: **warning** en `in-review` (así el agente las corrige antes de pedir revisión), **error** con `approved` o `done`, y no se revisan en `draft`. Las reglas de contenido de la spec (EVA025–029) y el gate de clarify (EVA070–071) se aplican con la spec aprobada. Las de verificación (EVA090–093), con la spec en `done`.

---

## Proyecto

### EVA001 — no es un proyecto eva
**Error.** Falta `.eva/config.json`. **Solución:** corre `eva init`.

### EVA010 — nombre de carpeta
**Warning.** Una carpeta de `specs/` no es `NNN-slug-en-kebab` (p. ej. `Login Feature`). Se ignora. **Solución:** renómbrala, p. ej. `specs/004-user-login`.

### EVA011 — número de feature repetido
**Error.** Dos carpetas comparten número (`004-login`, `004-signup`). **Solución:** renumera una; revisa también otras ramas (`git branch -a`).

## Documentos

### EVA020 — falta la spec
**Error.** Una carpeta de feature no tiene `spec.md`. **Solución:** corre `/eva-specify` o borra la carpeta.

### EVA021 — front-matter inválido
**Error.** El archivo no empieza con `---`, el bloque no se cierra o una línea no es `clave: valor`. Aplica a spec, clarify, plan, tasks y verification. **Solución:** copia el front-matter de `.eva/templates/`.

### EVA022 — el id no coincide con la carpeta
**Error.** `id: 001-b` dentro de `specs/001-a/`. **Solución:** pon en `id` el nombre de la carpeta.

### EVA023 — estado inválido
**Error.** El estado no es `draft`, `in-review`, `approved` ni `done`.

### EVA024 — aprobado sin aprobador
**Error.** `approved`/`done` con `approved_by` vacío. **Solución:** lo completa el dev que aprueba; los agentes nunca.

### EVA025 — preguntas abiertas en una spec aprobada
**Error.** Una spec aprobada todavía tiene `[NEEDS CLARIFICATION …]` (se ignora el texto entre backticks o en bloques de código). **Solución:** resuélvelo o vuelve la spec a `in-review`.

## Contenido de la spec (specs aprobadas)

### EVA026 — sin requisitos
**Error.** No hay ítems de lista que empiecen con `FR-n` bajo un título *Functional requirements*. **Solución:** agrega requisitos en §4; los FR en otras secciones son solo referencias.

### EVA027 — id de requisito repetido
**Error.** El mismo `FR-n` está definido dos veces. **Solución:** dale al nuevo el siguiente número libre; nunca renumeres.

### EVA028 — no es EARS
**Error.** Un requisito no empieza con The/When/While/If/Where, no tiene `shall <respuesta>` o tiene un `If` sin `, then`. **Solución:** reescríbelo, p. ej. `If the password is empty, then the app shall show "Required".`

### EVA029 — criterios de finalización
**Error.** No hay sección *Completion criteria* o no tiene ítems de lista (los comentarios no cuentan). **Solución:** lista criterios verificables en §7.

## Orden de fases

### EVA030 — plan antes de aprobar la spec
**Error.** Existe `plan.md` pero la spec no está aprobada. **Solución:** aprueba primero la spec o borra el plan.

### EVA031 — tasks antes de aprobar el plan
**Error.** Existe `tasks.md` pero el plan no está aprobado.

### EVA032 — tildadas antes de aprobar
**Error.** Hay tasks tildadas pero `tasks.md` no está aprobado.

### EVA033 — done con tasks abiertas
**Error.** La spec está en `done` pero `tasks.md` falta, está vacío o tiene ítems sin tildar.

## Gate por rama

### EVA040 — no es un repositorio git
**Warning.** Se saltea el gate. **Solución:** `git init` o usa `--no-git`.

### EVA041 — la feature no existe
**Error** si la rama tiene cambios de código, **warning** si no. La rama apunta a `NNN` pero no hay `specs/NNN-*`. **Solución:** crea la spec con `/eva-specify` o renombra la rama.

### EVA042 — código antes de aprobar
**Error.** La rama tiene cambios de código y la constitución o la spec, el plan o las tasks de la feature no están aprobados; el mensaje lista lo pendiente (p. ej. `tasks.md (in-review)`, `plan.md (missing)`). **Solución:** termina las fases o mueve el código a otra rama.

### EVA043 — código en una rama sin id de feature
**Warning** (error con `--strict`). Cambios de código en una rama que no es `…/NNN-slug`. **Solución:** usa `feat/NNN-slug`.

### EVA044 — no se encuentra la rama base
**Warning.** No existe ninguna de las bases candidatas; se saltea el gate. **Solución:** `--base <ref>` o `fetch-depth: 0` en CI.

### EVA045 — HEAD desacoplado
**Warning.** **Solución:** `--branch <nombre>`.

### EVA046 — falló git diff
**Warning.** El mensaje incluye el error de git; se saltea el gate.

## Constitución

### EVA050 — falta la constitución
**Error.** **Solución:** `eva init` la restaura.

### EVA051 — constitución inválida
**Error.** Front-matter inválido o un estado distinto de `draft`, `in-review`, `approved`.

### EVA052 — aprobada sin aprobador
**Error.** `approved_by` está vacío.

### EVA053 — constitución sin aprobar
**Warning** mientras no haya specs (un proyecto recién instalado pasa); **error** en cuanto existe una spec. **Solución:** `/eva-constitution` y después apruébala.

### EVA054 — quedan TODO
**Error.** Una constitución aprobada tiene `TODO` (fuera de código). **Solución:** reemplaza o borra cada placeholder.

## Tipo y fechas (todos los documentos)

### EVA060 — tipo
**Error.** Falta `type` o no corresponde al archivo: `constitution`, `spec`, `clarify`, `plan`, `tasks`, `verification`.

### EVA061 — formato de fecha
**Error.** Una fecha no es una fecha `YYYY-MM-DD` real (`23/10/2026`, `2026-02-30`).

### EVA062 — falta created
**Error.** `created` vacío fuera de `draft` (en clarify y verification siempre es obligatorio).

### EVA063 — approved_at
**Error.** `approved`/`done` sin `approved_at`, o `approved_at` puesto sin estar aprobado. **Solución:** al volver un documento a `in-review`, vacía `approved_by` y `approved_at`.

### EVA064 — closed_at
**Error.** `done` sin `closed_at`, `closed_at` puesto antes de `done`, o cualquier `closed_at` en la constitución.

### EVA065 — orden de fechas
**Error.** Dentro de un documento no se cumple `created ≤ approved_at ≤ closed_at`.

## Clarificación

### EVA070 — falta clarify
**Error.** La spec está aprobada pero no existe `clarify.md`. **Solución:** vuelve la spec a `in-review` y corre `/eva-clarify`.

### EVA071 — hallazgos abiertos
**Error.** La spec está aprobada y `clarify.md` tiene hallazgos sin tildar. **Solución:** respóndelos y corre `/eva-specify NNN`.

### EVA072 — id de hallazgo repetido
**Error.** Dos hallazgos usan el mismo `C-n`.

### EVA073 — categoría
**Error.** Un hallazgo no tiene categoría o tiene una distinta de `[ambiguity]`, `[contradiction]`, `[edge-case]`, `[constitution]`.

### EVA074 — tildado sin respuesta
**Error.** Un hallazgo está en `- [x]` pero su `Answer:` está vacío. Solo responde el dev.

## Plan y tasks (warning en revisión, error con aprobación)

### EVA080 — tipo de task
Una task no tiene tag `[feat]`, `[fix]` ni `[improve]`.

### EVA081 — requisito desconocido
El plan o una task cita un `FR-n` que la spec no define.

### EVA082 — requisito sin task
Ninguna task cita un `FR-n` de la spec.

### EVA083 — feat/fix sin requisito
Una task `[feat]` o `[fix]` no cita ningún `FR-n`. Si no cambia el comportamiento, es un `[improve]`.

### EVA084 — falta Files o Done when
Una task no tiene línea `Files:` con rutas entre backticks, o no tiene `Done when:`.

### EVA085 — tests
Una task `[feat]`/`[fix]` no tiene línea `Tests:` con un archivo de test, o la línea no dice qué `FR-n` verifican los tests.

### EVA086 — task demasiado grande
Más de 3 archivos de código es siempre **warning**; más de 5 sigue al estado (error con aprobación). Tests, `.md` y `.snap` no cuentan. **Solución:** divide la task.

### EVA087 — id de task repetido
Dos tasks comparten número `T`.

### EVA088 — trazabilidad del plan
El plan no tiene la sección *Requirements traceability* o le falta un `FR-n` de la spec.

## Verificación (spec en done)

### EVA090 — falta la verificación
**Error.** **Solución:** corre `/eva-verify`.

### EVA091 — requisito sin verificar
**Error.** Un `FR-n` no tiene fila en la tabla *Requirements traceability*.

### EVA092 — veredicto
**Error.** No hay línea `Verdict:` o no es `PASS` (un placeholder como `PASS | FAIL` cuenta como faltante).

### EVA093 — requisito fallido
**Error.** Una fila de FR está marcada ❌ o `FAIL` con la spec en `done`.
