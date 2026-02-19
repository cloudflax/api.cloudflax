# Cuentas y propiedad de los datos (multi-tenancy)

Este documento recoge el requisito de producto y técnico: **la Cuenta (empresa/tenant) es la dueña de los datos**, y cómo afecta al comportamiento de la API y al filtrado.

---

## Concepto

| Entidad           | Rol                                                                                                                                 |
|-------------------|-------------------------------------------------------------------------------------------------------------------------------------|
| **User**          | La persona (identidad). Puede pertenecer a varias Cuentas. Dentro de una Cuenta, los datos pueden estar asociados al User (quién emitió una factura, quién creó un recurso, etc.) para trazabilidad y filtrado. |
| **Account**       | Empresa/tenant en el SaaS. **Dueña de los datos.** Todos los datos de negocio (proyectos, recursos, facturación, etc.) pertenecen a una Cuenta, no al User. Dentro de esa segmentación, muchos recursos tienen además un **usuario responsable** (ej. emisor de factura, creador de proyecto). |
| **Miembros**      | Usuarios vinculados a una Cuenta (p. ej. con un rol). Cualquier miembro puede actuar en nombre de la Cuenta dentro de sus permisos. |

**Principio:** Los datos son propiedad de la **Cuenta**. Al atender una petición, la API debe **filtrar por Cuenta** y, dentro de esa Cuenta, **tener en cuenta al User**: quién realiza la acción, roles y permisos, y la asociación usuario–recurso (quién emitió, creó o modificó). El filtrado combina **Cuenta + User/rol** según el caso.

---

## Implicaciones

1. **Aislamiento de datos**  
   Todo recurso que sea “dato de negocio” (proyectos, ítems, suscripciones, etc.) debe estar asociado a una **Cuenta** (p. ej. `account_id`). Las consultas y respuestas deben estar siempre acotadas a una sola Cuenta.

2. **Contexto de la petición**  
   Cuando un miembro envía una petición desde el cliente:
   - La petición va ligada a **una Cuenta** (p. ej. claim en JWT `account_id`, header `X-Account-ID`, o workspace elegido en la app).
   - El backend debe **filtrar todas las lecturas y escrituras** por esa Cuenta y, cuando aplique, por el **User** (roles, quién emitió/creó el recurso):
     - **Lectura:** devolver solo datos de esa Cuenta; las listas y filtros pueden acotar además por usuario (ej. “facturas emitidas por este usuario”) o por rol.
     - **Escritura:** crear/actualizar/borrar solo dentro de esa Cuenta (y comprobar membresía/rol); asociar al User cuando corresponda (ej. `issued_by_user_id` en una factura).

3. **Autorización**  
   - Autenticar al **User** (quién hace la petición).  
   - Resolver el **Account** (cuenta) del request.  
   - Comprobar que el User es **miembro** de esa Cuenta (y opcionalmente el rol).  
   - Aplicar las reglas de acceso: **solo datos de esa Cuenta**, y cuando el recurso o la operación lo exijan, **considerar al User** (roles, atribución, filtros por “mis facturas”, “recursos que creé”, etc.).

4. **Usuarios con varias cuentas**  
   Un User puede ser miembro de varias Cuentas. El cliente (p. ej. Next.js) debe:
   - Permitir **elegir la Cuenta/workspace actual**.
   - Enviar ese contexto de Cuenta en cada petición (p. ej. en el token o en un header).  
   La API no “elige” la Cuenta; la **valida** y **filtra por ella**.

---

## Resumen

- **Account = tenant/empresa = dueña de los datos.**  
- **Filtrar por Cuenta** en toda petición y, dentro de esa Cuenta, **tener en cuenta al User**: roles, quién emitió/creó el recurso, y filtros por usuario cuando aplique (ej. facturas emitidas por un usuario concreto).  
- El **contexto de la petición** incluye Cuenta (y opcionalmente User); el backend valida membresía, acota por Cuenta y aplica reglas que combinan Cuenta + User/rol según el recurso.  
- Los recursos de negocio pueden llevar **atribución a usuario** (ej. `issued_by_user_id`) sin cambiar la propiedad: la factura es de la Cuenta; el User indica quién la emitió.  
- Este README es el placeholder del requisito; la implementación (modelos, middleware, scoping) vendrá después.

---

## Esquema de base de datos (modelo entidad-relación)

A continuación se describe cómo se traduce el modelo conceptual en tablas y relaciones. Parte del esquema ya existe (`users`, `refresh_tokens`); el resto es objetivo de diseño para la implementación de multi-tenancy.

### Entidades y atributos

| Entidad | Tabla | Atributos principales | Notas |
|--------|--------|------------------------|--------|
| **User** | `users` | `id` (PK), `name`, `email`, `password_hash`, `created_at`, `updated_at`, `deleted_at` | Identidad de la persona. Ya implementado. |
| **Account** | `accounts` | `id` (PK), `name`, `slug` (único), `created_at`, `updated_at`, `deleted_at` | Tenant/empresa; dueña de los datos. |
| **AccountMember** | `account_members` | `id` (PK), `account_id` (FK), `user_id` (FK), `role`, `created_at`, `updated_at` | Relación N:M User–Account con rol. UNIQUE(account_id, user_id). |
| **Recurso de negocio** (ej. factura, proyecto) | p. ej. `invoices` | `id` (PK), `account_id` (FK, NOT NULL), `issued_by_user_id` (FK, nullable), … | Todo recurso de negocio tiene `account_id`; opcionalmente un usuario responsable. |

### Relaciones

- **User ↔ Account:** muchos a muchos, a través de `account_members`. Un User puede pertenecer a varias Accounts; una Account tiene muchos Users (miembros).
- **Account → Recursos de negocio:** uno a muchos. Cada recurso (proyecto, factura, ítem, etc.) pertenece a una sola Account (`account_id`).
- **User → Recursos de negocio (atribución):** opcional. Campos como `issued_by_user_id`, `created_by_user_id` indican quién realizó la acción dentro de la Account; la propiedad del dato sigue siendo de la Account.

### Diagrama entidad-relación (Mermaid)

```mermaid
erDiagram
    users ||--o{ account_members : "es miembro"
    accounts ||--o{ account_members : "tiene miembros"
    accounts ||--o{ invoices : "posee"
    users ||--o{ invoices : "emite (atribución)"

    users {
        uuid id PK
        string name
        string email
        string password_hash
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
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
        uuid issued_by_user_id FK "nullable"
        string number
        decimal total
        timestamp created_at
        timestamp updated_at
    }
```

*Nota: `invoices` es un ejemplo de recurso de negocio; el mismo patrón aplica a proyectos, ítems, suscripciones, etc.: `account_id` obligatorio y, si aplica, FKs de atribución a usuario.*
