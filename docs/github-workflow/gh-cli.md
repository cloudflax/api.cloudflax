# GitHub CLI (`gh`) — apéndice (agente)

**Audiencia:** agentes de IA que operan en este repo. **Objetivo:** saber qué exige `gh` (credenciales = OAuth o **PAT** con **scopes** para **repo** y **GitHub Projects**, y org si aplica) y qué comandos usar **si** ya hay sesión válida en el entorno.

Índice del flujo humano: [GITHUB_WORKFLOW.md](../GITHUB_WORKFLOW.md).

## Auth

| Situación | Comando de referencia |
|-----------|------------------------|
| Comprobar si hay sesión y qué scopes tiene | `gh auth status -h github.com` |
| Ampliar scopes (p. ej. Projects) | `gh auth refresh -h github.com -s project,read:project,read:org` |
| Login interactivo (humano) | `gh auth login -h github.com` (HTTPS o SSH según el remoto) |
| PAT sin TTY (humano o CI) | Pipe del token a `gh auth login -h github.com --with-token` |

Scopes de referencia: `repo`; Projects v2: `read:project`, `project`; org: `read:org`. Si aparece `INSUFFICIENT_SCOPES` en GraphQL o `gh project`, hace falta refrescar scopes o credenciales — **avisá al usuario**; no inventes tokens.

**Reglas para el agente:** no completés vos el login (navegador, device code, pegar PAT): eso lo hace el **usuario** o el entorno (CI, contenedor ya autenticado). Antes de usar `gh issue`, `gh project`, etc., podés correr `gh auth status` para diagnosticar; si no hay sesión o faltan scopes, **pedí explícitamente** que configure auth o que ejecute los comandos que correspondan.

**Versión:** `gh` antiguo puede no incluir `gh project` — indicá actualizar o usar `gh api graphql` ([releases](https://github.com/cli/cli/releases)).

## Project (owner `cloudflax`)

Solo si `gh auth status` muestra sesión con scopes adecuados:

1. `gh project list --owner cloudflax`
2. `gh project field-list <n> --owner cloudflax`
3. Ítems: `gh project item-list …`

**Editar ítem:** `gh project item-edit --project-id … --id <item-id> --field-id …` y uno de `--single-select-option-id …`, `--number …`, `--date YYYY-MM-DD`, `--clear`.

**Issue:** `gh issue edit <n> --add-assignee cloudflax --add-label "<label>"`.

**Alineación con el repo:** no abras issue ni PR ni mutés el tablero sin indicación explícita del usuario; ver [GITHUB_WORKFLOW.md](../GITHUB_WORKFLOW.md) y [AGENT_RULES.md](../AGENT_RULES.md).
