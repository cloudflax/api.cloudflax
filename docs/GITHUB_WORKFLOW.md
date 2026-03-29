# GitHub — Issue, proyecto `@api.cloudflax`, rama y PR

Directrices para trabajo **trazable** en `cloudflax/api.cloudflax` (features, cambios de comportamiento, refactors relevantes). No hace falta seguir este flujo para ediciones mínimas acordadas.

---

## 1. Issue

Crea la tarea en `cloudflax/api.cloudflax` con objetivos y criterios de aceptación. Enlázala al project de organización **`@api.cloudflax`** con:

```bash
gh issue create -R cloudflax/api.cloudflax -p "@api.cloudflax"
```

El nombre del project debe coincidir **exactamente** con el de GitHub.

---

## 2. Rama

Formato: `feature/<número-de-issue>-<slug-corto-en-kebab>` (ej. `feature/3-agents-md-hub`).

**Rama dedicada:** Ábrela cuando aporte (varios commits, revisión aislada, riesgo en `main`, trabajo en paralelo). No fuerces rama para cada microcambio si el equipo acuerda lo contrario.

---

## 3. Asociación issue ↔ rama ↔ PR

Desarrolla en la rama vinculada a la issue. En el **PR**, incluye `Closes #N` o `Refs #N` en la descripción para vincular la issue.

Opcionalmente deja un comentario en la issue indicando el nombre de la rama.

---

## 4. Tablero (project)

Actualiza **Status** y demás campos del project (p. ej. *Ready* → *In review* → *Done*) según el estado real del trabajo.

---

## 5. Commits

- Mensajes en **inglés**.
- Conventional Commits cuando encaje (`feat`, `fix`, …).
- No reescribas historia en remoto con `--amend` salvo petición explícita.

---

## 6. CLI `gh` y permisos

Hacen falta scopes adecuados (`project`, `read:project`). Si falta permiso o login interactivo, ejecuta `gh auth refresh` (u otro paso que indique `gh`).
