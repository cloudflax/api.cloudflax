# GitHub Workflow & Git Flow (`cloudflax/api.cloudflax`)

**Git Flow:** integración por **merges** entre ramas (no PR por defecto). Issues/ramas solo si el usuario lo pide.

## Ramas

| Rama | Origen → destino |
|------|------------------|
| `main` | Solo `release/*` y `hotfix/*`. Producción. |
| `develop` | Integración hacia la próxima release. |
| `feature/<kebab>` | `develop` → trabajo → merge a `develop`. |
| `release/<versión>` | `develop` → `main` + de vuelta a `develop`. |
| `hotfix/<kebab>` | `main` → `main` + `develop`. |

Nombre: `feature/<ID>-<slug>` si hay issue (commits con el mismo `#ID`).

## Al empezar (preguntar)

- **Issue + rama:** `gh issue create …`, rama `feature/…` desde `develop` al día.
- **Solo Issue:** item en project; sin rama hasta orden.
- **Rama ya abierta:** solo commits; no duplicar issue/rama.

## Git

- Mensajes en **inglés**, **Conventional Commits**; pie `Refs #ID` / `Closes #ID` si aplica.
- Sin **`git push`**, **`amend`** o reescritura de remoto salvo petición explícita.
- Cierre **feature:** merge a **`develop`** (Git Flow).

## Issues

`gh issue create -R cloudflax/api.cloudflax -p "@api.cloudflax"`. Cuerpo: objetivos y criterios.

**Primera pasada:** `--assignee cloudflax`, `--label …`; en project **Priority** y **Size** obligatorios; **Estimate**/fechas solo si hay acuerdo (si no, decirlo en el cuerpo). Assignee alternativo si hace falta.

## Project (Status y campos)

**Status:** Backlog → Ready → In progress → In review (merge/revisión pendiente) → Done.

**CLI:** `gh project list --owner cloudflax` → `field-list` → `item-list` (item-id) → `project item-edit` (`--single-select-option-id` / `--number` / `--date` / `--clear`). Issue: `gh issue edit <n> --add-assignee … --add-label …`.

**Permisos `gh`:** `gh auth refresh` si falla `project` / `read:project`.
