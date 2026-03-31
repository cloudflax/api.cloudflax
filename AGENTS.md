## Agents | Cloudflax - Backend

Instrucciones para quien ejecuta tareas en este repo (agente). Prioriza la lista de obligaciones; la tabla enlaza el detalle.

Abre y lee el contenido de una **referencia solo cuando aplique** a la tarea (p. ej. auth → `AUTH_INTEGRATION`, GitHub → `GITHUB_WORKFLOW`). No recorras ni cargues documentación que no necesites.

### Acción → referencia

| Acción | Referencia |
|--------|------------|
| Entorno, `make`, árbol del repo | [README.md](./README.md) |
| Stack, capas, middleware, migraciones | [ARCHITECTURE.md](./ARCHITECTURE.md) |
| API, errores, JSON, tests | [CONVENTIONS.md](./CONVENTIONS.md) |
| Issues, ramas, Git Flow, commits | [docs/GITHUB_WORKFLOW.md](./docs/GITHUB_WORKFLOW.md) |
| Nivel y foco de tus respuestas | [SKILLS.md](./SKILLS.md) |
| Contrato auth (JWT, cookies) | [AUTH_INTEGRATION.md](./AUTH_INTEGRATION.md) |
| Cuentas y titularidad de datos | [docs/ACCOUNTS_AND_DATA_OWNERSHIP.md](./docs/ACCOUNTS_AND_DATA_OWNERSHIP.md) |
| Comentarios Go (sin En/Es en doc de `package`; En/Es solo en declaraciones, no en sentencias internas) | [`.cursor/rules/go-bilingual-comments.mdc`](./.cursor/rules/go-bilingual-comments.mdc) (siempre activo en Cursor) |

### Obligaciones

1. Ejecuta **`make lint`** y **`make test`** antes de considerar el trabajo terminado.
2. Si tocas el modelo GORM: registra la migración en **`database.RunMigrations()`** (`cmd/api/main.go`); el flujo está en [ARCHITECTURE.md](./ARCHITECTURE.md).
3. GitHub: sigue [docs/GITHUB_WORKFLOW.md](./docs/GITHUB_WORKFLOW.md) (Git Flow y matriz al inicio). No hagas `push` salvo petición explícita; integración por merges entre ramas según ese doc, no PR por defecto; commits en inglés, Conventional, con `Refs`/`Closes` cuando aplique.
4. Usa **`slog`** estructurado; no escribas passwords, tokens ni PII en logs.
5. Siempre que crees código, documenta según [`.cursor/rules/go-bilingual-comments.mdc`](./.cursor/rules/go-bilingual-comments.mdc).

### Si no abres otro doc

Mantén el **código** en inglés (identificadores, mensajes de error expuestos por la API, etc.); encadena errores con `fmt.Errorf("...: %w", err)`; secretos solo vía entorno; listados paginados y códigos HTTP según [CONVENTIONS.md](./CONVENTIONS.md).

### Cómo respondes al humano

Redacta en **español**, de forma concisa. Menciona explícitamente si el cambio afecta **`.env`** (o configuración por entorno) o exige **migración** de base de datos.

**Stack:** Go 1.25, Fiber v3, GORM, PostgreSQL, `slog` — ampliar en [ARCHITECTURE.md](./ARCHITECTURE.md).
