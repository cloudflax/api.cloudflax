# Setup del proyecto Cloudflax API - Progreso

Seguimiento paso a paso de la configuración inicial del proyecto.

---

## Checklist

### Corto plazo — Módulo User

- [x] **1. GET /users/me** — Devolver el usuario autenticado (userID desde token locals)
- [x] **2. PUT /users/me** — Actualizar el propio perfil sin necesitar el ID en la URL
- [x] **3. DELETE /users/me** — Eliminar el propio usuario sin necesitar el ID en la URL (userID desde token locals)
- [x] **4. Revocar refresh tokens al eliminar usuario** — Llamar a `RevokeAllByUserID` desde el servicio de user al hacer DELETE

### Corto plazo — Módulo Auth

- [ ] **5. Cleanup de refresh tokens expirados** — Eliminar registros con `expires_at` pasado (tarea periódica o al login)

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

- **Siguiente paso:** 5. Cleanup de refresh tokens expirados
- **Última actualización:** 2026-02-19

