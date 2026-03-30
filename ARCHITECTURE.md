# Architecture — Cloudflax API (Backend)

Estructura, patrones y flujo de datos del backend.

## Stack

Go (feature-driven), Fiber v3, GORM + PostgreSQL; tests unitarios e integración (PostgreSQL/SQLite).

## Directorios (`internal/`)

- **`bootstrap/`**: Config, servidor, arranque.
- **`{feature}/`**: `handler`, `service`, `repository`, `model`, `dto`, `routes`; opcional `validator` u helpers.
- **`shared/`**: `database` (conexión, migraciones), `middleware`, `pagination`, `filtering`, `errors`, `validator`, `verificationnotify` (Lambda correo verificación), utilidades.

No todo feature necesita todos los archivos; lo mínimo suele ser `handler`, `repository`, `model`, `routes`.

## Capas

**Flujo:** `Request → Middleware → Handler → Service → Repository → DB`.

| Capa | Rol |
|------|-----|
| **Handler** | HTTP (Fiber): parseo y respuestas. Sin negocio ni DB directa; delega al Service. |
| **Service** | Negocio y orquestación. Sin Fiber ni status HTTP; modelos y errores de dominio. |
| **Repository** | GORM. Sin respuestas HTTP. |
| **Validación** | DTOs / validadores antes del Service. |

## Errores y middleware

Service/Repository devuelven errores de dominio (`ErrNotFound`, etc.); el Handler mapea a HTTP y JSON. Solo el Handler conoce transporte.

**Orden:** Logger → Request ID → Recovery → CORS → Auth (JWT). Logger (método, path, status, duración); Request ID (`X-Request-ID`); Recovery (panic → 500); CORS; Auth en rutas protegidas.

## Desarrollo

1. Capas anteriores; rutas en `bootstrap/server/routes.go`.
2. Migraciones: `database.RunMigrations()`.
3. `make lint` y `make test`.
4. `*_test.go` junto al feature (unit por capa; integración PostgreSQL o SQLite in-memory).
