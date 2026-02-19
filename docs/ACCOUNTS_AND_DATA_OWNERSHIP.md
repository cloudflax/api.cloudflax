# Cuentas y propiedad de los datos

**Account** (cuenta/tenant) es la **dueña de los datos**. **User** es la identidad de la persona; puede pertenecer a varias Accounts. Toda la API filtra por Account y, cuando aplica, por User (roles, quién creó/emitió).

---

## 1. Registro de usuario (paso a paso)

Al registrarse, la API debe:

1. **Recibir datos del proveedor de autenticación**  
   No solo los que envía el cliente en el body: también los que devuelve el proveedor (email/contraseña hoy; OAuth mañana). Ejemplos: email, nombre, y para OAuth el `sub` (identificador externo) y el nombre del proveedor.

2. **Crear o vincular el User**  
   - Si es registro nuevo: crear fila en `users` con name, email (normalizado), y si aplica `password_hash`.  
   - Guardar el **proveedor** en `user_auth_providers`: `provider` (ej. `email`, `google`) y `provider_subject_id` (email para email; `sub` para OAuth). Así una misma identidad (User) puede tener varios proveedores y el login resuelve unívocamente.

3. **Obligar la validación de correo**  
   Tras el registro, el usuario **debe validar su correo electrónico** antes de poder completar el onboarding o usar la app con normalidad. La API:
   - Envía (o dispara) el correo de verificación con link o código.
   - Mantiene un estado de “email no verificado” (p. ej. `email_verified_at` en `users` o tabla de verificaciones).
   - No considera el onboarding completo ni permite crear Account hasta que el email esté verificado (o según política: bloquear todo hasta verificación).

**Resumen:** Registro = datos del cliente **+** datos del proveedor → crear/actualizar User y `user_auth_providers` → enviar verificación → usuario debe validar email.

---

## 2. Creación de la cuenta (paso a paso)

Solo cuando el User existe y (según política) **ha verificado su email**:

1. **Cliente envía nombre del equipo/empresa**  
   Ej. `POST /accounts` o `POST /onboarding/account` con `name` (y opcionalmente `slug`).

2. **API crea Account y AccountMember**  
   - Crea la Account (nombre + slug único).  
   - Crea AccountMember vinculando el `user_id` del JWT con la Account y rol por defecto (ej. `owner`).

3. **Cliente usa la app en contexto de Cuenta**  
   En cada petición envía `account_id` (o slug). La API valida membresía y filtra todo por esa Account.

---

## 3. Inicio de sesión (login) (paso a paso)

El login es el flujo por el que un User ya registrado obtiene un **JWT** para consumir la API:

1. **Cliente envía credenciales**  
   Ej. `POST /auth/login` (o `POST /login`) con email + password, o datos del proveedor OAuth (token/code).

2. **API valida contra User y UserAuthProvider**  
   - Resuelve la identidad con `user_auth_providers` (provider + provider_subject_id).  
   - Para `credentials`: verifica password contra `password_hash`.  
   - (Opcional) Exige `email_verified_at` no nulo para permitir login completo.

3. **API genera y devuelve el JWT**  
   - Firma un token con al menos `user_id` (y opcionalmente `email`, `account_id` activo, roles, expiración).  
   - Respuesta típica: `200 OK` con body `{ "access_token": "<JWT>", "token_type": "Bearer", "expires_in": ... }`.

4. **Cliente usa el JWT en peticiones posteriores**  
   Header `Authorization: Bearer <JWT>`; la API identifica al User y valida membresía cuando se envía `account_id`.

**Resumen:** Login = credenciales → validar User + proveedor → **generar JWT** → devolver token → cliente lo envía en cada petición.

---

## 4. Flujo de la API (resumen)

El flujo completo es: **Registro → Validación de email → Login (obtener JWT) → Crear cuenta → Uso con Account**.

### Cuándo se crea el JWT

| Momento | Qué pasa |
|--------|----------|
| **Registro** | Opcional: la API puede devolver un JWT limitado (p. ej. solo `user_id`, claim `email_verified: false`) para que el cliente pueda llamar a “reenviar verificación” o mostrar pantalla de verificación. Si no se devuelve JWT, el usuario debe hacer **login** después de verificar el email para obtenerlo. |
| **Login** | Siempre: tras validar credenciales (y política de email verificado), la API **genera el JWT**, lo firma y lo devuelve en la respuesta. Ese token es el que se usa para `POST /accounts`, para enviar `account_id` y para todas las peticiones autenticadas. |

En ambos casos el JWT se crea en la API (mismo servicio de firma/claims); la diferencia es **cuándo** se emite: en el registro (opcional) o en el login (habitual).

### Diagrama de secuencia del registro (Mermaid)

```mermaid
sequenceDiagram
  participant C as Cliente
  participant A as API
  participant S as Sistema (email)
  participant U as Usuario

  Note over C,U: 1. Registro
  C->>A: POST registro (email/password u OAuth)
  A->>A: Crear o actualizar User
  A->>A: Guardar en user_auth_providers
  A->>S: Disparar envío de correo de verificación
  S->>U: Email con link o código
  A->>C: 201 (opcional: JWT limitado o sin JWT)

  Note over C,U: 2. Validación de email
  U->>A: Clic en link o introduce código
  A->>A: Actualizar email_verified_at

  Note over C,U: 3. Login (obtener JWT)
  C->>A: POST /auth/login (credenciales)
  A->>A: Validar User + user_auth_providers
  A->>A: Generar JWT
  A->>C: 200 + access_token (JWT)

  Note over C,U: 4. Creación de cuenta
  C->>A: POST /accounts (nombre del equipo) [Authorization: Bearer JWT]
  A->>A: Crear Account + AccountMember (user_id del JWT)
  A->>C: 201 Created

  Note over C,U: 5. Uso con Account
  C->>A: Peticiones con header account_id + Bearer JWT
  A->>A: Validar JWT, membresía, filtrar por Account
  A->>C: Respuesta (datos en contexto de la Account)
```

### Diagrama de secuencia del login (Mermaid)

```mermaid
sequenceDiagram
  participant C as Cliente
  participant A as API

  C->>A: POST /auth/login (email + password u OAuth)
  A->>A: Resolver User por user_auth_providers
  A->>A: Validar password (si provider = credentials)
  alt Credenciales inválidas
    A->>C: 401 Unauthorized
  else OK (y email verificado si aplica)
    A->>A: Generar JWT (user_id, exp, etc.)
    A->>C: 200 OK { access_token, token_type, expires_in }
  end
  Note over C: Cliente guarda JWT y lo envía en Authorization
  C->>A: Peticiones con Authorization: Bearer <JWT>
  A->>A: Extraer user_id del JWT, validar firma/exp
  A->>C: Respuesta según recurso
```

| Paso | Quién | Qué hace la API |
|------|--------|------------------|
| 1 | Cliente | Registro (email/password u OAuth). | Crea/actualiza User; guarda proveedor en `user_auth_providers`; opcionalmente devuelve JWT limitado. |
| 2 | Sistema | — | Envía correo de verificación. Usuario debe validar. |
| 3 | Cliente | Login (credenciales). | Valida User + proveedor; **genera y devuelve JWT**. |
| 4 | Cliente | Envía nombre del equipo (crear cuenta) con JWT. | Crea Account + AccountMember (solo si email verificado; `user_id` del JWT). |
| 5 | Cliente | Peticiones con `account_id` y Bearer JWT. | Valida JWT y membresía; filtra por Account (y User cuando aplique). |

---

## 5. Modelo entidad-relación (base de datos)

### Entidades y atributos

| Entidad | Tabla | Atributos principales | Notas |
|--------|--------|------------------------|--------|
| **User** | `users` | `id` (PK), `name`, `email`, `password_hash`, `email_verified_at` (opcional), `created_at`, `updated_at`, `deleted_at` | Identidad. Vinculación con proveedores vía `user_auth_providers`. |
| **UserAuthProvider** | `user_auth_providers` | `id` (PK), `user_id` (FK), `provider`, `provider_subject_id`, `created_at`, `updated_at` | `provider` es enum (google, facebook, credentials, …). UNIQUE(provider, provider_subject_id). |
| **Account** | `accounts` | `id` (PK), `name`, `slug` (único), `created_at`, `updated_at`, `deleted_at` | Tenant; dueña de los datos. |
| **AccountMember** | `account_members` | `id` (PK), `account_id` (FK), `user_id` (FK), `role`, `created_at`, `updated_at` | UNIQUE(account_id, user_id). |
| **Recurso de negocio** | p. ej. `invoices` | `id` (PK), `account_id` (FK), `issued_by_user_id` (FK, nullable), … | Siempre `account_id`; opcional atribución a User. |

### Relaciones

- **User → UserAuthProvider:** 1:N. Un User puede tener varios proveedores (email, google, etc.).
- **User ↔ Account:** N:M mediante `account_members`.
- **Account → Recursos:** 1:N. Todo recurso tiene `account_id`.
- **User → Recursos (atribución):** opcional (`issued_by_user_id`, etc.); la propiedad es de la Account.

### Diagrama ER (Mermaid)

```mermaid
erDiagram
    users ||--o{ user_auth_providers : "tiene"
    users ||--o{ account_members : "es miembro"
    accounts ||--o{ account_members : "tiene miembros"
    accounts ||--o{ invoices : "posee"
    users ||--o{ invoices : "emite (atribución)"

    users {
        uuid id PK
        string name
        string email
        string password_hash
        timestamp email_verified_at
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    user_auth_providers {
        uuid id PK
        uuid user_id FK
        string provider "enum"
        string provider_subject_id
        timestamp created_at
        timestamp updated_at
    }

    accounts {
        uuid id PK
        string name
        string slug UK
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    account_members {
        uuid id PK
        uuid account_id FK
        uuid user_id FK
        string role
        timestamp created_at
        timestamp updated_at
    }

    invoices {
        uuid id PK
        uuid account_id FK
        uuid issued_by_user_id FK
        string number
        decimal total
        timestamp created_at
        timestamp updated_at
    }
```

**Nota:** En `user_auth_providers`, el campo `provider` es un **enum** con valores: `google`, `facebook`, `credentials`, etc.

---

## 6. Contexto de petición y autorización

- **Autenticación:** identificar al User (JWT).
- **Contexto de Cuenta:** el cliente envía `account_id` (o slug) en header o token; la API **valida membresía** y **filtra todas las lecturas/escrituras** por esa Account.
- **Atribución:** recursos pueden llevar `issued_by_user_id` / `created_by_user_id`; la propiedad sigue siendo de la Account.

Un User puede ser miembro de varias Accounts; el cliente elige la Account activa y la envía en cada petición.
