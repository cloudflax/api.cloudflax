# Arquitectura — Cloudflax API (Backend)

Documentación de la estructura, patrones y flujo de datos del backend.

---

## 1. Stack Tecnológico

- **Lenguaje**: Go (Estructura feature-driven).
- **Framework**: Fiber v3.
- **ORM**: GORM con PostgreSQL.
- **Testing**: Unitarios y de integración con PostgreSQL/SQLite.

---

## 2. Estructura de Directorios (Feature-driven)

Cada funcionalidad vive en su propia carpeta dentro de `internal/`:

- **`internal/bootstrap/`**: Configuración, servidor y arranque de la app.
- **`internal/{feature}/`**: Puede contener `handler.go`, `service.go`, `repository.go`, `model.go`, `dto.go`, `routes.go` y opcionalmente `validator.go` u otros helpers específicos del recurso.
- **`internal/shared/`**: Código común: `database` (conexión y migraciones), `middleware`, `pagination`, `filtering`, `errors`, `validator` y otras utilidades compartidas.

No todos los features requieren todos estos archivos; un recurso simple puede usar solo `handler.go`, `repository.go`, `model.go` y `routes.go`.

---

## 3. Patrones y Capas del Sistema

- **Flujo de Datos**: `Request → Middleware → Handler → Service → Repository → DB`.
- **Responsabilidades**:
  - **Handler**: Solo HTTP (Fiber); parseo de requests y formateo de respuestas.
  - **Service**: Lógica de negocio y orquestación; independiente de HTTP.
  - **Repository**: Abstracción de acceso a datos (GORM).
- **Validación**: Uso de DTOs y lógica de validación previa al Service.
- **Reglas de dependencia entre capas**:
  - **Handler**: No debe contener lógica de negocio ni acceder directamente a la base de datos; siempre delega en el Service.
  - **Service**: No debe depender de HTTP (ni Fiber ni códigos de estado); trabaja con modelos de dominio y errores de dominio.
  - **Repository**: No debe construir respuestas HTTP; solo se encarga del acceso a datos y de devolver errores de DB o de dominio.

---

## 4. Manejo de Errores y Seguridad

- **Errores**: Service/Repository devuelven errores de dominio (por ejemplo `ErrNotFound`, `ErrDuplicateEmail`); el Handler captura estos errores, los mapea a códigos HTTP coherentes y formatea la respuesta JSON.
- **Restricción**: Ni Service ni Repository deben crear respuestas HTTP ni depender de Fiber; solo el Handler conoce detalles de transporte.
- **Middleware Stack**: Logger → Request ID → Recovery → CORS → Auth (JWT).
  - **Logger**: Registra método, path, status y duración de cada request.
  - **Request ID**: Asigna y propaga un identificador (`X-Request-ID`) para facilitar el tracing entre logs.
  - **Recovery**: Captura `panic` y evita la caída del servidor, devolviendo un error 500 controlado.
  - **CORS**: Configura los encabezados para permitir el consumo desde el frontend autorizado.
  - **Auth**: Valida JWT y restringe el acceso a rutas protegidas según la configuración de autorización.

---

## 5. Workflow del Desarrollador

1. **Implementar**: Seguir el flujo de capas definido y registrar rutas en `bootstrap/server/routes.go`.
2. **Migrar**: Registrar cambios de modelos en `database.RunMigrations()`.
3. **Validar**: Ejecución obligatoria de `make lint` y `make test`.
4. **Testear**: Añadir tests unitarios por capa (Handler, Service, Repository) y tests de integración con base de datos (PostgreSQL o SQLite in-memory), ubicando los archivos `*_test.go` junto al código del feature.