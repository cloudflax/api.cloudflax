# GitHub Workflow & Traceability (`cloudflax/api.cloudflax`)

Agente integrado en este repo: trazabilidad vía GitHub Issues/Project y ramas. No ejecutes Issue + rama + PR sin señal del usuario.

## Inicio

Ante trabajo trazable (feature, cambio de comportamiento, refactor relevante), **pregunta** si quiere flujo de trazabilidad. Matriz:

- **Sí, rama + trazabilidad** → crear Issue en `@api.cloudflax`, anotar `#ID`, rama `feature/<ID>-<slug-kebab>`.
- **Solo crear tarea** → solo Issue en el project; sin rama hasta nueva orden.
- **Solo commits (rama ya ligada al issue)** → ya estás en `feature/<ID>-<slug>` (o equivalente donde el prefijo numérico sea el id del issue); no crear issue ni rama nueva. Commits en **inglés** y Conventional Commits, cuerpo o pie con **`Refs #ID`** o **`Closes #ID`** según cierre real; sin `push`/PR salvo petición explícita (igual que en [Cierre y commits](#cierre-y-commits)).

## Issues y ramas

Creación: `gh issue create -R cloudflax/api.cloudflax -p "@api.cloudflax"` (nombre del project exacto). Cuerpo con objetivos técnicos y criterios de aceptación. Rama: `feature/<ID_ISSUE>-<slug>`; el `#ID` debe existir antes. Ramas dedicadas para revisiones o paralelismo; no forzar para micro-cambios.

## Cierre y commits

No hagas `git push` ni PR al acabar código salvo petición explícita o acuerdo en sesión. PR: `Closes #ID` o `Refs #ID`. Commits y descripciones en **inglés**; Conventional Commits (`feat:`, `fix:`, …). Si la rama sigue `feature/<ID>-…`, usa ese **mismo `#ID`** en el mensaje para asociar el commit al issue. Sin `git commit --amend` ni reescritura de historial remoto salvo petición expresa.

## Project: Status

Mantén **Status** alineado con la realidad. Valores: **Backlog**, **Ready**, **In progress**, **In review**, **Done**. Flujo típico en ese orden.

## Project: Priority, Size, Estimate, fechas

Además de Status, rellenar desde la IA cuando aplique: **Priority** y **Size** (single select) — proponer tras entender alcance e impacto, alinear vocabulario con el tablero. **Estimate** (número), **Start date** y **Target date** — solo con acuerdo explícito del usuario o política del equipo; no inventar fechas ni números.

CLI: `gh project list --owner cloudflax` (project id); `gh project field-list <n> --owner cloudflax` (field ids y option ids); localizar `item-id` con `gh project item-list`. Un campo por comando: `gh project item-edit --project-id … --id <item-id> --field-id …` + `--single-select-option-id …` | `--number …` | `--date YYYY-MM-DD` | `--clear`.

Orden sugerido: crear issue → enlace al project → Priority/Size cuando el alcance esté claro → Estimate/fechas si hay criterio.

## Tooling

Errores de permisos o scopes (`project`, `read:project`): pedir al usuario `gh auth refresh` (u auth adecuada) antes de reintentar.
