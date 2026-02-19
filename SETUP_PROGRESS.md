# Setup del proyecto Cloudflax API - Progreso

Seguimiento paso a paso de la configuración inicial del proyecto.

---

## Checklist

### Corto plazo — Módulo User

- [ ] **1. GET /users/me** — Devolver el usuario autenticado (userID desde token locals)
- [ ] **2. PUT /users/me** — Actualizar el propio perfil sin necesitar el ID en la URL
- [ ] **3. Revocar refresh tokens al eliminar usuario** — Llamar a `RevokeAllByUserID` desde el servicio de user al hacer DELETE

### Corto plazo — Módulo Auth

- [ ] **4. Cleanup de refresh tokens expirados** — Eliminar registros con `expires_at` pasado (tarea periódica o al login)

### Corto plazo — Infraestructura

- [ ] **5. CORS** — Headers para que el frontend consuma la API
- [ ] **6. Request ID** — Propagar `X-Request-ID` para tracing
- [ ] **7. Error handler centralizado** — Fiber error handler global para respuestas de error consistentes

### Medio plazo

- [ ] **8. CI/CD** — GitHub Actions o GitLab CI
- [ ] **9. Paginación** — `?page=1&limit=10` en listas

### Largo plazo

- [ ] **10. Rate limiting** — Límite de requests por IP
- [ ] **11. API versioning** — Rutas bajo `/api/v1/`
- [ ] **12. Métricas y tracing** — Prometheus, OpenTelemetry
- [ ] **13. Documentación API** — OpenAPI/Swagger

---

## Estado actual

- **Siguiente paso:** 1. GET /users/me
- **Última actualización:** 2026-02-19
