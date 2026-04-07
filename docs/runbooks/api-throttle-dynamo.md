# Runbook | API throttle (DynamoDB)

## Qué hace

Con `API_THROTTLE_TABLE_NAME` apuntando a una tabla DynamoDB, la API aplica límites para:

- reenvío de email de verificación;
- forgot-password por email;
- `POST /auth/login` y `POST /auth/refresh` por IP (misma tabla, claves con prefijos distintos).

Si el nombre de tabla está **vacío**, esos límites **no se aplican** (comportamiento *fail-open* a propósito: la API sigue operativa sin Dynamo).

## Arranque

| Situación | Comportamiento por defecto |
|-----------|----------------------------|
| Tabla vacía | Sin guards; sin throttles anteriores. |
| Tabla definida, init OK | Guards activos. |
| Tabla definida, error al crear cliente Dynamo | Log `WARN` con `event=api_throttle_guard_init_failed` y `component` (`resend_verification`, `forgot_password`, `login_ip`, `refresh_ip`); el proceso **sigue** y ese guard concreto no se monta (*fail-open*). |

Con **`API_THROTTLE_STRICT_INIT=true`** y tabla no vacía, cualquier error de init de un guard hace **fallar el arranque** (*fail-closed* para no servir sin protección cuando se esperaba Dynamo).

Con **`API_THROTTLE_REQUIRED_IN_PRODUCTION=true`** y `APP_ENV=production`, un `API_THROTTLE_TABLE_NAME` vacío hace **fallar la validación de config** al arrancar.

## Peticiones en runtime

Si el guard está montado y **Dynamo falla** en `GetItem`/`PutItem`, los handlers de login/refresh (u otros) responden **500** y se registra el error. Eso es *fail-closed* por petición: no se omite el control por un fallo intermitente de almacenamiento.

## Alertas sugeridas

Basar reglas en logs estructurados (`slog`):

- **Init degradado:** `event=api_throttle_guard_init_failed` — indica credenciales, red, o política IAM incorrecta si esperabas throttling.
- Opcional: tasa de `login ip throttle` / `refresh ip throttle` / `resend verification throttle` / `forgot password throttle` con nivel `ERROR` (fallos de Dynamo en runtime).

## Comprobaciones

1. Variable `API_THROTTLE_TABLE_NAME` acorde al entorno (misma cuenta/región que la API).
2. IAM de la tarea/Lambda/EC2: `dynamodb:GetItem`, `dynamodb:PutItem` sobre la tabla (y condiciones si usáis ABAC).
3. Si usáis endpoint local (`AWS_ENDPOINT_URL`), que el servicio Dynamo esté accesible desde el runtime.

## Referencias

- Código: `internal/bootstrap/server/throttle_guards.go`, guards en `internal/auth/*_guard.go`.
- Variables: tabla en [`AUTH_INTEGRATION.md`](../AUTH_INTEGRATION.md).
