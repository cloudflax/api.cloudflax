# GitHub Project `@api.cloudflax` — Flujo de estados (referencia)

Documentación del **orden y significado** de los valores de **Status** en el tablero. La **sincronización** con ramas, PRs y merges suele hacerse con **workflows** configurados en GitHub Project; este archivo no pide al agente ni al humano que gestione el campo a mano salvo indicación explícita.

## Secuencia

**Backlog** → **In progress** → **In review** → **Staging** → **Done**

## Estados (texto del project)

Definiciones alineadas al campo **Status** en `@api.cloudflax` (mismo criterio que en la UI de GitHub Project).

| Status | Descripción |
|--------|---------------|
| **Backlog** | This item hasn't been started. |
| **In progress** | This is actively being worked on. |
| **In review** | This item is in review. |
| **Staging** | Merged to develop. |
| **Done** | This has been completed. |
