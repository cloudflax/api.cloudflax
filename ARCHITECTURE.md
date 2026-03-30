# Architecture | Cloudflax Backend

## Stack
Go (feature-driven), Fiber v3, GORM + PostgreSQL; tests unitarios e integración (PostgreSQL / SQLite in-memory).

## `internal/` (feature-driven)
| Ruta | Contenido |
|------|-----------|
| `bootstrap/` | Config, servidor, arranque |
| `{feature}/` | Por recurso (opcional según complejidad): `handler`, `service`, `repository`, `model`, `dto`, `routes`, `validator`, helpers |
| `shared/` | `database` (conexión + migraciones), `middleware`, `pagination`, `filtering`, `errors`, `validator`, `verificationnotify` (Lambda correo verificación), utilidades |

Recursos simples pueden omitir capas (p. ej. solo handler + repository + model + routes).

## Capas
Flujo: **Middleware → Handler → Service → Repository → DB**.

| Capa | Rol | No hace |
|------|-----|---------|
| Handler | HTTP Fiber: parseo requests, respuestas | Lógica de negocio, acceso directo a DB |
| Service | Negocio y orquestación | Depender de HTTP/Fiber ni códigos de estado |
| Repository | Acceso datos (GORM) | Respuestas HTTP |

Validación: DTOs y checks **antes** del Service.

## Errores y transporte
Service/Repository devuelven errores de dominio (p. ej. `ErrNotFound`, `ErrDuplicateEmail`). El Handler los mapea a HTTP + JSON. Solo el Handler conoce transporte; Service/Repository sin Fiber ni respuestas HTTP.

## Middleware (orden)
`Logger → Request ID → Recovery → CORS → Auth (JWT)`  
Logger: método, path, status, duración. Request ID: `X-Request-ID` para tracing. Recovery: `panic` → 500 controlado. CORS: frontend autorizado. Auth: JWT en rutas protegidas.

## Workflow
1. Implementar capas; rutas en `internal/bootstrap/server/routes.go`.
2. Cambios de modelo GORM: registrar en `database.RunMigrations()` (`cmd/api/main.go`).
3. Antes de cerrar: `make lint` y `make test`.
4. Tests: `*_test.go` junto al feature; unitarios por capa + integración con DB.
