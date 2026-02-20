# Auth Integration — Cloudflax API

Documento de referencia para integrar autenticación entre el frontend Next.js y este backend Go/Fiber.

---

## Estrategia de autenticación

| Token | Tipo | Duración | Almacenamiento recomendado (frontend) |
|-------|------|----------|---------------------------------------|
| Access token | JWT (HS256) | 15 minutos | Memoria (variable de estado / React context) |
| Refresh token | Token opaco (random hex) | 7 días | `httpOnly` cookie |

El refresh token se almacena **hasheado (SHA-256)** en la tabla `refresh_tokens` de PostgreSQL.
El access token **nunca** se persiste en base de datos.

---

## Endpoints

### Base URL

```
http://localhost:3000
```

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
| 422 | `VALIDATION_ERROR` | Email inválido o password < 8 chars |

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
| 401 | `TOKEN_INVALID` | Token inválido, ya usado o expirado |

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
| 401 | `UNAUTHORIZED` | Access token ausente o inválido |
| 401 | `TOKEN_INVALID` | Access token malformado o expirado |

---

## Rutas protegidas

Todas las rutas bajo `/users` requieren el header `Authorization: Bearer <access_token>`.

```http
GET    /users
GET    /users/:id
POST   /users
PUT    /users/:id
DELETE /users/:id
```

Sin el header o con token inválido el backend responde `401 UNAUTHORIZED`.

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
| `exp` | Timestamp de expiración (15 min) |

Algoritmo: **HS256**

---

## Flujo completo (Next.js ↔ Go Fiber)

```
[Next.js]                          [Fiber API]

1. Login
   POST /auth/login ──────────────► Valida credenciales
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

### Frontend (Next.js)

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `NEXT_PUBLIC_API_URL` | URL base del backend (para llamadas del cliente) | `http://localhost:3000` |
| `API_URL` | URL base del backend (para llamadas server-side) | `http://localhost:3000` |

---

## Códigos de error auth

| `error.code` | Status | Descripción |
|---|---|---|
| `INVALID_CREDENTIALS` | 401 | Email o password incorrecto en login |
| `UNAUTHORIZED` | 401 | Endpoint protegido sin token |
| `TOKEN_INVALID` | 401 | Token malformado, firmado con otro secret, o ya revocado |
| `TOKEN_EXPIRED` | 401 | Access token expirado (también devuelve `TOKEN_INVALID`) |

---

## CORS

El backend debe tener CORS configurado para aceptar requests desde el origen del frontend. Configurar en `internal/bootstrap/app/app.go` con el middleware de Fiber:

```go
import "github.com/gofiber/fiber/v3/middleware/cors"

app.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3001"}, // origen del frontend
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowCredentials: true, // necesario para cookies
}))
```

---

## Checklist de integración

### Backend (completado)
- [x] `POST /auth/login` — devuelve `access_token` + `refresh_token`
- [x] `POST /auth/refresh` — rota el refresh token
- [x] `POST /auth/logout` — revoca todos los refresh tokens del usuario
- [x] Middleware JWT — protege rutas `/users`
- [x] Refresh token rotation — el token anterior se invalida al renovar
- [x] Refresh tokens en DB — tabla `refresh_tokens` con hash SHA-256

### Frontend (pendiente)
- [ ] Función de login que llame a `POST /auth/login`
- [ ] Guardar `access_token` en memoria (context/store)
- [ ] Guardar `refresh_token` como `httpOnly` cookie (vía Route Handler)
- [ ] Interceptor HTTP que adjunte el `Authorization: Bearer` header
- [ ] Lógica de refresh automático cuando el server devuelve `401`
- [ ] Función de logout que llame a `POST /auth/logout` y limpie el estado
- [ ] Rutas protegidas en Next.js (middleware `middleware.ts` con verificación de token)
- [ ] Configurar CORS en el backend para el origen del frontend
