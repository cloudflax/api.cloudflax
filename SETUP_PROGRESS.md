# Setup del proyecto Cloudflax API - Progreso

Seguimiento del estado del proyecto. Lo ya implementado (modelos User/Account, registro, login JWT, cuentas y contexto de petición) está completo; este documento se centra en lo pendiente.

---

## Checklist pendiente

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

- **Completado:** Cuentas y propiedad de datos (docs/ACCOUNTS_AND_DATA_OWNERSHIP.md): modelos User/Account, registro, login JWT, cuentas, membresía, RequestContext y filtrado por account.
- **Completado:** Envío de email de verificación vía AWS SES v2 (`internal/shared/email`). El email se envía al registrar y al reenviar verificación. Configuración: `SES_FROM_ADDRESS`, `APP_URL`. Reutiliza `AWS_REGION`, `AWS_ENDPOINT_URL`, `AWS_ACCESS_KEY_ID` y `AWS_SECRET_ACCESS_KEY`.
- **Siguiente paso:** Corto plazo — CORS (tarea 6), Request ID (7) o error handler centralizado (8).
- **Última actualización:** 2026-02-21
