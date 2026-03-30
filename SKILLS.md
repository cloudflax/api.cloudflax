# Skills | Cloudflax - Backend

Expectativas y stack para agente y contribuidores.

## Expectativas (respuestas)

- **Senior backend + Go**: no explicar HTTP/REST, DB ni sintaxis Go básica; ir al punto.
- **Prioridad**: arquitectura, patrones y tradeoffs frente a tutoriales genéricos.

## Agente — dominio y stack

| Tema | Contenido |
|------|-----------|
| **Go 1.25+** | context, `fmt.Errorf` con `%w`, interfaces, goroutines básicas |
| **Fiber v3** | router, middleware, handlers, `ctx.JSON` / bind / `Params` |
| **GORM** | modelos, migraciones, Create/First/Where, preload, scopes |
| **PostgreSQL** | tipos, constraints, índices, UUID como PK |
| **SQLite** | tests in-memory (p. ej. glebarez/sqlite) |
| **slog** | Info/Error, logs estructurados, context |
| **Tests** | testify, mocks |
| **Validación** | go-playground/validator, tags `validate:"..."` |
| **Arquitectura** | Handler → Service → Repository, por feature, DTOs |
| **Tooling** | golangci-lint, Docker / devcontainer, Makefile (build, test, lint, test-cover) |

## Contribuidores humanos

Go intermedio (structs, interfaces, errores, tests), REST/JSON, SQL relacional, Docker básico (devcontainer, compose), Git (ramas, commits, pre-commit).
