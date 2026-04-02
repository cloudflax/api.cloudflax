# GitHub workflow — Acción 2: Push y pull request

**Cuándo aplica:** el usuario lo pidió **explícitamente** o hay acuerdo en la sesión ([AGENT_RULES.md](../AGENT_RULES.md)).

- `git push` y apertura de PR solo en esas condiciones.
- **Rama y base por defecto:** el PR sale siempre de la **rama actual** (head) hacia **`develop`** (base), salvo que el usuario indique otra base explícitamente.
- **Excepción:** hotfix u otra política acordada puede usar **`main`** (u otra base) en lugar de `develop`.
- Descripción del PR: `Closes #ID` o `Refs #ID` según corresponda.
