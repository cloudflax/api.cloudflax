# Roadmap | Auth y endurecimiento de seguridad

Documento maestro para **orden de ejecución**, **dependencias** y **qué posponer** sin perder el hilo. Cada fila puede tener su spec detallado en `docs/specs/auth-security/` (plantilla: `SPEC-TEMPLATE.md`).

## Leyenda de estado

| Estado        | Significado                                      |
|---------------|--------------------------------------------------|
| `pending`     | No iniciado; puedes saltarlo marcando `deferred` |
| `deferred`    | Consciente de que no se hará (por ahora)        |
| `spec_ready`  | Spec escrito / criterios claros                 |
| `done`        | Implementado y verificado en repo               |

## Principios de orden

1. **Menos debate de producto primero** (operativa, bugs claros, configuración).
2. Luego **abuso y coste** (rate limits, throttles).
3. Después **integridad de sesión** (refresh: carreras, reutilización).
4. **JWT y secretos** (claims, procedencia del secreto).
5. **Datos sensibles en BD** (tokens de verificación en claro).
6. **Filtraciones por canal HTTP** (GET + query, Referer).
7. **Cambios que rompen contrato o cruzan equipos** (cookie vs JSON, políticas de password/registro).

Ajusta el orden si tu frontend o infra obligan otra secuencia; mantén anotado el motivo en el spec.

---

## Fase 0 — Inventario y decisión (sin código obligatorio)

| ID   | Tema | Notas | Estado | Spec |
|------|------|--------|--------|------|
| **Z0** | Registrar ítems `deferred` | Por defecto: fases **D/E/F** y **C2** pueden esperar hasta priorizar producto/infra | `done` | — |
| **Z1** | Asegurar `APP_ENV=production` en prod y `ENABLE_AUTH_DEV_ENDPOINTS` desactivado | Staging: revisar flags explícitos | `pending` | opcional |

---

## Fase A — Operativa y superficie (bajo acoplamiento con el front)

| ID   | Tema | Dependencias | Estado | Spec |
|------|------|--------------|--------|------|
| **A1** | **Endpoint dev**: `ENABLE_AUTH_DEV_ENDPOINTS` + no producción; peek sin reenviar email | Ninguna | `done` | ver `AUTH_INTEGRATION.md` |
| **A2** | **Throttling “fail-open”**: documentar en runbook; opcionalmente fallar cerrado o alertar si falta tabla Dynamo | Config AWS | `pending` | `SPEC-A2.md` |
| **A3** | **Trust proxy Fiber** (`TRUST_PROXY`, `PROXY_HEADER`, rangos trust) | Infra | `done` | ver `.env.example` |
| **A4** | **Cabeceras de seguridad** (HSTS delegado a LB, `X-Content-Type-Options`, etc.) | Ninguna | `pending` | `SPEC-A4.md` |

---

## Fase B — Abuso: credenciales y refresh

| ID   | Tema | Dependencias | Estado | Spec |
|------|------|--------------|--------|------|
| **B1** | **Rate limit en `POST /auth/login`** (Dynamo por IP; misma tabla que throttle API) | A3 si usas IP | `done` | `internal/auth/ip_throttle_guard.go` |
| **B2** | **Rate limit en `POST /auth/refresh`** | Igual que B1 | `done` | `internal/auth/ip_throttle_guard.go` |
| **B3** | **Bloqueo / backoff por intentos fallidos de login** (por cuenta o email, distinto del límite por IP): contador, ventana, lockout temporal, desbloqueo (tiempo / email / admin) y mismo `INVALID_CREDENTIALS` en API | Producto + **migración** probable en `users` o store aparte; definir interacción con B1 | `pending` | `SPEC-B3.md` |

*Notas:* B1/B2 (por IP) y **B3** (fallos por cuenta/email) son complementarias. B1/B2 ya comparten implementación en código; B3 requiere spec propio (`SPEC-B3.md`) antes de implementar.

---

## Fase C — Refresh: integridad y sesión

| ID   | Tema | Dependencias | Estado | Spec |
|------|------|--------------|--------|------|
| **C1** | **Transacción / bloqueo en refresh** (`UPDATE` atómico: revocar + insertar refresh nuevo en una TX) | Ninguna crítica | `done` | `internal/auth/repository.go` `ConsumeRefreshByTokenHash` |
| **C2** | **Política ante reutilización de refresh** (ej. revocar cadena o todos los refresh del usuario al detectar token ya revocado) | C1 recomendado antes | `pending` | `SPEC-C2.md` |

---

## Fase D — JWT y secretos

| ID   | Tema | Dependencias | Estado | Spec |
|------|------|--------------|--------|------|
| **D1** | **`iss` / `aud` (y opcional `nbf`)** en access JWT; validación estricta al parsear | Coordinación si hay varios consumidores | `pending` | `SPEC-D1.md` |
| **D2** | **Política de `JWT_SECRET`**: longitud mínima en `Validate()`, y/o origen en Secrets Manager | Config / despliegue | `pending` | `SPEC-D2.md` |
| **D3** | **Opcional:** `jti` en access + denylist (solo si hay requisito de revocación antes de exp); suele ser P2/P3 | D1 | `pending` | `SPEC-D3.md` |

---

## Fase E — Verificación de email (BD y transporte)

| ID   | Tema | Dependencias | Estado | Spec |
|------|------|--------------|--------|------|
| **E1** | **Token de verificación: no almacenar en claro** (hash como refresh) o flujo de un solo uso equivalente | Migración posible | `pending` | `SPEC-E1.md` |
| **E2** | **Evitar token en query en GET** (p. ej. POST con body, o flujo front que no exponga a Referer) | **Frontend** | `pending` | `SPEC-E2.md` |

---

## Fase F — Contrato API / producto (valorar “deferred” largo)

| ID   | Tema | Dependencias | Estado | Spec |
|------|------|--------------|--------|------|
| **F1** | **Refresh: cookie `httpOnly` + `Set-Cookie`** vs solo JSON — alinear con `AUTH_INTEGRATION.md` | **Frontend**, CORS, SameSite | `pending` | `SPEC-F1.md` |
| **F2** | **Registro: no enumerar emails** (misma respuesta genérica que resend) | Producto / UX | `pending` | `SPEC-F2.md` |
| **F3** | **Política de contraseña** (complejidad, breach list, etc.) | Producto | `pending` | `SPEC-F3.md` |
| **F4** | **Claim `email` en JWT vs BD**: revalidar en acciones sensibles o acortar access + documentar | D1 opcional | `pending` | `SPEC-F4.md` |

---

## Convención de commits

Un commit (o PR) por **ID** cuando sea posible; si un cambio toca varios IDs, referencia todos en el cuerpo del commit.

Ejemplos:

- `feat(auth): add login rate limit (B1)`
- `fix(auth): transactional refresh token rotation (C1)`

---

## Relación con documentación existente

- Contrato actual front–back: [`AUTH_INTEGRATION.md`](./AUTH_INTEGRATION.md)
- Tras cambios de contrato, actualizar ese doc en el mismo PR o inmediatamente después.

---

## Checklist rápido antes de cerrar un ítem

- [ ] Spec con criterios de aceptación y pruebas
- [ ] `make lint` y `make test`
- [ ] `.env` / migraciones documentados si aplica
- [ ] Roadmap: estado `done` + enlace a commit o PR
- [ ] Sin tokens ni contraseñas en logs (regla del repo)
