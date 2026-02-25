# Cloudflax Backend 🚀

**Cloudflax** es el motor central de una plataforma de ecommerce integral. Este backend, construido con **Go**, gestiona desde la experiencia de compra del cliente hasta el control operativo de inventarios, métricas y logística de envío en tiempo real.

---

## 📦 Características Principales

* **Gestión de Catálogo:** API robusta para productos con filtros avanzados por categorías, tallas y colores.
* **Trazabilidad Extrema:** Seguimiento detallado del flujo logístico (Empacado → Despachado → En ruta → Entregado).
* **Gestión de Inventario (Stock):** Control automatizado de existencias para evitar sobreventas y optimizar el almacén.
* **Métricas y CRM:** Panel administrativo para el control de ventas, análisis de marketing y gestión centralizada de clientes.
* **Pasarela de Pagos:** Integración para procesamiento de transacciones seguras en línea.
* **Historial de Usuario:** Consulta de compras anteriores y seguimiento de envíos en tiempo real.

---

## 🛠️ Stack Tecnológico

* **Lenguaje:** [Go (Golang)](https://go.dev/)
* **Framework Web:** [Fiber](https://gofiber.io/)
* **Base de Datos:** PostgreSQL + [GORM](https://gorm.io/) (ORM)
* **Contenedores:** Docker & **DevContainers** (Entorno de desarrollo estandarizado).
* **Infraestructura (AWS):**
    * **EC2:** Hosting del servidor principal.
    * **RDS:** Instancia gestionada de PostgreSQL.
    * **S3:** Almacenamiento de imágenes y activos.
    * **CloudFront:** CDN para entrega de contenido global.
    * **Lambda:** Procesos asíncronos y tareas específicas.

---

## 📁 Estructura del Proyecto

Enfoque **feature-driven**: cada recurso tiene su carpeta con handler, service, repository, model, dto, routes. Lo común está en `shared/`.

```
├── cmd/api/              # Entry point de la API y migraciones
├── internal/
│   ├── bootstrap/        # Arranque y configuración
│   │   ├── app/          # Fiber app, Run()
│   │   ├── config/       # Variables de entorno y secrets
│   │   └── server/       # Router principal, montaje de rutas, handlers raíz
│   ├── shared/           # Código compartido
│   │   ├── database/     # Conexión GORM + PostgreSQL
│   │   ├── logger/       # slog (logging estructurado JSON)
│   │   ├── middleware/   # Auth, account, logger de requests
│   │   ├── validator/    # Validaciones comunes
│   │   ├── email/        # Envío de correos (SES)
│   │   └── ...
│   ├── auth/             # Autenticación (login, registro, tokens)
│   ├── user/             # Usuarios
│   ├── account/          # Cuentas/organizaciones
│   └── invoice/          # Facturas
├── postgres/             # Configuración SSL y certificados
├── scripts/              # Scripts de utilidad (certs, hooks)
├── docs/                 # Documentación adicional
├── Makefile              # Comandos: build, run, test, lint
└── docker-compose.yml
```

---

## 🚀 Instalación y Configuración

### 1. Clonar el repositorio

```bash
git clone https://github.com/cloudflax/api.cloudflax.git
cd api.cloudflax
```

### 2. Certificados SSL para PostgreSQL

Antes del primer `docker-compose up`, genera los certificados:

```bash
make db-certs
```

Ver [postgres/README.md](postgres/README.md) para más detalles.

### 3. Entorno de Desarrollo (DevContainer)

Este proyecto usa **Dev Containers** (Cursor / VS Code):

1. Instala **Docker** y la extensión **Dev Containers**.
2. Abre la carpeta y acepta `Reopen in Container`.
3. El contenedor incluye: Go, Air (hot reload), golangci-lint. El hook pre-commit ejecuta `make lint` antes de cada commit.

### 4. Variables de Entorno

En Docker, las variables se configuran en `docker-compose.yml`. Las variables se cargan desde el entorno. En Docker, vienen de `docker-compose`. Para desarrollo local, usa `.env.example` como referencia:

| Variable      | Descripción        | Default    |
|---------------|--------------------|------------|
| `PORT`        | Puerto de la API   | `3000`     |
| `DB_HOST`     | Host de PostgreSQL | `db`       |
| `DB_PORT`     | Puerto de PostgreSQL| `5432`    |
| `DB_USER`     | Usuario DB         | `postgres` |
| `DB_PASSWORD` | Contraseña DB      | —          |
| `DB_NAME`     | Nombre de la DB    | `cloudflax`|
| `DB_SSL_MODE` | Modo SSL: `require`, `verify-ca`, `verify-full`, `disable` | `disable` |
| `LOG_LEVEL`   | Nivel de log: `DEBUG`, `info`, `WARN`, `ERROR`            | `info`    |

#### 4.1 Configuración con AWS Secrets Manager (LocalStack)

Si usas **LocalStack** con Secrets Manager, puedes cargar las credenciales de la base de datos desde un secreto en lugar de variables de entorno. El secreto debe ser un JSON con: `dbname`, `host`, `password`, `port`, `username`.

1. Define en LocalStack un secreto (por ejemplo `db/cloudflax`) con el JSON de credenciales.
2. En `docker-compose` o en el entorno, configura:
   - `CONFIG_SOURCE=secrets`
   - `AWS_ENDPOINT_URL=http://localhost.localstack.cloud:4566` (o `http://host.docker.internal:4566` si LocalStack corre en tu máquina y la app en Docker)
   - `AWS_REGION=us-east-1`
   - `AWS_SECRET_NAME=db/cloudflax`
   - `AWS_ACCESS_KEY_ID=test` y `AWS_SECRET_ACCESS_KEY=test` (LocalStack acepta credenciales de prueba).

La aplicación cargará el secreto **solo al arranque** y usará esos valores para la conexión a la base de datos. El resto de la configuración (`PORT`, `LOG_LEVEL`) sigue leyéndose de variables de entorno.

#### 4.2 Ver correos enviados por SES en LocalStack

LocalStack no entrega correos reales; los guarda en memoria. Para inspeccionar los emails enviados (p. ej. verificación de cuenta):

**Requisito previo:** En LocalStack hay que **verificar la identidad del remitente** (el `SES_FROM_ADDRESS`) antes de que los envíos se acepten. Si no lo haces, los correos no se guardan y el registro puede fallar al enviar el email.

```bash
# En la máquina donde corre LocalStack (puerto 4566), verificar la identidad una vez:
AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 sesv2 create-email-identity --email-identity jose.guerrero@cloudflax.com --region us-east-1
```

Luego, para listar los correos enviados:

- **Desde la máquina donde corre LocalStack** (puerto 4566):
  ```bash
  curl -s "http://localhost:4566/_aws/ses" | jq .
  ```
- **Desde el DevContainer** (LocalStack en el host):
  ```bash
  curl -s "http://host.docker.internal:4566/_aws/ses" | jq .
  ```
- Opcional: filtrar por remitente con `?email=tu-ses-from@ejemplo.com`.

La respuesta incluye `Subject`, `Body` (text/html), `Destination`, `Source` y `Timestamp` de cada mensaje.

**Si `messages` sale vacío:** (1) Comprueba que ejecutaste la verificación de identidad arriba antes de registrar. (2) Prueba sin filtro: `curl -s "http://localhost:4566/_aws/ses" | jq .` para ver todos los mensajes. (3) Revisa los logs de la app al arrancar por si aparece "failed to initialise SES sender" (entonces se usa noop y no se envía nada).

### 5. Comandos (dentro del DevContainer)

```bash
make build      # Compilar
make run        # Ejecutar (requiere variables de entorno)
make test       # Tests
make test-cover # Tests con cobertura (genera coverage.html)
make lint       # golangci-lint
```

### 6. Endpoints

| Método | Ruta       | Descripción                              |
|--------|------------|------------------------------------------|
| GET    | `/`        | Info de la API                           |
| GET    | `/health`  | Health check (verifica conexión DB)      |
| GET    | `/users`   | Lista usuarios                           |
| GET    | `/users/:id` | Usuario por ID |

### 7. Ejecución manual (sin Docker)

```bash
go mod tidy
export DB_HOST=localhost DB_PASSWORD=postgres  # y el resto de vars
make run
```

## 🎯 Roadmap del Proyecto

- [ ] **Integración con AWS S3:** Implementación completa para la carga y gestión de imágenes de productos.
- [ ] **Webhooks de Pago:** Implementación para actualizaciones automáticas desde las pasarelas de pago.
- [ ] **Notificaciones Push:** Sistema de alertas para cambios en tiempo real del estado de los envíos.
- [ ] **Módulo de Analítica:** Generación de reportes avanzados para estrategias de marketing.