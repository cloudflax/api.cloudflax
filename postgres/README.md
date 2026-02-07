# PostgreSQL con SSL

Para habilitar SSL en la conexión con PostgreSQL:

1. Genera los certificados antes del primer `docker-compose up`:
   ```bash
   make db-certs
   ```

2. Reinicia los contenedores:
   ```bash
   docker-compose down && docker-compose up -d
   ```

Los certificados se guardan en `postgres/certs/` (no están en git).

## Variables de entorno

- `DB_SSL_MODE`: `require` (default), `verify-ca`, `verify-full`, o `disable`
  - `require`: cifra la conexión, no verifica el certificado
  - `verify-ca`: verifica el certificado contra la CA
  - `verify-full`: verifica certificado y hostname (requiere CA)
  - `disable`: sin SSL (solo desarrollo local)
