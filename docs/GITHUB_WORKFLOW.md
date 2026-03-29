# GitHub Workflow & Traceability (`cloudflax/api.cloudflax`)

Actúa como un agente de desarrollo senior integrado en el repositorio `cloudflax/api.cloudflax`. Tu objetivo es garantizar la trazabilidad total del trabajo siguiendo estrictamente este protocolo de comandos y jerarquía.

## 1. Protocolo de Inicio y Autonomía
* **Detección de Tarea:** Ante cualquier trabajo trazable (features, cambios de comportamiento o refactors relevantes), **pregunta obligatoriamente** al usuario si desea iniciar el flujo de trazabilidad.
* **Restricción de Ejecución:** No asumas ni ejecutes el flujo completo (Issue + Rama + PR) de forma automática. Sigue esta matriz de decisión:

| Input del Usuario | Acción Obligatoria del Agente |
| :--- | :--- |
| **"Sí, rama + trazabilidad"** | 1. Crear Issue en `@api.cloudflax`. <br> 2. Obtener el `#ID`. <br> 3. Crear rama `feature/<ID>-<slug>`. |
| **"Solo crear tarea"** | Crear únicamente la Issue en el Project. No crear rama hasta nueva orden. |

## 2. Gestion de Issues y Projects
* **Comando de Creación:** Usa `gh issue create -R cloudflax/api.cloudflax -p "@api.cloudflax"`.
* **Project Board:** Es obligatorio asignar la Issue al proyecto de la organización **`@api.cloudflax`**. El nombre debe coincidir exactamente.
* **Contenido:** La Issue debe incluir objetivos técnicos claros y criterios de aceptación antes de proceder.

## 3. Estándar de Ramas (Branching)
* **Formato:** `feature/<ID_ISSUE>-<slug-corto-en-kebab>` (Ejemplo: `feature/7-auth-fix`).
* **Dependencia:** El número de la Issue **debe existir antes** de nombrar y crear la rama.
* **Uso:** Prioriza ramas dedicadas para revisiones aisladas o trabajos en paralelo. No fuerces ramas para micro-ediciones mínimas.

## 4. Desarrollo y Cierre (Loop de Feedback)
* **Iteración:** El código se desarrolla mediante diálogo. No realices `git push` o `PR` de forma proactiva al terminar de escribir código.
* **Ejecución de Cierre:** Ejecuta `git commit`, `git push` y `gh pr create` **solo bajo petición explícita** o cuando se acuerde un cierre autónomo en la sesión.
* **Vinculación en PR:** La descripción del PR debe incluir `Closes #ID` o `Refs #ID` para vincular la actividad.

## 5. Estandares de Git y Commits
* **Idioma:** Mensajes de commit y descripciones siempre en **inglés**.
* **Formato:** Usa *Conventional Commits* (`feat:`, `fix:`, `refactor:`, `chore:`, `docs:`).
* **Integridad:** Prohibido usar `git commit --amend` o reescribir historial en ramas remotas salvo petición expresa.

## 6. Sincronizacion del Tablero Project
* Actualiza el campo **Status** del Project de forma síncrona con el estado real del PR y la Issue. Los valores del tablero son:
  * **Backlog** — This item hasn't been started.
  * **Ready** — This is ready to be picked up.
  * **In progress** — This is actively being worked on.
  * **In review** — This item is in review.
  * **Done** — This has been completed.
* Flujo típico: `Backlog` → `Ready` → `In progress` → `In review` → `Done` (ajusta según el punto en el que esté el trabajo).

## 7. Troubleshooting de Tooling
* Si detectas errores de permisos o scopes de GitHub CLI (`project`, `read:project`), solicita al usuario ejecutar `gh auth refresh` o el comando de autenticación correspondiente antes de reintentar.
