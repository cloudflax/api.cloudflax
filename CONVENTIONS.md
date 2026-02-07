# Convenciones de código — Cloudflax API

Reglas para mantener consistencia. En este archivo van:

- **Naming** — Cómo nombrar funciones, variables, archivos y recursos.
- **Estructura de carpetas** — Organización, rutas y layout de directorios.
- **Estilo de código** — Formato, imports, comentarios y buenas prácticas de sintaxis.
- **Formato de API** — Respuestas JSON y códigos de estado por operación.

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

### Nombres descriptivos (Clean Code)

Evitar abreviaciones y nombres cortos que obligan al lector a “traducir” mentalmente. Preferir nombres que revelen la intención.

- **No abreviar:** Preferir `userRepository`, `userService`, `userHandler` en lugar de `userRepo`, `userSvc`, `userHnd`.
- **En parámetros y variables:** Usar nombres completos como `repository`, `service`, `handler`, `user`, `request`, `response`.
- **Excepciones ampliamente conocidas:** `err` (error), `ctx` (context), `id`, `req`/`resp` en contexto HTTP — son convenciones estándar y legibles.
- **Receivers:** Go acepta receivers de 1–2 letras; usar nombres descriptivos cuando mejoren la claridad (ej. `repository` o `svc` en lugar de `r` si hay varios receivers).

---

## Formato de respuesta JSON

**Éxito:**
```json
{
  "data": { ... },
  "message": "opcional"
}
```

**Error:**
```json
{
  "error": "Descripción del error",
  "code": "ERROR_CODE",
  "status": 400
}
```

Para listas paginadas, incluir `meta` con `page`, `limit`, `total`.

---

## Status codes por operación

| Operación | Éxito | Errores frecuentes |
|-----------|-------|---------------------|
| List | 200 | 400 (query inválida) |
| Get by ID | 200 | 404 (no encontrado) |
| Create | 201 | 400 (validación), 409 (duplicado) |
| Update | 200 | 400 (validación), 404 (no encontrado) |
| Delete | 200 o 204 | 404 (no encontrado) |
| Login | 200 | 401 (credenciales inválidas) |
