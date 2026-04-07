# Specs | Auth y seguridad

Aquí viven las **especificaciones por ítem** (o por grupo acotado) antes y durante la implementación.

## Cómo usarlo

1. Abre [`docs/auth-security-roadmap.md`](../../auth-security-roadmap.md) y elige la **fase** y el **ID** en orden (o marca `deferred` lo que no toques aún).
2. Copia [`SPEC-TEMPLATE.md`](./SPEC-TEMPLATE.md) a `SPEC-<ID>.md` (ej. `SPEC-A3-jwt-claims.md`) solo para los ítems que vayas a implementar o revisar en serio.
3. Implementa → `make lint` y `make test` → un **commit** por ítem o por grupo cerrado (Conventional Commits).
4. Actualiza el roadmap: estado del ítem y enlace al spec o al PR.

Los specs son **vivos**: si algo queda fuera, márcalo en “Fuera de alcance” y enlaza un issue o un `deferred` en el roadmap.
