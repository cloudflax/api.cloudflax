## Agents — Cloudflax API (Backend)

Este documento establece las directrices operativas que deben seguir los agentes de Cursor para garantizar la robustez, seguridad y escalabilidad de la API de Cloudflax.

## Stack tecnológico

- **Lenguaje**: Go 1.25
- **Framework**: Fiber v3
- **ORM**: GORM (PostgreSQL)
- **Observabilidad**: slog (Structured Logging)

## Convenciones de Código y Naming

- **Idioma**: El código fuente, logs y mensajes de error deben ser estrictamente en **inglés**.
- **Naming CRUD**: Sigue el patrón `List{Resource}`, `Get{Resource}`, `Create{Resource}`, `Update{Resource}`.
- **Funciones**: Mantén una lógica simple; máximo ~50 líneas por función y no más de 4 parámetros.

## Arquitectura y Datos

- **Estructura de Carpetas**: Implementa cada recurso en `internal/{recurso}/` con sus capas correspondientes:
  - `model.go`: Definición de estructuras y tags de GORM.
  - `repository.go`: Consultas directas a la base de datos.
  - `service.go`: Lógica de negocio y validaciones.
  - `handler.go`: Controladores de Fiber (entrada/salida).
- **Gestión de Errores**: Usa `fmt.Errorf` con el verbo `%w` para mantener la trazabilidad. Nunca ignores un error.
- **Migraciones**: Si modificas un modelo, registra la migración en `database.RunMigrations()` dentro de `cmd/api/main.go`.

## Seguridad operativa y herramientas

- **Logs**: Usa `slog`. Queda terminantemente prohibido loguear datos sensibles (passwords, tokens, PII).
- **Makefile**: Usa `make lint` y `make test` como validación obligatoria antes de proponer cambios.
- **Entorno**: El entorno de trabajo es `/app` dentro de un Devcontainer.


## Handlers HTTP y respuestas

- **Consistencia en respuestas**: Estandariza la forma de responder errores y éxitos (status code + payload JSON con `message`, `data` opcional y/o `error`).
- **Validación de entrada**: Valida siempre parámetros de ruta, query y body en la capa `handler` o `service` antes de llamar al repositorio.
- **Errores HTTP**: Mapea los errores de negocio a códigos HTTP claros (`400` validación, `401`/`403` auth/autorización, `404` no encontrado, `409` conflicto, `500` errores inesperados).

## Logging y observabilidad

- **Niveles de log**: Usa `Debug` para detalle de desarrollo, `Info` para flujos exitosos importantes, `Warn` para situaciones anómalas recuperables y `Error` para fallos que requieren atención.
- **Contexto estructurado**: Incluye siempre campos relevantes (por ejemplo `request_id`, `user_id`, `resource`, `operation`) usando `slog` con atributos estructurados.
- **Trazabilidad**: Propaga un identificador de correlación por request (por ejemplo, `X-Request-ID`) y loguéalo en todos los puntos clave.
- **No PII**: Nunca incluyas en logs passwords, tokens, credenciales, ni datos sensibles de usuarios.

## Pruebas

- **Cobertura mínima**: Cada nueva funcionalidad debe incluir tests de unidad para servicios y, cuando aplique, tests de integración para repositorios y endpoints críticos.
- **make test**: Ejecuta `make test` antes de abrir un PR; no se deben introducir fallos en el pipeline de tests.
- **Aislamiento**: Evita dependencias externas no deterministas en tests (red, tiempos reales, etc.); usa dobles de prueba o fixtures controlados.

## Acceso a datos y rendimiento

- **Paginación obligatoria**: En listados (`List{Resource}`) implementa paginación por defecto para evitar lecturas masivas en memoria.
- **Consultas eficientes**: Revisa índices y uso de `SELECT` específicos; evita `SELECT *` en consultas críticas.
- **Transacciones**: Usa transacciones en el repositorio cuando varias operaciones deban ser atómicas.

## Seguridad

- **Validación y saneamiento**: Valida y normaliza toda entrada externa antes de usarla en queries o lógica sensible.
- **Principio de mínimo privilegio**: Diseña servicios y consultas asumiendo el menor conjunto posible de permisos y datos expuestos.
- **Gestión de secretos**: Nunca hardcodees secretos o credenciales en el repositorio; deben provenir de variables de entorno o sistemas seguros.

## Flujo de trabajo y PRs

- **Antes de abrir un PR**:
  - Ejecuta `make lint` y `make test` y asegúrate de que pasan.
  - Revisa que no haya logs de depuración temporales ni comentarios obsoletos.
- **Durante la revisión**:
  - Explica brevemente el propósito del cambio y cualquier decisión no obvia.
  - Acepta y responde comentarios en inglés dentro del código, manteniendo la descripción funcional en español en la conversación con el usuario.


## Flujo de Comunicación

- Explica los cambios en español de forma concisa.
- Informa proactivamente si los cambios afectan al archivo `.env` o requieren una migración de base de datos.
