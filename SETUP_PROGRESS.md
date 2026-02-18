# Setup del proyecto Cloudflax API - Progreso

Seguimiento paso a paso de la configuración inicial del proyecto.

---

## Checklist

### Prioridad inmediata — Estandar de errores (API user)

### Corto plazo — Usuario y acceso al SaaS

**Base (modelo y rutas públicas)**

**Autenticación**
- [ ] **4. POST /auth/login** — Login (email + password), devolver token
- [ ] **5. JWT** — Generación y validación de tokens
- [ ] **6. Middleware de autenticación** — Proteger rutas que requieren login

**Rutas protegidas y CRUD**
- [ ] **7. GET /users/me** — Usuario actual (token en header)
- [ ] **8. PUT /users/:id** — Actualizar usuario
- [ ] **9. DELETE /users/:id** — Eliminar usuario (soft delete)

**Infraestructura para frontend**
- [ ] **10. CORS** — Headers para que el frontend consuma la API
- [ ] **11. Request ID** — X-Request-ID para tracing
- [ ] **12. Error handler centralizado** — Respuestas de error consistentes

### Medio plazo

- [ ] **13. CI/CD** — GitHub Actions o GitLab CI
- [ ] **14. Paginación** — ?page=1&limit=10 en listas

### Largo plazo

- [ ] **15. Rate limiting** — Límite de requests por IP
- [ ] **16. API versioning** — Rutas bajo /api/v1/
- [ ] **17. Métricas y tracing** — Prometheus, OpenTelemetry
- [ ] **18. Documentación API** — OpenAPI/Swagger

---

## Estado actual

- **Siguiente paso:** 4. POST /auth/login
- **Última actualización:** 2026-02-18
