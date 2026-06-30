# 21-world-model-rollout-viewer

## 🧠 Descripción

Viewer ligero para visualizar rollouts de world models, predicciones de siguiente estado y error acumulado.

Este proyecto pertenece a la ruta:

```txt id="bp21-route"
Building Projects
```

y acompaña directamente al proyecto:

```txt id="bp21-match"
AI Engineer Proyecto 41 — world-models-planning-mini-lab
```

Mientras AI Engineer profundiza en world models, transition dataset, next-state prediction, planning, rollout simulation y error accumulation, este Building Project crea una herramienta visual para mostrar cómo se predicen estados futuros y cómo el error puede crecer.

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

* Qué estado observa el modelo.
* Qué acción se aplica.
* Qué siguiente estado predice.
* Cómo se encadenan predicciones.
* Cómo se acumula el error.
* Qué riesgos tiene planificar con un world model imperfecto.

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

* Estado actual.
* Acción.
* Siguiente estado real o esperado.
* Metadata.
* Explicación.

Pregunta central:

```txt id="bp21-q1"
¿Qué aprende un world model a predecir?
```

---

### Módulo 2 — Next-State Prediction Viewer

Mostrar predicción de siguiente estado.

Incluye:

* Estado actual.
* Acción.
* Estado predicho.
* Comparación con estado esperado.
* Error.

Pregunta central:

```txt id="bp21-q2"
¿Qué tan bien predice el modelo el próximo estado?
```

---

### Módulo 3 — Rollout Sequence Viewer

Encadenar predicciones.

Incluye:

* Paso 1.
* Paso 2.
* Paso 3.
* Estado predicho por paso.
* Trayectoria simulada.

Pregunta central:

```txt id="bp21-q3"
¿Qué ocurre cuando uso predicciones del modelo para imaginar el futuro?
```

---

### Módulo 4 — Error Comparison

Comparar predicción contra realidad.

Incluye:

* Error por paso.
* Error promedio.
* Diferencias visibles.
* Desviación acumulada.
* Advertencias.

Pregunta central:

```txt id="bp21-q4"
¿Cuánto se aleja la predicción del estado esperado?
```

---

### Módulo 5 — Accumulated Error Notes

Documentar acumulación de error.

Incluye:

* Error pequeño inicial.
* Error que crece.
* Riesgo de planificación.
* Límites del rollout.
* Interpretación.

Pregunta central:

```txt id="bp21-q5"
¿Por qué un pequeño error puede crecer durante varios pasos?
```

---

### Módulo 6 — Planning Risk Cards

Crear tarjetas de riesgo.

Incluye:

* Rollout confiable.
* Rollout incierto.
* Rollout degradado.
* Recomendación.
* Limitación.

Pregunta central:

```txt id="bp21-q6"
¿Cuándo planificar con el modelo puede ser riesgoso?
```

---

### Módulo 7 — Rollout Viewer

Crear vista final.

Puede ser:

* Streamlit simple.
* Notebook visual.
* `dashboard/README.md`.
* HTML ligero.

Debe mostrar:

* Estados.
* Acciones.
* Predicciones.
* Error.
* Acumulación.
* Riesgo.

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

Este proyecto puede generar:

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

Este proyecto debe demostrar que puedo convertir world models y planificación en una visualización clara, honesta y útil.
