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
* **Base de Datos:** PostgreSQL
* **Contenedores:** Docker & **DevContainers** (Entorno de desarrollo estandarizado).
* **Infraestructura (AWS):**
    * **EC2:** Hosting del servidor principal.
    * **RDS:** Instancia gestionada de PostgreSQL.
    * **S3:** Almacenamiento de imágenes y activos.
    * **CloudFront:** CDN para entrega de contenido global.
    * **Lambda:** Procesos asíncronos y tareas específicas.

---

## 🚀 Instalación y Configuración

### 1. Clonar el repositorio
```bash
git clone https://github.com/cloudflax/api.cloudflax.git
cd api.cloudflax
```

### 2. Entorno de Desarrollo (Recomendado)

Este proyecto incluye soporte para **VS Code DevContainers**. Para usarlo:

1. Asegúrate de tener instalado **Docker** y la extensión **Dev Containers** en VS Code.
2. Al abrir la carpeta en VS Code, acepta la opción `Reopen in Container`.
3. El entorno configurará automáticamente **Go** y las dependencias necesarias dentro de un contenedor dedicado.

### 3. Variables de Entorno

Configura un archivo `.env` en la raíz del proyecto con los siguientes parámetros:

```env
# Configuración de la Base de Datos
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=cloudflax

# Configuración de AWS
AWS_ACCESS_KEY=
AWS_SECRET_KEY=
AWS_S3_BUCKET=
```

### 4. Ejecución manual

Si prefieres ejecutarlo fuera de Docker, sigue estos pasos:

```bash
# Descargar y limpiar dependencias
go mod tidy

# Ejecutar la aplicación
go run main.go
```

## 🎯 Roadmap del Proyecto

- [ ] **Integración con AWS S3:** Implementación completa para la carga y gestión de imágenes de productos.
- [ ] **Webhooks de Pago:** Implementación para actualizaciones automáticas desde las pasarelas de pago.
- [ ] **Notificaciones Push:** Sistema de alertas para cambios en tiempo real del estado de los envíos.
- [ ] **Módulo de Analítica:** Generación de reportes avanzados para estrategias de marketing.