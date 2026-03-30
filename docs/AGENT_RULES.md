# Reglas del agente | Cloudflax Backend

Normas que aplican a quien implementa en este repo (Cursor u otro agente). El resumen enlaza desde [AGENTS.md](../AGENTS.md).

## Obligaciones

1. **Calidad antes de cerrar**: ejecuta `make lint` y `make test` antes de considerar el trabajo terminado.
2. **Modelos GORM**: si cambias o añades un modelo, incluye su tipo en `database.RunMigrations(...)` en `cmd/api/main.go` (AutoMigrate vía `internal/shared/database`). Flujo: [ARCHITECTURE.md](../ARCHITECTURE.md).
3. **GitHub**: sigue [GITHUB_WORKFLOW.md](./GITHUB_WORKFLOW.md). No hagas `push` ni abras PR salvo petición explícita. Commits en inglés, Conventional Commits, con `Refs` / `Closes` cuando aplique (detalle allí).
4. **Logs**: usa `slog` estructurado; no registres contraseñas, tokens ni PII.
5. **Código nuevo**: documenta según [`.cursor/rules/go-bilingual-comments.mdc`](../.cursor/rules/go-bilingual-comments.mdc).

## Si no consultas otro documento

Mantén el **código** en inglés (identificadores, mensajes de error expuestos por la API, etc.). Encadena errores con `fmt.Errorf("...: %w", err)`. Secretos solo vía variables de entorno. Listados paginados y códigos HTTP: [CONVENTIONS.md](../CONVENTIONS.md).

## Respuesta al humano

Redacta en **español**, de forma concisa. Indica explícitamente si el cambio afecta **`.env`** (u otra configuración por entorno) o si exige **migración** de base de datos.

## Stack (recordatorio)

Go 1.25, Fiber v3, GORM, PostgreSQL, `slog`. Detalle: [ARCHITECTURE.md](../ARCHITECTURE.md).
