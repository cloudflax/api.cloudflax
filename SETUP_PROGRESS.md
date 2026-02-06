# Setup del proyecto Cloudflax API - Progreso

Seguimiento paso a paso de la configuración inicial del proyecto.

---

## Checklist

### Corto plazo (setup inicial)

- [x] **1. Estructura de carpetas** — `cmd/api/`, `internal/app/`, `internal/handlers/`
- [ ] **2. Carga de `.env`** — godotenv o envconfig + validación al arranque
- [ ] **3. Makefile** — build, run, test, lint, docker-up, docker-down
- [ ] **4. golangci-lint** — linter + pre-commit o integración en devcontainer
- [ ] **5. Health check con DB** — `/health` que verifique conexión a PostgreSQL

### Medio plazo

- [ ] **6. Migraciones de base de datos** — golang-migrate o goose
- [ ] **7. Logging estructurado** — slog o zerolog
- [ ] **8. Tests unitarios** — cobertura básica
- [ ] **9. CI/CD** — GitHub Actions o GitLab CI

### Largo plazo

- [ ] **10. Métricas y tracing** — Prometheus, OpenTelemetry
- [ ] **11. Rate limiting y CORS** — seguridad
- [ ] **12. Documentación API** — OpenAPI/Swagger

---

## Estado actual

- **Siguiente paso:** 2. Carga de `.env`
- **Última actualización:** 2026-02-06
