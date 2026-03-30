# Convenciones | Cloudflax - Backend

Naming, carpetas, errores, handler/repository/service, DTOs, tests, estilo, docs de módulo, JSON y status HTTP.

## Nombres CRUD

Patrón `{Action}{Resource}` en handler, service, repository. **Resource** singular (`User`, `Country`). **Action:** `List`, `Get`, `Create`, `Update`, `Delete` (ej. `ListUser`, `GetCountry`, `DeleteOrder`).

## Estructura

- **Features:** `internal/{recurso}/` — singular, lowercase.
- **Shared:** `internal/shared/{módulo}/`.
- **Archivos:** `handler.go`, `service.go`, `repository.go`, `routes.go`, `dto.go`, `errors.go`, `model.go` según uso.

| Archivo | Rol |
|---------|-----|
| `model.go` | Dominio GORM |
| `repository.go` | `NewRepository(db)`, CRUD; traducir `gorm.ErrRecordNotFound` → errores del paquete; `fmt.Errorf("op: %w", err)` para el resto |
| `service.go` | `NewService(repo)`, mismos nombres CRUD; validar UUIDs; `ErrNotFound` si formato inválido; deps opcionales por interfaz + `if dep != nil` o `WithX()` |
| `handler.go` | Bind, `validator.Validate`, service, `runtimeerror` |
| `routes.go` | `Routes(router, handler, authMiddleware)` |
| `dto.go` | `Create/Update{Resource}Request`, tags `json`/`validate`, punteros en updates opcionales |
| `errors.go` | `var ErrX = fmt.Errorf(...)`; repo/service los devuelven; handler mapea con `errors.Is`/`errors.As` |

Tests: `handler_test.go` (éxito + errores); `model_test.go` si hay lógica. **Bootstrap:** `internal/bootstrap/server/routes.go` → repo → service → handler → `{recurso}.Routes(...)`.

## Errores HTTP

Handler: `runtimeerror.Respond` / `RespondWithDetails`; códigos de recurso en `runtimeerror` (ej. `CodeUserNotFound`). Validación: helper tipo `toErrorDetails(validator.ValidationErrors)`. Inesperados: `slog` + `CodeInternalServerError`.

## Flujo handler

Auth (`requestctx.UserOnly` si aplica) → `ctx.Bind().Body(&req)` → `validator.Validate` → service → `errors.Is`/`errors.As` → JSON éxito `{"data": ...}` o 201/204 según operación.

## DTOs / constructores

`NewRepository(db)`, `NewService(repository)`, `NewHandler(service)`. Opcionales en service vía métodos que devuelven `*Service`.

## Tests handler

Parámetro **`test *testing.T`** (no `t`). `Setup{Resource}HandlerTest(test)` → `database.InitForTesting()`, migraciones, repo → service → handler. `DecodeErrorResponse(test, body)` → `ErrorResponse` y `errorResponse.Error.Code`. Por endpoint: al menos un éxito y errores (401, 404, 422, 409, …) según aplique.

## Estilo

Imports: std → third → internal, líneas en blanco. Alias camelCase para paquetes de una palabra confusa (ej. `runtimeError "…/runtimeerror"`). camelCase / PascalCase según exportación. **Código y comentarios en inglés.**

Nombres: `userRepository` no `userRepo`; `repository`, `service`, `handler` en vars; excepciones: `err`, `ctx`, `id`, `req`/`resp`. Receivers descriptivos si hay ambigüedad.

## Docs módulo

1. Comentario antes de `package`: `// Package user provides ...` (godoc).
2. `README.md` en el feature: modelo, errores, HTTP, notas técnicas.

## JSON

Éxito: `{"data": {...}, "message": "opcional"}`. Error: `error` con `code`, `message`, `status`, `trace_id`; validación añade `details[]` (`field`, `message`). Paginación: `meta` (`page`, `limit`, `total`). Helpers en `runtimeerror`.

## Status

| Operación | OK | Errores típicos |
|-----------|-----|-----------------|
| List | 200 | 400 |
| Get | 200 | 404 |
| Create | 201 | 400, 409 |
| Update | 200 | 400, 404 |
| Delete | 200 o 204 | 404 |
| Login | 200 | 401 |
