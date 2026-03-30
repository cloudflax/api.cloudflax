## Conventions | Cloudflax - Backend

Naming · carpetas · errores/HTTP · handler/service/repository · DTOs · tests · estilo · docs de módulo · JSON · status.

## CRUD naming

Patrón global: `{Action}{Resource}`. **Resource** singular (`User`, `Order`). **Action:** `List`, `Get`, `Create`, `Update`, `Delete` → ej. `ListUser`, `GetUser`, … Aplica en `handler.go`, `service.go`, `repository.go` del feature.

## Carpetas y archivos

- **Feature:** `internal/{recurso}/` — singular, lowercase (`user`, `product`).
- **Shared:** `internal/shared/{módulo}/`.
- **Archivos:** singular, lowercase, sufijo por rol.

| Archivo | Rol |
|---------|-----|
| `model.go` | Dominio GORM, tabla |
| `repository.go` | `NewRepository(db)`; CRUD; sentinels del paquete |
| `service.go` | `NewService(repo)`; mismos nombres CRUD que handler |
| `handler.go` | Bind, validate, service, `runtimeerror` |
| `routes.go` | `Routes(router, handler, authMiddleware)` |
| `dto.go` | `CreateXRequest`, `UpdateXRequest`; tags `json`, `validate` |
| `errors.go` | `var ErrX = fmt.Errorf(...)` para repo/service; handler mapea a HTTP |

Tests: `model_test.go` si hay lógica; `handler_test.go` éxito + errores. Service/repo tests si aplica.

**Bootstrap:** `internal/bootstrap/server/routes.go` — repo → service (opciones) → handler → `{recurso}.Routes(app, handler, authMiddleware)`.

## Errores y HTTP

- **Sentinels** en `errors.go`; repo/service los devuelven (no GORM crudo).
- **Repo:** `gorm.ErrRecordNotFound` → `ErrNotFound`; resto `fmt.Errorf("op: %w", err)`.
- **Handler:** `errors.Is` / `errors.As`; `runtimeerror.Respond` / `RespondWithDetails`. Validación: helper tipo `toErrorDetails(validator.ValidationErrors)` → `[]runtimeError.ErrorDetail`.
- **Códigos de recurso** en `internal/shared/runtimeerror/runtimeerror.go` (ej. `CodeUserNotFound`). Inesperados: `slog` + `CodeInternalServerError`.

## Handler (orden)

1. Auth si aplica (`requestctx.UserOnly`).
2. `ctx.Bind().Body(&req)`.
3. `validator.Validate(req)` → `RespondWithDetails` si hay detalles por campo.
4. Llamar service.
5. Mapear error → status/código.
6. Éxito: `ctx.JSON(fiber.Map{"data": ...})` o `201` / `204` según operación.

## Service

- Validar IDs (ej. `uuid.Parse`); inválido → mismo error que no encontrado (`ErrNotFound`).
- Dependencias externas por **interfaz**; opcionales: `if dep != nil`; builder `WithTokenRevoker(tr) *Service` si hace falta.

## DTOs y constructores

- DTOs: `Create{Resource}Request`, `Update{Resource}Request` (o `UpdateMeRequest`). Updates: punteros para opcionales.
- `NewRepository(db)`, `NewService(repo)`, `NewHandler(service)`. Opcionales vía métodos `*Service`, no firma gigante.

## Tests de handler

- Parámetro **`test *testing.T`** (nunca `t`) en tests y helpers.
- `Setup{Resource}HandlerTest(test)` → `database.InitForTesting()`, migraciones del modelo, repo → service → handler.
- `DecodeErrorResponse(test, body)` → `runtimeError.ErrorResponse`; assert `errorResponse.Error.Code`.
- Por endpoint: al menos un éxito + errores relevantes (401, 404, 422, 409, …).

## Estilo

- Imports: std → third → internal; líneas en blanco entre grupos.
- Alias camelCase para paquetes de una palabra ambigua: `runtimeError "…/runtimeerror"`.
- Variables camelCase; exportados PascalCase.
- Código y comentarios en **inglés**.
- Nombres explícitos: `userRepository`, no `userRepo`; params `repository`, `service`, `handler`, `request`, `response`. OK: `err`, `ctx`, `id`, `req`/`resp` HTTP. Receivers descriptivos si hay varios.

## Documentación del módulo

1. **Godoc:** comentario antes de `package` empezando por `Package <name> …`.
2. **`README.md`** en el feature: arquitectura, modelo/datos, errores + HTTP, middlewares/validaciones/integraciones.

## JSON API

Códigos y helpers: `internal/shared/runtimeerror/runtimeerror.go`; `runtimeerror.Respond` y `runtimeerror.RespondWithDetails` construyen la forma de error.

**Éxito:**

```json
{
  "data": { ... },
  "message": "opcional"
}
```

**Error (único):**

```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "user not found",
    "status": 404,
    "trace_id": "abc-123"
  }
}
```

**Error (validación, varios campos):**

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "status": 422,
    "trace_id": "abc-123",
    "details": [
      { "field": "email", "message": "must be a valid email address" },
      { "field": "password", "message": "must be at least 8 characters" }
    ]
  }
}
```

**Lista paginada:** mismo contenedor de éxito; añadir `meta` con `page`, `limit`, `total`.

```json
{
  "data": [ ... ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 100
  }
}
```

## Status HTTP

| Operación | OK | Errores típicos |
|-----------|-----|-----------------|
| List | 200 | 400 query |
| Get | 200 | 404 |
| Create | 201 | 400, 409 duplicado |
| Update | 200 | 400, 404 |
| Delete | 200 o 204 | 404 |
| Login | 200 | 401 |
