# Auth Integration | Cloudflax - Backend

Documento de referencia para integrar autenticación entre el frontend Next.js y este backend Go/Fiber.

---

## Estrategia de autenticación

| Token | Tipo | Duración | Almacenamiento recomendado (frontend) |
|-------|------|----------|---------------------------------------|
| Access token | JWT (HS256) | 15 minutos | Memoria (variable de estado / React context) |
| Refresh token | Token opaco (random hex) | 7 días | `httpOnly` cookie |

El refresh token se almacena **hasheado (SHA-256)** en la tabla `refresh_tokens` de PostgreSQL.
El access token **nunca** se persiste en base de datos.

**Verificación de email:** `POST /auth/login` y `POST /auth/refresh` solo tienen éxito si el usuario tiene el email verificado. Si no, responden `403` con `EMAIL_VERIFICATION_REQUIRED`.

---

## Endpoints

### Base URL

URL base de **esta API** (no la del frontend Next.js), p. ej. según `PORT` / `APP_URL` en el entorno. En desarrollo suele coincidir con el valor de `.env.example` (p. ej. `http://localhost:3000` si `PORT=3000`).

---

### POST `/auth/register`

Crea usuario con credenciales email/contraseña y envía (si está configurado) el correo de verificación. No devuelve tokens.

**Request:**

```http
POST /auth/register
Content-Type: application/json

{
  "name": "Ada Lovelace",
  "email": "user@example.com",
  "password": "password123"
}
```

**Response 201 Created:**

```json
{
  "data": { "id": "…", "name": "…", "email": "…" },
  "meta": { "email_verification_required": true }
}
```

**Errores posibles:**

| Status | `error.code` | Causa |
|--------|-------------|-------|
| 400 | `INVALID_REQUEST_BODY` | Body no es JSON válido |
| 409 | `EMAIL_ALREADY_EXISTS` | Email ya registrado |
| 422 | `VALIDATION_ERROR` | Validación de campos |

---

### GET `/auth/verify-email`

Confirma el email con el token del enlace (query). El frontend puede redirigir a esta URL con `token` en la query.

**Request:**

```http
GET /auth/verify-email?token=<uuid>
```

**Response 200 OK:**

```json
{ "message": "Email verified successfully" }
```

**Errores posibles:**

| Status | `error.code` | Causa |
|--------|-------------|-------|
| 400 | `INVALID_REQUEST_BODY` / `VALIDATION_ERROR` | Query inválida |
| 422 | `INVALID_VERIFICATION_TOKEN` | Token inválido o expirado |

---

### POST `/auth/resend-verification`

Solicita un nuevo correo de verificación. Por diseño no revela si el email existe (`200` en ambos casos cuando el flujo es correcto).

**Request:**

```json
{ "email": "user@example.com" }
```

**Response 200 OK:**

```json
{ "message": "If the email exists, a verification link has been sent" }
```

**Errores posibles:**

| Status | `error.code` | Causa |
|--------|-------------|-------|
| 409 | `EMAIL_ALREADY_VERIFIED` | El email ya estaba verificado |

---

### POST `/auth/login`

Autentica al usuario y devuelve un par de tokens.

**Request:**

```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response 200 OK:**

```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "a3f8b2c1d4e5f6a7b8c9d0e1f2a3b4c5...",
    "expires_at": "2026-02-18T12:15:00Z"
  }
}
```

**Errores posibles:**

| Status | `error.code` | Causa |
|--------|-------------|-------|
| 400 | `INVALID_REQUEST_BODY` | Body no es JSON válido |
| 401 | `INVALID_CREDENTIALS` | Email o password incorrectos |
| 403 | `EMAIL_VERIFICATION_REQUIRED` | Cuenta sin email verificado |
| 422 | `VALIDATION_ERROR` | Email inválido o password con menos de 8 caracteres |
| 429 | `RATE_LIMITED` | Demasiados intentos desde la misma IP (si `API_THROTTLE_TABLE_NAME` está configurada). Cabecera `Retry-After`. |

---

### POST `/auth/refresh`

Intercambia un refresh token válido por un nuevo par de tokens. El refresh token anterior **queda invalidado** (rotación).

**Request:**

```http
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "a3f8b2c1d4e5f6a7b8c9d0e1f2a3b4c5..."
}
```

**Response 200 OK:** (mismo formato que `/auth/login`)

```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "b7e9c3d1f0a2e4c6b8d0f2a4c6e8b0d2...",
    "expires_at": "2026-02-18T12:30:00Z"
  }
}
```

**Errores posibles:**

| Status | `error.code` | Causa |
|--------|-------------|-------|
| 400 | `INVALID_REQUEST_BODY` | Body no es JSON válido |
| 400 | `REFRESH_TOKEN_WRONG_FORMAT` | Se envió el JWT de acceso en lugar del refresh opaco |
| 401 | `TOKEN_INVALID` | Refresh inválido, ya usado o expirado |
| 403 | `EMAIL_VERIFICATION_REQUIRED` | Usuario sin email verificado |
| 429 | `RATE_LIMITED` | Demasiados refresh desde la misma IP (misma condición que login). Cabecera `Retry-After`. |

---

### POST `/auth/logout`

Revoca todos los refresh tokens activos del usuario autenticado.

**Request:**

```http
POST /auth/logout
Authorization: Bearer <access_token>
```

**Response 204 No Content** (sin body)

**Errores posibles:**

| Status | `error.code` | Causa |
|--------|-------------|-------|
| 401 | `UNAUTHORIZED` | Sin header `Authorization` o formato distinto de `Bearer` |
| 401 | `TOKEN_INVALID` | JWT ausente en el sentido correcto, malformado, firma inválida o expirado |

---

### POST `/auth/dev/verify-email-token` (solo desarrollo explícito)

Solo se monta si **`ENABLE_AUTH_DEV_ENDPOINTS`** es verdadero (`true` / `1` / `yes` / `on`) **y** `APP_ENV` ≠ `production`. Devuelve el **token de verificación ya existente** (no envía correo ni lo rota). Si no hay token pendiente, responde `422` con `INVALID_VERIFICATION_TOKEN` y mensaje orientativo.

---

## Rutas protegidas

Requieren `Authorization: Bearer <access_token>` (middleware JWT). Sin token válido: `401` con `UNAUTHORIZED` o `TOKEN_INVALID`.

**Usuarios (perfil del usuario autenticado):**

```http
GET    /users/me
GET    /users/me/accounts
PUT    /users/me
DELETE /users/me
```

**Cuentas:**

```http
POST   /accounts
POST   /accounts/active
```

**Facturas:** prefijo `/invoices` con autenticación **y** pertenencia a la cuenta activa (middleware adicional). Ver [ARCHITECTURE.md](./ARCHITECTURE.md) / código de `invoice` para el detalle.

---

## Formato de error estándar

Todos los errores siguen esta estructura:

```json
{
  "error": {
    "code": "TOKEN_INVALID",
    "message": "Invalid or expired token",
    "status": 401,
    "trace_id": "req-abc-123"
  }
}
```

Para errores de validación, `error.details` contiene los campos fallidos:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "status": 422,
    "trace_id": "req-abc-123",
    "details": [
      { "field": "email", "message": "must be a valid email address" },
      { "field": "password", "message": "must be at least 8 characters" }
    ]
  }
}
```

---

## JWT — Estructura del access token

El payload del JWT contiene:

| Campo | Descripción |
|-------|-------------|
| `user_id` | UUID del usuario |
| `email` | Email del usuario |
| `sub` | Igual a `user_id` (estándar JWT) |
| `iat` | Timestamp de emisión |
| `exp` | Timestamp de expiración (por defecto 15 min; configurable con `JWT_ACCESS_TOKEN_DURATION_MINUTES`) |

Algoritmo: **HS256**

---

## Flujo completo (Next.js ↔ Go Fiber)

```
[Next.js]                          [Fiber API]

0. Registro y verificación (primera vez)
   POST /auth/register ───────────► Crea usuario, envía email (Lambda si está configurado)
   Usuario abre enlace ───────────► GET /auth/verify-email?token=...

1. Login
   POST /auth/login ──────────────► Valida credenciales y email verificado
                    ◄────────────── { access_token, refresh_token }

2. Guardar tokens
   access_token → React state / memory
   refresh_token → httpOnly cookie

3. Llamada autenticada
   GET /users
   Authorization: Bearer <access_token> ──► Middleware valida JWT
                                       ◄─── { data: [...] }

4. Access token expirado (interceptor en Next.js)
   POST /auth/refresh
   { refresh_token } ─────────────► Valida y rota token
                    ◄────────────── { access_token, refresh_token (nuevo) }

5. Logout
   POST /auth/logout
   Authorization: Bearer <access_token> ──► Revoca todos los refresh tokens
                                       ◄─── 204 No Content
   Limpiar estado local
```

---

## Implementación sugerida en Next.js

### Interceptor con Axios (o fetch wrapper)

```typescript
// lib/api.ts
import axios from 'axios'

const api = axios.create({ baseURL: process.env.NEXT_PUBLIC_API_URL })

// Adjuntar access token en cada request
api.interceptors.request.use((config) => {
  const token = getAccessToken() // desde React state / context
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// Renovar access token automáticamente cuando expire
api.interceptors.response.use(
  (res) => res,
  async (error) => {
    if (error.response?.status === 401 && !error.config._retry) {
      error.config._retry = true
      const newPair = await refreshTokens() // POST /auth/refresh con cookie
      setAccessToken(newPair.access_token)
      error.config.headers.Authorization = `Bearer ${newPair.access_token}`
      return api(error.config)
    }
    return Promise.reject(error)
  }
)
```

### Enviar refresh token como cookie httpOnly

El backend no maneja cookies directamente (recibe el token en el body). El frontend debe:

1. Al recibir el `refresh_token` en login, guardarlo como `httpOnly` cookie **desde un Server Action o Route Handler de Next.js** (no desde el cliente).
2. Al llamar a `/auth/refresh`, leer la cookie en el servidor y enviarla en el body.

```typescript
// app/api/auth/refresh/route.ts (Next.js Route Handler)
import { cookies } from 'next/headers'
import { NextResponse } from 'next/server'

export async function POST() {
  const refreshToken = cookies().get('refresh_token')?.value
  if (!refreshToken) return NextResponse.json({ error: 'No refresh token' }, { status: 401 })

  const res = await fetch(`${process.env.API_URL}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken }),
  })

  const data = await res.json()
  if (!res.ok) return NextResponse.json(data, { status: res.status })

  const response = NextResponse.json({ access_token: data.data.access_token })
  response.cookies.set('refresh_token', data.data.refresh_token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 7 * 24 * 60 * 60,
  })
  return response
}
```

---

## Variables de entorno

### Backend (Go Fiber)

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `JWT_SECRET` | Clave de firma del JWT. Mínimo 32 chars aleatorios. | `openssl rand -hex 32` |
| `FRONTEND_URL` | Origen del frontend: CORS (`AllowOrigins`) y enlaces `.../auth/verify-email?token=` en el correo. | `http://localhost:3001` |
| `JWT_ACCESS_TOKEN_DURATION_MINUTES` | Duración del access token (minutos). Por defecto `15`. | `15` |
| `LAMBDA_SEND_VERIFY_EMAIL_NAME` | Nombre de la función Lambda que envía el email de verificación; vacío → no se envía correo (notifier noop). | — |
| `APP_ENV` | Entorno; con `production` no se montan rutas `/auth/dev/*` aunque el flag de abajo esté en true. | `development` |
| `ENABLE_AUTH_DEV_ENDPOINTS` | `true` para exponer `POST /auth/dev/verify-email-token` fuera de producción. En producción debe ser `false` o ausente. | `true` (solo local) |
| `TRUST_PROXY` | `true` si la API está detrás de proxy y se debe tomar la IP del cliente desde `PROXY_HEADER` (p. ej. throttling por IP). | `false` |
| `PROXY_HEADER` | Cabecera de IP del cliente cuando `TRUST_PROXY=true`. | `X-Forwarded-For` |
| `TRUST_PROXY_TRUST_PRIVATE` | Confiar en proxies en rangos RFC1918 (típico detrás de ALB en VPC). | `true` |
| `TRUST_PROXY_TRUST_LOOPBACK` | Confiar en loopback (útil en algunos setups locales con proxy). | `false` |
| `API_THROTTLE_TABLE_NAME` | Tabla DynamoDB para límites: resend, forgot-password, **login y refresh por IP**. Vacío → esos throttles desactivados. | — |

### Frontend (Next.js)

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `NEXT_PUBLIC_API_URL` | URL base de la **API** (llamadas desde el cliente) | Debe coincidir con el host/puerto del backend |
| `API_URL` | URL base de la **API** (llamadas server-side / Route Handlers) | Mismo criterio |

---

## Códigos de error auth

| `error.code` | Status | Descripción |
|---|---|---|
| `INVALID_CREDENTIALS` | 401 | Email o password incorrecto en login |
| `EMAIL_VERIFICATION_REQUIRED` | 403 | Login o refresh con email aún no verificado |
| `UNAUTHORIZED` | 401 | Endpoint protegido sin `Authorization` o sin esquema `Bearer` |
| `TOKEN_INVALID` | 401 | JWT de acceso malformado, firma incorrecta o expirado; también refresh inválido/revocado en `/auth/refresh` |
| `REFRESH_TOKEN_WRONG_FORMAT` | 400 | Se envió un JWT como `refresh_token` en lugar del token opaco |
| `RATE_LIMITED` | 429 | Login o refresh con throttle por IP (Dynamo); incluye `Retry-After` |
| `TOKEN_EXPIRED` | — | Definido en la API; el middleware de acceso actual devuelve `TOKEN_INVALID` cuando el JWT expira |

---

## CORS

El origen permitido se toma de **`FRONTEND_URL`** y se aplica en el arranque vía `middleware.CORS` en `internal/bootstrap/app/app.go` (implementación en `internal/shared/middleware/cors.go`): un solo origen explícito, métodos GET/POST/PUT/PATCH/DELETE/OPTIONS, headers `Origin`, `Content-Type`, `Accept`, `Authorization`, `X-Requested-With`. Si `FRONTEND_URL` está vacío, no se permite ningún origen (fail-closed).

---

## Checklist de integración

### Backend (completado)
- [x] `POST /auth/register` — alta y flujo de verificación por email
- [x] `GET /auth/verify-email` — confirma email con token en query
- [x] `POST /auth/resend-verification` — reenvío de correo de verificación
- [x] `POST /auth/login` — devuelve `access_token` + `refresh_token` (requiere email verificado)
- [x] `POST /auth/refresh` — rota el refresh token (requiere email verificado)
- [x] `POST /auth/logout` — revoca todos los refresh tokens del usuario
- [x] Middleware JWT — protege rutas de usuario, cuenta e invoice según el router
- [x] Refresh token rotation — el token anterior se invalida al renovar
- [x] Refresh tokens en DB — tabla `refresh_tokens` con hash SHA-256
- [x] CORS — origen desde `FRONTEND_URL`
