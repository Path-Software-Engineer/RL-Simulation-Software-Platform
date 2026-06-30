# 19-gridworld-agent-visualizer

## 🧠 Descripción

Visualizador ligero para explicar un agente dentro de un entorno Gridworld.

Este proyecto pertenece a la ruta:

```txt id="bp19-route"
Building Projects
```

y acompaña directamente al proyecto:

```txt id="bp19-match"
AI Engineer Proyecto 37 — reinforcement-learning-foundations-gridworld
```

Mientras AI Engineer profundiza en los fundamentos de Reinforcement Learning, estados, acciones, recompensas, episodios, políticas y evaluación de agentes, este Building Project convierte esos conceptos en una herramienta visual y entendible.

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

Crear un visualizador de Gridworld que muestre estados, acciones, recompensas, trayectoria y política básica del agente.

El objetivo es explicar:

* Qué es un entorno.
* Dónde está el agente.
* Qué acciones puede tomar.
* Qué recompensa recibe.
* Qué camino siguió.
* Qué política básica usa.
* Qué errores o loops pueden aparecer.

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

Crear un entorno simple.

Incluye:

* Tamaño del grid.
* Posición inicial del agente.
* Objetivo.
* Obstáculos.
* Zonas de recompensa.
* Reglas básicas del entorno.

Pregunta central:

```txt id="bp19-q1"
¿Qué mundo está explorando el agente?
```

---

### Módulo 2 — State and Action Cards

Explicar estados y acciones.

Incluye:

* Estado actual.
* Acciones disponibles.
* Acción elegida.
* Acción inválida.
* Consecuencia de la acción.

Pregunta central:

```txt id="bp19-q2"
¿Qué puede hacer el agente desde cada estado?
```

---

### Módulo 3 — Reward Map

Mostrar recompensas.

Incluye:

* Recompensa por llegar al objetivo.
* Penalización por obstáculo.
* Penalización por paso.
* Recompensa acumulada.
* Interpretación visual.

Pregunta central:

```txt id="bp19-q3"
¿Cómo sabe el agente si una acción fue buena o mala?
```

---

### Módulo 4 — Episode Runner

Simular episodios.

Incluye:

* Inicio del episodio.
* Pasos ejecutados.
* Acciones tomadas.
* Recompensas recibidas.
* Fin del episodio.
* Éxito o fallo.

Pregunta central:

```txt id="bp19-q4"
¿Qué ocurre durante un episodio completo?
```

---

### Módulo 5 — Trajectory Viewer

Visualizar el camino del agente.

Incluye:

* Ruta recorrida.
* Número de pasos.
* Loops.
* Choques.
* Llegada al objetivo.
* Visual del recorrido.

Pregunta central:

```txt id="bp19-q5"
¿Qué decisiones tomó el agente durante el camino?
```

---

### Módulo 6 — Policy Notes

Explicar política básica.

Incluye:

* Random policy.
* Greedy policy conceptual.
* Flechas por estado.
* Limitaciones.
* Comparación simple.

Pregunta central:

```txt id="bp19-q6"
¿Qué regla usa el agente para elegir acciones?
```

---

### Módulo 7 — Visual Report

Crear reporte visual final.

Puede ser:

* `dashboard/README.md`.
* Streamlit simple.
* Notebook visual.
* HTML ligero.

Debe mostrar:

* Grid.
* Agente.
* Objetivo.
* Trayectoria.
* Recompensas.
* Política.
* Conclusiones.

Pregunta central:

```txt id="bp19-q7"
¿Puede alguien entender el comportamiento del agente sin abrir el código?
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

Este proyecto puede generar:

* Número de episodios.
* Pasos por episodio.
* Recompensa total.
* Ruta del agente.
* Número de choques.
* Éxito o fallo.
* Policy notes.
* Trajectory viewer.
* Visual report.
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

Este proyecto debe demostrar que puedo convertir fundamentos de Reinforcement Learning en una visualización clara y explicable.
