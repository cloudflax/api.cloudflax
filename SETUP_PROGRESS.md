# Setup del proyecto Cloudflax API - Progreso

Seguimiento del estado del proyecto.

---

## Completado — Migración a AWS dev real

Transición de servicios AWS simulados (Moto/LocalStack) a los servicios reales del entorno de desarrollo en AWS.

- [x] 1. Limpiar `docker-compose.yml` — migrado a `env_file` con `.env`
- [x] 2. Configurar credenciales AWS reales — `.env` con placeholders, `.env.example` como referencia
- [x] 3. Limpiar comentarios de moto/LocalStack en el código Go
- [x] 4. Actualizar README con el nuevo flujo de configuración

---

## Checklist pendiente

### Corto plazo — Infraestructura

- [ ] CORS — Headers para que el frontend consuma la API
- [ ] Request ID — Propagar `X-Request-ID` para tracing
- [ ] Error handler centralizado — Fiber error handler global para respuestas de error consistentes

### Medio plazo

- [ ] CI/CD — GitHub Actions o GitLab CI
- [ ] Paginación — `?page=1&limit=10` en listas

### Largo plazo

- [ ] Rate limiting — Límite de requests por IP
- [ ] API versioning — Rutas bajo `/api/v1/`
- [ ] Métricas y tracing — Prometheus, OpenTelemetry
- [ ] Documentación API — OpenAPI/Swagger

---

## Completado

- Cuentas y propiedad de datos: modelos User/Account, registro, login JWT, cuentas, membresía, RequestContext y filtrado por account.
- Envío de email de verificación vía AWS SES v2 (`internal/shared/email`).
- Invoices CRUD con filtrado por account.

---

**Última actualización:** 2026-03-07
