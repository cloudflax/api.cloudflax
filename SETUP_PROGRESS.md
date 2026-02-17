# Setup del proyecto Cloudflax API - Progreso

Seguimiento paso a paso de la configuración inicial del proyecto.

---

## Checklist

### Corto plazo — Usuario y acceso al SaaS

**Base (modelo y rutas públicas)**
- [x] **1. Extender modelo User** — Email (único), password (hash), campos para login
- [ ] **2. POST /users** — Registro de usuario (crear cuenta)
- [ ] **3. Validación de entrada** — Validar body (email, password) con go-playground/validator

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

### Integración avanzada con Secrets Manager (LocalStack)

- [x] **S1. Capa de caché para secretos** — Definir interfaz de caché (en memoria) y aplicar patrón Singleton con TTL para evitar llamadas repetitivas a Secrets Manager.
- [x] **S2. Manejo de rotación y reintentos** — Añadir lógica de reintento cuando haya errores de autenticación/credenciales (p.ej. invalid auth) e invalidar caché en esos casos.
- [x] **S3. Seguridad de memoria y parseo** — Asegurar que los secretos se parseen directamente a structs privados (sin pasar por variables de entorno adicionales), revisando que no se impriman ni se logueen.
- [ ] **S4. Pruebas y observabilidad** — Tests unitarios/mocks para la capa de caché + métricas/logs mínimos para monitorizar aciertos/fallos de caché (solo en desarrollo).

### Largo plazo

- [ ] **15. Rate limiting** — Límite de requests por IP
- [ ] **16. API versioning** — Rutas bajo /api/v1/
- [ ] **17. Métricas y tracing** — Prometheus, OpenTelemetry
- [ ] **18. Documentación API** — OpenAPI/Swagger

---

## Estado actual

- **Siguiente paso:** 2. POST /users (registro de usuario)
- **Última actualización:** 2026-02-17
