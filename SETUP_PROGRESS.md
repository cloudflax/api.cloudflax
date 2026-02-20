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

## Ruta de trabajo: Cuentas y propiedad de datos

Referencia: **docs/ACCOUNTS_AND_DATA_OWNERSHIP.md**. Las tareas están ordenadas para mínimo contexto por paso (una capa o un recurso a la vez).

### Fase 1 — Modelos y migraciones (DB)

| # | Tarea | Alcance | Depende de |
|---|--------|---------|-------------|
| ✅ A1 | **Modelos User y UserAuthProvider** | `internal/user/`: `model.go` con User (id, name, email, password_hash, email_verified_at, timestamps) y UserAuthProvider (user_id, provider, provider_subject_id). Registrar en migraciones. | — |
| ✅ A2 | **Modelos Account y AccountMember** | `internal/account/`: `model.go` con Account (id, name, slug UK) y AccountMember (account_id, user_id, role). UNIQUE(account_id, user_id). Registrar en migraciones. | — |

### Fase 2 — Registro de usuario

| # | Tarea | Alcance | Depende de |
|---|--------|---------|-------------|
| ✅ B1 | **Repository users y user_auth_providers** | `internal/user/repository.go`: CreateUser, GetUserByID, GetUserByEmail; CreateAuthProvider, FindByProviderAndSubject. | A1 |
| ✅ B2 | **Service Register** | `internal/user/service.go`: Register(datos + provider) → crear/actualizar User y UserAuthProvider. Sin envío de email aún. | B1 |
| ✅ B3 | **Handler POST /auth/register** | `internal/user/handler.go` + `routes.go`: recibir body, llamar Register, responder 201 (opcional: JWT limitado). | B2 |
| ✅ B4 | **Verificación de email** | Campo `email_verified_at` ya en User. Endpoint para “marcar verificado” (link o código) y/o reenviar verificación. Sin envío real de correo si no hay SMTP. | B3 |

### Fase 3 — Login y JWT

| # | Tarea | Alcance | Depende de |
|---|--------|---------|-------------|
| ✅ C1 | **Servicio JWT** | `internal/auth/service.go`: firmar token (user_id, email, exp); validar y extraer claims (`ValidateAccessToken`). Config (secret, expiración). | — |
|| ✅ C2 | **Service Login** | `internal/auth/service.go`: `Login()` resuelve User por email, valida password_hash, devuelve `TokenPair`. | B1, C1 |
|| ✅ C3 | **Handler POST /auth/login** | `internal/auth/handler.go`: body email+password; respuesta 200 con `access_token`, `refresh_token`, `expires_at`. 401 si credenciales inválidas. | C2 |
|| ✅ C4 | **Middleware de autenticación** | `internal/shared/middleware/auth.go`: `RequireAuth()` extrae Bearer JWT, valida firma/exp, inyecta userID y email en Fiber locals. 401 si inválido. | C1 |

### Fase 4 — Cuentas (Account) y membresía

| # | Tarea | Alcance | Depende de |
|---|--------|---------|-------------|
| ✅ D1 | **Repository accounts y account_members** | `internal/account/repository.go`: CreateAccount, GetByID, GetBySlug; CreateMember, GetMember(account_id, user_id), ListMembers. | A2 |
| ✅ D2 | **Service CreateAccount** | Crear Account (name, slug único) + AccountMember (user_id del JWT, rol `owner`). Comprobar que User existe y (según política) email verificado. | D1 |
| ✅ D3 | **Handler POST /accounts** | Ruta protegida con middleware auth. Body: name (y opcional slug). Crear cuenta y miembro; 201. | C4, D2 |
| D4 | **Contexto de cuenta (middleware/helper)** | Leer `account_id` o `slug` (header o query); validar membresía (AccountMember); inyectar account_id en contexto. Responder 403 si no miembro. | C4, D1 |

### Fase 5 — Contexto de petición y filtrado por Account

| # | Tarea | Alcance | Depende de |
|---|--------|---------|-------------|
| E1 | **Tipo RequestContext** | Struct (UserID, AccountID, etc.) y funciones para obtener desde Fiber context. Usar en handlers en lugar de leer JWT/header a mano. | C4, D4 |
| E2 | **Filtrado por Account en un recurso** | Aplicar filtro por `account_id` (del contexto) en un recurso existente (p. ej. invoices): list/get/create. Patrón para extender al resto. | E1, recurso existente |

### Resumen de dependencias (orden sugerido)

```
A1 → B1 → B2 → B3 → B4
A2 → D1 → D2 → D3
C1 → C2 → C3 ; C1 → C4
C4 + D1 → D4
C4 + D4 → E1 → E2
```

- **A1 y A2** se pueden hacer en paralelo (modelos independientes).
- **C1** es independiente; conviene hacerlo antes o en paralelo a B2.
- **E2** requiere al menos un recurso con `account_id`; si no existe, crear uno mínimo (ej. `internal/invoice/` con account_id) en una tarea previa o dentro de E2.

---

## Estado actual

- **Completado:** A1 → B1 → B2 → B3 → B4 → C1 → C2 → C3 → C4 → A2 → D1 → D2 → D3
- **Siguiente paso:** D4 (Contexto de cuenta — middleware/helper)
- **Última actualización:** 2026-02-20

