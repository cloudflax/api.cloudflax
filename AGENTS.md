## AGENTS — Cloudflax API

Este documento establece las directrices operativas que deben seguir los agentes de Cursor. El objetivo es garantizar la consistencia arquitectónica, la calidad del código y la velocidad de entrega en todo el ecosistema de Cloudflax.

---

## Entorno de desarrollo

- **Devcontainer:** El proyecto se ejecuta dentro de un devcontainer. Asume siempre que la ruta de trabajo es `/app`.
- **Operaciones:** Utiliza el `Makefile` para las tareas estándar del ciclo de vida:
  - `make build`: Compilar la aplicación.
  - `make test`: Ejecutar la suite de pruebas.
  - `make test-cover`: Generar reporte de cobertura (`coverage.html`).
  - `make lint`: Ejecutar el linter (`golangci-lint`).

---

## Estándares de Calidad (Antes de Commit)

Para asegurar la integridad del repositorio, el agente debe:
- Ejecutar obligatoriamente `make lint` y `make test` antes de proponer cambios.
- Confirmar que todos los tests pasen con éxito.
- Recordar que el *pre-commit hook* automatizado ya incluye la ejecución de `make lint`.

---

## Convenciones Generales de Código

- **Idioma:** El código fuente (nombres, variables, comentarios, logs y errores) debe ser estrictamente en **inglés**. La comunicación con el usuario se mantiene en español.
- **Gestión de Errores:** Usa `fmt.Errorf` con el verbo `%w` para el wrapping de errores. Nunca ignores errores con `_`.
- **Observabilidad:** Utiliza `slog` para el registro de eventos. Queda prohibido loguear datos sensibles como passwords o tokens.
- **Estructura de Funciones:** Mantén funciones concisas (máximo ~50 líneas) y limita los parámetros a 3 o 4 por firma.
- **Imports:** Ordena los paquetes por bloques: 1) Estándar, 2) Terceros, 3) Internos.
- **Constantes:** Evita los "magic numbers"; usa constantes descriptivas para valores repetidos.

---

## Guía de Consulta Rápida

| Si necesitas... | Consulta el archivo... |
| :--- | :--- |
| Entender capas, flujo o añadir un feature | `ARCHITECTURE.md` |
| Reglas de naming para funciones o variables | `CONVENTIONS.md` |
| Conocer las tareas pendientes | `SETUP_PROGRESS.md` |
| Revisar el stack y capacidades del agente | `SKILLS.md` |
| Obtener información general o estructura | `README.md` |

---

## Workflow: Implementación de Features

Para añadir una nueva funcionalidad, sigue este orden estricto:
1. **Directorio:** Crea `internal/{recurso}/` en singular y minúsculas.
2. **Componentes:** Define como mínimo `model.go`, `repository.go`, `service.go`, `handler.go` y `routes.go`.
3. **Naming CRUD:** Usa el formato `List{Resource}`, `Get{Resource}`, `Create{Resource}`.
4. **Rutas:** Registra el recurso llamando a `{recurso}.Routes()` en `internal/bootstrap/server/routes.go`.
5. **Persistencia:** Si hay cambios en modelos, registra las migraciones en `database.RunMigrations()` dentro de `cmd/api/main.go`.
6. **Validación:** Añade tests unitarios y de integración para cada capa modificada.

---

## Stack Tecnológico Core

- **Lenguaje:** Go 1.25
- **Web Framework:** Fiber v3
- **ORM:** GORM
- **Base de Datos:** PostgreSQL
- **Logging:** slog