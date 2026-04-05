# Setup del proyecto Cloudflax API — progreso y trazabilidad

Documento único para seguimiento de estado y requisitos en diseño o implementación.

---

## Forgot password (recuperación de contraseña)

**Estado:** implementado en API (GORM migra columnas en `users`; Lambda + SES en runtime).  
**Última revisión:** 2026-04-02.

### Entrega del correo (paridad con auth-verify-email)

El reset **no** llama a SES desde la API. El mecanismo es el **mismo que el correo de verificación de cuenta**:

1. La API construye un **JSON de evento** (destinatario, nombre, enlace, etc.).
2. Se invoca la **función Lambda dedicada** con **`InvocationType` Event** (asíncrono), con la misma configuración AWS que el resto de la app (`AWS_REGION`, credenciales, `AWS_ENDPOINT_URL` si aplica). Variable de entorno: **`LAMBDA_SEND_FORGOT_PASSWORD_EMAIL_NAME`** (dev: **`cloudflax-dev-send-forgot-password-email`**).
3. Esa **Lambda** llama a SES (**SendEmail** con **Template** + **TemplateData**).

**Plantilla SES en AWS (creada):** `auth-forgot-password`.  
La Lambda **`cloudflax-dev-send-forgot-password-email`** debe usar esa plantilla al enviar (paridad de patrón con la Lambda de verify, pero **función distinta**).

### Objetivo

Permitir que un usuario con credencial **email + contraseña** solicite recuperación cuando no puede iniciar sesión, usando el **mismo canal** que verificación: **Lambda asíncrona + plantilla SES** (`auth-forgot-password`).

### Alcance

| Incluido | Fuera de alcance (por ahora) |
|----------|------------------------------|
| Usuarios con proveedor `credentials` y email verificado (o política explícita si se permite antes de verificar) | Cambio de contraseña ya autenticado (`PUT /users/me`) — ya existe |
| Envío por correo con plantilla SES | SMS / OTP fuera de email |
| Respuesta API uniforme (no enumerar emails) | Detalle de implementación Lambda interna (solo contrato de entrada) |

### Flujo lógico (alto nivel)

1. Cliente llama a **solicitud de reset** (p. ej. `POST /auth/forgot-password`) con `email`.
2. API valida formato, aplica **rate limiting**, y si existe usuario elegible: genera **token opaco de un solo uso**, caducidad, guarda **hash** en BD, invalida intentos anteriores no consumidos del mismo usuario si aplica.
3. API invoca **`cloudflax-dev-send-forgot-password-email`** (config: `LAMBDA_SEND_FORGOT_PASSWORD_EMAIL_NAME`), en modo **evento** / asíncrono, con un JSON de carga útil acordado.
4. Lambda llama a SES **`SendEmail`** con plantilla **`auth-forgot-password`** y **TemplateData** (JSON con las variables del template).
5. Cliente abre el **enlace del correo** (front), envía **token + nueva contraseña** a **confirmación** (p. ej. `POST /auth/reset-password`).
6. API valida token, actualiza hash de contraseña, marca token consumido, **revoca refresh tokens** del usuario.

Respuesta de solicitud: siempre el mismo mensaje genérico (éxito aparente), independientemente de si el email existe en el sistema.

### Requisitos funcionales (trazabilidad)

| ID | Requisito | Criterio de aceptación |
|----|-----------|-------------------------|
| FR-FP-01 | Solicitud con email | Endpoint documentado; validación de email; respuesta única para “existe / no existe” |
| FR-FP-02 | Token seguro | Valor aleatorio opaco; solo hash almacenado; un solo uso; expiración configurable |
| FR-FP-03 | Correo SES plantilla | Igual que verify-email: API → Lambda async → SES; plantilla `auth-forgot-password` |
| FR-FP-04 | Confirmación de reset | Endpoint con token + nueva contraseña; política de contraseña alineada con registro |
| FR-FP-05 | Sesiones | Tras reset exitoso, sesiones previas invalidadas (revocación de refresh tokens) |
| FR-FP-06 | Abuso | Límite por IP y/o email (coherente con `resend-verification`) |

### Requisitos no funcionales

- **Privacidad:** no filtrar existencia de cuenta ni estado de verificación en mensajes públicos.
- **Coherencia:** nombres de variables de plantilla alineados con el flujo de **verificación de email** (`name`, `link`, `email`) para reducir fricción en Lambda y plantillas.
- **Operación:** plantilla SES **`auth-forgot-password`**; Lambda dedicada configurable con **`LAMBDA_SEND_FORGOT_PASSWORD_EMAIL_NAME`** (ej. `cloudflax-dev-send-forgot-password-email`).

### Variables para la plantilla SES `auth-forgot-password` (`TemplateData`)

El **JSON** que la Lambda pasa a SES como **TemplateData** para **`auth-forgot-password`** sigue el mismo estilo que el payload hacia Lambda en registro (`email`, `name`, `link`), más **`expiresIn`** si la plantilla lo usa.

| Variable (clave JSON) | Tipo | Descripción |
|----------------------|------|-------------|
| `email` | string | Dirección del destinatario (útil para Lambda y, si la plantilla lo muestra, “enviado a…”). |
| `name` | string | Nombre para saludo personalizado (mismo criterio que en registro). |
| `link` | string | URL absoluta al front para completar el reset, p. ej. `{FRONTEND_URL}/auth/reset-password?token={token_opaco}`. |
| `expiresIn` | string | Texto legible para el cuerpo del correo, p. ej. `"60 minutes"` / `"1 hour"` (idioma acordado con marketing). |

**Ejemplo de `TemplateData`** (cadena JSON para SES):

```json
{
  "email": "user@example.com",
  "name": "Ada",
  "link": "https://app.example.com/auth/reset-password?token=…",
  "expiresIn": "60 minutes"
}
```

**Nota:** En registro/verificación, la carga hacia Lambda hoy incluye `email`, `name`, `link` (ver `verificationPayload` en `internal/shared/verificationnotify/lambda.go`). Para reset, se **extiende** con `expiresIn` para que la plantilla SES pueda mostrar la caducidad sin calcularla en la plantilla.

### Dependencias y seguimiento

| Ítem | Estado |
|------|--------|
| Columnas `password_reset_token_hash` / `password_reset_expires_at` en `users` (AutoMigrate) | Listo |
| Lambda **`cloudflax-dev-send-forgot-password-email`** + plantilla SES **`auth-forgot-password`** | Infra AWS (operación) |
| `POST /auth/forgot-password`, `POST /auth/reset-password` | Listo |
| Throttle DynamoDB (`THROTTLE#FORGOT_PASSWORD#EMAIL#…`, misma tabla que resend) | Listo |
| Tests (servicio + handler) | Listo |

### Referencias en código

- Notifier reset: `internal/shared/verificationnotify/lambda.go` (`NotifyPasswordResetEmail`, payload `email`, `name`, `link`, `expiresIn`).
- Bootstrap: `internal/bootstrap/server/routes.go` (`newPasswordResetNotifier`, `NewDynamoForgotPasswordGuard`).
- Servicio y rutas: `internal/auth/service.go`, `internal/auth/handler.go`, `internal/auth/routes.go`.
