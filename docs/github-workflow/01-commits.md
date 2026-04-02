# GitHub workflow — Acción 1: Commits con trazabilidad

**Cuándo aplica:** hay cambios listos para commitear; convención del repo.

- Mensajes en **inglés**; [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, …).
- **Enlace al issue:** del nombre de rama obtené el `#ID` (ej. `feature/3-agents-md-hub` → `#3`). En el cuerpo o pie del mensaje: `Refs #ID` o `Closes #ID` según corresponda.
- Sin `git commit --amend` ni reescritura de historial remoto salvo petición explícita del usuario.
