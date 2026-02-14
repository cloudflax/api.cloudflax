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

```
├── cmd/api/           # Entry point de la API
├── internal/
│   ├── app/           # Configuración Fiber y rutas
│   ├── config/        # Carga y validación de variables de entorno
│   ├── db/            # GORM + conexión PostgreSQL + migraciones
│   ├── handlers/      # Handlers HTTP por ruta
│   ├── logger/        # slog (logging estructurado JSON)
│   ├── middleware/    # Logger de requests
│   └── models/        # Modelos GORM (User)
├── postgres/          # Configuración SSL y certificados
├── scripts/           # Scripts de utilidad (certs, hooks)
├── Makefile           # Comandos: build, run, test, lint
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
| `LOG_LEVEL`   | Nivel de log: `DEBUG`, `INFO`, `WARN`, `ERROR`            | `INFO`    |

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