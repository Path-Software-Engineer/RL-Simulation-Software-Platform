# 07-rl-simulation-control-platform

## Current implementation status — Azure + Neon release candidate

Sprint 1, **Gridworld Agent Visualizer**, is fixed at
`v0.1.0-sprint-01-gridworld-agent-visualizer`; Sprint 2, **DQN Training Dashboard**, is fixed at
`v0.2.0-sprint-02-dqn-training-dashboard`; Sprint 3, **World Model Rollout Viewer**, is fixed at
`v0.3.0-sprint-03-world-model-rollout-viewer`. The release work continues on
`release/p7-v1.0.0-azure-neon`. The active scope includes:

- Nuxt/Vue workspace with an accessible Gridworld, policy overlay and episode player.
- Go/Gin API with public bounded evidence reads, protected operator commands and idempotency.
- Public portfolio mode replays the latest persisted episode or world-model rollout without credentials.
- PostgreSQL/TimescaleDB as durable truth and Redis Streams with outbox/inbox/DLQ boundaries.
- Python runner restricted to registered, SHA-256-verified Q-Learning and SARSA artifacts.
- Durable `queued → running → succeeded` evidence, cooperative controls and audited feedback.
- OpenAPI 3.1, AsyncAPI 3.0, JSON Schema, ADRs, threat model, runbook and sprint evidence.
- A seeded one-hidden-layer DQN with replay buffer, epsilon-greedy exploration and target sync.
- Forty bounded training episodes with reward, moving average, loss, epsilon, success and action
  telemetry persisted through Redis Streams into TimescaleDB.
- A responsive Nuxt observatory with accessible SVG charts, exact data tables, progress, summary
  cards and action distribution.
- A SHA-256-verified empirical action-delta model fitted from 12 versioned transition examples.
- One 11-step autoregressive rollout compared against real Gridworld transitions, with predicted
  states, per-step Manhattan error, accumulated error and risk metrics persisted end to end.
- A responsive next-state comparison, horizontal rollout inspector, exact table, error chart and
  planning-risk cards.
- One scale-to-zero Azure Container App pulling five public immutable GHCR images anonymously.
- A transient same-replica Redis Streams transport and Neon PostgreSQL with distinct pooled runtime
  and direct migration connections; no Azure Managed Redis or stored Log Analytics.
- An atomic migration Job, immutable GHCR builds and a public three-profile release smoke.

The registered deterministic Q-Learning oracle reaches the goal in **10 transitions**, with
**0 collisions** and **9.64 total reward**. This validates the controlled teaching fixture only;
it is not a claim of generalization or production readiness.

Run the complete local acceptance gate from the project directory:

```powershell
Copy-Item .env.example .env
# Replace OPERATOR_TOKEN before sharing the workspace.
.\scripts\setup.ps1
.\scripts\run-quality-gate.ps1 -KeepRunning
```

Local routes:

- Web: `http://127.0.0.1:3000`
- API readiness: `http://127.0.0.1:8080/health/ready`
- Swagger UI: `http://127.0.0.1:8080/docs/`
- OpenAPI: `http://127.0.0.1:8080/openapi.json`

See [Azure + Neon release guide](docs/azure-neon-release.md),
[Sprint 3 evidence](docs/sprints/sprint-03-world-model-rollout-viewer/README.md),
[architecture](docs/architecture/architecture.md), [runbook](docs/runbook.md),
[user stories](docs/user-stories.md) and [technical stories](docs/technical-stories.md).

The direct Sprint 2 seed-11 profile completes 40 real episodes and 400 metric samples. Its best
episode reward is **7.64**, final 10-episode moving average is **-33.13** and success rate is
**37.5%**. Those deliberately modest results demonstrate observability and instability; they do
not establish convergence or generalization.

The direct Sprint 3 profile uses 12 observed transitions and an 11-action plan. The real path
reaches the goal with one obstacle collision, while the empirical model first diverges at step 2
and accumulates **9.00 cells of Manhattan error**. This exposes a planning limitation; it is not
evidence of reliable planning or generalization.

> Evidence boundary: all three sprints are tagged technical checkpoints. The complete Sprint 3
> containerized gate and three live cross-layer smokes passed on 2026-08-16. Azure/Neon release
> assets are present, but provisioning and public acceptance are not evidence until the authenticated
> deployment flow succeeds.

---


## 🧠 Descripción

**RL Simulation Control Platform** es una aplicación de software aplicada a Reinforcement Learning y simulación.

Este proyecto pertenece a:

```txt
Path Software Engineer
Plan 7 — RL & Simulation Software Platform
```

y acompaña directamente al:

```txt
Path AI Engineer
Plan 7 — Reinforcement Learning, World Models & Robotics Simulation
```

Este proyecto nace como evolución de los antiguos proyectos de Building Projects:

```txt
19-gridworld-agent-visualizer
20-dqn-training-dashboard-lite
21-world-model-rollout-viewer
```

Ahora esos proyectos ya no vivirán como repositorios separados.

Se convierten en **3 sprints principales** dentro de una sola aplicación de software más robusta, documentada y orientada a producto.

La plataforma busca mostrar cómo un agente actúa, aprende, se equivoca, mejora y cómo un modelo puede simular futuros posibles.

---

# 🎯 Objetivo general

Construir una plataforma visual de Reinforcement Learning y simulación que permita:

```txt
visualizar entornos
→ mostrar estados
→ ejecutar acciones
→ recibir recompensas
→ registrar trayectorias
→ observar entrenamiento DQN
→ analizar rewards, epsilon y loss
→ visualizar rollouts de world models
→ medir error acumulado
→ explicar riesgos de planificación
→ mostrar resultados en dashboard
```

El objetivo no es construir un framework avanzado de RL.

El objetivo es crear una aplicación clara y progresiva que conecte:

```txt
agente
→ entorno
→ comportamiento
→ entrenamiento
→ simulación
→ visualización
→ documentación
→ evidencia profesional
```

---

# 🔗 Relación con Path AI Engineer

Este proyecto acompaña los proyectos impares del Plan 7 de Path AI Engineer:

```txt
Path AI Engineer Proyecto 37
→ reinforcement-learning-foundations-gridworld

Path AI Engineer Proyecto 39
→ deep-q-network-gymnasium-lab

Path AI Engineer Proyecto 41
→ world-models-planning-mini-lab
```

Cada uno se convierte en un sprint dentro de esta plataforma.

```txt
Sprint 1 → Gridworld Agent Visualizer
Sprint 2 → DQN Training Dashboard
Sprint 3 → World Model Rollout Viewer
```

---

# 👤 Usuario objetivo

Esta plataforma está pensada para:

```txt
estudiante de Reinforcement Learning
AI Engineer en formación
persona interesada en agentes
equipo que necesita observar entrenamiento RL
reclutador técnico
constructor de portafolio aplicado
```

El usuario debe poder abrir la aplicación o el README y entender:

```txt
qué observa el agente
qué acción tomó
qué recompensa recibió
qué trayectoria siguió
cómo evolucionó el entrenamiento
qué predijo el world model
qué error acumuló
qué riesgos tiene planificar con predicciones
```

---

# 🏗️ Arquitectura general esperada

```txt
RL Simulation Control Platform
│
├── Frontend
│   └── Dashboards visuales
│
├── Backend
│   └── API para simulaciones, logs y resultados
│
├── AI Services
│   └── entornos, agentes, métricas, rollouts y análisis
│
├── Data Layer
│   └── estados, logs, trayectorias y ejemplos
│
├── Reports
│   └── gráficos, summaries, risk cards y outputs
│
└── Docs
    └── user stories, technical stories, decisiones y arquitectura
```

---

# 🔁 Flujo general de la plataforma

```txt
environment setup
→ state definition
→ action selection
→ reward assignment
→ episode execution
→ trajectory recording
→ training logs
→ reward dashboard
→ world model rollout
→ error comparison
→ planning risk cards
→ dashboard
→ reports
→ documentation
```

---

# 🧱 Módulos principales

## 1. Gridworld Environment Engine

Responsabilidad:

```txt
Definir entornos pequeños donde un agente pueda moverse, actuar y recibir recompensas.
```

Incluye:

```txt
grid
posición inicial
objetivo
obstáculos
reward map
acciones disponibles
reglas de transición
```

---

## 2. Agent Behavior Viewer

Responsabilidad:

```txt
Mostrar qué hace el agente dentro del entorno.
```

Incluye:

```txt
estado actual
acción elegida
acción inválida
siguiente estado
recompensa recibida
recompensa acumulada
```

---

## 3. Trajectory Viewer

Responsabilidad:

```txt
Visualizar el camino seguido por el agente.
```

Incluye:

```txt
ruta recorrida
pasos por episodio
loops
choques
llegada al objetivo
policy notes
```

---

## 4. DQN Training Dashboard

Responsabilidad:

```txt
Mostrar la evolución del entrenamiento DQN.
```

Incluye:

```txt
training logs
episode rewards
moving average
epsilon schedule
loss logs
action distribution
training summary cards
```

---

## 5. World Model Rollout Engine

Responsabilidad:

```txt
Simular predicciones de estados futuros usando un modelo del mundo.
```

Incluye:

```txt
estado actual
acción
estado predicho
rollout sequence
estado esperado
comparación
error por paso
```

---

## 6. Planning Risk Cards

Responsabilidad:

```txt
Convertir errores de rollout en notas de riesgo de planificación.
```

Incluye:

```txt
rollout confiable
rollout incierto
rollout degradado
error acumulado
limitaciones
recomendación responsable
```

---

## 7. Visual Dashboard

Responsabilidad:

```txt
Mostrar resultados de RL y simulación de forma clara para el usuario.
```

Debe incluir:

```txt
Gridworld view
Trajectory viewer
Reward charts
DQN training dashboard
World model rollout viewer
Risk cards
Limitations
```

---

# 🏃 Estructura por sprints

## Sprint 1 — Gridworld Agent Visualizer

### Match

```txt
Path AI Engineer Proyecto 37 — reinforcement-learning-foundations-gridworld
```

### Base anterior

```txt
19-gridworld-agent-visualizer
```

### Objetivo

Crear el primer módulo de la plataforma para explicar un agente dentro de un entorno Gridworld.

Flujo:

```txt
grid
→ agente
→ estado
→ acción
→ recompensa
→ trayectoria
→ política
→ visualización
```

### Resultado esperado

Al finalizar este sprint, la plataforma debe permitir ver:

```txt
entorno definido
posición del agente
acciones disponibles
reward map
episodio ejecutado
trayectoria
policy notes
visual report
```

### Labs esperados

```txt
tec-gridworld-setup-lab
tec-state-action-card-lab
tec-reward-map-lab
tec-episode-runner-lab
tec-trajectory-viewer-lab
tec-policy-notes-lab
docs-rl-foundations-storytelling-lab
cloud-gridworld-outputs-to-gcp-storage-lab
```

---

## Sprint 2 — DQN Training Dashboard

### Match

```txt
Path AI Engineer Proyecto 39 — deep-q-network-gymnasium-lab
```

### Base anterior

```txt
20-dqn-training-dashboard-lite
```

### Objetivo

Agregar un módulo para observar entrenamiento DQN mediante logs, métricas y gráficos.

Flujo:

```txt
run or simulate DQN training
→ log episodes
→ log rewards
→ log epsilon
→ log loss
→ summarize behavior
→ render dashboard
```

### Resultado esperado

Al finalizar este sprint, la plataforma debe permitir ver:

```txt
reward por episodio
reward promedio
epsilon
loss
pasos por episodio
acción dominante
mejor episodio
summary cards
dashboard de entrenamiento
```

### Labs esperados

```txt
tec-dqn-training-log-schema-lab
tec-reward-chart-lab
tec-epsilon-schedule-viewer-lab
tec-loss-viewer-lab
tec-action-distribution-lab
tec-training-summary-card-lab
docs-dqn-training-storytelling-lab
cloud-dqn-training-logs-to-aws-s3-lab
```

---

## Sprint 3 — World Model Rollout Viewer

### Match

```txt
Path AI Engineer Proyecto 41 — world-models-planning-mini-lab
```

### Base anterior

```txt
21-world-model-rollout-viewer
```

### Objetivo

Agregar un módulo para visualizar rollouts simulados, predicción de siguiente estado y error acumulado.

Flujo:

```txt
transition examples
→ current state
→ action
→ predicted next state
→ rollout sequence
→ expected state
→ error comparison
→ accumulated error
→ risk cards
→ viewer
```

### Resultado esperado

Al finalizar este sprint, la plataforma debe permitir ver:

```txt
transition dataset cards
next-state prediction viewer
rollout sequences
error por paso
error acumulado
planning risk cards
rollout viewer
limitaciones de planificación
```

### Labs esperados

```txt
tec-transition-dataset-card-lab
tec-next-state-prediction-viewer-lab
tec-rollout-sequence-lab
tec-error-comparison-lab
tec-accumulated-error-notes-lab
tec-planning-risk-card-lab
docs-world-model-storytelling-lab
cloud-rollout-results-to-azure-blob-lab
```

---

# 🧪 Labs del proyecto

## Sprint 1 — Gridworld Labs

```txt
tec-gridworld-setup-lab
tec-state-action-card-lab
tec-reward-map-lab
tec-episode-runner-lab
tec-trajectory-viewer-lab
tec-policy-notes-lab
docs-rl-foundations-storytelling-lab
docs-agent-trajectory-report-template-lab
cloud-gridworld-outputs-to-gcp-storage-lab
cloud-gridworld-outputs-to-aws-s3-lab
cloud-gridworld-outputs-to-azure-blob-lab
```

## Sprint 2 — DQN Labs

```txt
tec-dqn-training-log-schema-lab
tec-reward-chart-lab
tec-epsilon-schedule-viewer-lab
tec-loss-viewer-lab
tec-action-distribution-lab
tec-training-summary-card-lab
docs-dqn-training-storytelling-lab
docs-rl-dashboard-report-template-lab
cloud-dqn-training-logs-to-gcp-storage-lab
cloud-dqn-training-logs-to-aws-s3-lab
cloud-dqn-training-logs-to-azure-blob-lab
```

## Sprint 3 — World Model Labs

```txt
tec-transition-dataset-card-lab
tec-next-state-prediction-viewer-lab
tec-rollout-sequence-lab
tec-error-comparison-lab
tec-accumulated-error-notes-lab
tec-planning-risk-card-lab
docs-world-model-storytelling-lab
docs-rollout-report-template-lab
cloud-rollout-results-to-gcp-storage-lab
cloud-rollout-results-to-aws-s3-lab
cloud-rollout-results-to-azure-blob-lab
```

---

# 📊 Métricas y evidencia esperada

## Gridworld

```txt
número de episodios
pasos por episodio
recompensa total
ruta del agente
número de choques
éxito o fallo
policy notes
trajectory viewer
visual report
```

## DQN

```txt
reward por episodio
reward promedio
epsilon
loss
pasos por episodio
acción dominante
mejor episodio
training summary cards
dashboard
```

## World Models

```txt
estados iniciales
acciones
estados predichos
estados esperados
error por paso
error acumulado
rollout sequence
planning risk cards
rollout viewer
```

---

# 🖥️ Dashboard esperado

La plataforma debe incluir una vista visual con secciones como:

```txt
Overview
Gridworld
DQN Training
World Model Rollouts
Planning Risks
Reports
Limitations
```

Cada sección debe mostrar resultados claros sin obligar al usuario a leer el código.

---

# 📁 Estructura recomendada del repositorio

```txt
07-rl-simulation-control-platform/
│
├── README.md
├── project-structure.txt
├── .gitignore
├── .env.example
├── docker-compose.yml
│
├── frontend/
│   └── app/
│
├── backend/
│   └── api/
│
├── ai-services/
│   └── rl-simulation-service/
│       ├── README.md
│       ├── src/
│       │   ├── environments/
│       │   ├── agents/
│       │   ├── rewards/
│       │   ├── training/
│       │   ├── world_models/
│       │   ├── rollouts/
│       │   ├── analysis/
│       │   └── visualization/
│       └── checks/
│
├── data/
│   ├── raw/
│   ├── processed/
│   ├── logs/
│   └── transitions/
│
├── reports/
│   ├── figures/
│   ├── summaries/
│   ├── outputs/
│   ├── trajectory_cards/
│   ├── training_cards/
│   └── risk_cards/
│
├── docs/
│   ├── architecture.md
│   ├── decisions.md
│   ├── api-contract.md
│   ├── user-stories.md
│   ├── technical-stories.md
│   ├── sprint-01-gridworld-agent-visualizer.md
│   ├── sprint-02-dqn-training-dashboard.md
│   └── sprint-03-world-model-rollout-viewer.md
│
├── labs/
│   ├── tec-labs/
│   ├── cloud-labs/
│   └── docs-labs/
│
├── tests/
│   ├── frontend/
│   ├── backend/
│   └── ai-services/
│
├── scripts/
└── deployment/
```

Regla:

```txt
La estructura debe servir al proyecto.
No el proyecto a la estructura.
```

---

# 🧾 Sistema de documentación

Este proyecto se trabajará con documentación profesional basada en:

```txt
US — User Stories
TS — Technical Stories
AC — Acceptance Criteria
DoD — Definition of Done
Sprint Review
Sprint Retrospective
```

---

# ✅ Definition of Done

Una tarea no termina solo cuando el código funciona.

Termina cuando deja evidencia.

```txt
Código implementado
Prueba mínima realizada
Historia documentada
Decisión registrada
Resultado visible
README actualizado si aplica
Sin archivos basura
Sin responsabilidades mezcladas
```

---

# 🚀 Estado actual

Los tres sprints están entregados como **checkpoints técnicos versionados**. La rama release añade
Azure Container Apps con escala total a cero, imágenes públicas en GHCR, Redis transitorio
en la misma réplica, migración atómica y enlace seguro con Neon. El gate
containerizado de Sprint 3 pasó el 2026-08-16; el despliegue público permanece explícitamente
pendiente hasta ejecutar la aceptación cloud con una sesión Azure autenticada.

---

# 📌 Próximos pasos

## Sprint 1

```txt
Gate containerizado completo aprobado
queued → running → succeeded validado contra PostgreSQL y Redis reales
Smoke real aprobado: 10 transiciones, 9.64 reward, 0 colisiones, goal reached
Tag técnico de entrega creado antes de Sprint 2
Capturas y certificación visual explícitamente no incluidas en este checkpoint
```

## Sprint 2

```txt
Training log schema versionado y migración 0002 implementados
DQN real: MLP, replay buffer, epsilon-greedy y target network
Reward chart y promedio móvil implementados
Epsilon schedule y loss viewers implementados
Action distribution y summary cards implementados
Dashboard responsive con tablas accesibles implementado
Smoke aprobado: 40 episodios ordenados y 400 métricas persistidas
Gate containerizado completo aprobado el 2026-08-16
Tag técnico creado; certificación visual independiente no archivada
```

## Sprint 3

```txt
Transition examples versionados en artefacto SHA-256
Next-state prediction viewer implementado
Rollout sequence de 11 pasos implementada
Error por paso y acumulado persistidos
Planning risk cards derivadas de evidencia durable
Rollout viewer y tabla accesible implementados
Labs y ADR documentados
Gate containerizado y tres smokes aprobados el 2026-08-16
Tag técnico creado; Azure y Neon quedan para la rama release
```

## Release Azure + Neon

```txt
Topología Bicep sin registro Azure de costo fijo y cinco imágenes GHCR inmutables implementadas
Migración Neon directa, transaccional y protegida por SHA-256 implementada
API Neon pooled y Redis transitorio interno preparados mediante secretos
Smoke remoto de los tres perfiles implementado
Aprovisionamiento real y URLs públicas pendientes de sesión Azure autenticada
```

---

# ✅ Entregable final

Al terminar este proyecto debe existir:

```txt
Aplicación RL aplicada
Gridworld visualizer
DQN training dashboard
World model rollout viewer
Frontend
Backend/API si aplica
Servicio de simulación/IA
Dataset o ejemplos de estados
Pipeline reproducible si aplica
User stories
Technical stories
Acceptance criteria
Sprint docs
Reports
Trajectory cards
Training cards
Planning risk cards
Labs documentados
README profesional
Evidencia visual
Deploy o guía de deploy
```

---

# 🧠 Resultado esperado

Al terminar este proyecto podré decir:

```txt
Construí una aplicación de software aplicada a Reinforcement Learning.

No solo corrí episodios.
No solo miré rewards.
No solo simulé estados.

Integré entorno, agente, recompensa, trayectoria, entrenamiento, world models, dashboards, documentación, sprints e historias en una sola plataforma.
```

---

# 🧭 Regla final

```txt
Este proyecto no será un agente escondido en código.

Será una plataforma para observar comportamiento, aprendizaje y planificación.

Un agente se entiende viendo cómo actúa.

Un world model se entiende viendo sus rollouts y sus errores.

Path AI Engineer me da la profundidad.

Path Software Engineer convierte esa profundidad en producto.
```

---

# 👤 Autor

**Jean Franck Loa Rojas**

Path Software Engineer Builder
Reinforcement Learning • Gridworld • DQN • World Models • Simulation • Dashboards • Agent Visualization • Product Architecture
