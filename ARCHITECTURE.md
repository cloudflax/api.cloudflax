# Arquitectura — Cloudflax API

Documentación de la estructura y patrones del proyecto. En este archivo van:

- **Estructura del proyecto** — Enfoque feature-driven, layout de carpetas y archivos por feature.
- **Patrones utilizados** — Repository, Service, DTO, etc.
- **Capas del sistema** — Handler, Service, Repository y sus responsabilidades.
- **Flujo de responsabilidades** — Cómo fluyen los datos entre capas.
- **Qué puede y qué no puede tocar cada capa** — Reglas de dependencia.

---

## Enfoque: Feature-driven + Shared Kernel

Cada feature (user, product, order, etc.) vive en su propia carpeta con todos sus archivos. El código común se centraliza en `shared/`.

---

## Estructura objetivo

```
internal/
├── bootstrap/          # Arranque y configuración de la aplicación
│   ├── app/            # Creación del Fiber app, Run()
│   ├── config/         # Carga de configuración (env, secrets)
│   └── server/         # Router principal, Mount(), handlers raíz (/health, /)
│
├── user/
│   ├── handler.go      # HTTP, Fiber, request/response
│   ├── service.go      # Lógica de negocio
│   ├── repository.go   # Acceso a DB
│   ├── model.go        # Modelo User
│   ├── validator.go    # Validaciones del user
│   ├── dto.go          # Requests / responses
│   └── routes.go       # Endpoints del user
│
├── product/            # Ejemplo: feature más simple
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   ├── model.go
│   └── routes.go
│
└── shared/
    ├── database/       # Conexión, migraciones
    ├── jsonapi/        # Formato JSON API (si aplica)
    ├── pagination/     # Lógica de paginación
    ├── filtering/      # Filtros genéricos
    ├── errors/         # Errores y respuestas HTTP
    └── validator/      # Validaciones comunes (struct tags, helpers)
```

---

## Archivos por feature

| Archivo | Responsabilidad |
|---------|-----------------|
| `handler.go` | HTTP, Fiber, parsear request, formatear response |
| `service.go` | Lógica de negocio, orquestación |
| `repository.go` | Acceso a DB (GORM), queries |
| `model.go` | Modelo de dominio (GORM) |
| `validator.go` | Validaciones específicas del recurso |
| `dto.go` | Request/Response DTOs |
| `routes.go` | Definición de endpoints del recurso |

No todos los features requieren todos los archivos. Por ejemplo, un recurso simple puede tener solo `handler`, `repository`, `model` y `routes`.

---

## Patrones utilizados

(Patrones de diseño y arquitectura que aplicamos en el proyecto.)

| Patrón | Uso |
|--------|-----|
| **Repository** | Abstracción del acceso a datos en cada feature |
| **Service** | Lógica de negocio, orquestación entre repositorios |
| **DTO** | Request/Response separados del modelo de dominio |
| **Validator** | Validación de entrada antes de llegar al service |

---

## Capas

| Capa | Responsabilidad |
|------|-----------------|
| **Handler** | Solo HTTP: parsear request, llamar al service, formatear response. Sin lógica de negocio. |
| **Service** | Reglas de negocio, validaciones, orquestación. |
| **Repository** | Acceso a base de datos (GORM), queries. |

```
Handler → Service → Repository
```

---

## Reglas (qué puede y qué no puede cada capa)

- **Handler:** debe delegar al service. No debe contener lógica de negocio.
- **Service:** no debe depender de HTTP (no Fiber, no status codes). Usa errores de dominio.
- **Repository:** no debe devolver errores HTTP. Solo errores de DB o de dominio.

---

## Flujo de datos

```
Request → Middleware → Handler → Service → Repository → DB
                                 ↓
Response ← Handler ← Service ← Repository
```

---

## Middleware stack

Orden de ejecución (de fuera hacia dentro):

1. **Logger** — Registra request, método, path, duración
2. **Request ID** — Añade `X-Request-ID` para tracing
3. **Recovery** — Panic recovery, evita caída del servidor
4. **CORS** — Headers para consumo desde frontend
5. **Auth** — Validación JWT en rutas protegidas (solo donde aplique)

---

## Manejo de errores

- **Service/Repository:** Devuelven errores de dominio (p. ej. `ErrNotFound`, `ErrDuplicateEmail`).
- **Handler:** Captura errores del service, mapea a códigos HTTP y formatea la respuesta.
- **shared/errors:** Errores comunes y helpers para respuestas HTTP consistentes (futuro).

---

## Rutas

- **bootstrap/server/routes.go:** Monta las rutas de cada feature (`user.Routes()`, `product.Routes()`, etc.)
- **Públicas:** `/`, `/health`, `/auth/login`
- **Protegidas:** `/users`, `/users/:id`, `/users/me`
- **Prefijo (futuro):** `/api/v1/` para versionado

---

## Estrategia de testing

- **Unit tests:** Handler, Service, Repository por separado. Usar mocks para dependencias.
- **Integration tests:** Flujo completo con DB real (PostgreSQL) o SQLite in-memory para tests rápidos.
- **Ubicación:** `*_test.go` junto al archivo bajo test.
- **DB en tests:** `internal/db/testing.go` provee helpers para tests con SQLite o PostgreSQL.

---

## Migración

La estructura ha sido migrada a feature-driven. El módulo `user` usa la nueva estructura. La capa de arranque (app, config, server) vive en `internal/bootstrap/`. `shared/database` y `shared/middleware` centralizan código común.
