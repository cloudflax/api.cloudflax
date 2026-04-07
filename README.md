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

### 2. Entorno de Desarrollo (DevContainer)

Este proyecto usa **Dev Containers** (Cursor / VS Code):

1. Instala **Docker** y la extensión **Dev Containers**.
2. Abre la carpeta y acepta `Reopen in Container`.
3. El contenedor incluye: Go, Air (hot reload), golangci-lint. El hook pre-commit ejecuta `make lint` antes de cada commit.

### 3. Variables de Entorno

Las variables de entorno se gestionan mediante un archivo `.env` en la raíz del proyecto (no se commitea; está en `.gitignore`). Usa `.env.example` como referencia para crear el tuyo.

El `docker-compose.yml` del DevContainer carga automáticamente el `.env` mediante `env_file`.

#### Variables de aplicación

| Variable      | Descripción        | Default    |
|---------------|--------------------|------------|
| `APP_ENV`     | Entorno (`development`, `production`). En producción debe ser `production` (el arranque rechaza `ENABLE_AUTH_DEV_ENDPOINTS=true` con este valor). | (vacío en Load; localmente `development`) |
| `ENABLE_AUTH_DEV_ENDPOINTS` | Expone rutas `/auth/dev/*` fuera de producción. En producción: `false` o ausente. | `false` |
| `PORT`        | Puerto de la API   | `3000`     |
| `LOG_LEVEL`   | Nivel de log: `debug`, `info`, `warn`, `error` | `info` |
| `APP_URL`     | URL base de la aplicación | `http://localhost:3000` |
| `JWT_SECRET`  | Clave secreta para tokens JWT | — (requerido) |
| `DB_SSL_MODE` | Modo SSL de PostgreSQL: `require`, `verify-ca`, `verify-full`, `disable` | `disable` |

#### Variables de AWS

Las credenciales de base de datos se obtienen de **AWS Secrets Manager** al arranque. Los correos se envían mediante **AWS SES v2**.

| Variable              | Descripción | Default |
|-----------------------|-------------|---------|
| `AWS_REGION`          | Región de AWS | `us-east-1` |
| `AWS_SECRET_NAME`     | Nombre o ARN del secreto en Secrets Manager (JSON con `dbname`, `host`, `password`, `port`, `username`) | — (requerido) |
| `AWS_ACCESS_KEY_ID`   | Access key de AWS | — (requerido) |
| `AWS_SECRET_ACCESS_KEY` | Secret key de AWS | — (requerido) |
| `SES_FROM_ADDRESS`    | Dirección de remitente para emails (debe estar verificada en SES) | — (requerido) |

### 4. Comandos (dentro del DevContainer)

```bash
make build      # Compilar
make run        # Ejecutar (requiere variables de entorno)
make test       # Tests
make test-cover # Tests con cobertura (genera coverage.html)
make lint       # golangci-lint
```

### 5. Endpoints

| Método | Ruta       | Descripción                              |
|--------|------------|------------------------------------------|
| GET    | `/`        | Info de la API                           |
| GET    | `/health`  | Health check (verifica conexión DB)      |
| GET    | `/users`   | Lista usuarios                           |
| GET    | `/users/:id` | Usuario por ID |

### 6. Ejecución manual (sin Docker)

```bash
go mod tidy
cp .env.example .env   # edita .env con tus valores reales
source .env
make run
```

## 🎯 Roadmap del Proyecto

- [ ] **Integración con AWS S3:** Implementación completa para la carga y gestión de imágenes de productos.
- [ ] **Webhooks de Pago:** Implementación para actualizaciones automáticas desde las pasarelas de pago.
- [ ] **Notificaciones Push:** Sistema de alertas para cambios en tiempo real del estado de los envíos.
- [ ] **Módulo de Analítica:** Generación de reportes avanzados para estrategias de marketing.