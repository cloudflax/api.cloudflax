# Setup del proyecto Cloudflax API - Progreso

Seguimiento paso a paso de la configuración inicial del proyecto.

---

## Checklist

### Corto plazo (setup inicial)

- [ ] **1. POST /users** — Crear usuario
- [ ] **2. CORS** — Headers para frontend
- [ ] **3. Request ID** — X-Request-ID en cada request para tracing
- [ ] **4. Validación de entrada** — Validar body y params (go-playground/validator)
- [ ] **5. Error handler centralizado** — Respuestas de error consistentes
- [ ] **6. CRUD completo Users** — PUT, DELETE (completar con POST del punto 1)

### Medio plazo

- [ ] **7. CI/CD** — GitHub Actions o GitLab CI
- [ ] **8. Paginación** — ?page=1&limit=10 en listas

### Largo plazo

- [ ] **9. Rate limiting** — Límite de requests por IP
- [ ] **10. API versioning** — Rutas bajo /api/v1/
- [ ] **11. Métricas y tracing** — Prometheus, OpenTelemetry
- [ ] **12. Documentación API** — OpenAPI/Swagger

---

## Estado actual

- **Siguiente paso:** 1. POST /users
- **Última actualización:** 2026-02-07
