# Setup del proyecto Cloudflax API - Progreso

Seguimiento paso a paso de la configuración inicial del proyecto.

---

## Checklist

### Corto plazo (setup inicial)

- [x] **1. Estructura de carpetas** — `cmd/api/`, `internal/app/`, `internal/handlers/`
- [x] **2. Carga de `.env`** — godotenv + validación en internal/config
- [x] **3. Makefile** — build, run, test, lint (devcontainer)
- [x] **4. golangci-lint** — linter + IDE + pre-commit (devcontainer)
- [x] **5. Health check con DB** — `/health` verifica conexión a PostgreSQL

### Medio plazo

- [x] **6. Migraciones de base de datos** — GORM AutoMigrate + modelos User, Post
- [x] **7. Logging estructurado** — slog (JSON, middleware, LOG_LEVEL)
- [ ] **8. Tests unitarios** — cobertura básica
- [ ] **9. CI/CD** — GitHub Actions o GitLab CI

### Largo plazo

- [ ] **10. Métricas y tracing** — Prometheus, OpenTelemetry
- [ ] **11. Rate limiting y CORS** — seguridad
- [ ] **12. Documentación API** — OpenAPI/Swagger

---

## Estado actual

- **Siguiente paso:** 8. Tests unitarios
- **Última actualización:** 2026-02-06
