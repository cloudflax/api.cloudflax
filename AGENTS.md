## Agents | Cloudflax - Backend

Instrucciones para quien ejecuta tareas en este repo (agente). Prioriza la lista de obligaciones; la tabla enlaza el detalle.

Abre y lee el contenido de una **referencia solo cuando aplique** a la tarea (p. ej. auth → `AUTH_INTEGRATION`, GitHub → `GITHUB_WORKFLOW`). No recorras ni cargues documentación que no necesites.

### Acción → referencia

| Acción | Referencia |
|--------|------------|
| API, errores, JSON, tests | [CONVENTIONS.md](./CONVENTIONS.md) |
| Stack, capas, middleware, migraciones | [ARCHITECTURE.md](./ARCHITECTURE.md) |
| Entorno, `make`, árbol del repo | [README.md](./README.md) |
| Nivel y foco de tus respuestas | [SKILLS.md](./SKILLS.md) |
| Contrato auth (JWT, cookies) | [AUTH_INTEGRATION.md](./AUTH_INTEGRATION.md) |
| Cuentas y titularidad de datos | [docs/ACCOUNTS_AND_DATA_OWNERSHIP.md](./docs/ACCOUNTS_AND_DATA_OWNERSHIP.md) |
| Una feature concreta | `internal/{feature}/README.md` |
| Issues, ramas, PR, commits | [docs/GITHUB_WORKFLOW.md](./docs/GITHUB_WORKFLOW.md) |

### Obligaciones

1. Ejecuta **`make lint`** y **`make test`** antes de considerar el trabajo terminado.
2. Si tocas el modelo GORM: registra la migración en **`database.RunMigrations()`** (`cmd/api/main.go`); el flujo está en [ARCHITECTURE.md](./ARCHITECTURE.md).
3. GitHub: sigue [docs/GITHUB_WORKFLOW.md](./docs/GITHUB_WORKFLOW.md) (matriz al inicio). No hagas `push` ni abras PR salvo petición explícita; los commits, en inglés, Conventional, con `Refs`/`Closes` cuando aplique.
4. Usa **`slog`** estructurado; no escribas passwords, tokens ni PII en logs.

### Si no abres otro doc

Mantén el código en inglés; encadena errores con `fmt.Errorf("...: %w", err)`; secretos solo vía entorno; listados paginados y códigos HTTP según [CONVENTIONS.md](./CONVENTIONS.md).

### Cómo respondes al humano

Redacta en **español**, de forma concisa. Menciona explícitamente si el cambio afecta **`.env`** (o configuración por entorno) o exige **migración** de base de datos.

**Stack:** Go 1.25, Fiber v3, GORM, PostgreSQL, `slog` — ampliar en [ARCHITECTURE.md](./ARCHITECTURE.md).
