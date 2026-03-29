## Agents — Cloudflax API (Backend)

Este fichero es el **punto de entrada** para agentes de Cursor: prioridades operativas, checklist y enlaces al detalle. La normativa larga vive en los documentos enlazados; evita duplicar aquí lo que ya está cubierto allí.

---

### 1. Mapa de documentación (léelo según la tarea)

| Necesitas… | Documento |
|------------|-----------|
| Naming CRUD, estructura `internal/{feature}/`, handler → service → repository, errores de dominio, formato JSON de API, tests de handler | [CONVENTIONS.md](./CONVENTIONS.md) |
| Stack, capas, flujo de datos, middleware (logger, request ID, auth), workflow de implementación y migraciones | [ARCHITECTURE.md](./ARCHITECTURE.md) |
| Descripción del producto, árbol del repo, entorno, `make` y variables | [README.md](./README.md) |
| Nivel técnico esperado del agente y foco de respuestas | [SKILLS.md](./SKILLS.md) |
| Contrato de auth para frontends (JWT, refresh, cookies) | [AUTH_INTEGRATION.md](./AUTH_INTEGRATION.md) |
| Cuentas y titularidad de datos | [docs/ACCOUNTS_AND_DATA_OWNERSHIP.md](./docs/ACCOUNTS_AND_DATA_OWNERSHIP.md) |
| Detalle por feature (modelo, errores HTTP, notas) | `internal/{feature}/README.md` (p. ej. [internal/auth/README.md](./internal/auth/README.md), [internal/user/README.md](./internal/user/README.md)) |

---

### 2. Checklist antes de proponer cambios

- Ejecuta **`make lint`** y **`make test`**; no dejes el pipeline roto.
- Cambio de modelo GORM: registra la migración en **`database.RunMigrations()`** (`cmd/api/main.go`). Ver flujo en [ARCHITECTURE.md](./ARCHITECTURE.md).
- Trabajo **trazable** (feature, cambio de comportamiento, refactor relevante): issue en GitHub + rama asociada (ver § 4).
- **Logs:** `slog` estructurado; no registrar passwords, tokens ni PII. Niveles y correlación: ver capa de middleware en [ARCHITECTURE.md](./ARCHITECTURE.md) y respuestas de error en [CONVENTIONS.md](./CONVENTIONS.md).

---

### 3. Reglas compactas (si no abres otro doc)

- **Idioma en código:** inglés (fuente, comentarios, mensajes de error y logs).
- **Errores:** `fmt.Errorf("...: %w", err)`; no ignorar errores.
- **Listados:** paginación por defecto en `List{Resource}`; forma de respuesta y `meta` en [CONVENTIONS.md](./CONVENTIONS.md).
- **Secretos:** solo variables de entorno o sistemas seguros; nunca en el repositorio.
- **HTTP:** validación en handler/service; mapa de códigos (`400` / `401` / `403` / `404` / `409` / `500`) alineado con [CONVENTIONS.md](./CONVENTIONS.md).

---

### 4. GitHub: issue, proyecto `@api.cloudflax`, rama y PR

Para trabajo **trazable**, usa este flujo (no hace falta para ediciones mínimas acordadas).

1. **Issue:** Crea la tarea en `cloudflax/api.cloudflax` con objetivos y criterios de aceptación. Enlázala al project de organización **`@api.cloudflax`** con:

   `gh issue create -R cloudflax/api.cloudflax -p "@api.cloudflax"`

   (El nombre del project debe coincidir **exactamente** con el de GitHub.)

2. **Rama:** `feature/<número-de-issue>-<slug-corto-en-kebab>` (ej. `feature/3-agents-md-hub`).

3. **Asociación issue ↔ rama:** Desarrolla en esa rama; en el **PR** incluye `Closes #N` o `Refs #N` en la descripción para vincular la issue.

4. **Tablero:** Actualiza **Status** y demás campos del project (p. ej. *Ready* → *In review* → *Done*).

5. **Commits:** Mensajes en **inglés**; Conventional Commits cuando encaje (`feat`, `fix`, …). No reescribas historia en remoto con `--amend` salvo petición explícita.

6. **`gh`:** Hacen falta scopes adecuados (`project`, `read:project`). Si falta permiso o login interactivo, el usuario debe ejecutar `gh auth refresh` (u otro paso que indique `gh`).

**Rama dedicada:** Ábrela cuando aporte (varios commits, revisión aislada, riesgo en `main`, trabajo en paralelo). No fuerces rama para cada microcambio si el equipo acuerda lo contrario.

---

### 5. Flujo de comunicación con el usuario

- Explica los cambios en **español**, de forma concisa.
- Indica si el cambio toca **`.env`** o requiere **migración** de base de datos.

---

### 6. Stack (recordatorio)

Go 1.25, Fiber v3, GORM + PostgreSQL, observabilidad con **slog**. Detalle en [ARCHITECTURE.md](./ARCHITECTURE.md) y [README.md](./README.md).
