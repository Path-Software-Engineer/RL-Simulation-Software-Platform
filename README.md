# Building Projects Roadmap — Plan 7

## 🧠 RL & Simulation Visual Labs

Esta organización reúne los proyectos del **Plan 7 — RL & Simulation Visual Labs** dentro de **Building Projects**.

Este plan acompaña directamente al:

```txt id="bp7-ai-relation"
AI Engineer Plan 7 — Reinforcement Learning, World Models & Robotics Simulation
```

La idea central es construir visualizadores, tableros y herramientas pequeñas para explicar agentes que actúan, aprenden por recompensa, navegan en entornos, usan políticas y simulan consecuencias.

Mientras AI Engineer profundiza en Reinforcement Learning, Q-Learning, DQN, reward design, world models y robotics simulation, Building Projects convierte una parte de ese aprendizaje en evidencia visual.

```txt id="bp7-core"
entorno
→ estado
→ acción
→ recompensa
→ política
→ trayectoria
→ entrenamiento
→ visualización
→ reporte claro
```

Building Projects no reemplaza los proyectos profundos de AI Engineer.

Los acompaña con herramientas pequeñas que permitan mostrar cómo un agente aprende, decide, se equivoca y mejora.

---

# 🎯 Objetivo general

Construir herramientas visuales de RL y simulación capaces de:

* Mostrar estados, acciones y recompensas.
* Visualizar trayectorias de agentes.
* Explicar una política aprendida.
* Mostrar evolución de entrenamiento.
* Crear dashboards de rewards por episodio.
* Visualizar rollouts de world models.
* Mostrar error acumulado en predicciones de estado.
* Explicar fallos de planificación.
* Crear evidencia visual para GitHub.
* Acompañar la ruta principal sin inflar el alcance.

---

# 🔗 Regla de match del Plan 7

Building Projects hará match solo con los proyectos impares de AI Engineer.

```txt id="bp7-match-rule"
Proyecto 37 IA → Proyecto 19 Building
Proyecto 38 IA → Nada
Proyecto 39 IA → Proyecto 20 Building
Proyecto 40 IA → Nada
Proyecto 41 IA → Proyecto 21 Building
Proyecto 42 IA → Nada
```

Esto significa que este plan tendrá **3 proyectos**, no 6.

Cada proyecto Building toma como referencia la duración del proyecto IA correspondiente.

---

# 🗺️ Cronograma Plan 7

| Semana Building |                Proyecto Building | Match IA |  Duración | Objetivo                                                              |
| --------------- | -------------------------------: | -------: | --------: | --------------------------------------------------------------------- |
| 77-80           |  `19-gridworld-agent-visualizer` |    IA 37 | 4 semanas | Visualizar estados, acciones, recompensas y trayectorias              |
| 81-85           | `20-dqn-training-dashboard-lite` |    IA 39 | 5 semanas | Mostrar entrenamiento DQN, rewards y evolución por episodio           |
| 86-90           |  `21-world-model-rollout-viewer` |    IA 41 | 5 semanas | Visualizar rollouts, predicción de siguiente estado y error acumulado |

Duración total del Plan 7:

```txt id="bp7-duration"
14 semanas
```

---

# 🧭 Filosofía de trabajo

Reinforcement Learning puede parecer abstracto si solo se ve como código de entrenamiento.

Este plan existe para mostrar:

```txt id="bp7-philosophy"
qué observa el agente
qué acción elige
qué recompensa recibe
qué trayectoria sigue
qué política aprende
qué error comete
qué mejora con entrenamiento
```

Regla central:

```txt id="bp7-rule"
Un agente no se entiende solo por su recompensa final.
Debo ver sus estados, acciones, decisiones, fallos y trayectoria.
```

Un Building Project de RL y simulación debe ser:

```txt id="bp7-values"
visual
trazable
pequeño
explicable
documentado
terminable
```

No debe convertirse en un framework de RL.

Debe mostrar una pieza clara del comportamiento del agente.

---

# 🧩 Conceptos base

## Estado

El estado representa la situación actual del agente dentro del entorno.

Ejemplo:

```txt id="bp7-state-example"
posición actual
objetivo
obstáculo cercano
recompensa acumulada
acción disponible
```

---

## Acción

Una acción es lo que el agente decide hacer.

Ejemplos:

* mover arriba;
* mover abajo;
* girar;
* avanzar;
* quedarse quieto;
* tomar decisión de control.

---

## Recompensa

La recompensa indica si una acción acercó o alejó al agente del objetivo.

Ejemplo:

```txt id="bp7-reward-example"
llegar al objetivo → recompensa positiva
chocar con obstáculo → recompensa negativa
dar paso inútil → pequeña penalización
```

---

## Política

La política define qué acción toma el agente ante un estado.

```txt id="bp7-policy-example"
estado
→ política
→ acción elegida
```

---

## Trajectory

Una trayectoria muestra el camino recorrido por el agente.

Permite ver:

* decisiones;
* errores;
* loops;
* avances;
* obstáculos;
* llegada al objetivo.

---

## World Model Rollout

Un rollout de world model simula posibles estados futuros.

Ejemplo:

```txt id="bp7-rollout-example"
estado actual
→ acción
→ estado predicho
→ siguiente acción
→ siguiente estado predicho
```

El riesgo es que el error se acumule si el modelo predice mal.

---

# 📁 Proyectos del Plan 7

---

## 19 — gridworld-agent-visualizer

### Match

```txt id="bp19-match"
AI Engineer Proyecto 37 — reinforcement-learning-foundations-gridworld
```

### Duración

```txt id="bp19-duration"
4 semanas
```

---

## 🧠 Descripción

Visualizador ligero para explicar un agente en Gridworld.

Este proyecto acompaña al proyecto de AI Engineer donde se estudian fundamentos de Reinforcement Learning, entorno Gridworld, estados, acciones, recompensas, episodios y políticas básicas.

Mientras AI Engineer trabaja el fundamento técnico, este Building Project convierte el ciclo RL en una herramienta visual.

La idea es mostrar:

```txt id="bp19-core"
grid
→ agente
→ estado
→ acción
→ recompensa
→ trayectoria
→ política
→ visualización
```

Este proyecto no busca crear un agente RL avanzado.

Busca demostrar que puedo explicar cómo un agente actúa dentro de un entorno simple.

---

## 🎯 Objetivo

Crear un visualizador de Gridworld que muestre estados, acciones, recompensas y trayectorias.

El objetivo es explicar:

* qué es un entorno;
* dónde está el agente;
* qué acciones puede tomar;
* qué recompensa recibe;
* qué camino siguió;
* qué política básica usa;
* qué fallos aparecen.

---

## 👤 Usuario objetivo

* Estudiante de Reinforcement Learning.
* AI Engineer en formación.
* Persona que quiere entender agentes visualmente.
* Reclutador técnico viendo evidencia de RL aplicado.
* Yo mismo como constructor de portafolio visual.

---

## 🧱 Arquitectura esperada

```txt id="bp19-architecture"
Gridworld Environment
      ↓
Agent Position
      ↓
Available Actions
      ↓
Reward Map
      ↓
Episode Runner
      ↓
Trajectory Viewer
      ↓
Policy Notes
      ↓
Visual Report
```

---

## 🔁 Flujo técnico

```txt id="bp19-flow"
define grid
→ place agent
→ define goal and obstacles
→ choose action
→ update state
→ assign reward
→ store trajectory
→ render visualizer
```

---

## 🧩 Módulos

### Módulo 1 — Gridworld Setup

Crear entorno simple.

Incluye:

* tamaño del grid;
* posición inicial;
* objetivo;
* obstáculos;
* zonas de recompensa;
* reglas básicas.

Pregunta central:

```txt id="bp19-q1"
¿Qué mundo está explorando el agente?
```

---

### Módulo 2 — State and Action Cards

Explicar estados y acciones.

Incluye:

* estado actual;
* acciones disponibles;
* acción elegida;
* acción inválida;
* consecuencia.

Pregunta central:

```txt id="bp19-q2"
¿Qué puede hacer el agente desde cada estado?
```

---

### Módulo 3 — Reward Map

Mostrar recompensas.

Incluye:

* recompensa por objetivo;
* penalización por obstáculo;
* penalización por paso;
* recompensa acumulada;
* interpretación.

Pregunta central:

```txt id="bp19-q3"
¿Cómo sabe el agente si una acción fue buena o mala?
```

---

### Módulo 4 — Episode Runner

Simular episodios.

Incluye:

* inicio;
* pasos;
* acciones;
* recompensas;
* fin del episodio;
* éxito o fallo.

Pregunta central:

```txt id="bp19-q4"
¿Qué ocurre durante un episodio completo?
```

---

### Módulo 5 — Trajectory Viewer

Visualizar camino del agente.

Incluye:

* ruta recorrida;
* número de pasos;
* loops;
* choques;
* llegada al objetivo;
* visual del recorrido.

Pregunta central:

```txt id="bp19-q5"
¿Qué decisiones tomó el agente durante el camino?
```

---

### Módulo 6 — Policy Notes

Explicar política básica.

Incluye:

* random policy;
* greedy policy conceptual;
* flechas por estado;
* limitaciones;
* comparación simple.

Pregunta central:

```txt id="bp19-q6"
¿Qué regla usa el agente para elegir acciones?
```

---

## 🧪 Labs

### tec-labs

* `tec-gridworld-setup-lab`
* `tec-state-action-card-lab`
* `tec-reward-map-lab`
* `tec-episode-runner-lab`
* `tec-trajectory-viewer-lab`
* `tec-policy-notes-lab`

### docs-labs

* `docs-rl-foundations-storytelling-lab`
* `docs-agent-trajectory-report-template-lab`

### cloud-labs

* `cloud-gridworld-outputs-to-gcp-storage-lab`
* `cloud-gridworld-outputs-to-aws-s3-lab`
* `cloud-gridworld-outputs-to-azure-blob-lab`

---

## 📊 Métricas / Evidencia

* Número de episodios.
* Pasos por episodio.
* Recompensa total.
* Ruta del agente.
* Número de choques.
* Éxito / fallo.
* Policy notes.
* Trajectory viewer.
* Capturas.
* README profesional.

---

## 🚀 Estado actual

Pendiente / por iniciar.

---

## 🧭 Ciclo de trabajo

```txt id="bp19-cycle"
Semana 1 → Gridworld setup, estado, acciones y reward map
Semana 2 → Episode runner, trayectoria y policy notes
Semana 3 → Visualizer, reportes, errores y docs-labs
Semana 4 → Cloud-labs, README, capturas y cierre
```

---

## 📌 Próximos pasos

* Definir grid.
* Definir agente.
* Definir objetivo.
* Definir obstáculos.
* Crear reward map.
* Crear acciones disponibles.
* Simular episodio.
* Guardar trayectoria.
* Crear visualizador.
* Crear policy notes.
* Documentar labs.
* Agregar capturas.
* Publicar repo.

---

## ✅ Entregable final

Al terminar este proyecto debe existir:

* Gridworld visualizer.
* Entorno definido.
* Estado del agente.
* Acciones disponibles.
* Reward map.
* Episode runner.
* Trajectory viewer.
* Policy notes.
* Labs documentados.
* README profesional.
* Capturas u outputs visibles.
* Conexión clara con `reinforcement-learning-foundations-gridworld`.

---

## 🧭 Regla final

```txt id="bp19-rule"
Un agente no se entiende solo por llegar al objetivo.
Debo ver qué estados visitó, qué acciones tomó, qué recompensas recibió y qué camino siguió.
```

---

# 20 — dqn-training-dashboard-lite

### Match

```txt id="bp20-match"
AI Engineer Proyecto 39 — deep-q-network-gymnasium-lab
```

### Duración

```txt id="bp20-duration"
5 semanas
```

---

## 🧠 Descripción

Dashboard ligero para visualizar entrenamiento de un Deep Q-Network.

Este proyecto acompaña al proyecto de AI Engineer donde se estudia DQN, Gymnasium, replay buffer, epsilon-greedy, target network concept y training loop.

Mientras AI Engineer profundiza en el entrenamiento técnico, este Building Project crea un dashboard para mostrar evolución del agente por episodios.

La idea es mostrar:

```txt id="bp20-core"
entorno
→ DQN agent
→ episodio
→ reward
→ epsilon
→ loss
→ política
→ dashboard
```

Este proyecto no busca crear un benchmark avanzado de RL.

Busca demostrar que puedo observar, registrar y explicar el entrenamiento de un agente DQN.

---

## 🎯 Objetivo

Crear un dashboard que muestre rewards, episodios, acciones, epsilon, loss y evolución del entrenamiento DQN.

El objetivo es explicar:

* cómo evoluciona el reward;
* cómo cambia epsilon;
* qué ocurre durante episodios;
* cuándo el agente mejora;
* cuándo se estanca;
* qué limitaciones tiene el entrenamiento.

---

## 👤 Usuario objetivo

* Estudiante de Reinforcement Learning.
* AI Engineer en formación.
* Persona interesada en DQN.
* Reclutador técnico viendo evidencia de entrenamiento RL.
* Yo mismo como constructor de portafolio aplicado.

---

## 🧱 Arquitectura esperada

```txt id="bp20-architecture"
Training Logs
      ↓
Episode Rewards
      ↓
Epsilon Schedule
      ↓
Loss Logs
      ↓
Action Distribution
      ↓
Training Summary
      ↓
Dashboard Lite
```

---

## 🔁 Flujo técnico

```txt id="bp20-flow"
run or simulate dqn training
→ log episodes
→ log rewards
→ log epsilon
→ log loss
→ summarize behavior
→ render dashboard
```

---

## 🧩 Módulos

### Módulo 1 — Training Log Schema

Definir estructura de logs.

Incluye:

* episodio;
* reward;
* pasos;
* epsilon;
* loss;
* acción dominante;
* estado del entrenamiento.

Pregunta central:

```txt id="bp20-q1"
¿Qué debo registrar para entender el entrenamiento DQN?
```

---

### Módulo 2 — Reward Chart

Visualizar recompensas.

Incluye:

* reward por episodio;
* promedio móvil;
* tendencia;
* picos;
* estancamiento.

Pregunta central:

```txt id="bp20-q2"
¿El agente está mejorando con el tiempo?
```

---

### Módulo 3 — Epsilon Schedule Viewer

Mostrar exploración vs explotación.

Incluye:

* epsilon inicial;
* decay;
* exploración;
* explotación;
* interpretación.

Pregunta central:

```txt id="bp20-q3"
¿Cuándo el agente explora y cuándo empieza a explotar lo aprendido?
```

---

### Módulo 4 — Loss Viewer

Mostrar pérdida del entrenamiento.

Incluye:

* loss por paso o episodio;
* estabilidad;
* ruido;
* advertencia de interpretación;
* limitaciones.

Pregunta central:

```txt id="bp20-q4"
¿Qué me dice la loss y qué no me dice en RL?
```

---

### Módulo 5 — Action Distribution

Mostrar distribución de acciones.

Incluye:

* acciones tomadas;
* frecuencia;
* cambios durante entrenamiento;
* posibles sesgos;
* notas.

Pregunta central:

```txt id="bp20-q5"
¿Qué acciones está prefiriendo el agente?
```

---

### Módulo 6 — Training Summary Cards

Crear tarjetas de resumen.

Incluye:

* mejor episodio;
* reward promedio;
* episodio final;
* estabilidad;
* comportamiento observado;
* limitaciones.

Pregunta central:

```txt id="bp20-q6"
¿Cómo resumo el entrenamiento de forma entendible?
```

---

### Módulo 7 — Dashboard Lite

Crear vista final.

Puede ser:

* Streamlit simple;
* notebook visual;
* `dashboard/README.md`;
* HTML ligero.

Debe mostrar:

* reward chart;
* epsilon;
* loss;
* action distribution;
* summary cards;
* limitaciones.

Pregunta central:

```txt id="bp20-q7"
¿Puede alguien entender cómo entrenó el agente sin abrir el código?
```

---

## 🧪 Labs

### tec-labs

* `tec-dqn-training-log-schema-lab`
* `tec-reward-chart-lab`
* `tec-epsilon-schedule-viewer-lab`
* `tec-loss-viewer-lab`
* `tec-action-distribution-lab`
* `tec-training-summary-card-lab`

### docs-labs

* `docs-dqn-training-storytelling-lab`
* `docs-rl-dashboard-report-template-lab`

### cloud-labs

* `cloud-dqn-training-logs-to-gcp-storage-lab`
* `cloud-dqn-training-logs-to-aws-s3-lab`
* `cloud-dqn-training-logs-to-azure-blob-lab`

---

## 📊 Métricas / Evidencia

* Reward por episodio.
* Reward promedio.
* Epsilon.
* Loss.
* Pasos por episodio.
* Acción dominante.
* Mejor episodio.
* Summary cards.
* Dashboard.
* Capturas.
* README profesional.

---

## 🚀 Estado actual

Pendiente / por iniciar.

---

## 🧭 Ciclo de trabajo

```txt id="bp20-cycle"
Semana 1 → Log schema, episodios y reward chart
Semana 2 → Epsilon schedule, loss viewer y action distribution
Semana 3 → Summary cards, dashboard base y limitaciones
Semana 4 → Docs-labs, visual report y capturas
Semana 5 → Cloud-labs, README final y cierre
```

---

## 📌 Próximos pasos

* Definir logs reales o simulados.
* Crear esquema de entrenamiento.
* Registrar episodios.
* Graficar rewards.
* Mostrar epsilon.
* Mostrar loss.
* Mostrar acciones.
* Crear summary cards.
* Crear dashboard.
* Documentar labs.
* Agregar capturas.
* Publicar repo.

---

## ✅ Entregable final

Al terminar este proyecto debe existir:

* DQN training dashboard.
* Training logs.
* Reward chart.
* Epsilon viewer.
* Loss viewer.
* Action distribution.
* Summary cards.
* Labs documentados.
* README profesional.
* Capturas u outputs visibles.
* Conexión clara con `deep-q-network-gymnasium-lab`.

---

## 🧭 Regla final

```txt id="bp20-rule"
En RL no basta con decir que el agente entrenó.
Debo mostrar rewards, exploración, acciones, estabilidad y límites del aprendizaje.
```

---

# 21 — world-model-rollout-viewer

### Match

```txt id="bp21-match"
AI Engineer Proyecto 41 — world-models-planning-mini-lab
```

### Duración

```txt id="bp21-duration"
5 semanas
```

---

## 🧠 Descripción

Viewer ligero para visualizar rollouts de world models, predicciones de siguiente estado y error acumulado.

Este proyecto acompaña al proyecto de AI Engineer donde se estudian world models, transition dataset, next-state prediction, planning, rollout simulation y error accumulation.

Mientras AI Engineer profundiza en la construcción técnica del world model, este Building Project crea una herramienta visual para mostrar cómo se predicen estados futuros y cómo el error puede crecer.

La idea es mostrar:

```txt id="bp21-core"
estado actual
→ acción
→ estado predicho
→ rollout
→ estado real / esperado
→ error
→ acumulación
→ viewer
```

Este proyecto no busca crear un sistema avanzado de planificación.

Busca demostrar que puedo explicar world models mediante rollouts visuales y análisis de error.

---

## 🎯 Objetivo

Crear un viewer que muestre rollouts simulados, predicción de siguiente estado y error acumulado.

El objetivo es explicar:

* qué estado observa el modelo;
* qué acción se aplica;
* qué siguiente estado predice;
* cómo se encadenan predicciones;
* cómo se acumula el error;
* qué riesgos tiene planificar con un world model imperfecto.

---

## 👤 Usuario objetivo

* Estudiante de RL avanzado.
* AI Engineer en formación.
* Persona interesada en world models.
* Reclutador técnico viendo evidencia de planificación y simulación.
* Yo mismo como constructor de portafolio visual.

---

## 🧱 Arquitectura esperada

```txt id="bp21-architecture"
Transition Dataset
      ↓
Current State
      ↓
Action
      ↓
Predicted Next State
      ↓
Rollout Sequence
      ↓
Error Comparison
      ↓
Accumulated Error Notes
      ↓
Rollout Viewer
```

---

## 🔁 Flujo técnico

```txt id="bp21-flow"
load transition examples
→ select current state
→ apply action
→ predict next state
→ repeat rollout
→ compare with expected state
→ calculate error
→ render viewer
```

---

## 🧩 Módulos

### Módulo 1 — Transition Dataset Cards

Crear ejemplos de transición.

Incluye:

* estado actual;
* acción;
* siguiente estado real o esperado;
* metadata;
* explicación.

Pregunta central:

```txt id="bp21-q1"
¿Qué aprende un world model a predecir?
```

---

### Módulo 2 — Next-State Prediction Viewer

Mostrar predicción de siguiente estado.

Incluye:

* estado actual;
* acción;
* estado predicho;
* comparación con estado esperado;
* error.

Pregunta central:

```txt id="bp21-q2"
¿Qué tan bien predice el modelo el próximo estado?
```

---

### Módulo 3 — Rollout Sequence Viewer

Encadenar predicciones.

Incluye:

* paso 1;
* paso 2;
* paso 3;
* estado predicho por paso;
* trayectoria simulada.

Pregunta central:

```txt id="bp21-q3"
¿Qué ocurre cuando uso predicciones del modelo para imaginar el futuro?
```

---

### Módulo 4 — Error Comparison

Comparar predicción contra realidad.

Incluye:

* error por paso;
* error promedio;
* diferencias visibles;
* desviación acumulada;
* advertencias.

Pregunta central:

```txt id="bp21-q4"
¿Cuánto se aleja la predicción del estado esperado?
```

---

### Módulo 5 — Accumulated Error Notes

Documentar acumulación de error.

Incluye:

* error pequeño inicial;
* error que crece;
* riesgo de planificación;
* límites del rollout;
* interpretación.

Pregunta central:

```txt id="bp21-q5"
¿Por qué un pequeño error puede crecer durante varios pasos?
```

---

### Módulo 6 — Planning Risk Cards

Crear tarjetas de riesgo.

Incluye:

* rollout confiable;
* rollout incierto;
* rollout degradado;
* recomendación;
* limitación.

Pregunta central:

```txt id="bp21-q6"
¿Cuándo planificar con el modelo puede ser riesgoso?
```

---

### Módulo 7 — Rollout Viewer

Crear vista final.

Puede ser:

* Streamlit simple;
* notebook visual;
* `dashboard/README.md`;
* HTML ligero.

Debe mostrar:

* estados;
* acciones;
* predicciones;
* error;
* acumulación;
* riesgo.

Pregunta central:

```txt id="bp21-q7"
¿Puede alguien entender un world model viendo sus rollouts?
```

---

## 🧪 Labs

### tec-labs

* `tec-transition-dataset-card-lab`
* `tec-next-state-prediction-viewer-lab`
* `tec-rollout-sequence-lab`
* `tec-error-comparison-lab`
* `tec-accumulated-error-notes-lab`
* `tec-planning-risk-card-lab`

### docs-labs

* `docs-world-model-storytelling-lab`
* `docs-rollout-report-template-lab`

### cloud-labs

* `cloud-rollout-results-to-gcp-storage-lab`
* `cloud-rollout-results-to-aws-s3-lab`
* `cloud-rollout-results-to-azure-blob-lab`

---

## 📊 Métricas / Evidencia

* Estados iniciales.
* Acciones.
* Estados predichos.
* Estados esperados.
* Error por paso.
* Error acumulado.
* Rollout sequence.
* Planning risk cards.
* Viewer visual.
* Capturas.
* README profesional.

---

## 🚀 Estado actual

Pendiente / por iniciar.

---

## 🧭 Ciclo de trabajo

```txt id="bp21-cycle"
Semana 1 → Transition dataset cards y next-state prediction viewer
Semana 2 → Rollout sequences y error comparison
Semana 3 → Accumulated error notes y planning risk cards
Semana 4 → Viewer, docs-labs y capturas
Semana 5 → Cloud-labs, README final y cierre
```

---

## 📌 Próximos pasos

* Crear ejemplos de transición.
* Definir estados y acciones.
* Crear next-state prediction viewer.
* Crear rollout sequence.
* Comparar predicción vs esperado.
* Calcular error por paso.
* Documentar error acumulado.
* Crear planning risk cards.
* Crear viewer.
* Agregar capturas.
* Publicar repo.

---

## ✅ Entregable final

Al terminar este proyecto debe existir:

* World model rollout viewer.
* Transition dataset cards.
* Next-state prediction viewer.
* Rollout sequences.
* Error comparison.
* Accumulated error notes.
* Planning risk cards.
* Labs documentados.
* README profesional.
* Capturas u outputs visibles.
* Conexión clara con `world-models-planning-mini-lab`.

---

## 🧭 Regla final

```txt id="bp21-rule"
Un world model no solo predice el próximo estado.
Permite imaginar futuros posibles, pero sus errores también pueden crecer.

Planificar con predicciones exige medir incertidumbre y error acumulado.
```

---

# 🧱 Ciclo general de cada proyecto

Cada proyecto del Plan 7 sigue este ciclo:

```txt id="bp7-cycle-general"
1. Definir demo de agente o simulación.
2. Definir usuario.
3. Definir qué debe observarse.
4. Elegir entorno pequeño.
5. Crear README inicial.
6. Crear estructura mínima.
7. Crear primera visualización.
8. Agregar tarjetas explicativas.
9. Crear labs pequeños.
10. Probar si aplica.
11. Documentar decisiones.
12. Agregar capturas.
13. Preparar demo o evidencia.
14. Escribir aprendizajes.
15. Definir limitaciones.
16. Definir siguiente paso.
17. Publicar en GitHub.
18. Conectar con el proyecto IA correspondiente.
```

---

# 🗂️ Estructura recomendada del repositorio

```txt id="bp7-repo-structure"
RL-and-Simulation-Visual-Labs/
├── 19-gridworld-agent-visualizer/
│   ├── data/
│   ├── src/
│   ├── reports/
│   ├── visuals/
│   ├── dashboard/
│   ├── docs/
│   ├── labs/
│   ├── scripts/
│   └── README.md
│
├── 20-dqn-training-dashboard-lite/
│   ├── data/
│   ├── src/
│   ├── reports/
│   ├── logs/
│   ├── dashboard/
│   ├── docs/
│   ├── labs/
│   └── README.md
│
├── 21-world-model-rollout-viewer/
│   ├── data/
│   ├── src/
│   ├── reports/
│   ├── rollouts/
│   ├── dashboard/
│   ├── docs/
│   ├── labs/
│   └── README.md
│
└── README.md
```

---

# 📊 Nivel esperado al terminar Plan 7

| Área                               | Nivel esperado |
| ---------------------------------- | -------------: |
| RL visual explanation              |           8/10 |
| State/action/reward mapping        |         8.5/10 |
| Gridworld visualization            |         8.5/10 |
| Trajectory viewer                  |           8/10 |
| Policy notes                       |           8/10 |
| DQN training dashboard             |           8/10 |
| Reward chart interpretation        |           8/10 |
| Epsilon schedule explanation       |           8/10 |
| Loss and action distribution notes |         7.5/10 |
| World model rollout visualization  |           8/10 |
| Error accumulation notes           |           8/10 |
| Planning risk cards                |         8.5/10 |
| README profesional                 |         8.5/10 |
| Evidencia visual de aprendizaje    |           9/10 |

---

# 🧠 Resultado esperado del Plan 7

Al completar este plan, podré decir:

```txt id="bp7-result"
Sé visualizar agentes en entornos simples.
Sé mostrar estados, acciones, recompensas y trayectorias.
Sé explicar políticas básicas.
Sé crear dashboards de entrenamiento DQN.
Sé mostrar rewards, epsilon, loss y acciones.
Sé visualizar rollouts de world models.
Sé explicar error acumulado y riesgo de planificación.
Sé convertir RL y simulación en evidencia visual clara.
```

---

# 🧭 Regla final de avance

```txt id="bp7-final-rule"
Un agente se entiende viendo su comportamiento,
no solo leyendo su recompensa final.

Estado, acción, recompensa, trayectoria y error deben quedar visibles.
```

Frase guía:

```txt id="bp7-final-phrase"
AI Engineer me enseña agentes que aprenden.
Building Projects me obliga a mostrar cómo actúan, fallan y mejoran.
```

---

# 👤 Autor

**Jean Franck Loa Rojas**

Building Projects Path Builder
Reinforcement Learning • Gridworld • DQN • World Models • Simulation • Agent Visualization • Technical Storytelling
