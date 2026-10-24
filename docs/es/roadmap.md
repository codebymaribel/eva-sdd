# Roadmap y decisiones

[English](../roadmap.md) · **Español** · [Índice](README.md)

## v1

| # | Feature | Estado |
|---|---------|--------|
| 1 | Constitución con estado y versión, `/eva-constitution`, gate (EVA050–054) | ✅ hecha |
| 2 | Spec: `FR-n` en EARS, entrevista (máx. 6 preguntas, de a una), criterios de finalización, `type`/`created`/`approved_at`/`closed_at` en todos los documentos, validación de fechas (EVA026–029, EVA060–065) | ✅ hecha |
| 3 | Clarificación: `/eva-clarify` escribe `clarify.md`, solo detecta; no se aprueba con hallazgos abiertos (EVA070–074) | ✅ hecha |
| 4 | Trazabilidad del plan a los FR; tasks tipadas de un commit con *Done when*; tests primero; verificación FR por FR con veredicto (EVA080–093) | ✅ hecha |
| 5 | Cambio: `/eva-change` reabre una spec no mergeada (vuelve a `in-review`) o, si ya está en main, crea una feature nueva con `amends: NNN-slug`; `eva check` valida `amends` | planificada |
| 6 | Fixes, improvements y mantenimiento: con la feature abierta, una task nueva `[fix]`/`[improve]` en su `tasks.md` (vuelve a `in-review`); ya en main, `specs/NNN/fixes/NN-slug.md` con `restores: FR-n` y `specs/NNN/improvements/NN-slug.md`; `specs/maintenance/` para lo que no pertenece a ninguna feature; ramas `fix/NNN-NN-slug`, `improve/NNN-NN-slug` y `fix/maintenance-NN-slug` con su gate; validación del tipo en los commits | planificada |
| 7 | Reglas en `AGENTS.md`: sección obligatoria "Al terminar cualquier tarea" con verificación, y "no modificar `specs/` sin pedido explícito" | planificada |
| 8 | **Hashes de templates:** `eva init` guarda un SHA-256 de cada archivo que escribe en `.eva/eva-lock.json` (generado, con un campo `_generated` que lo dice, commiteado y marcado `linguist-generated` en un bloque gestionado de `.gitattributes`). Los archivos intactos cuyo template cambió se actualizan; los editados se saltean y `eva init --dry-run --diff` muestra qué cambió en el template; `eva init --prune` borra archivos obsoletos solo si están intactos; los proyectos sin lock guardan las huellas en la primera corrida; los finales de línea se normalizan antes de calcular el hash. Los bloques de `AGENTS.md`/`CLAUDE.md` y `config.json` siguen como hoy | ✅ hecha |
| 9 | **Placeholders de templates:** el texto que quedó de los templates (tomado automáticamente de los templates embebidos, todo `<…>`, más "Feature title", "Short imperative description" y `path/to/…`) es warning en revisión y error con aprobación, en todos los documentos de la feature y en la constitución. Las tablas de guía de los templates (como los patrones EARS) pasan a comentarios HTML | planificada |
| 10 | **File map:** cada ruta de `Files:` de una task tiene que estar en el file map del plan (warning en revisión, error con aprobación); las entradas son rutas exactas o carpetas terminadas en `/`; las rutas del file map, de `Files:` y de `covers` son desde la raíz del repo. Gate EVA047: el código cambiado en la rama fuera del file map es un único warning agrupado (error con `--strict`), solo con las tasks aprobadas; los lockfiles y los patrones `ignore` de `.eva/config.json` no cuentan; cuando aparece, `/eva-implement` frena y pregunta si se agrega el archivo al plan o se revierte | planificada |

Cada feature actualiza el README, esta documentación (los dos idiomas) y los goldens.

## Decisiones de diseño

- **Un solo espacio de nombres: eva.** Los comandos son `/eva-<fase>`, la carpeta del proyecto es `.eva/` y el manifest `.eva/config.json`, igual que `eva init` / `eva check`. Los nombres genéricos `sdd` podían chocar con otras herramientas de SDD, y eva nunca pisa un archivo que escribió o editó otra persona.
- **Independiente del agente.** `AGENTS.md` es la única fuente de verdad; los atajos de cada agente solo apuntan a `.eva/commands/`.
- **Solo personas aprueban.** Los agentes pasan documentos a `in-review`; `approved`, `done`, `approved_by`, `approved_at` y `closed_at` son solo humanos. La aprobación se lee del archivo, nunca del chat.
- **Solo EARS** para los requisitos (sin Gherkin): un enunciado comprobable por `FR-n`; los ids nunca se reutilizan.
- **Clarify detecta, specify aplica.** Clarify nunca edita la spec; las respuestas son del dev; `/eva-specify` las aplica y tilda el hallazgo.
- **La severidad sigue al estado.** Las reglas de plan y tasks son warnings en revisión (el agente se corrige solo) y errores con aprobación; los borradores son libres.
- **Fechas por documento.** `created ≤ approved_at ≤ closed_at` se revisa dentro de cada documento, no entre documentos, así reabrir y reaprobar una spec no invalida planes anteriores. Las fechas alimentan `eva status` y `eva log NNN` en la v2.
- **Fix vs cambio.** Una spec aprobada nunca se edita para registrar un fix. Si cambia *qué* hace el sistema, es un cambio (feature 5); si no, un fix o un improvement (feature 6).
- **Tamaño de task por alcance**, no por tiempo: un commit, ≤ 3 archivos de código (límite 5), sin contar tests ni docs.
- **Tests primero** se exige por estructura (archivos de test y FR declarados en cada task) y por instrucción (se anota el paso en rojo); eva no revisa el historial de git para esto.
- **Solo biblioteca estándar** en la v1, así `go install` no necesita módulos de terceros (el asistente de la v2 es la excepción prevista).
- **Templates en inglés**, embebidos en el binario.

## Limitaciones conocidas

- **Archivos de antes del lock.** Los proyectos instalados antes de la feature 8 de la v1 no tienen registro de lo que escribió eva: los archivos que difieren de los templates se tratan como editados, y `--prune` no detecta los obsoletos de esa época (por ejemplo los viejos atajos `sdd-*.md`).
- **File map.** `eva check` no verifica que los archivos de las tasks pertenezcan al file map del plan (la regla de `[improve]` queda solo como instrucción). Lo resuelve la feature 10 de la v1.
- **Placeholders del template.** Una spec aprobada que todavía tiene placeholders `<…>` del template pasa. Lo resuelve la feature 9 de la v1.
- **Número de rama.** El gate toma el primer `NNN-` del nombre de la rama, así que las futuras ramas `fix/NNN-NN-slug` ya se asocian a la feature `NNN`; las reglas propias llegan con la feature 6.

## v2 (planificada)

| # | Feature | Estado |
|---|---------|--------|
| 1 | **`eva status`**: una tabla con todas las features del proyecto: número y título, estado de spec / plan / tasks, hallazgos de clarify abiertos, tasks hechas/total, veredicto de la verificación y `created` / `closed_at`. Funciona en CI (texto plano) y en la terminal | planificada |
| 2 | **`eva log NNN`**: historial cronológico de una feature a partir de las fechas del front-matter: spec, fixes, improvements y las features que la enmiendan (`amends`), con el evento (`created`, `approved`, `closed`) y el estado. Necesita las features 5 y 6 de la v1 | planificada |
| 3 | **Asistente interactivo**: `eva init` con menús navegables con flechas, selección múltiple y confirmación (`charmbracelet/huh`) detrás de `prompt.Prompter`; las preguntas por línea actuales quedan como alternativa fuera de una terminal y en los tests | planificada |
| 4 | **Proyectos existentes (brownfield), modo "spec al tocar"**: `eva adopt` (Go, determinístico) muestra el inventario: stack según `package.json`, features candidatas a partir de las rutas de Expo Router / Next.js App Router, tests existentes por feature y qué features ya tienen spec (`covers`). `/eva-adopt <feature>` escribe una spec base (`type: baseline`, `origin: reverse`) a partir del código; los FR dudosos quedan marcados `[review]` con `Evidence:`, y cada uno se confirma, se convierte en fix (`restores: FR-n`) o se descarta. Una spec base no se aprueba mientras queden `[review]`; los tests faltantes son warnings | planificada |
| 5 | **Propiedad del código**: `covers` (globs desde la raíz del repo) dice qué código le pertenece a una feature: en el front-matter del plan en las features normales (`/eva-verify` lo actualiza al cerrar) y en las specs base. `eva owner <ruta>` dice a qué feature pertenece un archivo, y `/eva-plan` lo usa para llenar una columna nueva **Owner** en el file map. Warnings: un archivo del file map que es de otra feature con el Owner vacío o equivocado; dos features que reclaman el mismo archivo; una rama que toca código de otra feature sin declararlo | planificada |
| 6 | **Hook para Claude Code** (opcional, por agente): antes de cada edición, un hook que instala eva bloquea los archivos que no están en `Files:` de la task en curso y pide aprobación primero, convirtiendo la detección en prevención. Otros agentes solo si ofrecen algo equivalente | planificada |

Decisiones de la v2:

- Las specs base usan ramas `docs/NNN-slug` y commits `docs(NNN):` (un tipo estándar de Conventional Commits); una rama `docs/` no puede tener cambios de código.
- El inventario no se versiona: `eva adopt` lo muestra (`--out <archivo>` lo escribe para leerlo). Las decisiones manuales (`ignore`, `group`) van en `.eva/adopt.json`, que sí se commitea.
- `covers` es obligatorio en las specs base y opcional en las features normales, donde vive en el plan porque la spec nunca nombra archivos. Solo genera warnings.
- El asistente suma el primer módulo de terceros (`huh`); aceptado, con las preguntas por línea como alternativa.

## v3 (ideas)

- Brownfield en modo "inventario completo": lotes y progreso sobre todo el inventario, y una constitución inferida a partir del código existente.

## Ideas para más adelante

- Binarios precompilados, para que ni Homebrew ni Go tengan que compilar eva (requiere firma y notarización de macOS).
- Más agentes (Cursor, GitHub Copilot, Gemini CLI) después de la v1, a medida que los pidan.
