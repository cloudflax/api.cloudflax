# GitHub Workflow & Traceability (`cloudflax/api.cloudflax`)

Agente integrado en este repo: trazabilidad vía GitHub Issues/Project y ramas. No ejecutes Issue + rama + PR sin señal del usuario.

## Inicio

Ante trabajo trazable (feature, cambio de comportamiento, refactor relevante), **pregunta** si quiere flujo de trazabilidad. Matriz:

- **Sí, rama + trazabilidad** → crear Issue en `@api.cloudflax`, anotar `#ID`, rama `feature/<ID>-<slug-kebab>`.
- **Solo crear tarea** → solo Issue en el project; sin rama hasta nueva orden.
- **Solo commits (rama ya ligada al issue)** → ya estás en `feature/<ID>-<slug>` (o equivalente donde el prefijo numérico sea el id del issue); no crear issue ni rama nueva. Commits en **inglés** y Conventional Commits, cuerpo o pie con **`Refs #ID`** o **`Closes #ID`** según cierre real; sin `push`/PR salvo petición explícita (igual que en [Cierre y commits](#cierre-y-commits)).

## Issues y ramas

Creación: `gh issue create -R cloudflax/api.cloudflax -p "@api.cloudflax"` (nombre del project exacto). Cuerpo con objetivos y criterios de aceptación. Rama: `feature/<ID_ISSUE>-<slug>`; el `#ID` debe existir antes. Ramas dedicadas para revisiones o paralelismo; no forzar para micro-cambios.

**Primera creación (misma pasada):** issue con `--assignee cloudflax` y `--label …` si el CLI lo permite; en el project, **Priority** y **Size** siempre con valor (no dejar el select vacío). **Estimate** y fechas solo con criterio o acuerdo; si no hay, indicarlo en el cuerpo — no inventar. Assignee alternativo si `cloudflax` no aplica en el org.

## Cierre y commits

No hagas `git push` ni PR al acabar código salvo petición explícita o acuerdo en sesión. PR: `Closes #ID` o `Refs #ID`. Commits y descripciones en **inglés**; Conventional Commits (`feat:`, `fix:`, …). Si la rama sigue `feature/<ID>-…`, usa ese **mismo `#ID`** en el mensaje para asociar el commit al issue. Sin `git commit --amend` ni reescritura de historial remoto salvo petición expresa.

## Project: Status

Mantén **Status** alineado con la realidad. Valores: **Backlog**, **Ready**, **In progress**, **In review**, **Done**. Flujo típico en ese orden.

## Project: Priority, Size, Estimate, fechas

CLI: `gh project list --owner cloudflax` → `gh project field-list <n> --owner cloudflax` → `item-id` vía `gh project item-list`. Editar: `gh project item-edit --project-id … --id <item-id> --field-id …` + `--single-select-option-id …` | `--number …` | `--date YYYY-MM-DD` | `--clear`. Issue: `gh issue edit <n> --add-assignee cloudflax --add-label "<label>"`.

## Tooling

Errores de permisos o scopes (`project`, `read:project`): pedir al usuario `gh auth refresh` (u auth adecuada) antes de reintentar.
