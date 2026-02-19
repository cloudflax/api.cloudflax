# Setup del proyecto Cloudflax API - Progreso

Seguimiento paso a paso de la configuración inicial del proyecto.

---

## Checklist

### Corto plazo — Infraestructura

- [ ] **6. CORS** — Headers para que el frontend consuma la API
- [ ] **7. Request ID** — Propagar `X-Request-ID` para tracing
- [ ] **8. Error handler centralizado** — Fiber error handler global para respuestas de error consistentes

### Medio plazo

- [ ] **9. CI/CD** — GitHub Actions o GitLab CI
- [ ] **10. Paginación** — `?page=1&limit=10` en listas

### Largo plazo

- [ ] **11. Rate limiting** — Límite de requests por IP
- [ ] **12. API versioning** — Rutas bajo `/api/v1/`
- [ ] **13. Métricas y tracing** — Prometheus, OpenTelemetry
- [ ] **14. Documentación API** — OpenAPI/Swagger

---

## Estado actual

- **Siguiente paso:** 6. CORS
- **Última actualización:** 2026-02-19

