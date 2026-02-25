# Convenciones de código — Cloudflax API

Reglas para mantener consistencia. En este archivo van:

- **Naming** — Cómo nombrar funciones, variables, archivos y recursos.
- **Estructura de carpetas** — Organización, archivos estándar por módulo y registro en bootstrap.
- **Errores de dominio** — Sentinel errors por módulo y mapeo a respuestas HTTP.
- **Handler, Repository, Service** — Flujo del handler, patrones de repository/service, DTOs y constructores.
- **Tests** — Setup de handler tests, decodificación de errores y casos a cubrir.
- **Estilo de código** — Formato, imports, comentarios y buenas prácticas de sintaxis.
- **Documentación de módulos** — Comentario de paquete y README por feature.
- **Formato de API** — Respuestas JSON y códigos de estado por operación.

---

## Nombres CRUD

**Todas las acciones CRUD usan el mismo patrón en todos los features y en todas las capas:**

```
{Action}{Resource}
```

- **Resource** siempre en singular: `User`, `Country`, `Order`
- **Action** de CRUD: `List`, `Get`, `Create`, `Update`, `Delete`

| Acción | Ejemplo User | Ejemplo Country | Ejemplo Order |
|--------|--------------|-----------------|---------------|
| Listar | `ListUser` | `ListCountry` | `ListOrder` |
| Obtener uno | `GetUser` | `GetCountry` | `GetOrder` |
| Crear | `CreateUser` | `CreateCountry` | `CreateOrder` |
| Actualizar | `UpdateUser` | `UpdateCountry` | `UpdateOrder` |
| Eliminar | `DeleteUser` | `DeleteCountry` | `DeleteOrder` |

**Aplica en:** `handler.go`, `service.go`, `repository.go` y cualquier archivo del feature.

---

## Estructura de carpetas

- **Features:** `internal/{recurso}/` — nombre en singular, lowercase (ej: `user`, `product`).
- **Shared:** `internal/shared/{módulo}/` — código reutilizable entre features.
- **Archivos:** Nombres en singular, lowercase, con sufijo por tipo (`handler.go`, `service.go`, `repository.go`).

### Archivos estándar por módulo

Cada módulo en `internal/{recurso}/` debe incluir al menos los archivos que use según responsabilidad:

| Archivo | Uso |
|---------|-----|
| `model.go` | Modelo de dominio (GORM), tabla, métodos del modelo |
| `repository.go` | Acceso a datos: `NewRepository(db)`, Get/Create/Update/Delete |
| `service.go` | Lógica de negocio: `NewService(repository)`, mismos nombres CRUD que handler |
| `handler.go` | HTTP: bind, validación, llamada al service, respuesta con `runtimeerror` |
| `routes.go` | Función `Routes(router, handler, authMiddleware)` que monta los endpoints |
| `dto.go` | Request/Response DTOs (ej. `CreateUserRequest`, `UpdateMeRequest`) con tags `json` y `validate` |
| `errors.go` | Errores sentinel del dominio (ej. `ErrNotFound`, `ErrDuplicateEmail`) para que service/repository los devuelvan y el handler los mapee a códigos HTTP |

Tests: `model_test.go` cuando el modelo tenga lógica a probar; `handler_test.go` con casos de éxito y de error. Opcionalmente tests de service o repository si la lógica lo justifica.

Registro del módulo: en `internal/bootstrap/server/routes.go` crear repository → service (con opciones si aplica) → handler y llamar `{recurso}.Routes(app, handler, authMiddleware)`.

---

## Errores de dominio y mapeo HTTP

- **Por módulo:** Definir en `errors.go` errores sentinel con `var ErrX = fmt.Errorf("...")` (ej. `ErrNotFound`, `ErrDuplicateEmail`). El **repository** y el **service** devuelven estos errores.
- **Handler:** Mapear con `errors.Is(err, ErrNotFound)` (o `errors.As` para validación) y responder con `runtimeerror.Respond` / `runtimeerror.RespondWithDetails` usando los códigos definidos en `internal/shared/runtimeerror/runtimeerror.go`. Para validación, usar un helper privado tipo `toErrorDetails(validator.ValidationErrors)` que devuelva `[]runtimeError.ErrorDetail`.
- **Códigos por recurso:** Los códigos específicos del recurso (ej. `CodeUserNotFound`, `CodeEmailAlreadyExists`) se definen en el paquete `runtimeerror`; el handler los usa al responder. Errores inesperados se loguean con `slog` y se responde con `CodeInternalServerError`.

---

## Flujo del handler

En cada endpoint: (1) extraer contexto de auth si aplica (`requestctx.UserOnly`), (2) bind del body con `ctx.Bind().Body(&req)`, (3) validar con `validator.Validate(req)` y en caso de error usar `RespondWithDetails` si hay detalles por campo, (4) llamar al service, (5) según el error devuelto (`errors.Is`/`errors.As`) responder con el status y código adecuados, (6) en éxito devolver `ctx.JSON(fiber.Map{"data": ...})` o `ctx.Status(201).JSON(...)` / `204` según la operación.

---

## Repository y Service

- **Repository:** Traducir `gorm.ErrRecordNotFound` a los errores del paquete (ej. `ErrNotFound`). Envolver demás errores con `fmt.Errorf("operación: %w", err)`. No devolver errores de GORM crudos al service.
- **Service:** Validar identificadores (ej. UUID con `uuid.Parse`) antes de llamar al repository; si el formato es inválido, devolver el mismo error que “no encontrado” (ej. `ErrNotFound`) para no revelar información. Dependencias externas (ej. revocador de tokens) inyectadas por **interfaz** y opcionales: comprobar `if service.dependency != nil` antes de usarlas; se pueden exponer con métodos tipo `WithTokenRevoker(tr) *Service` para mantener el constructor simple.

---

## DTOs y constructores

- **DTOs:** Nombres `Create{Resource}Request`, `Update{Resource}Request` (o `UpdateMeRequest` cuando aplique). Usar tags `json` y `validate`; en updates usar punteros para campos opcionales.
- **Constructores:** `NewRepository(db *gorm.DB)`, `NewService(repository *Repository)`, `NewHandler(service *Service)`. Dependencias opcionales del service mediante métodos que devuelven `*Service` (builder style) para no complicar la firma del constructor.

---

## Tests de handler

- **Parámetro de test:** En todas las funciones de test y helpers de test usar siempre `test *testing.T` (no `t *testing.T`). Ejemplo: `func SetupUserHandlerTest(test *testing.T) *Handler`, `func TestGetMe(test *testing.T)`.
- **Setup:** Un helper `Setup{Resource}HandlerTest(test *testing.T)` que inicialice DB para tests (`database.InitForTesting()`), ejecute migraciones del modelo del módulo, cree repository → service → handler y devuelva el handler.
- **Errores:** Helper `DecodeErrorResponse(test, body)` que decodifique el body en `runtimeError.ErrorResponse` para afirmar `errorResponse.Error.Code`.
- **Casos:** Por endpoint, incluir al menos un test de éxito y tests por tipo de error (401, 404, 422, 409, etc.) según aplique.

---

## Estilo de código

- **Imports:** Ordenar: estándar → terceros → internal. Agrupar con líneas en blanco.
- **Alias de import:** Para paquetes cuyo nombre es una sola palabra en minúsculas que se lee como compuesta (ej. `runtimeerror`), usar alias en camelCase para mejorar legibilidad: `runtimeError "github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"`.
- **Variables:** camelCase. Constantes y tipos exportados: PascalCase.
- **Comentarios:** Siempre en inglés. Todo el código fuente debe estar íntegramente en inglés.

### Nombres descriptivos (Clean Code)

Evitar abreviaciones y nombres cortos que obligan al lector a “traducir” mentalmente. Preferir nombres que revelen la intención.

- **No abreviar:** Preferir `userRepository`, `userService`, `userHandler` en lugar de `userRepo`, `userSvc`, `userHnd`.
- **En parámetros y variables:** Usar nombres completos como `repository`, `service`, `handler`, `user`, `request`, `response`.
- **Excepciones ampliamente conocidas:** `err` (error), `ctx` (context), `id`, `req`/`resp` en contexto HTTP — son convenciones estándar y legibles.
- **Receivers:** Go acepta receivers de 1–2 letras; usar nombres descriptivos cuando mejoren la claridad (ej. `repository` o `svc` en lugar de `r` si hay varios receivers).

---

## Documentación de módulos

Cada módulo en `internal/{recurso}/` debe documentarse de dos formas:

1. **Comentario de paquete (godoc):** En al menos un archivo `.go` del paquete, incluir un comentario inmediatamente antes de `package ...` que empiece por `Package <nombre>`. Es el texto que muestra `go doc` y pkg.go.dev. Ejemplo:
   ```go
   // Package user provides user identity management: profile CRUD, password
   // hashing, soft delete, and session revocation.
   package user
   ```

2. **README.md:** En la carpeta del módulo, un `README.md` con arquitectura del feature, modelo de datos (tablas/campos), diccionario de errores y códigos HTTP, y consideraciones técnicas (middlewares, validaciones, integraciones). Sirve como documentación para el equipo y onboarding.

Los comentarios de tipos y funciones exportados siguen las reglas de **Estilo de código** (en inglés).

---

## Formato de respuesta JSON

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

**Error (múltiples campos — validación):**
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

Los códigos de error se definen en `internal/shared/runtimeerror/runtimeerror.go`. Los helpers `runtimeerror.Respond` y `runtimeerror.RespondWithDetails` construyen siempre esta estructura.

Para listas paginadas, incluir `meta` con `page`, `limit`, `total`.

---

## Status codes por operación

| Operación | Éxito | Errores frecuentes |
|-----------|-------|---------------------|
| List | 200 | 400 (query inválida) |
| Get by ID | 200 | 404 (no encontrado) |
| Create | 201 | 400 (validación), 409 (duplicado) |
| Update | 200 | 400 (validación), 404 (no encontrado) |
| Delete | 200 o 204 | 404 (no encontrado) |
| Login | 200 | 401 (credenciales inválidas) |
