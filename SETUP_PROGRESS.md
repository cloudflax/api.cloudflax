# Setup del proyecto Cloudflax API - Progreso

Seguimiento paso a paso de la configuración inicial del proyecto.

---

## Checklist

### Prioridad inmediata — Estandar de errores (API user)

- [ ] **0. Definir contrato de error** — Estructura única: `error.code`, `error.message`, `error.status`, `error.trace_id`, `error.details[]`
- [ ] **0.1 Catálogo de códigos de error de user** — Enumerar códigos estables (p. ej. `USER_NOT_FOUND`, `VALIDATION_ERROR`, `EMAIL_ALREADY_EXISTS`)
- [ ] **0.2 Política de HTTP status** — Mapear casos de user a `400/401/403/404/409/422/500`
- [ ] **0.3 Formato para errores simples** — Respuesta estándar cuando exista un único error de negocio o sistema
- [ ] **0.4 Formato para errores múltiples** — Respuesta estándar para validaciones con varios campos en una misma petición
- [ ] **0.5 Normalizar validaciones en handlers user** — Unificar mensajes y códigos de validación por campo
- [ ] **0.6 Error handler central para user** — Adaptar handlers para delegar al formato unificado
- [ ] **0.7 Actualizar tests de handler user** — Cubrir caso de error único y múltiples errores en una misma request
- [ ] **0.8 Documentar ejemplos de error** — Añadir ejemplos JSON (1 error y N errores) para frontend/integraciones

### Corto plazo — Usuario y acceso al SaaS

**Base (modelo y rutas públicas)**
- [x] **1. Validación de entrada** — Validar body (email, password) con go-playground/validator

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

- **Siguiente paso:** 2. POST /users (registro de usuario)
- **Última actualización:** 2026-02-17
