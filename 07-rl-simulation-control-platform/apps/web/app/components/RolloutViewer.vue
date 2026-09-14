<script setup lang="ts">
import type { Transition } from '~/types/api'

const props = defineProps<{ transitions: Transition[]; autoplayKey?: number }>()
const emit = defineEmits<{ change: [transition: Transition] }>()
const activeStep = ref(props.transitions[0]?.stepIndex ?? 0)
const playing = ref(false)
const speed = 800
let timer: ReturnType<typeof setInterval> | undefined

watch(() => props.transitions, (items) => {
  if (!items.some(item => item.stepIndex === activeStep.value)) activeStep.value = items[0]?.stepIndex ?? 0
}, { deep: true })

function select(transition: Transition) {
  activeStep.value = transition.stepIndex
  emit('change', transition)
}

function stop() {
  if (timer) clearInterval(timer)
  timer = undefined
  playing.value = false
}

function start() {
  stop()
  playing.value = true
  timer = setInterval(() => {
    const index = props.transitions.findIndex(item => item.stepIndex === activeStep.value)
    const next = props.transitions[index + 1]
    if (!next) return stop()
    select(next)
  }, speed)
}

function toggle() { if (playing.value) return stop(); start() }

watch(() => props.autoplayKey, (key) => {
  if (key) {
    activeStep.value = props.transitions[0]?.stepIndex ?? 0
    if (props.transitions[0]) emit('change', props.transitions[0])
    start()
  }
})
onBeforeUnmount(stop)
</script>

<template>
  <article class="rollout-viewer" aria-labelledby="rollout-viewer-title">
    <div class="section-heading"><div><p class="eyebrow">Autoregressive sequence</p><h2 id="rollout-viewer-title">Rollout step inspector</h2></div><div class="rollout-heading-actions"><span>{{ transitions.length }} persisted steps</span><button class="button secondary" type="button" @click="toggle">{{ playing ? 'Pause demo' : 'Play demo' }}</button></div></div>
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
