# Cloudflax API — Instrucciones para el agente

Este archivo guía al agente de IA con el contexto y las instrucciones del proyecto.

---

## Entorno de desarrollo

- **Devcontainer:** El proyecto se ejecuta dentro de un devcontainer. Asume que estás en `/app`.
- **Comandos:** Usa el Makefile para las operaciones habituales.

```bash
make build      # Compilar
make test       # Ejecutar tests
make test-cover # Tests con cobertura (genera coverage.html)
make lint       # Linter (golangci-lint)
```

---

## Antes de commit

- Ejecutar `make lint` y `make test` antes de hacer commit.
- Los tests deben pasar.
- El pre-commit hook ya ejecuta `make lint` automáticamente.

---

## Convenciones generales

- **Clean code y buenas prácticas:** Siempre aplicar en el código que escribas.
- **Idioma:** Todo el código fuente en inglés (nombres, variables, comentarios, mensajes de error, logs). La única comunicación en español será el chat con el usuario.
- **Tests:** Añadir o actualizar tests para el código que modifiques o crees.
- **Imports:** Ordenar: estándar → terceros → internal. Agrupar con líneas en blanco.
- **Errores:** Usar `fmt.Errorf` con `%w` para envolver errores. No ignorar errores con `_`.
- **Logging:** Usar slog para errores y eventos relevantes. No loguear datos sensibles (passwords, tokens).
- **Funciones:** Máximo ~50 líneas. Si crece, extraer. Evitar más de 3–4 parámetros.
- **Constantes:** Preferir constantes sobre magic numbers o strings literales repetidos.

---

## Cuándo consultar cada archivo

| Situación | Archivo |
|-----------|---------|
| Añadir/modificar un feature, entender capas o flujo | ARCHITECTURE.md |
| Nombrar funciones, archivos, variables | CONVENTIONS.md |
| Saber qué falta por hacer | SETUP_PROGRESS.md |
| Capacidades y stack del proyecto | SKILLS.md |
| Información general del proyecto | README.md |

---

## Workflow para añadir un feature nuevo

1. **Crear carpeta** — `internal/{recurso}/` (singular, lowercase).
2. **Archivos mínimos** — `model.go`, `repository.go`, `service.go`, `handler.go`, `routes.go`.
3. **Nombres CRUD** — `List{Resource}`, `Get{Resource}`, `Create{Resource}`, etc. (ver CONVENTIONS.md).
4. **Registrar rutas** — En `internal/bootstrap/server/routes.go` montar `{recurso}.Routes()`.
5. **Tests** — Añadir tests para handler, service y repository.
6. **Migraciones** — Si hay modelos nuevos, registrarlos en `database.RunMigrations()` en `cmd/api/main.go`.

---

## Referencias

- **ARCHITECTURE.md**
  - Estructura del proyecto, patrón por capas y organización por recurso.
  - Consultar al diseñar o añadir nuevos recursos.
- **CONVENTIONS.md**
  - Nombres, estilo de código y patrones de naming.
  - Consultar al escribir código.
- **SETUP_PROGRESS.md**
  - Estado actual del proyecto y lista de tareas pendientes.
  - Consultar antes de empezar una tarea nueva.
- **README.md**
  - Información general del proyecto y estructura.
- **SKILLS.md**
  - Capacidades del agente, del equipo y tecnologías del stack.

---

## Stack

- **Go 1.25** | **Fiber v3** | **GORM** | **PostgreSQL** | **slog**
