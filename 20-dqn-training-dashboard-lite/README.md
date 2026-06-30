# 20-dqn-training-dashboard-lite

## 🧠 Descripción

Dashboard ligero para visualizar entrenamiento de un Deep Q-Network.

Este proyecto pertenece a la ruta:

```txt id="bp20-route"
Building Projects
```

y acompaña directamente al proyecto:

```txt id="bp20-match"
AI Engineer Proyecto 39 — deep-q-network-gymnasium-lab
```

Mientras AI Engineer profundiza en DQN, Gymnasium, replay buffer, epsilon-greedy, target network concept y training loop, este Building Project crea un dashboard para mostrar la evolución del agente por episodios.

La idea es mostrar:

```txt id="bp20-core"
entorno
→ DQN agent
→ episodio
→ reward
→ epsilon
→ loss
→ acciones
→ dashboard
```

Este proyecto no busca crear un benchmark avanzado de RL.

Busca demostrar que puedo observar, registrar y explicar el entrenamiento de un agente DQN.

---

## 🎯 Objetivo

Crear un dashboard que muestre rewards, episodios, acciones, epsilon, loss y evolución del entrenamiento DQN.

El objetivo es explicar:

* Cómo evoluciona el reward.
* Cómo cambia epsilon.
* Qué ocurre durante los episodios.
* Cuándo el agente mejora.
* Cuándo se estanca.
* Qué acciones prefiere.
* Qué limitaciones tiene el entrenamiento.

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

* Episodio.
* Reward.
* Pasos.
* Epsilon.
* Loss.
* Acción dominante.
* Estado del entrenamiento.

Pregunta central:

```txt id="bp20-q1"
¿Qué debo registrar para entender el entrenamiento DQN?
```

---

### Módulo 2 — Reward Chart

Visualizar recompensas.

Incluye:

* Reward por episodio.
* Promedio móvil.
* Tendencia.
* Picos.
* Estancamiento.

Pregunta central:

```txt id="bp20-q2"
¿El agente está mejorando con el tiempo?
```

---

### Módulo 3 — Epsilon Schedule Viewer

Mostrar exploración vs explotación.

Incluye:

* Epsilon inicial.
* Decay.
* Exploración.
* Explotación.
* Interpretación.

Pregunta central:

```txt id="bp20-q3"
¿Cuándo el agente explora y cuándo empieza a explotar lo aprendido?
```

---

### Módulo 4 — Loss Viewer

Mostrar pérdida del entrenamiento.

Incluye:

* Loss por paso o episodio.
* Estabilidad.
* Ruido.
* Advertencia de interpretación.
* Limitaciones.

Pregunta central:

```txt id="bp20-q4"
¿Qué me dice la loss y qué no me dice en RL?
```

---

### Módulo 5 — Action Distribution

Mostrar distribución de acciones.

Incluye:

* Acciones tomadas.
* Frecuencia.
* Cambios durante entrenamiento.
* Posibles sesgos.
* Notas.

Pregunta central:

```txt id="bp20-q5"
¿Qué acciones está prefiriendo el agente?
```

---

### Módulo 6 — Training Summary Cards

Crear tarjetas de resumen.

Incluye:

* Mejor episodio.
* Reward promedio.
* Episodio final.
* Estabilidad.
* Comportamiento observado.
* Limitaciones.

Pregunta central:

```txt id="bp20-q6"
¿Cómo resumo el entrenamiento de forma entendible?
```

---

### Módulo 7 — Dashboard Lite

Crear vista final.

Puede ser:

* Streamlit simple.
* Notebook visual.
* `dashboard/README.md`.
* HTML ligero.

Debe mostrar:

* Reward chart.
* Epsilon.
* Loss.
* Action distribution.
* Summary cards.
* Limitaciones.

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

Este proyecto puede generar:

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

Este proyecto debe demostrar que puedo convertir entrenamiento DQN en evidencia visual y trazable.
