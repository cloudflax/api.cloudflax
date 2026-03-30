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
| Issue, project `@api.cloudflax`, matriz de trazabilidad (issue+rama / solo issue / **solo commits** en `feature/<ID>-…` con `Refs`/`Closes #ID`), PRs y **cadencia en equipo** | [docs/GITHUB_WORKFLOW.md](./docs/GITHUB_WORKFLOW.md) |

---

### 2. Checklist antes de proponer cambios

- Ejecuta **`make lint`** y **`make test`**; no dejes el pipeline roto.
- Cambio de modelo GORM: registra la migración en **`database.RunMigrations()`** (`cmd/api/main.go`). Ver flujo en [ARCHITECTURE.md](./ARCHITECTURE.md).
- Trabajo **trazable:** [docs/GITHUB_WORKFLOW.md](./docs/GITHUB_WORKFLOW.md): pregunta al inicio la **matriz** (issue+rama, solo issue, o solo commits en rama ya existente `feature/<ID>-…` vinculando `#ID` en el mensaje). Si hace falta issue nuevo: **issue primero y rama después**. **No** hagas `push` ni PR hasta petición explícita (el revisor puede iterar en el código contigo); los **commits** cuando lo pidas o encaje el escenario acordado (inglés, Conventional Commits, `Refs`/`Closes` al `#ID` de la rama si aplica).
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

El detalle operativo (comandos `gh`, nombres de rama, PR, tablero), la **matriz de inicio** (incluida la vía *solo commits* con `#ID` tomado de `feature/<ID>-…`) y la **cadencia colaborativa** (confirmación, orden issue→rama, commits/`push`/PR cuando el usuario lo pida) están en **[docs/GITHUB_WORKFLOW.md](./docs/GITHUB_WORKFLOW.md)**.

---

### 5. Flujo de comunicación con el usuario

- Explica los cambios en **español**, de forma concisa.
- Indica si el cambio toca **`.env`** o requiere **migración** de base de datos.

---

### 6. Stack (recordatorio)

Go 1.25, Fiber v3, GORM + PostgreSQL, observabilidad con **slog**. Detalle en [ARCHITECTURE.md](./ARCHITECTURE.md) y [README.md](./README.md).
