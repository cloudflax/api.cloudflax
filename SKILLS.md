# Skills — Cloudflax API

Capacidades y conocimientos esperados para trabajar en este proyecto.

---

## Skill Expectations

- **Assume senior backend knowledge** — No explicar conceptos básicos de backend, HTTP o bases de datos.
- **Do not explain basic Go syntax** — Asumir dominio de Go. Ir directo al punto.
- **Focus on architecture and tradeoffs** — Priorizar decisiones de diseño, patrones y pros/contras.

---

## Skills del agente

El agente de IA debe dominar:

| Área | Conocimiento requerido |
|------|------------------------|
| **Go** | Sintaxis 1.25+, context, error handling con `%w`, interfaces, goroutines básicas |
| **Fiber v3** | Router, handlers, middleware, `ctx.JSON()`, `ctx.Params()`, request bind |
| **GORM** | Modelos, migraciones, `db.Create()`, `db.First()`, `db.Where()`, preload, scopes |
| **PostgreSQL** | Tipos, constraints, índices, UUID como PK |
| **slog** | `slog.Info`, `slog.Error`, structured logging, context |
| **Testing** | testify, mocks, SQLite in-memory para tests rápidos |
| **Validación** | go-playground/validator, tags `validate:"required,email"` |
| **Arquitectura** | Handler → Service → Repository, feature-driven, DTOs |

---

## Skills del equipo

Quien contribuya debe conocer:

- **Go** — Nivel intermedio (structs, interfaces, error handling, tests)
- **HTTP/REST** — Verbos, status codes, JSON
- **Bases de datos relacionales** — Modelado, relaciones, queries
- **Docker** — Básico (devcontainer, docker-compose)
- **Git** — Branching, commits, pre-commit

---

## Skills del proyecto

Tecnologías y prácticas del stack:

| Tecnología | Uso |
|------------|-----|
| **Go 1.25** | Lenguaje principal |
| **Fiber v3** | Framework HTTP |
| **GORM** | ORM, migraciones |
| **PostgreSQL** | Base de datos principal |
| **SQLite** | Tests (glebarez/sqlite in-memory) |
| **slog** | Logging estructurado |
| **testify** | Assertions y mocks |
| **golangci-lint** | Linter |
| **go-playground/validator** | Validación de entrada |
| **Docker / Devcontainer** | Entorno de desarrollo |
| **Makefile** | build, test, lint, test-cover |
