## Agents | Cloudflax - Backend

Instrucciones para quien ejecuta tareas en este repo (agente). Prioriza [docs/AGENT_RULES.md](./docs/AGENT_RULES.md) y la tabla siguiente; cada fila enlaza el detalle cuando aplica.

Abre una **referencia solo cuando aplique** a la tarea (tabla siguiente). No cargues documentación que no necesites.

### Acción → referencia

| Acción | Referencia |
|--------|------------|
| Entorno, `make`, árbol del repo | [README.md](./README.md) |
| Stack, capas, middleware, migraciones | [ARCHITECTURE.md](./ARCHITECTURE.md) |
| API, errores, JSON, tests | [CONVENTIONS.md](./CONVENTIONS.md) |
| Issues, ramas, Git Flow, commits | [docs/GITHUB_WORKFLOW.md](./docs/GITHUB_WORKFLOW.md) |
| Nivel y foco de tus respuestas | [SKILLS.md](./SKILLS.md) |
| Contrato auth (JWT, cookies) | [AUTH_INTEGRATION.md](./docs/AUTH_INTEGRATION.md) |
| Cuentas y titularidad de datos | [docs/ACCOUNTS_AND_DATA_OWNERSHIP.md](./docs/ACCOUNTS_AND_DATA_OWNERSHIP.md) |
| Comentarios Go (sin En/Es en doc de `package`; En/Es solo en declaraciones, no en sentencias internas) | [.cursor/rules/go-bilingual-comments.mdc](./.cursor/rules/go-bilingual-comments.mdc) (siempre activo en Cursor) |
| Obligaciones del agente (lint/test, GORM, GitHub, logs, docs; respuesta al humano; stack) | [docs/AGENT_RULES.md](./docs/AGENT_RULES.md) |

**Obligaciones y respuesta al humano:** lista y detalle en [docs/AGENT_RULES.md](./docs/AGENT_RULES.md) (incluye qué hacer si no abres otro doc).
