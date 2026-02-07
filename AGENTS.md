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
- **Idioma:** Código en inglés (nombres, variables, comentarios). Documentación y mensajes al usuario en español.
- **Tests:** Añadir o actualizar tests para el código que modifiques o crees.
- **Imports:** Ordenar: estándar → terceros → internal. Agrupar con líneas en blanco.
- **Errores:** Usar `fmt.Errorf` con `%w` para envolver errores. No ignorar errores con `_`.
- **Logging:** Usar slog para errores y eventos relevantes. No loguear datos sensibles (passwords, tokens).
- **Funciones:** Máximo ~50 líneas. Si crece, extraer. Evitar más de 3–4 parámetros.
- **Constantes:** Preferir constantes sobre magic numbers o strings literales repetidos.

---

## Referencias

- **SETUP_PROGRESS.md**
  - Estado actual del proyecto y lista de tareas pendientes.
  - Consultar antes de empezar una tarea nueva.
- **README.md**
  - Información general del proyecto y estructura.

---

## Stack

- **Go 1.25** | **Fiber v3** | **GORM** | **PostgreSQL** | **slog**
