# Primeros pasos

[English](../getting-started.md) · **Español** · [Índice](README.md)

## Requisitos

- Homebrew, o Go 1.24 o superior, para instalar eva (tu proyecto no necesita Go).
- Git, para el gate por rama de `eva check`.
- Un agente de código. Claude Code y OpenCode reciben atajos `/eva-*`; cualquier otro agente puede seguir `AGENTS.md` y `.eva/commands/*.md`.

## Instalar

Con Homebrew (macOS y Linux):

```sh
brew install codebymaribel/tap/eva
brew upgrade eva      # más adelante, para actualizar
```

Homebrew compila eva desde el código del release y compara la descarga con la huella (checksum) de la receta. Instala Go por su cuenta como dependencia de compilación.

Con Go:

```sh
go install github.com/codebymaribel/eva-sdd/cmd/eva@latest
```

Desde un clon local (durante el desarrollo):

```sh
cd eva-sdd
make install          # go install ./cmd/eva
```

`go install` y `make install` dejan el binario en `$(go env GOPATH)/bin`. Si la terminal dice `command not found: eva`, agrega esa carpeta al `PATH`:

```sh
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc   # o ~/.bashrc
source ~/.zshrc
eva -h
```

`eva version` muestra la versión instalada (`dev` en builds locales). Los templates van embebidos en el binario: instalar un eva nuevo también trae templates nuevos.

## Agregar eva a un proyecto

Córrelo en la raíz del proyecto. Haz commit antes, así los cambios de eva son fáciles de revisar.

```sh
eva init
```

En una terminal, un asistente pregunta los stacks, la carpeta de cada app (solo si hay varias) y los agentes, muestra el plan de archivos y pide confirmación. Sin terminal (CI, scripts) usa flags:

```sh
eva init --stack rn --agent claude                                   # una app
eva init --stack rn,next --agent claude,opencode \
         --path rn=apps/mobile,next=apps/web                         # monorepo
eva init --dry-run                                                   # solo vista previa
```

| Flag | Descripción |
|------|-------------|
| `--stack` | `react-native` (alias `rn`, `expo`), `nextjs` (alias `next`, `next.js`), separados por coma |
| `--agent` | `claude` (alias `claude-code`), `opencode` (alias `open-code`), separados por coma |
| `--path` | carpeta de cada app, p. ej. `rn=apps/mobile`; por defecto `.` con un stack, `apps/mobile` / `apps/web` con varios |
| `--name` | nombre del proyecto (por defecto, el de la carpeta) |
| `--dir` | carpeta del proyecto (por defecto `.`) |
| `--dry-run` | muestra el plan sin escribir |
| `--force` | sobrescribe archivos que difieren de los templates, aunque los hayas editado |
| `--prune` | borra archivos que eva instaló antes, ya no trae y nadie editó |
| `--diff` | muestra en qué difiere del template nuevo cada archivo editado o actualizado |
| `--yes` | no pide confirmación |

Códigos de salida: `0` éxito, `1` falla o cancelado, `2` flags o argumentos inválidos.

## Qué se instala

```
AGENTS.md                         # bloque de eva: fases, gate de aprobación, reglas de implementación
CLAUDE.md                         # (claude) bloque de eva que importa @AGENTS.md
.claude/commands/eva-*.md         # (claude) /eva-constitution … /eva-verify
.opencode/commands/eva-*.md       # (opencode) /eva-constitution … /eva-verify
.eva/
  config.json                     # qué instaló eva: stacks, rutas, agentes, versión
  eva-lock.json                   # hash de cada archivo instalado (generado: se commitea, no se edita)
  constitution.md                 # principios del proyecto, aprobados una vez por un dev
  stacks/react-native.md          # (rn) base, estructura, comandos, checklist de performance
  stacks/nextjs.md                # (next) base, estructura, comandos, checklist de performance
  commands/*.md                   # instrucciones de cada fase, independientes del agente
  templates/{spec,clarify,plan,tasks,verification,adr}.md
specs/README.md                   # convenciones de carpetas, ciclo de estados, reglas
docs/adr/README.md                # registros de decisiones de arquitectura
.gitattributes                    # bloque de eva: marca eva-lock.json como generado para que los PRs lo plieguen
```

Después de instalar:

1. En tu agente, corre `/eva-constitution`, revisa y aprueba `.eva/constitution.md` (ver [el tutorial](tutorial.md)).
2. Revisa `.eva/stacks/*.md` (comandos, estructura, presupuestos) y ajústalos al proyecto.
3. Haz commit de los archivos generados.
4. Empieza una feature con `/eva-specify <lo que quieres construir>`.

## Volver a correrlo y actualizar

Correr `eva init` de nuevo es seguro, y es la forma de traer los templates nuevos después de actualizar eva:

- Los stacks y agentes instalados se mantienen; los nuevos se agregan.
- `AGENTS.md`, `CLAUDE.md` y `.gitattributes` solo reciben un bloque gestionado por eva (`<!-- eva:start -->…<!-- eva:end -->`, o `# eva:start`…`# eva:end` en `.gitattributes`). Tu contenido fuera del bloque nunca se toca; el bloque siempre se actualiza.
- `.eva/config.json` y `.eva/eva-lock.json` son de eva y siempre se reescriben.
- Para cualquier otro archivo, eva lo compara con el hash que guardó en `.eva/eva-lock.json` cuando lo escribió:

| El archivo… | `eva init` |
|-------------|------------|
| no existe | lo crea |
| es igual al template nuevo | lo deja igual |
| coincide con el hash guardado (nunca lo editaste) | **lo actualiza** al template nuevo |
| lo editaste | lo conserva y lo avisa; si además cambió el template, `eva init --dry-run --diff` muestra la diferencia |
| eva ya no lo trae | lo avisa; `--prune` lo borra si nunca lo editaste |

`--force` sobrescribe todos los archivos que difieren de su template, editados o no: haz commit antes y recupera lo que quieras con `git checkout -p`.

**`.eva/eva-lock.json`** es generado: se commitea para que todo el equipo comparta el mismo registro, y no se edita. `eva init` lo marca como generado en `.gitattributes`, así GitHub lo pliega en los pull requests. Los hashes ignoran los finales de línea, así que un checkout que convierte `\n` en `\r\n` no cuenta como edición.

**Proyectos instalados antes de que existiera el lock:** en el primer `eva init` con un eva más nuevo, los archivos iguales a los templates se registran, y los que difieren se tratan como editados (eva no puede saber su origen): se conservan, con `--diff` para comparar. Los archivos que eva instaló antes del lock y ya no trae (por ejemplo los viejos atajos `sdd-*.md`) no se detectan; bórralos tú.

## Chequear el proyecto

```sh
eva check                 # documentos + gate por rama
eva check --no-git        # solo documentos
eva check --base develop  # rama base personalizada
eva check --strict        # también falla con código en ramas sin id de feature
```

`eva check` sale con `1` cuando hay errores; los warnings nunca fallan. Cada hallazgo tiene un código: ver la [referencia de `eva check`](check-reference.md).

## Correrlo en CI (GitHub Actions)

```yaml
name: eva
on: pull_request
jobs:
  eva-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0 # el gate compara contra la rama base
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - run: go install github.com/codebymaribel/eva-sdd/cmd/eva@latest
      - run: eva check
```

`GITHUB_HEAD_REF` y `GITHUB_BASE_REF` se leen solos. Para asegurar que solo personas aprueben, agrega una regla `CODEOWNERS` para `specs/**` y `.eva/constitution.md` y exige revisión de code owners.

Siguiente: [Tutorial](tutorial.md).
