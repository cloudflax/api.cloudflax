# Convenciones de código — Cloudflax API

Reglas para mantener consistencia. En este archivo van:

- **Naming** — Cómo nombrar funciones, variables, archivos y recursos.
- **Estructura de carpetas** — Organización, rutas y layout de directorios.
- **Estilo de código** — Formato, imports, comentarios y buenas prácticas de sintaxis.

---

## Nombres CRUD

**Todas las acciones CRUD usan el mismo patrón en todos los features y en todas las capas:**

```
{Action}{Resource}
```

- **Resource** siempre en singular: `User`, `Country`, `Order`
- **Action** de CRUD: `List`, `Get`, `Create`, `Update`, `Delete`

| Acción | Ejemplo User | Ejemplo Country | Ejemplo Order |
|--------|--------------|-----------------|---------------|
| Listar | `ListUser` | `ListCountry` | `ListOrder` |
| Obtener uno | `GetUser` | `GetCountry` | `GetOrder` |
| Crear | `CreateUser` | `CreateCountry` | `CreateOrder` |
| Actualizar | `UpdateUser` | `UpdateCountry` | `UpdateOrder` |
| Eliminar | `DeleteUser` | `DeleteCountry` | `DeleteOrder` |

**Aplica en:** `handler.go`, `service.go`, `repository.go` y cualquier archivo del feature.

---

## Estructura de carpetas

- **Features:** `internal/{recurso}/` — nombre en singular, lowercase (ej: `user`, `product`).
- **Shared:** `internal/shared/{módulo}/` — código reutilizable entre features.
- **Archivos:** Nombres en singular, lowercase, con sufijo por tipo (`handler.go`, `service.go`, `repository.go`).

---

## Estilo de código

- **Imports:** Ordenar: estándar → terceros → internal. Agrupar con líneas en blanco.
- **Variables:** camelCase. Constantes y tipos exportados: PascalCase.
- **Comentarios:** En inglés para código. En español para documentación de usuario.
