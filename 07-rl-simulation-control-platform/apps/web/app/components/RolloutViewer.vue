<script setup lang="ts">
import type { Transition } from '~/types/api'

const props = defineProps<{ transitions: Transition[] }>()
const emit = defineEmits<{ change: [transition: Transition] }>()
const activeStep = ref(props.transitions[0]?.stepIndex ?? 0)

watch(() => props.transitions, (items) => {
  if (!items.some(item => item.stepIndex === activeStep.value)) activeStep.value = items[0]?.stepIndex ?? 0
}, { deep: true })

function select(transition: Transition) {
  activeStep.value = transition.stepIndex
  emit('change', transition)
}
</script>

<template>
  <article class="rollout-viewer" aria-labelledby="rollout-viewer-title">
    <div class="section-heading"><div><p class="eyebrow">Autoregressive sequence</p><h2 id="rollout-viewer-title">Rollout step inspector</h2></div><span>{{ transitions.length }} persisted steps</span></div>
    <ol class="rollout-track" aria-label="World-model rollout steps">
      <li v-for="transition in transitions" :key="transition.id">
        <button type="button" :aria-pressed="activeStep === transition.stepIndex" @click="select(transition)">
          <small>STEP {{ transition.stepIndex + 1 }}</small>
          <strong>{{ transition.action }}</strong>
          <span>error {{ transition.stepError?.toFixed(0) ?? '—' }}</span>
        </button>
      </li>
    </ol>
    <details>
      <summary>View exact rollout table</summary>
      <div class="table-scroll"><table><thead><tr><th>Step</th><th>Action</th><th>Expected</th><th>Predicted</th><th>Error</th><th>Accumulated</th></tr></thead><tbody><tr v-for="transition in transitions" :key="`table-${transition.id}`"><td>{{ transition.stepIndex + 1 }}</td><td>{{ transition.action }}</td><td>({{ transition.nextState.row }}, {{ transition.nextState.column }})</td><td v-if="transition.predictedNextState">({{ transition.predictedNextState.row }}, {{ transition.predictedNextState.column }})</td><td v-else>—</td><td>{{ transition.stepError?.toFixed(0) ?? '—' }}</td><td>{{ transition.accumulatedError?.toFixed(0) ?? '—' }}</td></tr></tbody></table></div>
    </details>
  </article>
</template>
