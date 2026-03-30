# Agent Rules | Cloudflax - Backend

[AGENTS.md](../AGENTS.md) apunta aquí para obligaciones; abre otros docs solo si la tarea lo exige.

**Antes de cerrar:** `make lint` y `make test`.

**GORM:** modelo nuevo o modificado → registrar el tipo en `database.RunMigrations(...)` (`cmd/api/main.go`). Detalle: [ARCHITECTURE.md](../ARCHITECTURE.md).

**GitHub:** [GITHUB_WORKFLOW.md](./GITHUB_WORKFLOW.md). Sin `push` ni PR salvo petición explícita. Commits en inglés, Conventional Commits; `Refs` / `Closes` cuando toque.

**Logs:** `slog` estructurado; nunca contraseñas, tokens ni PII.

**Comentarios Go:** [`.cursor/rules/go-bilingual-comments.mdc`](../.cursor/rules/go-bilingual-comments.mdc).

**Código (siempre):** inglés en identificadores y errores expuestos por la API; encadenar con `fmt.Errorf("…: %w", err)`; secretos solo vía env. Paginación y códigos HTTP: [CONVENTIONS.md](../CONVENTIONS.md).

**Respuesta al humano:** español, conciso. Indicar si el cambio afecta **`.env`** (u otra config por entorno) o exige **migración** de base de datos.

**Stack:** Go 1.25, Fiber v3, GORM, PostgreSQL, `slog` — [ARCHITECTURE.md](../ARCHITECTURE.md).
