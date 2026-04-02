# GitHub: flujo y trazabilidad (`cloudflax/api.cloudflax`)

El detalle de cada paso vive en archivos bajo [`docs/github-workflow/`](./github-workflow/); este archivo es el **índice**: secuencia y enlaces.

**Regla previa:** no abras issue, rama ni PR **sin indicación explícita del usuario**. Calidad de código (`make lint`, `make test`, GORM, etc.): [AGENT_RULES.md](./AGENT_RULES.md).

---

## Secuencia de acciones

Ejecutá **solo** las acciones que el usuario incluyó en el alcance de la sesión, **en este orden** cuando varias apliquen.

| # | Acción | Depende de |
|---|--------|------------|
| 1 | [Commits con trazabilidad](./github-workflow/01-commits.md) | Hay cambios listos; convención del repo. |
| 2 | [Push y pull request](./github-workflow/02-push-pr.md) | El usuario lo pidió explícitamente o hay acuerdo en la sesión. |

Si el usuario **no** pidió una acción, **no** la ejecutes (misma línea que [AGENT_RULES.md](./AGENT_RULES.md) para `push`/PR).

---

## Tablero (solo referencia)

[Flujo de estados del project `@api.cloudflax`](./github-workflow/03-tablero.md): qué significa cada **Status** y en qué orden van. No es una acción del flujo de este repo; el tablero se alinea al trabajo vía **workflows** del project salvo que alguien pida otra cosa.

---

## Apéndice (herramientas)

- [GitHub CLI: auth y `gh project`](./github-workflow/gh-cli.md)
