# GitHub: flujo y trazabilidad (`cloudflax/api.cloudflax`)

Este documento define **qué hacer en orden** cuando el trabajo debe quedar enlazado a GitHub (issues, proyecto, ramas, commits). Sirve para humanos y para agentes (IA).

**Regla previa:** no abras issue, rama ni PR **sin indicación del usuario**. Calidad de código (`make lint`, `make test`, GORM, etc.): [AGENT_RULES.md](./AGENT_RULES.md).

---

## 1. Modelo de ramas (referencia)

| Rama / prefijo | Rol | En este repo |
|----------------|-----|--------------|
| `develop` | Integración habitual | Base de `feature/…` y **destino del PR** en trabajo normal. |
| `main` | Estable / despliegue | Los features **no** apuntan aquí; PR va a `develop`. Hotfixes según acuerdo del equipo. |
| `feature/<ID>-<slug>` | Tarea acotada al issue | `<ID>` = número del issue en GitHub. |
| `release/<versión>` | Preparación de versión (opcional) | Política de equipo; suele mezclar en `main` y `develop`. |
| `hotfix/<ID>-<slug>` | Urgencia sobre producción | Desde `main`; PR típico a `main` y **reintegrar en `develop`**. |

---

## 2. Decisión con el usuario (antes de tocar GitHub)

El trabajo en GitHub es una **tubería** de piezas opcionales, en este orden lógico: **issue/project → rama → commits → PR** (y **estado en el tablero** al cerrar el ciclo). No ejecutes una fase si el usuario no la incluyó en el alcance de la sesión.

### 2.1 Prompt copiable (checklist en el chat)

Pedí al usuario que **pegue este bloque** en el mensaje y marque con `[x]` lo que aplica (en Cursor y en muchos visores de Markdown podés alternar la casilla con un clic). Lo no marcado se interpreta como **fuera de alcance** salvo que aclare otra cosa.

```markdown
**Alcance GitHub para esta tarea** (marcá con [x]):

- [ ] **Issue + project**: crear o actualizar issue en `@api.cloudflax` y campos del tablero (Fase 3).
- [ ] **Rama**: crear/publicar `feature/<ID>-<slug>` desde `develop` (Fase 4). Si no marcás esto, trabajo en la rama actual.
- [ ] **Commits**: dejar cambios commiteados con Conventional Commits y `Refs`/`Closes` (Fase 5).
- [ ] **PR**: push y abrir o actualizar PR hacia la base acordada (Fase 6).
- [ ] **Tablero al cerrar**: ajustar `Status` del item cuando corresponda (Fase 7).

Contexto breve (opcional): …
```

Si el usuario no usa la checklist, **preguntá en una sola franja** qué marca de la lista anterior necesita; no asumas trazabilidad completa.

### 2.2 De la checklist a las fases

| Qué está marcado | Fases (en orden) | Notas |
|------------------|------------------|-------|
| Issue + Rama + Commits + PR (+ tablero) | 3 → 4 → 5 → 6 → 7 | “Trazabilidad completa” típica. |
| Rama + Commits + PR (+ tablero si aplica) | 4 → 5 → 6 → (7) | Issue ya existe o no se pide en esta sesión. |
| Commits + PR | 5 → 6 | Ya estás en la rama correcta; no abras rama nueva. |
| Solo PR | 6 | `push` solo si hace falta; no reescribas historia remota sin orden explícita. |
| Solo issue / project | 3 | No crees rama hasta nueva orden. |
| Issue + Rama (sin commits todavía) | 3 → 4 | Dejá Fase 5 para cuando haya cambios listos. |
| Ya hay rama ligada al issue (`feature/…`, `hotfix/…`) y solo sigue el trabajo | 5 → 6 → (7) | No abras issue ni rama nuevas; alineá con lo marcado en la checklist. |

**Regla:** si algo no está marcado, **no** hagas esa acción en GitHub (misma línea que la regla previa del documento y [AGENT_RULES.md](./AGENT_RULES.md) para `push`/PR).

---

## 3. Crear el issue y ubicarlo en el project

1. **Issue** en el repo `cloudflax/api.cloudflax`, en el project **`@api.cloudflax`** (nombre exacto).  
   Comando típico: `gh issue create -R cloudflax/api.cloudflax -p "@api.cloudflax"`.  
   **Cuerpo en inglés:** objetivos y criterios de aceptación.
2. **Asignación y etiquetas:** si el CLI lo permite, `--assignee cloudflax` y `--label …`; si hace falta otro assignee, aplícalo.
3. **Tablero (sin campos vacíos obligatorios):** `Status`, `Priority`, `Size` y `Estimate` siempre con valor. **Start/Target date** solo con criterio o acuerdo; si no aplican, indícalo en el cuerpo del issue (no inventes fechas).

**Checklist mínimo al dar por creada la pieza en GitHub:** issue en inglés y asignado · labels adecuados · en el project: Status, Priority, Size, Estimate · item enlazado al issue.

---

## 4. Crear la rama (solo si el usuario pidió rama + trazabilidad)

1. Actualiza **`develop`** localmente.
2. Crea **`feature/<ID_ISSUE>-<slug-kebab>`** desde **`develop`**. El `#ID` debe existir antes en GitHub.  
   Puedes usar `gh issue develop` si enlaza o crea la rama con base correcta; si la base no es `develop`, créala a mano desde `develop`.
3. Publica la rama remota y mantenla alineada con el issue/project.

---

## 5. Mientras trabajas: commits

- **Idioma:** mensajes en **inglés**; [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, …).
- **Enlace al issue:** antes de commitear, **lee el nombre de la rama** para obtener el `#ID` (ej. `feature/3-agents-md-hub` → `#3`). En el **cuerpo o pie** del mensaje usa `Refs #ID` o `Closes #ID` según si el commit cierra el issue o solo lo referencia.
- **Historial:** sin `git commit --amend` ni reescritura de historial remoto **salvo** petición explícita del usuario.

---

## 6. Push y pull request

- **`git push` y apertura de PR:** solo si el usuario lo pide **explícitamente** o hay acuerdo en la sesión (misma regla que en [AGENT_RULES.md](./AGENT_RULES.md)).
- **Base del PR:** **`develop`** para features. **`main`** solo para hotfix (u otra política que el equipo deje por escrito).
- **Descripción del PR:** incluye `Closes #ID` o `Refs #ID` según corresponda.

---

## 7. Estado en el project durante y después del review

- **Status** debe reflejar la realidad. Valores habituales (orden típico): **Backlog** → **Ready** → **In progress** → **In review** → **Done**.
- **In review:** con PR abierto (p. ej. hacia `develop`).
- **Done:** cuando el código está mergeado en la rama de integración acordada o liberado según política del equipo.

---

## Apéndice

### Autenticación con GitHub CLI (`gh`)

**Quién autentica:** `gh auth login`, `gh auth refresh`, el flujo en navegador, el device code (`github.com/login/device`) y el uso de un PAT **los ejecutás y completás vos** en tu entorno (local o devcontainer). Un agente de IA no puede finalizar el login en tu cuenta; si falta sesión o scopes, hacelo vos y recién después seguí con issues/project desde la herramienta que uses.

**Login inicial** (navegador o device code):

```bash
gh auth login -h github.com
```

Elegí **HTTPS** si el remoto es `https://github.com/…`, o **SSH** si usás URL SSH.

**Token personal (PAT)** sin flujo interactivo:

```bash
echo 'YOUR_TOKEN' | gh auth login -h github.com --with-token
```

**Comprobar sesión y scopes** (issues, PR y labels suelen alcanzar con `repo`; tablero **Projects v2** en la org: **`read:project`**, **`project`**; recursos de org: **`read:org`**):

```bash
gh auth status -h github.com
```

**Ampliar permisos** cuando fallen consultas GraphQL a `projectsV2`, comandos `gh project` o visibilidad de la org (`INSUFFICIENT_SCOPES`). Un solo refresh añade lo necesario para tablero y organización:

```bash
gh auth refresh -h github.com -s project,read:project,read:org
```

`gh auth refresh` puede abrir el navegador o mostrar un código de un solo uso y la URL `https://github.com/login/device`; completá el flujo en la misma máquina donde ejecutaste el comando.

**Versión del CLI:** en paquetes del sistema a veces `gh` es antiguo y **no incluye** `gh project`. Para listar/editar el tablero desde terminal hace falta una versión reciente ([releases de GitHub CLI](https://github.com/cli/cli/releases)) o usar `gh api graphql` con los scopes anteriores.

### Edición avanzada del project con `gh`

1. `gh project list --owner cloudflax`
2. `gh project field-list <n> --owner cloudflax`
3. Obtener `item-id`: `gh project item-list`

Edición: `gh project item-edit --project-id … --id <item-id> --field-id …` con `--single-select-option-id …`, `--number …`, `--date YYYY-MM-DD` o `--clear`. Para el issue: `gh issue edit <n> --add-assignee cloudflax --add-label "<label>"`.
